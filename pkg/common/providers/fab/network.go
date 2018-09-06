/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"crypto/x509"

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
	MSPID string
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
