/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	config "github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
)

// MockProviderContext holds core providers to enable mocking.
type MockProviderContext struct {
	config          config.Config
	cryptoSuite     core.CryptoSuite
	signingManager  core.SigningManager
	stateStore      core.KVStore
	identityManager map[string]core.IdentityManager
}

// NewMockProviderContext creates a MockProviderContext consisting of defaults
func NewMockProviderContext() *MockProviderContext {

	im := make(map[string]core.IdentityManager)
	im[""] = &MockIdentityManager{}

	context := MockProviderContext{
		config:          NewMockConfig(),
		signingManager:  NewMockSigningManager(),
		cryptoSuite:     &MockCryptoSuite{},
		stateStore:      &MockStateStore{},
		identityManager: im,
	}
	return &context
}

// NewMockProviderContextCustom creates a MockProviderContext consisting of the arguments
func NewMockProviderContextCustom(config config.Config, cryptoSuite core.CryptoSuite, signer core.SigningManager, stateStore core.KVStore, identityManager map[string]core.IdentityManager) *MockProviderContext {
	context := MockProviderContext{
		config:          config,
		signingManager:  signer,
		cryptoSuite:     cryptoSuite,
		stateStore:      stateStore,
		identityManager: identityManager,
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
func (pc *MockProviderContext) CryptoSuite() core.CryptoSuite {
	return pc.cryptoSuite
}

// SigningManager returns the mock signing manager.
func (pc *MockProviderContext) SigningManager() core.SigningManager {
	return pc.signingManager
}

// StateStore returns the mock state store
func (pc *MockProviderContext) StateStore() core.KVStore {
	return pc.stateStore
}

// IdentityManager returns the identity manager
func (pc *MockProviderContext) IdentityManager(orgName string) (core.IdentityManager, bool) {
	mgr, ok := pc.identityManager[strings.ToLower(orgName)]
	return mgr, ok
}

// MockContext holds core providers and identity to enable mocking.
type MockContext struct {
	*MockProviderContext
	context.Identity
}

// NewMockContext creates a MockContext consisting of defaults and an identity
func NewMockContext(ic context.Identity) *MockContext {
	ctx := MockContext{
		MockProviderContext: NewMockProviderContext(),
		Identity:            ic,
	}
	return &ctx
}

// MockTransactionHeader supplies a transaction ID and metadata.
type MockTransactionHeader struct {
	MockID        fab.TransactionID
	MockCreator   []byte
	MockNonce     []byte
	MockChannelID string
}

// TransactionID returns the transaction's computed identifier.
func (th *MockTransactionHeader) TransactionID() fab.TransactionID {
	return fab.TransactionID(th.MockID)
}

// Creator returns the transaction creator's identity bytes.
func (th *MockTransactionHeader) Creator() []byte {
	return th.MockCreator
}

// Nonce returns the transaction's generated nonce.
func (th *MockTransactionHeader) Nonce() []byte {
	return th.MockNonce
}

// ChannelID returns the transaction's target channel identifier.
func (th *MockTransactionHeader) ChannelID() string {
	return th.MockChannelID
}

// NewMockTransactionHeader creates mock TxnID based on mock user.
func NewMockTransactionHeader(channelID string) (fab.TransactionHeader, error) {
	user := NewMockUser("test")

	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, err
	}

	creator, err := user.SerializedIdentity()
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	id, err := computeTxnID(nonce, creator, h)
	if err != nil {
		return nil, err
	}

	txnID := MockTransactionHeader{
		MockID:        fab.TransactionID(id),
		MockCreator:   creator,
		MockNonce:     nonce,
		MockChannelID: channelID,
	}

	return &txnID, nil
}

func computeTxnID(nonce, creator []byte, h hash.Hash) (string, error) {
	b := append(nonce, creator...)

	_, err := h.Write(b)
	if err != nil {
		return "", err
	}
	digest := h.Sum(nil)
	id := hex.EncodeToString(digest)

	return id, nil
}
