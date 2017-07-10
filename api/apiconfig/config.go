/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apiconfig

import (
	"crypto/x509"

	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
)

// Config fabric-sdk-go configuration interface
type Config interface {
	CAConfig(org string) (*CAConfig, error)
	CAServerCertFiles(org string) ([]string, error)
	CAClientKeyFile(org string) (string, error)
	CAClientCertFile(org string) (string, error)
	MspID(org string) (string, error)
	OrderersConfig() ([]OrdererConfig, error)
	RandomOrdererConfig() (*OrdererConfig, error)
	OrdererConfig(name string) (*OrdererConfig, error)
	PeersConfig(org string) ([]PeerConfig, error)
	NetworkConfig() (*NetworkConfig, error)
	IsTLSEnabled() bool
	SetTLSCACertPool(*x509.CertPool)
	TLSCACertPool(tlsCertificate string) (*x509.CertPool, error)
	IsSecurityEnabled() bool
	TcertBatchSize() int
	SecurityAlgorithm() string
	SecurityLevel() int
	KeyStorePath() string
	CAKeyStorePath() string
	CryptoConfigPath() string
	CSPConfig() *bccspFactory.FactoryOpts
}
