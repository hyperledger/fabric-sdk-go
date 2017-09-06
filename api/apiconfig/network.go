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

type ClientConfig struct {
	Organization string
	Logging      LoggingType
	CryptoConfig CCType
	Tls          TlsType
	// currently not used by GO-SDK
	CredentialStore CredentialStoreType
}

type LoggingType struct {
	Level string
}

type CCType struct {
	Path string
}
type TlsType struct {
	Enabled bool
}

type CredentialStoreType struct {
	Path        string
	CryptoStore struct {
		Path string
	}
	Wallet string
}

type ChannelConfig struct {
	// Orderers list of ordering service nodes
	Orderers []string
	// Peers a list of peer-channels that are part of this organization
	// to get the real Peer config object, use the Name field and fetch NetworkConfig.Peers[Name]
	Peers map[string]PeerChannelConfig
	// Chaincodes list of services
	Chaincodes []string
}

type PeerChannelConfig struct {
	EndorsingPeer  bool
	ChaincodeQuery bool
	LedgerQuery    bool
	EventSource    bool
}

type OrganizationConfig struct {
	MspID                  string
	Peers                  []string
	CertificateAuthorities []string
	AdminPrivateKey        TLSConfig
	SignedCert             TLSConfig
}

type OrdererConfig struct {
	URL         string
	GrpcOptions map[string]interface{}
	TlsCACerts  TLSConfig
}

type PeerConfig struct {
	Url         string
	EventUrl    string
	GrpcOptions map[string]interface{}
	TlsCACerts  TLSConfig
}

type CAConfig struct {
	Url         string
	HttpOptions map[string]interface{}
	TlsCACerts  MutualTLSConfig
	Registrar   struct {
		EnrollId     string
		EnrollSecret string
	}
	CaName string
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
	// Certfiles root certificates for TLS validation (Comma serparated path list)
	Path string
	// Client client TLS information
	Client struct {
		// Keyfile client key path
		Keyfile string
		// Certfile client cert path
		Certfile string
	}
}
