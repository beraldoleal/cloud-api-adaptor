// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package vsphere

import (
	"github.com/confidential-containers/cloud-api-adaptor/pkg/util"
)

type Config struct {
	VcenterURL   string
	UserName     string
	Password     string
	Thumbprint   string
	Datacenter   string
	Vcluster     string
	Datastore    string
	Resourcepool string
	Deployfolder string
	Template     string
}

func (c Config) Redact() Config {
	return *util.RedactStruct(&c, "UserName", "Password", "Thumbprint").(*Config)
}
