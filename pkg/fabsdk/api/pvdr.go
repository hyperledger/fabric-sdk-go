/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
)

// FabricProvider enables access to fabric objects such as peer and user based on config or
type FabricProvider interface {
	CreateChannelLedger(ic context.IdentityContext, name string) (fab.ChannelLedger, error)
	CreateChannelConfig(user context.IdentityContext, name string) (fab.ChannelConfig, error)
	CreateResourceClient(user context.IdentityContext) (api.Resource, error)
	CreateChannelTransactor(ic context.IdentityContext, cfg fab.ChannelCfg) (fab.Transactor, error)
	CreateChannelMembership(cfg fab.ChannelCfg) (fab.ChannelMembership, error)
	CreateEventHub(ic context.IdentityContext, name string) (fab.EventHub, error)

	CreatePeerFromConfig(peerCfg *core.NetworkPeer) (fab.Peer, error)
	CreateOrdererFromConfig(cfg *core.OrdererConfig) (fab.Orderer, error)
}

// Providers represents the SDK configured providers context.
type Providers interface {
	CoreProviders
	SvcProviders
}

// CoreProviders represents the SDK configured core providers context.
type CoreProviders interface {
	CryptoSuite() core.CryptoSuite
	StateStore() contextApi.KVStore
	Config() core.Config
	SigningManager() contextApi.SigningManager
	FabricProvider() FabricProvider
}

// SvcProviders represents the SDK configured service providers context.
type SvcProviders interface {
	DiscoveryProvider() fab.DiscoveryProvider
	SelectionProvider() fab.SelectionProvider
	ChannelProvider() fab.ChannelProvider
}
