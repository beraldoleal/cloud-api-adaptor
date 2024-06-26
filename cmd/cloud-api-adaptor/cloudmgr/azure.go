//go:build azure

// (C) Copyright Confidential Containers Contributors
// SPDX-License-Identifier: Apache-2.0

package cloudmgr

import (
	"flag"

	"github.com/confidential-containers/cloud-api-adaptor/pkg/adaptor/cloud"
	"github.com/confidential-containers/cloud-api-adaptor/pkg/adaptor/cloud/azure"
)

func init() {
	cloudTable["azure"] = &azureMgr{}
}

var azurecfg azure.Config

type azureMgr struct{}

func (_ *azureMgr) ParseCmd(flags *flag.FlagSet) {
	flags.StringVar(&azurecfg.ClientId, "clientid", "", "Client Id, defaults to `AZURE_CLIENT_ID`")
	flags.StringVar(&azurecfg.ClientSecret, "secret", "", "Client Secret, defaults to `AZURE_CLIENT_SECRET`")
	flags.StringVar(&azurecfg.TenantId, "tenantid", "", "Tenant Id, defaults to `AZURE_TENANT_ID`")
	flags.StringVar(&azurecfg.ResourceGroupName, "resourcegroup", "", "Resource Group")
	flags.StringVar(&azurecfg.Zone, "zone", "", "Zone")
	flags.StringVar(&azurecfg.Region, "region", "", "Region")
	flags.StringVar(&azurecfg.SubnetId, "subnetid", "", "Network Subnet Id")
	flags.StringVar(&azurecfg.SecurityGroupId, "securitygroupid", "", "Security Group Id")
	flags.StringVar(&azurecfg.Size, "instance-size", "", "Instance size")
	flags.StringVar(&azurecfg.ImageId, "imageid", "", "Image Id")
	flags.StringVar(&azurecfg.SubscriptionId, "subscriptionid", "", "Subscription ID")
	flags.StringVar(&azurecfg.SSHKeyPath, "ssh-key-path", "$HOME/.ssh/id_rsa.pub", "Path to SSH public key")
}

func (_ *azureMgr) LoadEnv() {
	defaultToEnv(&azurecfg.ClientId, "AZURE_CLIENT_ID")
	defaultToEnv(&azurecfg.ClientSecret, "AZURE_CLIENT_SECRET")
	defaultToEnv(&azurecfg.TenantId, "AZURE_TENANT_ID")

}

func (_ *azureMgr) NewProvider() (cloud.Provider, error) {
	return azure.NewProvider(&azurecfg)
}
