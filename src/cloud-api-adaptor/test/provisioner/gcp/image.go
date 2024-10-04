// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package gcp

import (
	"google.golang.org/api/compute/v1"
)

type GCPImage struct {
	Client *compute.Service
}

func NewGCPImage(credentialsPath string) *GCPImage {
	client, err := compute.NewService(context.TODO(), option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("GKE: failed to create GCP compute service: %v", err)
	}

	return &GCPImage{
		Client: client,
	}
}
