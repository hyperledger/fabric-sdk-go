/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// InfraProvider enables access to fabric objects such as peer and user based on config or
type InfraProvider interface {
	CreateChannelLedger(ic IdentityContext, name string) (ChannelLedger, error)
	CreateChannelConfig(user IdentityContext, name string) (ChannelConfig, error)
	CreateChannelTransactor(ic IdentityContext, cfg ChannelCfg) (Transactor, error)
	CreateChannelMembership(cfg ChannelCfg) (ChannelMembership, error)
	CreateEventHub(ic IdentityContext, name string) (EventHub, error)
	CreatePeerFromConfig(peerCfg *core.NetworkPeer) (Peer, error)
	CreateOrdererFromConfig(cfg *core.OrdererConfig) (Orderer, error)
}

// SelectionProvider is used to select peers for endorsement
type SelectionProvider interface {
	NewSelectionService(channelID string) (SelectionService, error)
}

// SelectionService selects peers for endorsement and commit events
type SelectionService interface {
	// GetEndorsersForChaincode returns a set of peers that should satisfy the endorsement
	// policies of all of the given chaincodes
	GetEndorsersForChaincode(channelPeers []Peer, chaincodeIDs ...string) ([]Peer, error)
}

// DiscoveryProvider is used to discover peers on the network
type DiscoveryProvider interface {
	NewDiscoveryService(channelID string) (DiscoveryService, error)
}

// DiscoveryService is used to discover eligible peers on specific channel
type DiscoveryService interface {
	GetPeers() ([]Peer, error)
}

// TargetFilter allows for filtering target peers
type TargetFilter interface {
	// Accept returns true if peer should be included in the list of target peers
	Accept(peer Peer) bool
}

// Providers represents the SDK configured service providers context.
type Providers interface {
	DiscoveryProvider() DiscoveryProvider
	SelectionProvider() SelectionProvider
	ChannelProvider() ChannelProvider
	FabricProvider() InfraProvider
}
