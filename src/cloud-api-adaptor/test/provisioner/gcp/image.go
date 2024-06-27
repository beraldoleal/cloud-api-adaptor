// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package gcp

import (
	"google.golang.org/api/compute/v1"
)

type GCPImage struct {
	Service     *compute.Service
	Description string
	Name        string
}

func NewGCPImage(srv *compute.Service, name string) (*GCPImage, error) {
	return &GCPImage{
		Service:     srv,
		Description: "Peer Pod VM image",
		Name:        name,
	}, nil
}
