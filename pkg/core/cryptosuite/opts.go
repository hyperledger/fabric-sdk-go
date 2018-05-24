/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/pkg/errors"
)

// CryptoConfigOptions represents CryptoConfig interface with overridable interface functions
// if a function is not overridden, the default CryptoConfig implementation will be used.
type CryptoConfigOptions struct {
	isSecurityEnabled
	securityAlgorithm
	securityLevel
	securityProvider
	softVerify
	securityProviderLibPath
	securityProviderPin
	securityProviderLabel
	keyStorePath
}

type applier func()
type predicate func() bool
type setter struct{ isSet bool }

// isSecurityEnabled interface allows to uniquely override CryptoConfig interface's IsSecurityEnabled() function
type isSecurityEnabled interface {
	IsSecurityEnabled() bool
}

// securityAlgorithm interface allows to uniquely override CryptoConfig interface's SecurityAlgorithm() function
type securityAlgorithm interface {
	SecurityAlgorithm() string
}

// securityLevel interface allows to uniquely override CryptoConfig interface's SecurityLevel() function
type securityLevel interface {
	SecurityLevel() int
}

// securityProvider interface allows to uniquely override CryptoConfig interface's SecurityProvider() function
type securityProvider interface {
	SecurityProvider() string
}

// softVerify interface allows to uniquely override CryptoConfig interface's SoftVerify() function
type softVerify interface {
	SoftVerify() bool
}

// securityProviderLibPath interface allows to uniquely override CryptoConfig interface's SecurityProviderLibPath() function
type securityProviderLibPath interface {
	SecurityProviderLibPath() string
}

// securityProviderPin interface allows to uniquely override CryptoConfig interface's SecurityProviderPin() function
type securityProviderPin interface {
	SecurityProviderPin() string
}

// securityProviderLabel interface allows to uniquely override CryptoConfig interface's SecurityProviderLabel() function
type securityProviderLabel interface {
	SecurityProviderLabel() string
}

// keyStorePath interface allows to uniquely override CryptoConfig interface's KeyStorePath() function
type keyStorePath interface {
	KeyStorePath() string
}

// BuildCryptoSuiteConfigFromOptions will return an CryptoConfig instance pre-built with Optional interfaces
// provided in fabsdk's WithConfigCrypto(opts...) call
func BuildCryptoSuiteConfigFromOptions(opts ...interface{}) (core.CryptoSuiteConfig, error) {
	// build a new CryptoConfig with overridden function implementations
	c := &CryptoConfigOptions{}
	for _, option := range opts {
		err := setCryptoConfigWithOptionInterface(c, option)
		if err != nil {
			return nil, err
		}
	}

	return c, nil

}

// UpdateMissingOptsWithDefaultConfig will verify if any functions of the CryptoConfig were not updated with fabsdk's
// WithConfigCrypto(opts...) call, then use default CryptoConfig interface for these functions instead
func UpdateMissingOptsWithDefaultConfig(c *CryptoConfigOptions, d core.CryptoSuiteConfig) core.CryptoSuiteConfig {
	s := &setter{}

	s.set(c.isSecurityEnabled, nil, func() { c.isSecurityEnabled = d })
	s.set(c.securityAlgorithm, nil, func() { c.securityAlgorithm = d })
	s.set(c.securityLevel, nil, func() { c.securityLevel = d })
	s.set(c.securityProvider, nil, func() { c.securityProvider = d })
	s.set(c.softVerify, nil, func() { c.softVerify = d })
	s.set(c.securityProviderLibPath, nil, func() { c.securityProviderLibPath = d })
	s.set(c.securityProviderPin, nil, func() { c.securityProviderPin = d })
	s.set(c.securityProviderLabel, nil, func() { c.securityProviderLabel = d })
	s.set(c.keyStorePath, nil, func() { c.keyStorePath = d })

	return c
}

// IsCryptoConfigFullyOverridden will return true if all of the argument's sub interfaces is not nil
// (ie CryptoSuiteConfig interface not fully overridden)
func IsCryptoConfigFullyOverridden(c *CryptoConfigOptions) bool {
	return !anyNil(c.isSecurityEnabled, c.securityAlgorithm, c.securityLevel, c.securityProvider, c.softVerify, c.securityProviderLibPath, c.securityProviderPin, c.securityProviderLabel, c.keyStorePath)
}

// will override CryptoSuiteConfig interface with functions provided by o (option)
func setCryptoConfigWithOptionInterface(c *CryptoConfigOptions, o interface{}) error {
	s := &setter{}

	s.set(c.isSecurityEnabled, func() bool { _, ok := o.(isSecurityEnabled); return ok }, func() { c.isSecurityEnabled = o.(isSecurityEnabled) })
	s.set(c.securityAlgorithm, func() bool { _, ok := o.(securityAlgorithm); return ok }, func() { c.securityAlgorithm = o.(securityAlgorithm) })
	s.set(c.securityLevel, func() bool { _, ok := o.(securityLevel); return ok }, func() { c.securityLevel = o.(securityLevel) })
	s.set(c.securityProvider, func() bool { _, ok := o.(securityProvider); return ok }, func() { c.securityProvider = o.(securityProvider) })
	s.set(c.softVerify, func() bool { _, ok := o.(softVerify); return ok }, func() { c.softVerify = o.(softVerify) })
	s.set(c.securityProviderLibPath, func() bool { _, ok := o.(securityProviderLibPath); return ok }, func() { c.securityProviderLibPath = o.(securityProviderLibPath) })
	s.set(c.securityProviderPin, func() bool { _, ok := o.(securityProviderPin); return ok }, func() { c.securityProviderPin = o.(securityProviderPin) })
	s.set(c.securityProviderLabel, func() bool { _, ok := o.(securityProviderLabel); return ok }, func() { c.securityProviderLabel = o.(securityProviderLabel) })
	s.set(c.keyStorePath, func() bool { _, ok := o.(keyStorePath); return ok }, func() { c.keyStorePath = o.(keyStorePath) })

	if !s.isSet {
		return errors.Errorf("option %#v is not a sub interface of CryptoSuiteConfig, at least one of its functions must be implemented.", o)
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
