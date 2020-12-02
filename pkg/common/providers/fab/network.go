/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"crypto/x509"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
)

// NetworkConfig provides a static definition of endpoint configuration network
type NetworkConfig struct {
	Channels      map[string]ChannelEndpointConfig
	Organizations map[string]OrganizationConfig
	Orderers      map[string]OrdererConfig
	Peers         map[string]PeerConfig
}

// ChannelEndpointConfig provides the definition of channels for the network
type ChannelEndpointConfig struct {
	// Orderers list of ordering service nodes
	Orderers []string
	// Peers a list of peer-channels that are part of this organization
	// to get the real Peer config object, use the Name field and fetch NetworkConfig.Peers[Name]
	Peers map[string]PeerChannelConfig
	//Policies list of policies for channel
	Policies ChannelPolicies
}

//ChannelPolicies defines list of policies defined for a channel
type ChannelPolicies struct {
	//Policy for querying channel block
	QueryChannelConfig QueryChannelConfigPolicy
	Discovery          DiscoveryPolicy
	Selection          SelectionPolicy
	EventService       EventServicePolicy
}

//QueryChannelConfigPolicy defines policy for channelConfigBlock
type QueryChannelConfigPolicy struct {
	MinResponses int
	MaxTargets   int
	RetryOpts    retry.Opts
}

//DiscoveryPolicy defines policy for discovery
type DiscoveryPolicy struct {
	MinResponses int
	MaxTargets   int
	RetryOpts    retry.Opts
}

// SelectionSortingStrategy is the endorser selection sorting strategy
type SelectionSortingStrategy string

const (
	// BlockHeightPriority (default) is a load-balancing selection sorting strategy
	// which also prioritizes peers at a block height that is above a certain "lag" threshold.
	BlockHeightPriority SelectionSortingStrategy = "BlockHeightPriority"

	// Balanced is a load-balancing selection sorting strategy
	Balanced SelectionSortingStrategy = "Balanced"
)

// BalancerType is the load-balancer type
type BalancerType string

const (
	// RoundRobin (default) chooses endorsers in a round-robin fashion
	RoundRobin BalancerType = "RoundRobin"

	// Random chooses endorsers randomly
	Random BalancerType = "Random"
)

//SelectionPolicy defines policy for selection
type SelectionPolicy struct {
	// SortingStrategy is the endorser sorting strategy to use
	SortingStrategy SelectionSortingStrategy

	// BalancerType is the balancer to use in order to load-balance calls to endorsers
	Balancer BalancerType

	// BlockHeightLagThreshold is the number of blocks from the highest block number of a group of peers
	// that a peer can lag behind and still be considered to be up-to-date. These peers will be sorted
	// using the given Balancer. If a peer's block height falls behind this threshold then it will be
	// demoted to a lower priority list of peers which will be sorted according to block height.
	// Note: This property only applies to BlockHeightPriority sorter
	BlockHeightLagThreshold int
}

// PeerChannelConfig defines the peer capabilities
type PeerChannelConfig struct {
	EndorsingPeer  bool
	ChaincodeQuery bool
	LedgerQuery    bool
	EventSource    bool
}

// ChannelPeer combines channel peer info with raw peerConfig info
type ChannelPeer struct {
	PeerChannelConfig
	NetworkPeer
}

// NetworkPeer combines peer info with MSP info
type NetworkPeer struct {
	PeerConfig
	MSPID      string
	Properties map[Property]interface{}
}

// OrganizationConfig provides the definition of an organization in the network
type OrganizationConfig struct {
	MSPID                  string
	CryptoPath             string
	Users                  map[string]CertKeyPair
	Peers                  []string
	CertificateAuthorities []string
}

// OrdererConfig defines an orderer configuration
type OrdererConfig struct {
	URL         string
	GRPCOptions map[string]interface{}
	TLSCACert   *x509.Certificate
}

// PeerConfig defines a peer configuration
type PeerConfig struct {
	URL         string
	GRPCOptions map[string]interface{}
	TLSCACert   *x509.Certificate
}

// CertKeyPair contains the private key and certificate
type CertKeyPair struct {
	Cert []byte
	Key  []byte
}

// ResolverStrategy is the peer resolver type
type ResolverStrategy string

const (
	// BalancedStrategy is a peer resolver strategy that chooses peers based on a configured load balancer
	BalancedStrategy ResolverStrategy = "Balanced"

	// MinBlockHeightStrategy is a peer resolver strategy that chooses the best peer according to a block height lag threshold.
	// The maximum block height of all peers is determined and the peers whose block heights are under the maximum height but above
	// a provided "lag" threshold are load balanced. The other peers are not considered.
	MinBlockHeightStrategy ResolverStrategy = "MinBlockHeight"

	// PreferOrgStrategy is a peer resolver strategy that determines which peers are suitable based on block height lag threshold,
	// although will prefer the peers in the current org (as long as their block height is above a configured threshold).
	// If none of the peers from the current org are suitable then a peer from another org is chosen.
	PreferOrgStrategy ResolverStrategy = "PreferOrg"
)

// MinBlockHeightResolverMode specifies the behaviour of the MinBlockHeight resolver strategy.
type MinBlockHeightResolverMode string

const (
	// ResolveByThreshold resolves to peers based on block height lag threshold.
	ResolveByThreshold MinBlockHeightResolverMode = "ResolveByThreshold"

	// ResolveLatest resolves to peers with the most up-to-date block height
	ResolveLatest MinBlockHeightResolverMode = "ResolveLatest"
)

// EnabledDisabled specifies whether or not a feature is enabled
type EnabledDisabled string

const (
	// Enabled indicates that the feature is enabled.
	Enabled EnabledDisabled = "Enabled"

	// Disabled indicates that the feature is disabled.
	Disabled EnabledDisabled = "Disabled"
)

// EventServicePolicy specifies the policy for the event service
type EventServicePolicy struct {
	// ResolverStrategy returns the peer resolver strategy to use when connecting to a peer
	// Default: MinBlockHeightPeerResolver
	ResolverStrategy ResolverStrategy

	// Balancer is the balancer to use when choosing a peer to connect to
	Balancer BalancerType

	// MinBlockHeightResolverMode specifies the behaviour of the MinBlockHeight resolver. Note that this
	// parameter is used when ResolverStrategy is either MinBlockHeightStrategy or PreferOrgStrategy.
	// ResolveByThreshold (default): resolves to peers based on block height lag threshold, as specified by BlockHeightLagThreshold.
	// MinBlockHeightResolverMode: then only the peers with the latest block heights are chosen.
	MinBlockHeightResolverMode MinBlockHeightResolverMode

	// BlockHeightLagThreshold returns the block height lag threshold. This value is used for choosing a peer
	// to connect to. If a peer is lagging behind the most up-to-date peer by more than the given number of
	// blocks then it will be excluded from selection.
	BlockHeightLagThreshold int

	// PeerMonitor indicates whether or not to enable the peer monitor.
	PeerMonitor EnabledDisabled

	// ReconnectBlockHeightLagThreshold - if >0 then the event client will disconnect from the peer if the peer's
	// block height falls behind the specified number of blocks and will reconnect to a better performing peer.
	// If set to 0 (default) then the peer will not disconnect based on block height.
	// NOTE: Setting this value too low may cause the event client to disconnect/reconnect too frequently, thereby
	// affecting performance.
	ReconnectBlockHeightLagThreshold int

	// PeerMonitorPeriod is the period in which the connected peer is monitored to see if
	// the event client should disconnect from it and reconnect to another peer.
	// If set to 0 then the peer will not be monitored and will not be disconnected.
	PeerMonitorPeriod time.Duration
}
