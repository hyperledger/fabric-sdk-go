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
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"

	"strings"

	"fmt"

	"github.com/pkg/errors"
)

// MockProviderContext holds core providers to enable mocking.
type MockProviderContext struct {
	cryptoSuiteConfig      core.CryptoSuiteConfig
	endpointConfig         fab.EndpointConfig
	identityConfig         msp.IdentityConfig
	cryptoSuite            core.CryptoSuite
	signingManager         core.SigningManager
	userStore              msp.UserStore
	identityManager        map[string]msp.IdentityManager
	privateKey             core.Key
	identity               msp.SigningIdentity
	localDiscoveryProvider fab.LocalDiscoveryProvider
	infraProvider          fab.InfraProvider
	channelProvider        fab.ChannelProvider
}

// ProviderUsersOptions ...
type ProviderUsersOptions struct {
	// first map is the org map
	// second map is the user map (within org)
	users    map[string]map[string]msp.SigningIdentity
	identity msp.SigningIdentity
}

// ProviderOption describes a functional parameter for the New constructor
type ProviderOption func(*ProviderUsersOptions) error

// WithProviderUser option
func WithProviderUser(username string, org string) ProviderOption {
	return func(opts *ProviderUsersOptions) error {

		si := mspmocks.NewMockSigningIdentity(username, org)
		opts.identity = si

		userMap := make(map[string]msp.SigningIdentity)
		userMap[username] = si
		orgMap := make(map[string]map[string]msp.SigningIdentity)
		orgMap[org] = userMap
		opts.users = orgMap

		return nil
	}
}

// NewMockProviderContext creates a MockProviderContext consisting of defaults
func NewMockProviderContext(userOpts ...ProviderOption) *MockProviderContext {

	users := ProviderUsersOptions{}
	for _, param := range userOpts {
		err := param(&users)
		if err != nil {
			panic(fmt.Errorf("error creating MockProviderContext: %s", err))
		}
	}

	im := make(map[string]msp.IdentityManager)
	for org := range users.users {
		im[org] = NewMockIdentityManager(WithUsers(users.users[org]))
	}

	context := MockProviderContext{
		cryptoSuiteConfig: NewMockCryptoConfig(),
		endpointConfig:    NewMockEndpointConfig(),
		identityConfig:    NewMockIdentityConfig(),
		signingManager:    mocks.NewMockSigningManager(),
		cryptoSuite:       &MockCryptoSuite{},
		userStore:         &mspmocks.MockUserStore{},
		identityManager:   im,
		infraProvider:     &MockInfraProvider{},
		channelProvider:   &MockChannelProvider{},
		identity:          users.identity,
	}
	return &context
}

// NewMockProviderContextCustom creates a MockProviderContext consisting of the arguments
func NewMockProviderContextCustom(cryptoConfig core.CryptoSuiteConfig, endpointConfig fab.EndpointConfig, identityConfig msp.IdentityConfig, cryptoSuite core.CryptoSuite, signer core.SigningManager, userStore msp.UserStore, identityManager map[string]msp.IdentityManager) *MockProviderContext {
	context := MockProviderContext{
		cryptoSuiteConfig: cryptoConfig,
		endpointConfig:    endpointConfig,
		identityConfig:    identityConfig,
		signingManager:    signer,
		cryptoSuite:       cryptoSuite,
		userStore:         userStore,
		identityManager:   identityManager,
	}
	return &context
}

// SetCryptoSuiteConfig sets the mock cryptosuite configuration.
func (pc *MockProviderContext) SetCryptoSuiteConfig(config core.CryptoSuiteConfig) {
	pc.cryptoSuiteConfig = config
}

// SetEndpointConfig sets the mock endpoint configuration.
func (pc *MockProviderContext) SetEndpointConfig(config fab.EndpointConfig) {
	pc.endpointConfig = config
}

// SetIdentityConfig sets the mock msp identity configuration.
func (pc *MockProviderContext) SetIdentityConfig(config msp.IdentityConfig) {
	pc.identityConfig = config
}

// CryptoSuite returns the mock crypto suite.
func (pc *MockProviderContext) CryptoSuite() core.CryptoSuite {
	return pc.cryptoSuite
}

// CryptoSuiteConfig ...
func (pc *MockProviderContext) CryptoSuiteConfig() core.CryptoSuiteConfig {
	return pc.cryptoSuiteConfig
}

// SigningManager returns the mock signing manager.
func (pc *MockProviderContext) SigningManager() core.SigningManager {
	return pc.signingManager
}

// UserStore returns the mock usser store
func (pc *MockProviderContext) UserStore() msp.UserStore {
	return pc.userStore
}

//IdentityConfig returns the Identity config
func (pc *MockProviderContext) IdentityConfig() msp.IdentityConfig {
	return pc.identityConfig
}

// IdentityManager returns the identity manager
func (pc *MockProviderContext) IdentityManager(orgName string) (msp.IdentityManager, bool) {
	mgr, ok := pc.identityManager[strings.ToLower(orgName)]
	return mgr, ok
}

// PrivateKey returns the crypto suite representation of the private key
func (pc *MockProviderContext) PrivateKey() core.Key {
	return pc.privateKey
}

// PublicVersion returns the public parts of this identity
func (pc *MockProviderContext) PublicVersion() msp.Identity {
	return pc.identity
}

// Sign the message
func (pc *MockProviderContext) Sign(msg []byte) ([]byte, error) {
	return nil, nil
}

//LocalDiscoveryProvider returns a local discovery provider
func (pc *MockProviderContext) LocalDiscoveryProvider() fab.LocalDiscoveryProvider {
	return pc.localDiscoveryProvider
}

//ChannelProvider returns channel provider
func (pc *MockProviderContext) ChannelProvider() fab.ChannelProvider {
	return pc.channelProvider
}

//SetCustomChannelProvider sets custom channel provider for unit-test purposes
func (pc *MockProviderContext) SetCustomChannelProvider(customChannelProvider fab.ChannelProvider) {
	pc.channelProvider = customChannelProvider
}

//InfraProvider returns fabric provider
func (pc *MockProviderContext) InfraProvider() fab.InfraProvider {
	return pc.infraProvider
}

//EndpointConfig returns mock end point config
func (pc *MockProviderContext) EndpointConfig() fab.EndpointConfig {
	return pc.endpointConfig
}

//SetCustomInfraProvider sets custom fabric provider for unit-test purposes
func (pc *MockProviderContext) SetCustomInfraProvider(customInfraProvider fab.InfraProvider) {
	pc.infraProvider = customInfraProvider
}

// GetMetrics not used in this mockcontext
func (pc *MockProviderContext) GetMetrics() *metrics.ClientMetrics {
	return &metrics.ClientMetrics{}
}

// MockContext holds core providers and identity to enable mocking.
type MockContext struct {
	*MockProviderContext
	SigningIdentity msp.SigningIdentity
}

// NewMockContext creates a MockContext consisting of defaults and an identity
func NewMockContext(si msp.SigningIdentity) *MockContext {
	ctx := MockContext{
		MockProviderContext: NewMockProviderContext(),
		SigningIdentity:     si,
	}
	return &ctx
}

// Identifier returns the identifier of that identity
func (m MockContext) Identifier() *msp.IdentityIdentifier {
	return &msp.IdentityIdentifier{ID: m.SigningIdentity.Identifier().ID, MSPID: m.SigningIdentity.Identifier().MSPID}
}

// Verify a signature over some message using this identity as reference
func (m MockContext) Verify(msg []byte, sig []byte) error {
	if m.SigningIdentity == nil {
		return errors.New("anonymous countext")
	}
	return m.SigningIdentity.Verify(msg, sig)
}

// Serialize converts an identity to bytes
func (m MockContext) Serialize() ([]byte, error) {
	if m.SigningIdentity == nil {
		return nil, errors.New("anonymous countext")
	}
	return m.SigningIdentity.Serialize()
}

// EnrollmentCertificate Returns the underlying ECert representing this user’s identity.
func (m MockContext) EnrollmentCertificate() []byte {
	if m.SigningIdentity == nil {
		return nil
	}
	return m.SigningIdentity.EnrollmentCertificate()
}

// Sign the message
func (m MockContext) Sign(msg []byte) ([]byte, error) {
	if m.SigningIdentity == nil {
		return nil, errors.New("anonymous countext")
	}
	return m.SigningIdentity.Sign(msg)
}

// PublicVersion returns the public parts of this identity
func (m MockContext) PublicVersion() msp.Identity {
	if m.SigningIdentity == nil {
		return nil
	}
	return m.SigningIdentity.PublicVersion()
}

// PrivateKey returns the crypto suite representation of the private key
func (m MockContext) PrivateKey() core.Key {
	if m.SigningIdentity == nil {
		return nil
	}
	return m.SigningIdentity.PrivateKey()
}

// MockChannelContext holds the client context plus channel-specific entities
type MockChannelContext struct {
	*MockContext
	channelID string
	Channel   fab.ChannelService
}

// NewMockChannelContext returns a new MockChannelContext
func NewMockChannelContext(context *MockContext, channelID string) *MockChannelContext {
	return &MockChannelContext{
		MockContext: context,
		channelID:   channelID,
	}
}

// ChannelService returns the ChannelService
func (c *MockChannelContext) ChannelService() fab.ChannelService {
	return c.Channel
}

// ChannelID returns the channel ID
func (c *MockChannelContext) ChannelID() string {
	return c.channelID
}

// GetMetrics not used in this mockcontext
func (c *MockChannelContext) GetMetrics() *metrics.ClientMetrics {
	return &metrics.ClientMetrics{}
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
	return th.MockID
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
	user := mspmocks.NewMockSigningIdentity("test", "Org1MSP")

	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, err
	}

	creator, err := user.Serialize()
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
