/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	commtls "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm/tls"
	logApi "github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
)

// Context is the context required by MSP services
type Context interface {
	core.Providers
	Providers
}

// IdentityManagerProvider provides identity management services
type IdentityManagerProvider interface {
	IdentityManager(orgName string) (IdentityManager, bool)
}

//IdentityConfig contains identity configurations
type IdentityConfig interface {
	Client() *ClientConfig
	CAConfig(caID string) (*CAConfig, bool)
	CAServerCerts(caID string) ([][]byte, bool)
	CAClientKey(caID string) ([]byte, bool)
	CAClientCert(caID string) ([]byte, bool)
	TLSCACertPool() commtls.CertPool
	CAKeyStorePath() string
	CredentialStorePath() string
}

// ClientConfig provides the definition of the client configuration
type ClientConfig struct {
	Organization    string
	Logging         logApi.LoggingType
	CryptoConfig    CCType
	TLSKey          []byte
	TLSCert         []byte
	CredentialStore CredentialStoreType
}

// CCType defines the path to crypto keys and certs
type CCType struct {
	Path string
}

// CredentialStoreType defines pluggable KV store properties
type CredentialStoreType struct {
	Path        string
	CryptoStore struct {
		Path string
	}
}

// EnrollCredentials holds credentials used for enrollment
type EnrollCredentials struct {
	EnrollID     string
	EnrollSecret string
}

// CAConfig defines a CA configuration
type CAConfig struct {
	ID               string
	URL              string
	GRPCOptions      map[string]interface{}
	Registrar        EnrollCredentials
	CAName           string
	TLSCAServerCerts [][]byte
	TLSCAClientCert  []byte
	TLSCAClientKey   []byte
}

// Providers represents a provider of MSP service.
type Providers interface {
	UserStore() UserStore
	IdentityManagerProvider
	IdentityConfig() IdentityConfig
}
