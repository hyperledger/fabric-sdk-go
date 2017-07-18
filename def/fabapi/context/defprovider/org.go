/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defprovider

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fabca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
)

// OrgClientFactory represents the default org provider factory.
type OrgClientFactory struct{}

// NewOrgClientFactory returns the default org provider factory.
func NewOrgClientFactory() *OrgClientFactory {
	f := OrgClientFactory{}
	return &f
}

// NewMSPClient returns a new default implmentation of the MSP client
func (f *OrgClientFactory) NewMSPClient(orgName string, config apiconfig.Config) (fabca.FabricCAClient, error) {
	mspClient, err := fabricCAClient.NewFabricCAClient(config, orgName)
	if err != nil {
		return nil, fmt.Errorf("NewFabricCAClient returned error: %v", err)
	}

	return mspClient, nil
}
