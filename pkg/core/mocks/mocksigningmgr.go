/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
)

// MockSigningManager is mock signing manager
type MockSigningManager struct {
}

// NewMockSigningManager Constructor for a mock signing manager.
func NewMockSigningManager() core.SigningManager {
	return &MockSigningManager{}
}

// Sign will sign the given object using provided key
func (mgr *MockSigningManager) Sign(object []byte, key core.Key) ([]byte, error) {
	return object, nil
}
