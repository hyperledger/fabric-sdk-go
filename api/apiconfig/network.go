/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apiconfig

// NetworkConfig provides a static definition of a Hyperledger Fabric network
type NetworkConfig struct {
	// Orderers list of ordering service nodes
	Orderers map[string]OrdererConfig
	// Organizations map of member organizations forming the network
	Organizations map[string]OrganizationConfig
}

// OrganizationConfig defines a member organization in the fabric network
type OrganizationConfig struct {
	// MspID Membership Service Provider ID for this organization
	MspID string
	// CA config defines the fabric-ca instance that issues identities for this org
	CA CAConfig
	// Peers a list of peers that are part of this organization
	Peers map[string]PeerConfig
}

// PeerConfig A set of configurations required to connect to a Fabric peer
type PeerConfig struct {
	// Host peer host
	Host string
	// Port peer port
	Port int
	// EventHost peer event host
	EventHost string
	// EventPort peer event port
	EventPort int
	// Primary is the the primary peer for the organization
	Primary bool
	// TLS configurations
	TLS TLSConfig
}

// OrdererConfig A set of configurations required to connect to an
// Ordering Service node
type OrdererConfig struct {
	// Host orderer host
	Host string
	// Port orderer port
	Port int
	// TLS configurations
	TLS TLSConfig
}

// CAConfig A set of configurations required to connect to fabric-ca
type CAConfig struct {
	// TLSEnabled flag
	TLSEnabled bool
	// Name CA name
	Name string
	// ServerURL server URL
	ServerURL string
	// TLS configurations
	TLS MutualTLSConfig
}

// TLSConfig TLS configurations
type TLSConfig struct {
	// Certificate root certificate path
	Certificate string
	// ServerHostOverride override host name for certificate validation.
	// For testing only.
	ServerHostOverride string
}

// MutualTLSConfig Mutual TLS configurations
type MutualTLSConfig struct {
	// Certfiles root certificates for TLS validation (Comma serparated path list)
	Certfiles string
	// Client client TLS information
	Client struct {
		// Keyfile client key path
		Keyfile string
		// Certfile client cert path
		Certfile string
	}
}
