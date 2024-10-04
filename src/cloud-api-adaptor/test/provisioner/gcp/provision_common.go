// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package gcp

import (
	"context"
	"errors"
	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var GCPProps = &GCPProvisioner{}

// MandatoryProperties holds the required GCP fields
var mandatoryProperties = []string{
	"gcpProjectID",
	"gcpCredentialsPath",
}

// GCPProvisioner implements the CloudProvisioner interface.
type GCPProvisioner struct {
	Properties   map[string]string
	GkeCluster   *GKECluster
	GcpVPC       *GCPVPC
	PodvmImage   *GCPImage
}

// NewGCPProvisioner creates a new GCPProvisioner with the given properties.
func NewGCPProvisioner(properties map[string]string) (pv.CloudProvisioner, error) {
	if err := ValidateProperties(properties); err != nil {
		return nil, err
	}

	gkeCluster, err := NewGKECluster(properties)
	if err != nil {
		return nil, err
	}

	gcpVPC, err := NewGCPVPC(properties)
	if err != nil {
		return nil, err
	}

	gcpImage, err := NewGCPImage(properties)
	if err != nil {
		return nil, err
	}

	GCPProps = &GCPProvisioner{
		Properties:   properties,
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
	return p.GcpVPC.Create(ctx)
}

// DeleteCluster deletes the GKE cluster.
func (p *GCPProvisioner) DeleteCluster(ctx context.Context, cfg *envconf.Config) error {
	return p.GkeCluster.Delete(ctx)
}

// DeleteVPC deletes the VPC in Google Cloud.
func (p *GCPProvisioner) DeleteVPC(ctx context.Context, cfg *envconf.Config) error {
	return p.GcpVPC.Delete(ctx)
}

func (p *GCPProvisioner) GetProperties(ctx context.Context, cfg *envconf.Config) map[string]string {
	return p.Properties
}

func (p *GCPProvisioner) UploadPodvm(imagePath string, ctx context.Context, cfg *envconf.Config) error {
	// To be Implemented
	return nil
}

// ValidateProperties will catch any missing mandatory property.
func ValidateProperties(properties map[string]string) error {
	for _, property := range mandatoryProperties {
		if _, ok := properties[property]; !ok {
			return errors.New("GCP: Missing mandatory property: " + property)
		}
	}

	return nil
}
