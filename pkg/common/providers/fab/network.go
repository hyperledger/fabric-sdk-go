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
	Name          string
	Description   string
	Version       string
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
}

//QueryChannelConfigPolicy defines opts for channelConfigBlock
type QueryChannelConfigPolicy struct {
	MinResponses int
	MaxTargets   int
	RetryOpts    retry.Opts
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
	EventURL    string
	GRPCOptions map[string]interface{}
	TLSCACert   *x509.Certificate
}

// MatchConfig contains match pattern and substitution pattern
// for pattern matching of network configured hostnames or channel names with static config
type MatchConfig struct {
	Pattern string

	// these are used for hostname mapping
	URLSubstitutionExp                  string
	EventURLSubstitutionExp             string
	SSLTargetOverrideURLSubstitutionExp string
	MappedHost                          string

	// this is used for Name mapping instead of hostname mappings
	MappedName string
}

// CertKeyPair contains the private key and certificate
type CertKeyPair struct {
	Cert []byte
	Key  []byte
}
