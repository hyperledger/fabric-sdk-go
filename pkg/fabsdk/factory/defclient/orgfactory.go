/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	credentialMgr "github.com/hyperledger/fabric-sdk-go/pkg/fab/credentialmgr"
)

// OrgClientFactory represents the default org provider factory.
type OrgClientFactory struct{}

// NewOrgClientFactory returns the default org provider factory.
func NewOrgClientFactory() *OrgClientFactory {
	f := OrgClientFactory{}
	return &f
}

/*
// CreateMSPClient returns a new default implementation of the MSP client
// TODO: duplicate of core factory method (remove one) or at least call the core one like in sessfactory
func (f *OrgClientFactory) CreateMSPClient(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (fabca.FabricCAClient, error) {
	return fabricCAClient.New(orgName, config, cryptoProvider)
}
*/

// CreateCredentialManager returns a new default implementation of the credential manager
func (f *OrgClientFactory) CreateCredentialManager(orgName string, config core.Config, cryptoProvider core.CryptoSuite) (api.CredentialManager, error) {
	return credentialMgr.New(orgName, config, cryptoProvider)
}
