/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
)

// Identity supplies the serialized identity and key reference.
type Identity interface {
	MspID() string
	SerializedIdentity() ([]byte, error)
	PrivateKey() core.Key
}

// Session primarily represents the session and identity context
type Session interface {
	Identity
}

// Client supplies the configuration and signing identity to client objects.
type Client interface {
	core.Providers
	Identity
}

// Providers represents the SDK configured providers context.
type Providers interface {
	core.Providers
	fab.Providers
}

//FabContext implementation for Providers interface
type FabContext struct {
	config            core.Config
	stateStore        core.KVStore
	cryptoSuite       core.CryptoSuite
	discoveryProvider fab.DiscoveryProvider
	selectionProvider fab.SelectionProvider
	signingManager    core.SigningManager
	identityManager   map[string]core.IdentityManager
	fabricProvider    fab.InfraProvider
	channelProvider   fab.ChannelProvider
}

//SDKContext implementation for Providers interface.
type SDKContext struct {
	FabContext
}

// Config returns the Config provider of sdk.
func (c *FabContext) Config() core.Config {
	return c.config
}

// CryptoSuite returns the BCCSP provider of sdk.
func (c *FabContext) CryptoSuite() core.CryptoSuite {
	return c.cryptoSuite
}

// IdentityManager returns identity manager for organization
func (c *FabContext) IdentityManager(orgName string) (core.IdentityManager, bool) {
	mgr, ok := c.identityManager[strings.ToLower(orgName)]
	return mgr, ok
}

// SigningManager returns signing manager
func (c *FabContext) SigningManager() core.SigningManager {
	return c.signingManager
}

// StateStore returns state store
func (c *FabContext) StateStore() core.KVStore {
	return c.stateStore
}

// DiscoveryProvider returns discovery provider
func (c *SDKContext) DiscoveryProvider() fab.DiscoveryProvider {
	return c.discoveryProvider
}

// SelectionProvider returns selection provider
func (c *SDKContext) SelectionProvider() fab.SelectionProvider {
	return c.selectionProvider
}

// ChannelProvider provides channel services.
func (c *SDKContext) ChannelProvider() fab.ChannelProvider {
	return c.channelProvider
}

// FabricProvider provides fabric objects such as peer and user
func (c *SDKContext) FabricProvider() fab.InfraProvider {
	return c.fabricProvider
}

//FabContextParams parameter for creating FabContext
type FabContextParams func(opts *FabContext)

//WithConfig sets config to FabContext
func WithConfig(config core.Config) FabContextParams {
	return func(ctx *FabContext) {
		ctx.config = config
	}
}

//WithStateStore sets state store to FabContext
func WithStateStore(stateStore core.KVStore) FabContextParams {
	return func(ctx *FabContext) {
		ctx.stateStore = stateStore
	}
}

//WithCryptoSuite sets cryptosuite parameter to FabContext
func WithCryptoSuite(cryptoSuite core.CryptoSuite) FabContextParams {
	return func(ctx *FabContext) {
		ctx.cryptoSuite = cryptoSuite
	}
}

//WithDiscoveryProvider sets discoveryProvider to FabContext
func WithDiscoveryProvider(discoveryProvider fab.DiscoveryProvider) FabContextParams {
	return func(ctx *FabContext) {
		ctx.discoveryProvider = discoveryProvider
	}
}

//WithSelectionProvider sets selectionProvider to FabContext
func WithSelectionProvider(selectionProvider fab.SelectionProvider) FabContextParams {
	return func(ctx *FabContext) {
		ctx.selectionProvider = selectionProvider
	}
}

//WithSigningManager sets signingManager to FabContext
func WithSigningManager(signingManager core.SigningManager) FabContextParams {
	return func(ctx *FabContext) {
		ctx.signingManager = signingManager
	}
}

//WithIdentityManager sets identityManagers maps to FabContext
func WithIdentityManager(identityManagers map[string]core.IdentityManager) FabContextParams {
	return func(ctx *FabContext) {
		ctx.identityManager = identityManagers
	}
}

//WithFabricProvider sets fabricProvider maps to FabContext
func WithFabricProvider(fabricProvider fab.InfraProvider) FabContextParams {
	return func(ctx *FabContext) {
		ctx.fabricProvider = fabricProvider
	}
}

//WithChannelProvider sets channelProvider to FabContext
func WithChannelProvider(channelProvider fab.ChannelProvider) FabContextParams {
	return func(ctx *FabContext) {
		ctx.channelProvider = channelProvider
	}
}

//CreateFabContext creates FabContext using parameters passed
func CreateFabContext(params ...FabContextParams) *FabContext {
	fabCtx := FabContext{}
	for _, param := range params {
		param(&fabCtx)
	}
	return &fabCtx
}
