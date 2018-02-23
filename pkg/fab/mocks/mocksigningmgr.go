/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// MockSigningManager is mock signing manager
type MockSigningManager struct {
	cryptoProvider core.CryptoSuite
	hashOpts       core.HashOpts
	signerOpts     core.SignerOpts
}

// NewMockSigningManager Constructor for a mock signing manager.
func NewMockSigningManager() api.SigningManager {
	return &MockSigningManager{}
}

// Sign will sign the given object using provided key
func (mgr *MockSigningManager) Sign(object []byte, key core.Key) ([]byte, error) {
	return object, nil
}
