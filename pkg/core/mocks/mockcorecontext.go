/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// MockCoreContext is a mock core context
type MockCoreContext struct {
	MockConfig         core.Config
	MockCcryptoSuite   core.CryptoSuite
	MockStateStore     core.KVStore
	MockSigningManager core.SigningManager
}

// Config ...
func (m *MockCoreContext) Config() core.Config {
	return m.MockConfig
}

// CryptoSuite ...
func (m *MockCoreContext) CryptoSuite() core.CryptoSuite {
	return m.MockCcryptoSuite
}

// StateStore ...
func (m *MockCoreContext) StateStore() core.KVStore {
	return m.MockStateStore
}

// SigningManager ...
func (m *MockCoreContext) SigningManager() core.SigningManager {
	return m.MockSigningManager
}
