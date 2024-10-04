// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package gcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	kconf "sigs.k8s.io/e2e-framework/klient/conf"
)

// GKECluster implements the basic GKE Cluster client operations.
type GKECluster struct {
	Srv        *container.Service
	Properties map[string]string
}

// DefaultClusterProperties holds the fields with default values
var defaultClusterProperties = map[string]string{
	"gcpZone":            "us-central1",
	"clusterName":        "peerpods-cluster",
	"clusterNodeCount":   "2",
	"clusterMachineType": "e2-medium",
}

// NewGKECluster creates a new GKECluster with the given properties
func NewGKECluster(properties map[string]string) (*GKECluster, error) {
	srv, err := container.NewService(
		context.Background(),
		option.WithCredentialsFile(properties["gcpCredentialsPath"]),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create GKE service: %w", err)
	}

	return &GKECluster{
		Srv:        srv,
		Properties: properties,
	}, nil
}

// Create creates the GKE cluster
func (g *GKECluster) Create(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	cluster := &container.Cluster{
		Name:             g.getProperty("clusterName"),
		InitialNodeCount: g.getInt64Property("clusterNodeCount"),
		NodeConfig: &container.NodeConfig{
			MachineType: g.getProperty("clusterMachineType"),
			// Default OS has a ro fs and can´t be used.
			ImageType: "UBUNTU_CONTAINERD",
		},
	}

	req := &container.CreateClusterRequest{
		Cluster: cluster,
	}

	op, err := g.Srv.Projects.Zones.Clusters.Create(
		g.getProperty("gcpProjectID"), g.getProperty("gcpZone"), req,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("GKE: Projects.Zones.Clusters.Create: %v", err)
	}

	log.Infof("GKE: Cluster creation operation: %v\n", op.Name)

	_, err = g.WaitForActivation(ctx, 30*time.Minute)
	if err != nil {
		return fmt.Errorf("GKE: Error waiting for cluster to become active: %v", err)
	}

	err = g.ApplyNodeLabels(ctx)
	if err != nil {
		return fmt.Errorf("GKE: Error applying node labels: %v", err)
	}
	return nil
}

// Delete deletes the GKE cluster
func (g *GKECluster) Delete(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	op, err := g.Srv.Projects.Zones.Clusters.Delete(
		g.getProperty("gcpProjectID"),
		g.getProperty("gcpZone"),
		g.getProperty("clusterName"),
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("GKE: Projects.Zones.Clusters.Delete: %v", err)
	}

	log.Infof("GKE: Cluster deletion operation: %v\n", op.Name)

	// Wait for the cluster to be deleted
	activationTimeout := 30 * time.Minute
	err = g.WaitForDeletion(ctx, activationTimeout)
	if err != nil {
		return fmt.Errorf("GKE: error waiting for cluster to be deleted: %v", err)
	}
	return nil
}

// WaitForActivation waits until the GKE cluster is active
func (g *GKECluster) WaitForActivation(
	ctx context.Context, activationTimeout time.Duration,
) (*container.Cluster, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx, activationTimeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("GKE: Reached timeout waiting for cluster.")
		case <-ticker.C:
			cluster, err := g.Srv.Projects.Zones.Clusters.Get(
				g.getProperty("gcpProjectID"),
				g.getProperty("gcpZone"),
				g.getProperty("gcpClusterName"),
			).Context(ctx).Do()
			if err != nil {
				return nil, fmt.Errorf("GKE: Projects.Zones.Clusters.Get: %v", err)
			}

			if cluster.Status == "RUNNING" {
				log.Info("GKE: Cluster is now active")
				return cluster, nil
			}

			log.Info("GKE: Waiting for cluster to become active...")
		}
	}
}

// WaitForDeletion waits until the GKE cluster is deleted
func (g *GKECluster) WaitForDeletion(
	ctx context.Context, activationTimeout time.Duration,
) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, activationTimeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("GKE: timeout waiting for cluster deletion")
		case <-ticker.C:
			_, err := g.Srv.Projects.Zones.Clusters.Get(
				g.getProperty("gcpProjectID"),
				g.getProperty("gcpZone"),
				g.getProperty("gcpClusterName"),
			).Context(ctx).Do()
			if err != nil {
				if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
					log.Info("GKE: Cluster deleted successfully")
					return nil
				}
				return fmt.Errorf("GKE: Projects.Zones.Clusters.Get: %v", err)
			}

			log.Info("GKE: Waiting for cluster to be deleted...")
		}
	}
}

// ApplyNodeLabels apply worker node labels
func (g *GKECluster) ApplyNodeLabels(ctx context.Context) error {
	kubeconfigPath, err := g.GetKubeconfigFile(ctx)
	if err != nil {
		return err
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %v", err)
	}

	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %v", err)
	}

	for _, node := range nodes.Items {
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			n, err := clientset.CoreV1().Nodes().Get(ctx, node.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get node: %v", err)
			}

			n.Labels["node.kubernetes.io/worker"] = ""
			n.Labels["node-role.kubernetes.io/worker"] = ""
			_, err = clientset.CoreV1().Nodes().Update(ctx, n, metav1.UpdateOptions{})
			return err
		})
		if err != nil {
			return fmt.Errorf("Failed to label node %s: %v\n", node.Name, err)
		}
		log.Infof("Successfully labeled node %s\n", node.Name)
	}
	return nil
}

// GetKubeconfigFile retrieves the path to the kubeconfig file
func (g *GKECluster) GetKubeconfigFile(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx,
		"gcloud", "container", "clusters", "get-credentials", g.getProperty("clusterName"),
		"--zone", g.getProperty("gcpZone"),
		"--project", g.getProperty("gcpProjectID"))
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("Failed to get credentials: %v\nOutput: %s", err, output)
	}

	kubeconfigPath := kconf.ResolveKubeConfigFile()
	_, err = os.Stat(kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("Failed to resolve KubeConfigfile: %v", err)
	}
	return kubeconfigPath, nil
}

// getProperty will return the property or default value. Always assume the
// property is a string. And since this is an object method, the object
// required properties are present because validation.
func (g *GKECluster) getProperty(key string) string {
	value, exists := g.Properties[key]
	if !exists || value == "" {
		value = defaultClusterProperties[key]
	}
	return value
}

// getInt64Property retrieves the integer value for the specified key.
// Returns -1 if conversion fails.
func (g *GKECluster) getInt64Property(key string) int64 {
	value := g.getProperty(key)
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Error("GCP: invalid integer for %s (%s).\n", key, value)
		return -1
	}
	return intValue
}
