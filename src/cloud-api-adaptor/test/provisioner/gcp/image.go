// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package gcp

import (
	"context"
	"fmt"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type GCPImage struct {
	Srv        *compute.Service
	Properties map[string]string
}

// DefaultImageProperties holds the fields with default values
var defaultImageProperties = map[string]string{
	"imageName":	"podvm-image",
}

func NewGCPImage(properties map[string]string) (*GCPImage, error) {
	srv, err := compute.NewService(
		context.TODO(),
		option.WithCredentialsFile(properties["gcpCredentialsPath"]),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP compute service: %w", err)
	}

	return &GCPImage{
		Srv:        srv,
		Properties: properties,
	}, nil
}

// getProperty will return the property or default value. Always assume the
// property is a string. And since this is an object method, the object
// required properties are present because validation.
func (g *GCPImage) getProperty(key string) string {
	value, exists := g.Properties[key]
	if !exists || value == "" {
		value = defaultImageProperties[key]
	}
	return value
}
