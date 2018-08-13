/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	reqContext "context"
	"crypto/tls"
	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"

	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
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

// CommManager enables network communication.
type CommManager interface {
	DialContext(ctx reqContext.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	ReleaseConn(conn *grpc.ClientConn)
}

//EndpointConfig contains endpoint network configurations
type EndpointConfig interface {
	Timeout(TimeoutType) time.Duration
	OrderersConfig() []OrdererConfig
	OrdererConfig(nameOrURL string) (*OrdererConfig, bool)
	PeersConfig(org string) ([]PeerConfig, bool)
	PeerConfig(nameOrURL string) (*PeerConfig, bool)
	NetworkConfig() *NetworkConfig
	NetworkPeers() []NetworkPeer
	ChannelConfig(name string) (*ChannelEndpointConfig, bool)
	ChannelPeers(name string) ([]ChannelPeer, bool)
	ChannelOrderers(name string) ([]OrdererConfig, bool)
	TLSCACertPool() CertPool
	EventServiceConfig() EventServiceConfig
	TLSClientCerts() []tls.Certificate
	CryptoConfigPath() string
}

// EventServiceConfig specifies configuration options for the event service
type EventServiceConfig interface {
	// BlockHeightLagThreshold returns the block height lag threshold. This value is used for choosing a peer
	// to connect to. If a peer is lagging behind the most up-to-date peer by more than the given number of
	// blocks then it will be excluded from selection.
	// If set to 0 then only the most up-to-date peers are considered.
	// If set to -1 then all peers (regardless of block height) are considered for selection.
	BlockHeightLagThreshold() int

	// ReconnectBlockHeightLagThreshold - if >0 then the event client will disconnect from the peer if the peer's
	// block height falls behind the specified number of blocks and will reconnect to a better performing peer.
	// If set to 0 (default) then the peer will not disconnect based on block height.
	// NOTE: Setting this value too low may cause the event client to disconnect/reconnect too frequently, thereby
	// affecting performance.
	ReconnectBlockHeightLagThreshold() int

	// BlockHeightMonitorPeriod is the period in which the connected peer's block height is monitored. Note that this
	// value is only relevant if reconnectBlockHeightLagThreshold >0.
	BlockHeightMonitorPeriod() time.Duration
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
}

// CertPool is a thread safe wrapper around the x509 standard library
// cert pool implementation.
type CertPool interface {
	// Get returns the cert pool, optionally adding the provided certs
	Get() (*x509.CertPool, error)
	//Add allows adding certificates to CertPool
	//Call Get() after Add() to get the updated certpool
	Add(certs ...*x509.Certificate)
}
