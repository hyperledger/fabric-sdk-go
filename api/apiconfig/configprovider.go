/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apiconfig

import (
	"crypto/x509"
	"time"

	bccspFactory "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp/factory"
)

// Config fabric-sdk-go configuration interface
type Config interface {
	Client() (*ClientConfig, error)
	CAConfig(org string) (*CAConfig, error)
	CAServerCertFiles(org string) ([]string, error)
	CAClientKeyFile(org string) (string, error)
	CAClientCertFile(org string) (string, error)
	TimeoutOrDefault(TimeoutType) time.Duration
	MspID(org string) (string, error)
	OrderersConfig() ([]OrdererConfig, error)
	RandomOrdererConfig() (*OrdererConfig, error)
	OrdererConfig(name string) (*OrdererConfig, error)
	PeersConfig(org string) ([]PeerConfig, error)
	PeerConfig(org string, name string) (*PeerConfig, error)
	NetworkConfig() (*NetworkConfig, error)
	ChannelConfig(name string) (*ChannelConfig, error)
	ChannelPeers(name string) ([]ChannelPeer, error)
	SetTLSCACertPool(*x509.CertPool)
	TLSCACertPool(tlsCertificate string) (*x509.CertPool, error)
	IsSecurityEnabled() bool
	SecurityAlgorithm() string
	SecurityLevel() int
	SecurityProvider() string
	Ephemeral() bool
	SoftVerify() bool
	SecurityProviderPin() string
	SecurityProviderLabel() string
	KeyStorePath() string
	CAKeyStorePath() string
	CryptoConfigPath() string
	CSPConfig() *bccspFactory.FactoryOpts
}

// TimeoutType enumerates the different types of outgoing connections
type TimeoutType int

const (
	// Endorser connection timeout
	Endorser TimeoutType = iota
	// EventHub connection timeout
	EventHub
	// EventReg connection timeout
	EventReg
	// Query timeout
	Query
	// ExecuteTx timeout
	ExecuteTx
	// OrdererConnection orderer connection timeout
	OrdererConnection
	// OrdererResponse orderer response timeout
	OrdererResponse
)
