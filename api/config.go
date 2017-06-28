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

// Config fabric-sdk-go configuration interface
type Config interface {
	GetCAConfig(org string) (*CAConfig, error)
	GetCAServerCertFiles(org string) ([]string, error)
	GetCAClientKeyFile(org string) (string, error)
	GetCAClientCertFile(org string) (string, error)
	GetMspID(org string) (string, error)
	GetFabricClientViper() *viper.Viper
	GetRandomOrdererConfig() (*OrdererConfig, error)
	GetOrdererConfig(name string) (*OrdererConfig, error)
	GetPeersConfig(org string) ([]PeerConfig, error)
	GetNetworkConfig() (*NetworkConfig, error)
	IsTLSEnabled() bool
	GetTLSCACertPool(tlsCertificate string) (*x509.CertPool, error)
	GetTLSCACertPoolFromRoots(ordererRootCAs [][]byte) (*x509.CertPool, error)
	IsSecurityEnabled() bool
	TcertBatchSize() int
	GetSecurityAlgorithm() string
	GetSecurityLevel() int
	GetKeyStorePath() string
	GetCAKeyStorePath() string
	GetCryptoConfigPath() string
	GetCSPConfig() *bccspFactory.FactoryOpts
}
