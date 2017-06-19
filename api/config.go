/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"crypto/x509"

	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	"github.com/spf13/viper"
)

// Config ...
type Config interface {
	GetServerURL() string
	GetServerCertFiles() []string
	GetFabricCAClientKeyFile() string
	GetFabricCAClientCertFile() string
	GetFabricCATLSEnabledFlag() bool
	GetFabricClientViper() *viper.Viper
	GetPeersConfig() ([]PeerConfig, error)
	IsTLSEnabled() bool
	GetTLSCACertPool(tlsCertificate string) (*x509.CertPool, error)
	GetTLSCACertPoolFromRoots(ordererRootCAs [][]byte) (*x509.CertPool, error)
	IsSecurityEnabled() bool
	TcertBatchSize() int
	GetSecurityAlgorithm() string
	GetSecurityLevel() int
	GetOrdererHost() string
	GetOrdererPort() string
	GetOrdererTLSServerHostOverride() string
	GetOrdererTLSCertificate() string
	GetFabricCAID() string
	GetFabricCAName() string
	GetKeyStorePath() string
	GetFabricCAHomeDir() string
	GetFabricCAMspDir() string
	GetCryptoConfigPath() string
	GetCSPConfig() *bccspFactory.FactoryOpts
}

// PeerConfig A set of configurations required to connect to a Fabric peer
type PeerConfig struct {
	Host      string
	Port      int
	EventHost string
	EventPort int
	Primary   bool
	TLS       struct {
		Certificate        string
		ServerHostOverride string
	}
}
