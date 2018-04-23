/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

// EndpointConfigOptions represents EndpointConfig interface with overridable interface functions
// if a function is not overridden, the default EndpointConfig implementation will be used.
type EndpointConfigOptions struct {
	timeout
	mspID
	peerMSPID
	orderersConfig
	ordererConfig
	peersConfig
	peerConfig
	networkConfig
	networkPeers
	channelConfig
	channelPeers
	channelOrderers
	tlsCACertPool
	eventServiceType
	tlsClientCerts
	cryptoConfigPath
}

type applier func()
type predicate func() bool
type setter struct{ isSet bool }

// timeout interface allows to uniquely override EndpointConfig interface's Timeout() function
type timeout interface {
	Timeout(fab.TimeoutType) time.Duration
}

// mspID interface allows to uniquely override EndpointConfig interface's MSPID() function
type mspID interface {
	MSPID(org string) (string, error)
}

// peerMSPID interface allows to uniquely override EndpointConfig interface's PeerMSPID() function
type peerMSPID interface {
	PeerMSPID(name string) (string, error)
}

// orderersConfig interface allows to uniquely override EndpointConfig interface's OrderersConfig() function
type orderersConfig interface {
	OrderersConfig() ([]fab.OrdererConfig, error)
}

// ordererConfig interface allows to uniquely override EndpointConfig interface's OrdererConfig() function
type ordererConfig interface {
	OrdererConfig(name string) (*fab.OrdererConfig, error)
}

// peersConfig interface allows to uniquely override EndpointConfig interface's PeersConfig() function
type peersConfig interface {
	PeersConfig(org string) ([]fab.PeerConfig, error)
}

// peerConfig interface allows to uniquely override EndpointConfig interface's PeerConfig() function
type peerConfig interface {
	PeerConfig(nameOrURL string) (*fab.PeerConfig, error)
}

// networkConfig interface allows to uniquely override EndpointConfig interface's NetworkConfig() function
type networkConfig interface {
	NetworkConfig() (*fab.NetworkConfig, error)
}

// networkPeers interface allows to uniquely override EndpointConfig interface's NetworkPeers() function
type networkPeers interface {
	NetworkPeers() ([]fab.NetworkPeer, error)
}

// channelConfig interface allows to uniquely override EndpointConfig interface's ChannelConfig() function
type channelConfig interface {
	ChannelConfig(name string) (*fab.ChannelNetworkConfig, error)
}

// channelPeers interface allows to uniquely override EndpointConfig interface's ChannelPeers() function
type channelPeers interface {
	ChannelPeers(name string) ([]fab.ChannelPeer, error)
}

// channelOrderers interface allows to uniquely override EndpointConfig interface's ChannelOrderers() function
type channelOrderers interface {
	ChannelOrderers(name string) ([]fab.OrdererConfig, error)
}

// tlsCACertPool interface allows to uniquely override EndpointConfig interface's TLSCACertPool() function
type tlsCACertPool interface {
	TLSCACertPool(certConfig ...*x509.Certificate) (*x509.CertPool, error)
}

// eventServiceType interface allows to uniquely override EndpointConfig interface's EventServiceType() function
type eventServiceType interface {
	EventServiceType() fab.EventServiceType
}

// tlsClientCerts interface allows to uniquely override EndpointConfig interface's TLSClientCerts() function
type tlsClientCerts interface {
	TLSClientCerts() ([]tls.Certificate, error)
}

// cryptoConfigPath interface allows to uniquely override EndpointConfig interface's CryptoConfigPath() function
type cryptoConfigPath interface {
	CryptoConfigPath() string
}

// BuildConfigEndpointFromOptions will return an EndpointConfig instance pre-built with Optional interfaces
// provided in fabsdk's WithEndpointConfig(opts...) call
func BuildConfigEndpointFromOptions(opts ...interface{}) (fab.EndpointConfig, error) {
	// build a new EndpointConfig with overridden function implementations
	c := &EndpointConfigOptions{}
	for i, option := range opts {
		logger.Debugf("option %d: %#v", i, option)
		err := setEndpointConfigWithOptionInterface(c, option)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// UpdateMissingOptsWithDefaultConfig will verify if any functions of the EndpointConfig were not updated with fabsdk's
// WithConfigEndpoint(opts...) call, then use default EndpointConfig interface for these functions instead
func UpdateMissingOptsWithDefaultConfig(c *EndpointConfigOptions, d fab.EndpointConfig) fab.EndpointConfig {
	s := &setter{}

	s.set(c.timeout, nil, func() { c.timeout = d })
	s.set(c.mspID, nil, func() { c.mspID = d })
	s.set(c.peerMSPID, nil, func() { c.peerMSPID = d })
	s.set(c.orderersConfig, nil, func() { c.orderersConfig = d })
	s.set(c.ordererConfig, nil, func() { c.ordererConfig = d })
	s.set(c.peersConfig, nil, func() { c.peersConfig = d })
	s.set(c.peerConfig, nil, func() { c.peerConfig = d })
	s.set(c.networkConfig, nil, func() { c.networkConfig = d })
	s.set(c.networkPeers, nil, func() { c.networkPeers = d })
	s.set(c.channelConfig, nil, func() { c.channelConfig = d })
	s.set(c.channelPeers, nil, func() { c.channelPeers = d })
	s.set(c.channelOrderers, nil, func() { c.channelOrderers = d })
	s.set(c.tlsCACertPool, nil, func() { c.tlsCACertPool = d })
	s.set(c.eventServiceType, nil, func() { c.eventServiceType = d })
	s.set(c.tlsClientCerts, nil, func() { c.tlsClientCerts = d })
	s.set(c.cryptoConfigPath, nil, func() { c.cryptoConfigPath = d })

	return c
}

// IsEndpointConfigFullyOverridden will return true if all of the argument's sub interfaces is not nil
// (ie EndpointConfig interface not fully overridden)
func IsEndpointConfigFullyOverridden(c *EndpointConfigOptions) bool {
	return !anyNil(c.timeout, c.mspID, c.peerMSPID, c.orderersConfig, c.ordererConfig, c.peersConfig, c.peerConfig, c.networkConfig,
		c.networkPeers, c.channelConfig, c.channelPeers, c.channelOrderers, c.tlsCACertPool, c.eventServiceType, c.tlsClientCerts, c.cryptoConfigPath)
}

// will override EndpointConfig interface with functions provided by o (option)
func setEndpointConfigWithOptionInterface(c *EndpointConfigOptions, o interface{}) error {
	s := &setter{}

	s.set(c.timeout, func() bool { _, ok := o.(timeout); return ok }, func() { c.timeout = o.(timeout) })
	s.set(c.mspID, func() bool { _, ok := o.(mspID); return ok }, func() { c.mspID = o.(mspID) })
	s.set(c.peerMSPID, func() bool { _, ok := o.(peerMSPID); return ok }, func() { c.peerMSPID = o.(peerMSPID) })
	s.set(c.orderersConfig, func() bool { _, ok := o.(orderersConfig); return ok }, func() { c.orderersConfig = o.(orderersConfig) })
	s.set(c.ordererConfig, func() bool { _, ok := o.(ordererConfig); return ok }, func() { c.ordererConfig = o.(ordererConfig) })
	s.set(c.peersConfig, func() bool { _, ok := o.(peersConfig); return ok }, func() { c.peersConfig = o.(peersConfig) })
	s.set(c.peerConfig, func() bool { _, ok := o.(peerConfig); return ok }, func() { c.peerConfig = o.(peerConfig) })
	s.set(c.networkConfig, func() bool { _, ok := o.(networkConfig); return ok }, func() { c.networkConfig = o.(networkConfig) })
	s.set(c.networkPeers, func() bool { _, ok := o.(networkPeers); return ok }, func() { c.networkPeers = o.(networkPeers) })
	s.set(c.channelConfig, func() bool { _, ok := o.(channelConfig); return ok }, func() { c.channelConfig = o.(channelConfig) })
	s.set(c.channelPeers, func() bool { _, ok := o.(channelPeers); return ok }, func() { c.channelPeers = o.(channelPeers) })
	s.set(c.channelOrderers, func() bool { _, ok := o.(channelOrderers); return ok }, func() { c.channelOrderers = o.(channelOrderers) })
	s.set(c.tlsCACertPool, func() bool { _, ok := o.(tlsCACertPool); return ok }, func() { c.tlsCACertPool = o.(tlsCACertPool) })
	s.set(c.eventServiceType, func() bool { _, ok := o.(eventServiceType); return ok }, func() { c.eventServiceType = o.(eventServiceType) })
	s.set(c.tlsClientCerts, func() bool { _, ok := o.(tlsClientCerts); return ok }, func() { c.tlsClientCerts = o.(tlsClientCerts) })
	s.set(c.cryptoConfigPath, func() bool { _, ok := o.(cryptoConfigPath); return ok }, func() { c.cryptoConfigPath = o.(cryptoConfigPath) })

	if !s.isSet {
		return errors.Errorf("option %#v is not a sub interface of EndpointConfig, at least one of its functions must be implemented.", o)
	}

	return nil
}

// needed to avoid meta-linter errors (too many if conditions)
func (o *setter) set(current interface{}, check predicate, apply applier) {
	if current == nil && (check == nil || check()) {
		apply()
		o.isSet = true
	}
}

// will verify if any of objs element is nil, also needed to avoid meta-linter errors
func anyNil(objs ...interface{}) bool {
	for _, p := range objs {
		if p == nil {
			return true
		}
	}
	return false
}
