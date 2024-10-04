// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package gcp

import (
	"context"
	"fmt"
	"google.golang.org/api/option"

	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner"
	// log "github.com/sirupsen/logrus"
	"google.golang.org/api/compute/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var GCPProps = &GCPProvisioner{}

// GCPProvisioner implements the CloudProvisioner interface.
type GCPProvisioner struct {
	GkeCluster   *GKECluster
	GcpVPC       *GCPVPC
	PodvmImage   *GCPImage
}

// NewGCPProvisioner creates a new GCPProvisioner with the given properties.
func NewGCPProvisioner(properties map[string]string) (pv.CloudProvisioner, error) {
	credentials_path = properties['credentials_path']

	gkeCluster, err := NewGKECluster(credentials_path)
	if err != nil {
		return nil, err
	}

	gcpVPC, err := NewGCPVPC(credentials_path, properties['vpc_name'])
	if err != nil {
		return nil, err
	}

	gcpImage, err := NewGCPImage(credentials_path)
	if err != nil {
		return nil, err
	}

	GCPProps = &GCPProvisioner{
		GkeCluster:   gkeCluster,
		GcpVPC:       gcpVPC,
		PodvmImage:   gcpImage,
	}
	return GCPProps, nil
}

// CreateCluster creates a new GKE cluster.
func (p *GCPProvisioner) CreateCluster(ctx context.Context, cfg *envconf.Config) error {
	err := p.GkeCluster.Create(ctx)
	if err != nil {
		return err
	}

	kubeconfigPath, err := p.GkeCluster.GetKubeconfigFile(ctx)
	if err != nil {
		return err
	}
	*cfg = *envconf.NewWithKubeConfig(kubeconfigPath)

	return nil
}

// CreateVPC creates a new VPC in Google Cloud.
func (p *GCPProvisioner) CreateVPC(ctx context.Context, cfg *envconf.Config) error {
	return p.GcpVPC.Create(ctx, cfg)
}

// DeleteCluster deletes the GKE cluster.
func (p *GCPProvisioner) DeleteCluster(ctx context.Context, cfg *envconf.Config) error {
	return p.GkeCluster.Delete(ctx)
}

// DeleteVPC deletes the VPC in Google Cloud.
func (p *GCPProvisioner) DeleteVPC(ctx context.Context, cfg *envconf.Config) error {
	return p.GcpVPC.Delete(ctx, cfg)
}

func (p *GCPProvisioner) GetProperties(ctx context.Context, cfg *envconf.Config) map[string]string {
	return map[string]string{
		"podvm_image_name": p.PodvmImage.Name,
		"machine_type":     p.GkeCluster.machineType,
		"project_id":       p.GkeCluster.projectID,
		"zone":             p.GkeCluster.zone,
		"network":          p.GcpVPC.vpcName,
	}
}

func (p *GCPProvisioner) UploadPodvm(imagePath string, ctx context.Context, cfg *envconf.Config) error {
	// To be Implemented
	return nil
}
