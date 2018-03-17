/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package core

import (
	"crypto/tls"
	"crypto/x509"
	"time"
)

// TODO - the Config (and related) interfaces need to be refactored.
// E.g., The interface should be minimized and split across providers.

// Config fabric-sdk-go configuration interface
type Config interface {
	Client() (*ClientConfig, error)
	CAConfig(org string) (*CAConfig, error)
	CAServerCertPems(org string) ([]string, error)
	CAServerCertPaths(org string) ([]string, error)
	CAClientKeyPem(org string) (string, error)
	CAClientKeyPath(org string) (string, error)
	CAClientCertPem(org string) (string, error)
	CAClientCertPath(org string) (string, error)
	TimeoutOrDefault(TimeoutType) time.Duration
	Timeout(TimeoutType) time.Duration
	MSPID(org string) (string, error)
	PeerMSPID(name string) (string, error)
	OrderersConfig() ([]OrdererConfig, error)
	RandomOrdererConfig() (*OrdererConfig, error)
	OrdererConfig(name string) (*OrdererConfig, error)
	PeersConfig(org string) ([]PeerConfig, error)
	PeerConfig(org string, name string) (*PeerConfig, error)
	PeerConfigByURL(url string) (*PeerConfig, error)
	NetworkConfig() (*NetworkConfig, error)
	NetworkPeers() ([]NetworkPeer, error)
	ChannelConfig(name string) (*ChannelConfig, error)
	ChannelPeers(name string) ([]ChannelPeer, error)
	ChannelOrderers(name string) ([]OrdererConfig, error)
	TLSCACertPool(certConfig ...*x509.Certificate) (*x509.CertPool, error)
	IsSecurityEnabled() bool
	SecurityAlgorithm() string
	SecurityLevel() int
	SecurityProvider() string
	Ephemeral() bool
	SoftVerify() bool
	SecurityProviderLibPath() string
	SecurityProviderPin() string
	SecurityProviderLabel() string
	KeyStorePath() string
	CAKeyStorePath() string
	CryptoConfigPath() string
	TLSClientCerts() ([]tls.Certificate, error)
	CredentialStorePath() string
	EventServiceType() EventServiceType
}

// ConfigProvider enables creation of a Config instance
type ConfigProvider func() (Config, error)

// TimeoutType enumerates the different types of outgoing connections
type TimeoutType int

const (
	// EndorserConnection connection timeout
	EndorserConnection TimeoutType = iota
	// EventHubConnection connection timeout
	EventHubConnection
	// EventReg connection timeout
	EventReg
	// Query timeout
	Query
	// Execute timeout
	Execute
	// OrdererConnection orderer connection timeout
	OrdererConnection
	// OrdererResponse orderer response timeout
	OrdererResponse
	// DiscoveryGreylistExpiry discovery Greylist expiration period
	DiscoveryGreylistExpiry
	// ConnectionIdle is the timeout for closing idle connections
	ConnectionIdle
	// CacheSweepInterval is the duration between cache sweeps
	CacheSweepInterval
	// EventServiceIdle is the timeout for closing the event service connection
	EventServiceIdle
	// PeerResponse peer response timeout
	PeerResponse
	// ResMgmt timeout is default overall timeout for all resource management operations
	ResMgmt
	// ChannelConfigRefresh channel configuration refresh interval
	ChannelConfigRefresh
	// ChannelMembershipRefresh channel membership refresh interval
	ChannelMembershipRefresh
)

// EventServiceType specifies the type of event service to use
type EventServiceType int

const (
	// DeliverEventServiceType uses the Deliver Service for block and filtered-block events
	DeliverEventServiceType EventServiceType = iota
	// EventHubEventServiceType uses the Event Hub for block events
	EventHubEventServiceType
)

// Providers represents the SDK configured core providers context.
type Providers interface {
	CryptoSuite() CryptoSuite
	Config() Config
	SigningManager() SigningManager
}
