/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	reqContext "context"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"google.golang.org/grpc"
)

// ClientContext contains the client context
// TODO: This is a duplicate of context.Client since importing context.Client causes
// a circular import error. This problem should be addressed in a future patch.
type ClientContext interface {
	core.Providers
	msp.Providers
	Providers
	msp.SigningIdentity
}

// InfraProvider enables access to fabric objects such as peer and user based on config or
type InfraProvider interface {
	CreateChannelConfig(name string) (ChannelConfig, error)
	CreateChannelCfg(ctx ClientContext, channelID string) (ChannelCfg, error)
	CreateChannelTransactor(reqCtx reqContext.Context, cfg ChannelCfg) (Transactor, error)
	CreateChannelMembership(ctx ClientContext, channelID string) (ChannelMembership, error)
	CreateEventService(ctx ClientContext, channelID string) (EventService, error)
	CreatePeerFromConfig(peerCfg *core.NetworkPeer) (Peer, error)
	CreateOrdererFromConfig(cfg *core.OrdererConfig) (Orderer, error)
	CommManager() CommManager
	Close()
}

// SelectionProvider is used to select peers for endorsement
type SelectionProvider interface {
	CreateSelectionService(channelID string) (SelectionService, error)
}

// SelectionService selects peers for endorsement and commit events
type SelectionService interface {
	// GetEndorsersForChaincode returns a set of peers that should satisfy the endorsement
	// policies of all of the given chaincodes.
	// A set of options may be provided to the selection service. Note that the type of options
	// may vary depending on the specific selection service implementation.
	GetEndorsersForChaincode(chaincodeIDs []string, opts ...options.Opt) ([]Peer, error)
}

// DiscoveryProvider is used to discover peers on the network
type DiscoveryProvider interface {
	CreateDiscoveryService(channelID string) (DiscoveryService, error)
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

// CommManager enables network communication.
type CommManager interface {
	DialContext(ctx reqContext.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	ReleaseConn(conn *grpc.ClientConn)
}

// Providers represents the SDK configured service providers context.
type Providers interface {
	DiscoveryProvider() DiscoveryProvider
	SelectionProvider() SelectionProvider
	ChannelProvider() ChannelProvider
	InfraProvider() InfraProvider
}
