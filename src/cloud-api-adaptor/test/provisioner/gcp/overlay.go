// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package gcp

import (
	"context"
	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"fmt"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

type GCPInstallOverlay struct {
	Overlay  *pv.KustomizeOverlay
}

func NewGCPInstallOverlay(installDir, provider string) (pv.InstallOverlay, error) {
	overlay, err := pv.NewKustomizeOverlay(filepath.Join(installDir, "overlays", provider))
	if err != nil {
		return nil, err
	}

	return &GCPInstallOverlay{
		Overlay: overlay,
	}, nil
}

func (a *GCPInstallOverlay) Apply(ctx context.Context, cfg *envconf.Config) error {
	return a.Overlay.Apply(ctx, cfg)
}

func (a *GCPInstallOverlay) Delete(ctx context.Context, cfg *envconf.Config) error {
	return a.Overlay.Delete(ctx, cfg)
}

func (a *GCPInstallOverlay) Edit(ctx context.Context, cfg *envconf.Config, properties map[string]string) error {
	image := properties["caaImageName"]
	log.Infof("Updating caa image with %s", image)
	if image != "" {
		err := a.Overlay.SetKustomizeImage("cloud-api-adaptor", "newName", image)
		if err != nil {
			return err
		}
	}

	// Mapping the internal properties to ConfigMapGenerator properties.
	mapProps := map[string]string{
		// GCP
		"gcpProjectID":       "GCP_PROJECT_ID",
		"gcpZone":            "GCP_ZONE",

		// GKE Cluster
		"clusterMachineType": "GCP_MACHINE_TYPE",

		// Image
		"imageName":          "PODVM_IMAGE_NAME",

		// VPC
		"vpcName":            "GCP_NETWORK",

		// TODO:
		// "pause_image":      "PAUSE_IMAGE",
		// "vxlan_port":       "VXLAN_PORT",
	}

	for k, v := range mapProps {
		newValue := properties[k]
		if newValue == "" {
			continue
		}

		// Network name is a special case
		if k == "vpcName" {
		    newValue = fmt.Sprintf("global/networks/%s", properties[k])
		}

		if err := a.Overlay.SetKustomizeConfigMapGeneratorLiteral("peer-pods-cm", v, newValue); err != nil {
		    return err
		}
	}

	// Setting custom credentials file
	credentialsPath := properties["gcpCredentialsPath"]
	if credentialsPath != "" {
		credentialsFileName := filepath.Base(credentialsPath)
		if err := a.Overlay.SetKustomizeSecretGeneratorFile("peer-pods-secret",
			credentialsFileName); err != nil {
			return err
		}
	}

	if err := a.Overlay.YamlReload(); err != nil {
		return err
	}

	return nil
}
