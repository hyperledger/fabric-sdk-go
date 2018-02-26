/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identitymgr"
)

// OrgClientFactory represents the default org provider factory.
type OrgClientFactory struct{}

// NewOrgClientFactory returns the default org provider factory.
func NewOrgClientFactory() *OrgClientFactory {
	f := OrgClientFactory{}
	return &f
}

// CreateCredentialManager returns a new default implementation of the credential manager
func (f *OrgClientFactory) CreateCredentialManager(orgName string, config core.Config, cryptoProvider core.CryptoSuite) (api.CredentialManager, error) {
	return identitymgr.New(orgName, config, cryptoProvider)
}
