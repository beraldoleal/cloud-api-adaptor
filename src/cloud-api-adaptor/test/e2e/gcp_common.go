// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"testing"
	"strings"
	"time"

	pv "github.com/confidential-containers/cloud-api-adaptor/src/cloud-api-adaptor/test/provisioner/gcp"
)

// GCPAssert implements the CloudAssert interface.
type GCPAssert struct {
	Vpc *pv.GCPVPC
}

func NewGCPAssert() GCPAssert {
	return GCPAssert{
		Vpc: pv.GCPProps.GcpVPC,
	}
}

func (aa GCPAssert) DefaultTimeout() time.Duration {
	return 1 * time.Minute
}

func (aa GCPAssert) HasPodVM(t *testing.T, id string) {
    podvmPrefix := "podvm-" + id

    // Create a request to list instances in the specified project and zone.
    req := aa.Vpc.Srv.Instances.List(
			aa.Vpc.Properties["gcpProjectID"],
			aa.Vpc.Properties["gcpZone"],
		)
    instances, err := req.Do()
    if err != nil {
        t.Errorf("Failed to list instances: %v", err)
        return
    }

    found := false
    for _, instance := range instances.Items {
        if instance.Status != "TERMINATED" && strings.HasPrefix(instance.Name, podvmPrefix) {
            found = true
            break
        }
    }

    if !found {
        t.Errorf("Podvm name=%s not found", id)
    }
}

func (aa GCPAssert) GetInstanceType(t *testing.T, podName string) (string, error) {
	return "", nil
}
