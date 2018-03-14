/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package core

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
)

// NetworkConfig provides a static definition of a Hyperledger Fabric network
type NetworkConfig struct {
	Name                   string
	Xtype                  string
	Description            string
	Version                string
	Client                 ClientConfig
	Channels               map[string]ChannelConfig
	Organizations          map[string]OrganizationConfig
	Orderers               map[string]OrdererConfig
	Peers                  map[string]PeerConfig
	CertificateAuthorities map[string]CAConfig
	EntityMatchers         map[string][]MatchConfig
}

// ClientConfig provides the definition of the client configuration
type ClientConfig struct {
	Organization    string
	Logging         LoggingType
	CryptoConfig    CCType
	TLS             TLSType
	TLSCerts        MutualTLSConfig
	CredentialStore CredentialStoreType
}

// LoggingType defines the level of logging
type LoggingType struct {
	Level string
}

// CCType defines the path to crypto keys and certs
type CCType struct {
	Path string
}

// TLSType defines whether or not TLS is enabled
type TLSType struct {
	Enabled bool
}

// CredentialStoreType defines pluggable KV store properties
type CredentialStoreType struct {
	Path        string
	CryptoStore struct {
		Path string
	}
	Wallet string
}

// ChannelConfig provides the definition of channels for the network
type ChannelConfig struct {
	// Orderers list of ordering service nodes
	Orderers []string
	// Peers a list of peer-channels that are part of this organization
	// to get the real Peer config object, use the Name field and fetch NetworkConfig.Peers[Name]
	Peers map[string]PeerChannelConfig
	// Chaincodes list of services
	Chaincodes []string
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
	Users                  map[string]TLSKeyPair
	Peers                  []string
	CertificateAuthorities []string
	AdminPrivateKey        endpoint.TLSConfig
	SignedCert             endpoint.TLSConfig
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
	EventURL    string
	GRPCOptions map[string]interface{}
	TLSCACerts  endpoint.TLSConfig
}

// CAConfig defines a CA configuration
type CAConfig struct {
	URL         string
	HTTPOptions map[string]interface{}
	TLSCACerts  MutualTLSConfig
	Registrar   EnrollCredentials
	CAName      string
}

// EnrollCredentials holds credentials used for enrollment
type EnrollCredentials struct {
	EnrollID     string
	EnrollSecret string
}

// MutualTLSConfig Mutual TLS configurations
type MutualTLSConfig struct {
	Pem []string
	// Certfiles root certificates for TLS validation (Comma separated path list)
	Path string

	//Client TLS information
	Client TLSKeyPair
}

// TLSKeyPair contains the private key and certificate for TLS encryption
type TLSKeyPair struct {
	Key  endpoint.TLSConfig
	Cert endpoint.TLSConfig
}

// MatchConfig contains match pattern and substitution pattern
// for pattern matching of network configured hostnames with static config
type MatchConfig struct {
	Pattern                             string
	URLSubstitutionExp                  string
	EventURLSubstitutionExp             string
	SSLTargetOverrideURLSubstitutionExp string
	MappedHost                          string
}
