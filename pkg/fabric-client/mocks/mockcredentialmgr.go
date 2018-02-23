/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// MockCredentialManager is a mock CredentialManager
type MockCredentialManager struct {
}

// NewMockCredentialManager Constructor for a credential manager.
func NewMockCredentialManager(orgName string, config core.Config, cryptoProvider core.CryptoSuite) (api.CredentialManager, error) {
	mcm := MockCredentialManager{}
	return &mcm, nil
}

// GetSigningIdentity will sign the given object with provided key,
func (mgr *MockCredentialManager) GetSigningIdentity(userName string) (*api.SigningIdentity, error) {

	si := api.SigningIdentity{}
	return &si, nil
}
