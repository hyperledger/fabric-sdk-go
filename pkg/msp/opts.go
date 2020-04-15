/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	commtls "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm/tls"
	"github.com/pkg/errors"
)

// IdentityConfigOptions represents IdentityConfig interface with overridable interface functions
// if a function is not overridden, the default IdentityConfig implementation will be used.
type IdentityConfigOptions struct {
	client
	caConfig
	caServerCerts
	caClientKey
	caClientCert
	caKeyStorePath
	credentialStorePath
	tlsCACertPool
}

type applier func()
type predicate func() bool
type setter struct{ isSet bool }

// client interface allows to uniquely override IdentityConfig interface's Client() function
type client interface {
	Client() *msp.ClientConfig
}

// caConfig interface allows to uniquely override IdentityConfig interface's CAConfig() function
type caConfig interface {
	CAConfig(caID string) (*msp.CAConfig, bool)
}

// caServerCerts interface allows to uniquely override IdentityConfig interface's CAServerCerts() function
type caServerCerts interface {
	CAServerCerts(caID string) ([][]byte, bool)
}

// caClientKey interface allows to uniquely override IdentityConfig interface's CAClientKey() function
type caClientKey interface {
	CAClientKey(caID string) ([]byte, bool)
}

// caClientCert interface allows to uniquely override IdentityConfig interface's CAClientCert() function
type caClientCert interface {
	CAClientCert(caID string) ([]byte, bool)
}

// caKeyStorePath interface allows to uniquely override IdentityConfig interface's CAKeyStorePath() function
type caKeyStorePath interface {
	CAKeyStorePath() string
}

// credentialStorePath interface allows to uniquely override IdentityConfig interface's CredentialStorePath() function
type credentialStorePath interface {
	CredentialStorePath() string
}

// tlsCACertPool interface allows to uniquely override IdentityConfig interface's TLSCACertPool() function
type tlsCACertPool interface {
	TLSCACertPool() commtls.CertPool
}

// BuildIdentityConfigFromOptions will return an IdentityConfig instance pre-built with Optional interfaces
// provided in fabsdk's WithConfigIdentity(opts...) call
func BuildIdentityConfigFromOptions(opts ...interface{}) (msp.IdentityConfig, error) {
	// build a new IdentityConfig with overridden function implementations
	c := &IdentityConfigOptions{}
	for _, option := range opts {
		err := setIdentityConfigWithOptionInterface(c, option)
		if err != nil {
			return nil, err
		}
	}

	return c, nil

}

// UpdateMissingOptsWithDefaultConfig will verify if any functions of the IdentityConfig were not updated with fabsdk's
// WithConfigIdentity(opts...) call, then use default IdentityConfig interface for these functions instead
func UpdateMissingOptsWithDefaultConfig(c *IdentityConfigOptions, d msp.IdentityConfig) msp.IdentityConfig {
	s := &setter{}

	s.set(c.client, nil, func() { c.client = d })
	s.set(c.caConfig, nil, func() { c.caConfig = d })
	s.set(c.caServerCerts, nil, func() { c.caServerCerts = d })
	s.set(c.caClientKey, nil, func() { c.caClientKey = d })
	s.set(c.caClientCert, nil, func() { c.caClientCert = d })
	s.set(c.caKeyStorePath, nil, func() { c.caKeyStorePath = d })
	s.set(c.credentialStorePath, nil, func() { c.credentialStorePath = d })
	s.set(c.tlsCACertPool, nil, func() { c.tlsCACertPool = d })

	return c
}

// IsIdentityConfigFullyOverridden will return true if all of the argument's sub interfaces is not nil
// (ie IdentityConfig interface not fully overridden)
func IsIdentityConfigFullyOverridden(c *IdentityConfigOptions) bool {
	return !anyNil(c.client, c.caConfig, c.caServerCerts, c.caClientKey, c.caClientCert, c.caKeyStorePath, c.credentialStorePath)
}

// will override IdentityConfig interface with functions provided by o (option)
func setIdentityConfigWithOptionInterface(c *IdentityConfigOptions, o interface{}) error {
	s := &setter{}

	s.set(c.client, func() bool { _, ok := o.(client); return ok }, func() { c.client = o.(client) })
	s.set(c.caConfig, func() bool { _, ok := o.(caConfig); return ok }, func() { c.caConfig = o.(caConfig) })
	s.set(c.caServerCerts, func() bool { _, ok := o.(caServerCerts); return ok }, func() { c.caServerCerts = o.(caServerCerts) })
	s.set(c.caClientKey, func() bool { _, ok := o.(caClientKey); return ok }, func() { c.caClientKey = o.(caClientKey) })
	s.set(c.caClientCert, func() bool { _, ok := o.(caClientCert); return ok }, func() { c.caClientCert = o.(caClientCert) })
	s.set(c.caKeyStorePath, func() bool { _, ok := o.(caKeyStorePath); return ok }, func() { c.caKeyStorePath = o.(caKeyStorePath) })
	s.set(c.credentialStorePath, func() bool { _, ok := o.(credentialStorePath); return ok }, func() { c.credentialStorePath = o.(credentialStorePath) })
	s.set(c.tlsCACertPool, func() bool { _, ok := o.(tlsCACertPool); return ok }, func() { c.tlsCACertPool = o.(tlsCACertPool) })

	if !s.isSet {
		return errors.Errorf("option %#v is not a sub interface of IdentityConfig, at least one of its functions must be implemented.", o)
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
