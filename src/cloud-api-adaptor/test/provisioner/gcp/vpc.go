// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package gcp

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

// GCPVPC implements the Google Compute VPC interface.
type GCPVPC struct {
	Client          *compute.Service
	CredentialsPath string
	Name            string
}

// NewGCPVPC creates a new GCPVPC object.
func NewGCPVPC(credentialsPath string, name string) (*GCPVPC, error) {
	client, err := compute.NewService(context.TODO(), option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("GKE: failed to create GCP compute service: %v", err)
	}

	return &GCPVPC{
		Client:          client,
		CredentialsPath: credentialsPath,
		Name:            name,
	}, nil
}

// Create creates a new VPC in Google Cloud.
func (g *GCPVPC) Create(
	ctx context.Context, cfg *envconf.Config,
) error {
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	_, err = g.Client.Networks.Get(g.projectID, g.Name).Context(ctx).Do()
	if err == nil {
		log.Infof("GKE: Using existing VPC %s.\n", g.Name)
		return nil
	}

	network := &compute.Network{
		Name:                  g.Name,
		AutoCreateSubnetworks: true,
	}

	op, err := g.Client.Networks.Insert(g.projectID, network).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("GKE: Networks.Insert: %v", err)
	}

	log.Infof("GKE: VPC creation operation started: %v\n", op.Name)

	err = g.WaitForCreation(ctx, 30*time.Minute)
	if err != nil {
		return fmt.Errorf("GKE: Error waiting for VPC to be created: %v", err)
	}
	return nil
}

// Delete deletes a VPC in Google Cloud.
func (g *GCPVPC) Delete(ctx context.Context, cfg *envconf.Config) error {
	op, err := g.Client.Networks.Delete(g.projectID, g.Name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("GKE: Networks.Delete: %v", err)
	}

	log.Infof("GKE: VPC deletion operation started: %v\n", op.Name)

	err = g.WaitForDeleted(ctx, 30*time.Minute)
	if err != nil {
		return fmt.Errorf("GKE: Error waiting for VPC to be deleted: %v", err)
	}

	return nil
}

// WaitForCreation waits until the VPC is created and available.
func (g *GCPVPC) WaitForCreation(
	ctx context.Context, timeout time.Duration,
) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for VPC creation")
		case <-ticker.C:
			network, err := g.Client.Networks.Get(g.projectID, g.Name).Context(ctx).Do()
			if err != nil {
				if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
					log.Info("Waiting for VPC to be created...")
					continue
				}
				return fmt.Errorf("Networks.Get: %v", err)
			}
			if network.SelfLink != "" {
				log.Info("VPC created successfully")
				return nil
			}
		}
	}
}

// WaitForDeleted waits until the VPC is deleted.
func (g *GCPVPC) WaitForDeleted(
	ctx context.Context, timeout time.Duration,
) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("GKE: timeout waiting for VPC deletion")
		case <-ticker.C:
			_, err := g.Client.Networks.Get(g.projectID, g.Name).Context(ctx).Do()
			if err != nil {
				if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
					log.Info("GKE: VPC deleted successfully")
					return nil
				}
				return fmt.Errorf("GKE: Networks.Get: %v", err)
			}
			log.Info("GKE: Waiting for VPC to be deleted...")
		}
	}
}
