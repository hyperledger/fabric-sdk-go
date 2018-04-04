/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package core

//CryptoSuiteConfig contains sdk configuration items for cryptosuite.
type CryptoSuiteConfig interface {
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
}

// Providers represents the SDK configured core providers context.
type Providers interface {
	CryptoSuite() CryptoSuite
	SigningManager() SigningManager
}

//ConfigProvider provides config backend for SDK
type ConfigProvider func() (ConfigBackend, error)

//LookupOpts contains options for looking up key in config backend
type LookupOpts struct {
	UnmarshalType interface{}
}

//LookupOption option to lookup key in config backend
type LookupOption func(opts *LookupOpts)

//ConfigBackend backend for all config types in SDK
type ConfigBackend interface {
	//TODO lookupOption should be removed, unmarshal option should be handled externally
	Lookup(key string, opts ...LookupOption) (interface{}, bool)
}

//WithUnmarshalType lookup option which can be used to unmarshal lookup value to provided type
func WithUnmarshalType(unmarshalType interface{}) LookupOption {
	return func(opts *LookupOpts) {
		opts.UnmarshalType = unmarshalType
	}
}
