/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apiconfig

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
}

// ClientConfig provides the definition of the client configuration
type ClientConfig struct {
	Organization string
	Logging      LoggingType
	CryptoConfig CCType
	TLS          TLSType
	TLSCerts     MutualTLSConfig

	// currently not used by GO-SDK
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
	MspID string
}

// OrganizationConfig provides the definition of an organization in the network
type OrganizationConfig struct {
	MspID                  string
	CryptoPath             string
	Users                  map[string]TLSKeyPair
	Peers                  []string
	CertificateAuthorities []string
	AdminPrivateKey        TLSConfig
	SignedCert             TLSConfig
}

// OrdererConfig defines an orderer configuration
type OrdererConfig struct {
	URL         string
	GRPCOptions map[string]interface{}
	TLSCACerts  TLSConfig
}

// PeerConfig defines a peer configuration
type PeerConfig struct {
	URL         string
	EventURL    string
	GRPCOptions map[string]interface{}
	TLSCACerts  TLSConfig
}

// CAConfig defines a CA configuration
type CAConfig struct {
	URL         string
	HTTPOptions map[string]interface{}
	TLSCACerts  MutualTLSConfig
	Registrar   struct {
		EnrollID     string
		EnrollSecret string
	}
	CAName string
}

// TLSConfig TLS configurations
type TLSConfig struct {
	// the following two fields are interchangeable.
	// If Path is available, then it will be used to load the cert
	// if Pem is available, then it has the raw data of the cert it will be used as-is
	// Certificate root certificate path
	Path string
	// Certificate actual content
	Pem string
}

// MutualTLSConfig Mutual TLS configurations
type MutualTLSConfig struct {
	Pem []string
	// Certfiles root certificates for TLS validation (Comma separated path list)
	Path string
	// Client client TLS information
	Client struct {
		KeyPem string
		// Keyfile client key path
		Keyfile string
		CertPem string
		// Certfile client cert path
		Certfile string
	}
}

// TLSKeyPair contains the private key and certificate for TLS encryption
type TLSKeyPair struct {
	Key  TLSConfig
	Cert TLSConfig
}
