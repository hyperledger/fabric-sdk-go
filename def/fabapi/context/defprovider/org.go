/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defprovider

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	fabca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
	credentialMgr "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/credentialmgr"
)

// OrgClientFactory represents the default org provider factory.
type OrgClientFactory struct{}

// NewOrgClientFactory returns the default org provider factory.
func NewOrgClientFactory() *OrgClientFactory {
	f := OrgClientFactory{}
	return &f
}

// NewMSPClient returns a new default implmentation of the MSP client
func (f *OrgClientFactory) NewMSPClient(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (fabca.FabricCAClient, error) {
	mspClient, err := fabricCAClient.NewFabricCAClient(orgName, config, cryptoProvider)
	if err != nil {
		return nil, errors.WithMessage(err, "NewFabricCAClient failed")
	}

	return mspClient, nil
}

// NewCredentialManager returns a new default implmentation of the credential manager
func (f *OrgClientFactory) NewCredentialManager(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (fab.CredentialManager, error) {

	credentialMgr, err := credentialMgr.NewCredentialManager(orgName, config, cryptoProvider)
	if err != nil {
		return nil, errors.WithMessage(err, "NewCredentialManager failed")
	}

	return credentialMgr, nil
}
