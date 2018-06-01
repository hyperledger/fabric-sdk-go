/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	m0 = &Config{}
	m1 = &mockIsSecurityEnabled{}
	m2 = &mockSecurityAlgorithm{}
	m3 = &mockSecurityLevel{}
	m4 = &mockSecurityProvider{}
	m5 = &mockSoftVerify{}
	m6 = &mockSecurityProviderLibPath{}
	m7 = &mockSecurityProviderPin{}
	m8 = &mockSecurityProviderLabel{}
	m9 = &mockKeyStorePath{}
)

func TestCreateCustomFullCryptotConfig(t *testing.T) {
	var opts []interface{}
	opts = append(opts, m0)
	// try to build with the overall interface (m0 is the overall interface's default implementation)
	cryptoConfigOption, err := BuildCryptoSuiteConfigFromOptions(opts...)
	if err != nil {
		t.Fatalf("BuildCryptoSuiteConfigFromOptions returned unexpected error %s", err)
	}
	if cryptoConfigOption == nil {
		t.Fatal("BuildCryptoSuiteConfigFromOptions call returned nil")
	}
}

func TestCreateCustomCryptoConfig(t *testing.T) {
	// try to build with partial implementations
	cryptoConfigOption, err := BuildCryptoSuiteConfigFromOptions(m1, m2, m3, m4)
	if err != nil {
		t.Fatalf("BuildCryptoSuiteConfigFromOptions returned unexpected error %s", err)
	}
	var cco *CryptoConfigOptions
	var ok bool
	if cco, ok = cryptoConfigOption.(*CryptoConfigOptions); !ok {
		t.Fatalf("BuildCryptoSuiteConfigFromOptions did not return an Options instance %T", cryptoConfigOption)
	}
	require.NotNil(t, cco, "build ConfigCryptoOption returned is nil")

	// test m1 implementation
	i := cco.IsSecurityEnabled()
	require.True(t, i, "IsSecurityEnabled returned must return true")

	// test m2 implementation
	a := cco.SecurityAlgorithm()
	require.Equal(t, "SHA2", a, "SecurityAlgorithm did not return expected interface value")

	// test m3 implementation
	s := cco.SecurityLevel()
	require.Equal(t, 256, s, "SecurityLevel did not return expected interface value")

	// test m4 implementation
	a = cco.SecurityProvider()
	require.Equal(t, "SW", a, "SecurityProvider did not return expected interface value")

	// verify if an interface was not passed as an option but was not nil, it should be nil (ie these implementations should not be populated in cco: m5, m6, m7, m8 and m9)
	require.Nil(t, cco.softVerify, "softVerify created with nil interface but got non nil one: %s. Expected nil interface", cco.softVerify)
	require.Nil(t, cco.securityProviderLibPath, "securityProviderLibPath created with nil interface but got non nil one: %s. Expected nil interface", cco.securityProviderLibPath)
	require.Nil(t, cco.securityProviderPin, "securityProviderPin created with nil interface but got non nil one: %s. Expected nil interface", cco.securityProviderPin)
	require.Nil(t, cco.securityProviderLabel, "securityProviderLabel created with nil interface but got non nil one: %s. Expected nil interface", cco.securityProviderLabel)
	require.Nil(t, cco.keyStorePath, "keyStorePath created with nil interface but got non nil one: %s. Expected nil interface", cco.keyStorePath)
}

func TestCreateCustomCryptoConfigRemainingFunctions(t *testing.T) {
	// try to build with the remaining implementations not tested above
	cryptoConfigOption, err := BuildCryptoSuiteConfigFromOptions(m5, m6, m7, m8, m9)
	if err != nil {
		t.Fatalf("BuildCryptoSuiteConfigFromOptions returned unexpected error %s", err)
	}
	var cco *CryptoConfigOptions
	var ok bool
	if cco, ok = cryptoConfigOption.(*CryptoConfigOptions); !ok {
		t.Fatalf("BuildCryptoSuiteConfigFromOptions did not return an Options instance %T", cryptoConfigOption)
	}
	require.NotNil(t, cco, "build ConfigCryptoOption returned is nil")

	// test m5 implementation
	b := cco.SoftVerify()
	require.True(t, true, b, "SoftVerify did not return expected interface value")

	// test m6 implementation
	s := cco.SecurityProviderLibPath()
	require.Equal(t, "test/sec/provider/lib/path", s, "SecurityProviderLibPath did not return expected interface value")

	// test m7 implementation
	s = cco.SecurityProviderPin()
	require.Equal(t, "1234", s, "SecurityProviderPin did not return expected interface value")

	// test m8 implementation
	s = cco.SecurityProviderLabel()
	require.Equal(t, "xyz", s, "SecurityProviderLabel did not return expected interface value")

	// test m9 implementation
	s = cco.KeyStorePath()
	require.Equal(t, "test/keystore/path", s, "KeyStorePath did not return expected interface value")

	// verify if an interface was not passed as an option but was not nil, it should be nil (ie these implementations should not be populated in cco: m1, m2, m3 and m4)
	require.Nil(t, cco.isSecurityEnabled, "isSecurityEnabled created with nil interface but got non nil one: %s. Expected nil interface", cco.isSecurityEnabled)
	require.Nil(t, cco.securityAlgorithm, "securityAlgorithm created with nil interface but got non nil one: %s. Expected nil interface", cco.securityAlgorithm)
	require.Nil(t, cco.securityLevel, "securityLevel created with nil interface but got non nil one: %s. Expected nil interface", cco.securityLevel)
	require.Nil(t, cco.securityProvider, "securityProvider created with nil interface but got non nil one: %s. Expected nil interface", cco.securityProvider)

	// now try with a non related interface to test if an error returns
	var badType interface{}
	_, err = BuildCryptoSuiteConfigFromOptions(m4, m5, m7, badType)
	require.Error(t, err, "BuildCryptoSuiteConfigFromOptions did not return error with badType")

}

func TestCreateCustomCryptoConfigWithSomeDefaultFunctions(t *testing.T) {
	// try to build with partial interfaces
	cryptoConfigOption, err := BuildCryptoSuiteConfigFromOptions(m1, m2, m3, m4)
	if err != nil {
		t.Fatalf("BuildCryptoSuiteConfigFromOptions returned unexpected error %s", err)
	}
	var cco *CryptoConfigOptions
	var ok bool
	if cco, ok = cryptoConfigOption.(*CryptoConfigOptions); !ok {
		t.Fatalf("BuildCryptoSuiteConfigFromOptions did not return an Options instance %T", cryptoConfigOption)
	}
	require.NotNil(t, cco, "build ConfigCryptoOption returned is nil")

	// now check if implementations that were not injected when building the config (ref first line in this function) are nil at this point
	// ie, verify these implementations should be nil: m5, m6, m7, m8 and m9
	require.Nil(t, cco.softVerify, "caClientCert created with nil interface but got non nil one: %s. Expected nil interface", cco.softVerify)
	require.Nil(t, cco.securityProviderLibPath, "securityProviderLibPath created with nil interface but got non nil one: %s. Expected nil interface", cco.securityProviderLibPath)
	require.Nil(t, cco.securityProviderPin, "securityProviderPin created with nil interface but got non nil one: %s. Expected nil interface", cco.securityProviderPin)
	require.Nil(t, cco.securityProviderLabel, "securityProviderLabel created with nil interface but got non nil one: %s. Expected nil interface", cco.securityProviderLabel)
	require.Nil(t, cco.keyStorePath, "keyStorePath created with nil interface but got non nil one: %s. Expected nil interface", cco.keyStorePath)

	// do the same test using IsCryptoConfigFullyOverridden() call
	require.False(t, IsCryptoConfigFullyOverridden(cco), "IsCryptoConfigFullyOverridden is supposed to return false with an Options instance not implementing all the interface functions")

	// now inject default interfaces (using m0 as default full implementation for the sake of this test) for the ones that were not overridden by options above
	cryptoConfigOptionWithSomeDefaults := UpdateMissingOptsWithDefaultConfig(cco, m0)

	// test implementations m1-m4 are still working

	// test m1 implementation
	i := cryptoConfigOptionWithSomeDefaults.IsSecurityEnabled()
	require.True(t, i, "IsSecurityEnabled returned must return true")

	// test m2 implementation
	a := cryptoConfigOptionWithSomeDefaults.SecurityAlgorithm()
	require.Equal(t, "SHA2", a, "SecurityAlgorithm did not return expected interface value")

	// test m3 implementation
	s := cryptoConfigOptionWithSomeDefaults.SecurityLevel()
	require.Equal(t, 256, s, "SecurityLevel did not return expected interface value")

	// test m4 implementation
	a = cryptoConfigOptionWithSomeDefaults.SecurityProvider()
	require.Equal(t, "SW", a, "SecurityProvider did not return expected interface value")

	if cco, ok = cryptoConfigOptionWithSomeDefaults.(*CryptoConfigOptions); !ok {
		t.Fatal("UpdateMissingOptsWithDefaultConfig() call did not return an implementation of CryptoConfigOptions")
	}

	// now check if implementations that were not injected when building the config (ref first line in this function) are defaulted with m0 this time
	// ie, verify these implementations should now be populated in cco: m5, m6, m7, m8 and m9
	require.NotNil(t, cco.softVerify, "softVerify should be populated with default interface but got nil one: %s. Expected default interface", cco.softVerify)
	require.NotNil(t, cco.securityProviderLibPath, "securityProviderLibPath should be populated with default interface but got nil one: %s. Expected default interface", cco.securityProviderLibPath)
	require.NotNil(t, cco.securityProviderPin, "securityProviderPin should be populated with default interface but got nil one: %s. Expected default interface", cco.securityProviderPin)
	require.NotNil(t, cco.securityProviderLabel, "securityProviderLabel should be populated with default interface but got nil one: %s. Expected default interface", cco.securityProviderLabel)
	require.NotNil(t, cco.keyStorePath, "keyStorePath should be populated with default interface but got nil one: %s. Expected default interface", cco.keyStorePath)

	// do the same test using IsCryptoConfigFullyOverridden() call
	require.True(t, IsCryptoConfigFullyOverridden(cco), "IsCryptoConfigFullyOverridden is supposed to return true since all the interface functions should be implemented")
}

type mockIsSecurityEnabled struct{}

func (m *mockIsSecurityEnabled) IsSecurityEnabled() bool {
	return true
}

type mockSecurityAlgorithm struct{}

func (m *mockSecurityAlgorithm) SecurityAlgorithm() string {
	return "SHA2"
}

type mockSecurityLevel struct{}

func (m *mockSecurityLevel) SecurityLevel() int {
	return 256
}

type mockSecurityProvider struct{}

func (m *mockSecurityProvider) SecurityProvider() string {
	return "SW"
}

type mockSoftVerify struct{}

func (m *mockSoftVerify) SoftVerify() bool {
	return true
}

type mockSecurityProviderLibPath struct{}

func (m *mockSecurityProviderLibPath) SecurityProviderLibPath() string {
	return "test/sec/provider/lib/path"
}

type mockSecurityProviderPin struct{}

func (m *mockSecurityProviderPin) SecurityProviderPin() string {
	return "1234"
}

type mockSecurityProviderLabel struct{}

func (m *mockSecurityProviderLabel) SecurityProviderLabel() string {
	return "xyz"
}

type mockKeyStorePath struct{}

func (m *mockKeyStorePath) KeyStorePath() string {
	return "test/keystore/path"
}
