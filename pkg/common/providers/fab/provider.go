/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	reqContext "context"
	"crypto/tls"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	commtls "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm/tls"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics"
	"google.golang.org/grpc"
)

// ClientContext contains the client context
type ClientContext interface {
	core.Providers
	msp.Providers
	Providers
	msp.SigningIdentity
}

// InfraProvider enables access to fabric objects such as peer and user based on config or
type InfraProvider interface {
	CreatePeerFromConfig(peerCfg *NetworkPeer) (Peer, error)
	CreateOrdererFromConfig(cfg *OrdererConfig) (Orderer, error)
	CommManager() CommManager
	Close()
}

// ChaincodeCall contains the ID of the chaincode as well
// as an optional set of private data collections that may be
// accessed by the chaincode.
type ChaincodeCall struct {
	ID          string
	Collections []string
}

// SelectionService selects peers for endorsement and commit events
type SelectionService interface {
	// GetEndorsersForChaincode returns a set of peers that should satisfy the endorsement
	// policies of all of the given chaincodes.
	// A set of options may be provided to the selection service. Note that the type of options
	// may vary depending on the specific selection service implementation.
	GetEndorsersForChaincode(chaincodes []*ChaincodeCall, opts ...options.Opt) ([]Peer, error)
}

// DiscoveryService is used to discover eligible peers on specific channel
type DiscoveryService interface {
	GetPeers() ([]Peer, error)
}

// LocalDiscoveryProvider is used to discover peers in the local MSP
type LocalDiscoveryProvider interface {
	CreateLocalDiscoveryService(mspID string) (DiscoveryService, error)
}

// TargetFilter allows for filtering target peers
type TargetFilter interface {
	// Accept returns true if peer should be included in the list of target peers
	Accept(peer Peer) bool
}

// TargetSorter allows for sorting target peers
type TargetSorter interface {
	// Returns the sorted peers
	Sort(peers []Peer) []Peer
}

// PrioritySelector determines how likely a peer is to be
// selected over another peer
type PrioritySelector interface {
	// A positive return value means peer1 is selected
	// A negative return value means the peer2 is selected
	// Zero return value means their priorities are the same
	Compare(peer1, peer2 Peer) int
}

// CommManager enables network communication.
type CommManager interface {
	DialContext(ctx reqContext.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	ReleaseConn(conn *grpc.ClientConn)
}

//EndpointConfig contains endpoint network configurations
type EndpointConfig interface {
	Timeout(TimeoutType) time.Duration
	OrderersConfig() []OrdererConfig
	OrdererConfig(nameOrURL string) (*OrdererConfig, bool, bool)
	PeersConfig(org string) ([]PeerConfig, bool)
	PeerConfig(nameOrURL string) (*PeerConfig, bool)
	NetworkConfig() *NetworkConfig
	NetworkPeers() []NetworkPeer
	ChannelConfig(name string) *ChannelEndpointConfig
	ChannelPeers(name string) []ChannelPeer
	ChannelOrderers(name string) []OrdererConfig
	TLSCACertPool() commtls.CertPool
	TLSClientCerts() []tls.Certificate
	CryptoConfigPath() string
}

// TimeoutType enumerates the different types of outgoing connections
type TimeoutType int

const (
	// PeerConnection connection timeout
	PeerConnection TimeoutType = iota
	// EventReg connection timeout
	EventReg
	// Query timeout
	Query
	// Execute timeout
	Execute
	// OrdererConnection orderer connection timeout
	OrdererConnection
	// OrdererResponse orderer response timeout
	OrdererResponse
	// DiscoveryGreylistExpiry discovery Greylist expiration period
	DiscoveryGreylistExpiry
	// ConnectionIdle is the timeout for closing idle connections
	ConnectionIdle
	// CacheSweepInterval is the duration between cache sweeps
	CacheSweepInterval
	// EventServiceIdle is the timeout for closing the event service connection
	EventServiceIdle
	// PeerResponse peer response timeout
	PeerResponse
	// ResMgmt timeout is default overall timeout for all resource management operations
	ResMgmt
	// ChannelConfigRefresh channel configuration refresh interval
	ChannelConfigRefresh
	// ChannelMembershipRefresh channel membership refresh interval
	ChannelMembershipRefresh
	// DiscoveryConnection discovery connection timeout
	DiscoveryConnection
	// DiscoveryResponse discovery response timeout
	DiscoveryResponse
	// DiscoveryServiceRefresh discovery service refresh interval
	DiscoveryServiceRefresh
	// SelectionServiceRefresh selection service refresh interval
	SelectionServiceRefresh
)

// Providers represents the SDK configured service providers context.
type Providers interface {
	LocalDiscoveryProvider() LocalDiscoveryProvider
	ChannelProvider() ChannelProvider
	InfraProvider() InfraProvider
	EndpointConfig() EndpointConfig
	MetricsProvider
}

// MetricsProvider represents a provider of metrics.
type MetricsProvider interface {
	GetMetrics() *metrics.ClientMetrics
}
