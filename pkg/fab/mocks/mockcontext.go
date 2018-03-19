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
	config "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"

	"strings"

	"fmt"

	"github.com/pkg/errors"
)

// MockProviderContext holds core providers to enable mocking.
type MockProviderContext struct {
	config            config.Config
	cryptoSuite       core.CryptoSuite
	signingManager    core.SigningManager
	userStore         msp.UserStore
	identityManager   map[string]msp.IdentityManager
	privateKey        core.Key
	identity          msp.SigningIdentity
	discoveryProvider fab.DiscoveryProvider
	selectionProvider fab.SelectionProvider
	infraProvider     fab.InfraProvider
	channelProvider   fab.ChannelProvider
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
			panic(fmt.Errorf("error creating MockProviderContext: %v", err))
		}
	}

	im := make(map[string]msp.IdentityManager)
	for org := range users.users {
		im[org] = NewMockIdentityManager(WithUsers(users.users[org]))
	}

	context := MockProviderContext{
		config:            NewMockConfig(),
		signingManager:    mocks.NewMockSigningManager(),
		cryptoSuite:       &MockCryptoSuite{},
		userStore:         &mspmocks.MockUserStore{},
		identityManager:   im,
		discoveryProvider: &MockStaticDiscoveryProvider{},
		selectionProvider: &MockSelectionProvider{},
		infraProvider:     &MockInfraProvider{},
		channelProvider:   &MockChannelProvider{},
		identity:          users.identity,
	}
	return &context
}

// NewMockProviderContextCustom creates a MockProviderContext consisting of the arguments
func NewMockProviderContextCustom(config config.Config, cryptoSuite core.CryptoSuite, signer core.SigningManager, userStore msp.UserStore, identityManager map[string]msp.IdentityManager) *MockProviderContext {
	context := MockProviderContext{
		config:          config,
		signingManager:  signer,
		cryptoSuite:     cryptoSuite,
		userStore:       userStore,
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

// UserStore returns the mock usser store
func (pc *MockProviderContext) UserStore() msp.UserStore {
	return pc.userStore
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

//DiscoveryProvider returns discovery provider
func (pc *MockProviderContext) DiscoveryProvider() fab.DiscoveryProvider {
	return pc.discoveryProvider
}

//SelectionProvider returns selection provider
func (pc *MockProviderContext) SelectionProvider() fab.SelectionProvider {
	return pc.selectionProvider
}

//ChannelProvider returns channel provider
func (pc *MockProviderContext) ChannelProvider() fab.ChannelProvider {
	return pc.channelProvider
}

//InfraProvider returns fabric provider
func (pc *MockProviderContext) InfraProvider() fab.InfraProvider {
	return pc.infraProvider
}

//SetCustomInfraProvider sets custom fabric provider for unit-test purposes
func (pc *MockProviderContext) SetCustomInfraProvider(customInfraProvider fab.InfraProvider) {
	pc.infraProvider = customInfraProvider
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

// NewMockContextWithCustomDiscovery creates a MockContext consisting of defaults and an identity
func NewMockContextWithCustomDiscovery(ic msp.SigningIdentity, discPvdr fab.DiscoveryProvider) *MockContext {
	mockCtx := NewMockProviderContext(WithProviderUser(ic.Identifier().ID, ic.Identifier().MSPID))
	mockCtx.discoveryProvider = discPvdr
	ctx := MockContext{
		MockProviderContext: mockCtx,
		SigningIdentity:     ic,
	}
	return &ctx
}

// MockChannelContext holds the client context plus channel-specific entities
type MockChannelContext struct {
	*MockContext
	channelID string
	Discovery fab.DiscoveryService
	Selection fab.SelectionService
	Channel   fab.ChannelService
}

// NewMockChannelContext returns a new MockChannelContext
func NewMockChannelContext(context *MockContext, channelID string) *MockChannelContext {
	return &MockChannelContext{
		MockContext: context,
		channelID:   channelID,
	}
}

// DiscoveryService returns a discovery service
func (c *MockChannelContext) DiscoveryService() fab.DiscoveryService {
	return c.Discovery
}

// SelectionService returns the selection service
func (c *MockChannelContext) SelectionService() fab.SelectionService {
	return c.Selection
}

// ChannelService returns the ChannelService
func (c *MockChannelContext) ChannelService() fab.ChannelService {
	return c.Channel
}

// ChannelID returns the channel ID
func (c *MockChannelContext) ChannelID() string {
	return c.channelID
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
