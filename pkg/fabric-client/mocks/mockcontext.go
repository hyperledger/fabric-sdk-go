/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
)

// MockProviderContext holds core providers to enable mocking.
type MockProviderContext struct {
	config         config.Config
	cryptoSuite    apicryptosuite.CryptoSuite
	signingManager fab.SigningManager
}

// NewMockProviderContext creates a MockProviderContext consisting of defaults
func NewMockProviderContext() *MockProviderContext {
	context := MockProviderContext{
		config:         NewMockConfig(),
		signingManager: NewMockSigningManager(),
		cryptoSuite:    &MockCryptoSuite{},
	}
	return &context
}

// NewMockProviderContextCustom creates a MockProviderContext consisting of the arguments
func NewMockProviderContextCustom(config config.Config, cryptoSuite apicryptosuite.CryptoSuite, signer fab.SigningManager) *MockProviderContext {
	context := MockProviderContext{
		config:         config,
		signingManager: signer,
		cryptoSuite:    cryptoSuite,
	}
	return &context
}

// Config returns the mock configuration.
func (pc *MockProviderContext) Config() config.Config {
	return pc.config
}

// SetConfig sets the mock configuration.
func (pc *MockProviderContext) SetConfig(config config.Config) {
	pc.config = config
}

// CryptoSuite returns the mock crypto suite.
func (pc *MockProviderContext) CryptoSuite() apicryptosuite.CryptoSuite {
	return pc.cryptoSuite
}

// SigningManager returns the mock signing manager.
func (pc *MockProviderContext) SigningManager() fab.SigningManager {
	return pc.signingManager
}

// MockContext holds core providers and identity to enable mocking.
type MockContext struct {
	*MockProviderContext
	fab.IdentityContext
}

// NewMockContext creates a MockContext consisting of defaults and an identity
func NewMockContext(ic fab.IdentityContext) *MockContext {
	ctx := MockContext{
		MockProviderContext: NewMockProviderContext(),
		IdentityContext:     ic,
	}
	return &ctx
}

// NewMockTxnID creates mock TxnID based on mock user.
func NewMockTxnID() (fab.TransactionID, error) {
	user := NewMockUser("test")

	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return fab.TransactionID{}, err
	}

	creator, err := user.Identity()
	if err != nil {
		return fab.TransactionID{}, err
	}

	id, err := protos_utils.ComputeProposalTxID(nonce, creator)
	if err != nil {
		return fab.TransactionID{}, err
	}

	txnID := fab.TransactionID{
		ID:    id,
		Nonce: nonce,
	}

	return txnID, nil
}
