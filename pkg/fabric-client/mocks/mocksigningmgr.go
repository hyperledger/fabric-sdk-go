/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// MockSigningManager is mock signing manager
type MockSigningManager struct {
	cryptoProvider bccsp.BCCSP
	hashOpts       bccsp.HashOpts
	signerOpts     bccsp.SignerOpts
}

// NewMockSigningManager Constructor for a mock signing manager.
func NewMockSigningManager() apifabclient.SigningManager {
	return &MockSigningManager{}
}

// Sign will sign the given object using provided key
func (mgr *MockSigningManager) Sign(object []byte, key bccsp.Key) ([]byte, error) {
	return object, nil
}
