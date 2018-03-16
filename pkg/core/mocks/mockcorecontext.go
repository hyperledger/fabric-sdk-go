/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

// MockCoreContext is a mock core context
type MockCoreContext struct {
	MockConfig         core.Config
	MockCcryptoSuite   core.CryptoSuite
	MockUserStore      msp.UserStore
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

// UserStore ...
func (m *MockCoreContext) UserStore() msp.UserStore {
	return m.MockUserStore
}

// SigningManager ...
func (m *MockCoreContext) SigningManager() core.SigningManager {
	return m.MockSigningManager
}
