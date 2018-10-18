/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
)

// ClientConfig provides the definition of the client configuration
type ClientConfig struct {
	Organization string
	TLSCerts     ClientTLSConfig
}

// ClientTLSConfig contains the client TLS configuration
type ClientTLSConfig struct {
	//Client TLS information
	Client endpoint.TLSKeyPair
}

// OrdererConfig defines an orderer configuration
type OrdererConfig struct {
	URL         string
	GRPCOptions map[string]interface{}
	TLSCACerts  endpoint.TLSConfig
}

// PeerConfig defines a peer configuration
type PeerConfig struct {
	URL         string
	GRPCOptions map[string]interface{}
	TLSCACerts  endpoint.TLSConfig
}

// OrganizationConfig provides the definition of an organization in the network
type OrganizationConfig struct {
	MSPID                  string
	CryptoPath             string
	Users                  map[string]endpoint.TLSKeyPair
	Peers                  []string
	CertificateAuthorities []string
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
	//Policy for querying discovery
	Discovery DiscoveryPolicy
	//Policy for endorser selection
	Selection SelectionPolicy
	//Policy for event service
	EventService EventServicePolicy
}

//QueryChannelConfigPolicy defines opts for channelConfigBlock
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

// EventServicePolicy specifies the policy for the event service
type EventServicePolicy struct {
	ResolverStrategy                 string
	MinBlockHeightResolverMode       string
	Balancer                         BalancerType
	BlockHeightLagThreshold          int
	PeerMonitor                      string
	ReconnectBlockHeightLagThreshold int
	PeerMonitorPeriod                time.Duration
}

// PeerChannelConfig defines the peer capabilities
type PeerChannelConfig struct {
	EndorsingPeer  bool
	ChaincodeQuery bool
	LedgerQuery    bool
	EventSource    bool
}

// MatchConfig contains match pattern and substitution pattern
// for pattern matching of network configured hostnames or channel names with static config
type MatchConfig struct {
	Pattern string

	// these are used for hostname mapping
	URLSubstitutionExp                  string
	SSLTargetOverrideURLSubstitutionExp string
	MappedHost                          string

	// this is used for Name mapping instead of hostname mappings
	MappedName string

	//IgnoreEndpoint option to exclude given entity from any kind of search or from entity list
	IgnoreEndpoint bool
}
