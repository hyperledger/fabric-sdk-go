/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockfab"

	commtls "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm/tls"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	logApi "github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	"github.com/stretchr/testify/require"
)

var (
	m0 = &IdentityConfig{}
	m1 = &mockClient{}
	m2 = &mockCaConfig{}
	m3 = &mockCaServerCerts{}
	m4 = &mockCaClientKey{}
	m5 = &mockCaClientCert{}
	m6 = &mockCaKeyStorePath{}
	m7 = &mockCredentialStorePath{}
	m8 = &mockTLSCACertPool{}
)

func TestCreateCustomFullIdentitytConfig(t *testing.T) {
	var opts []interface{}
	opts = append(opts, m0)
	// try to build with the overall interface (m0 is the overall interface's default implementation)
	identityConfigOption, err := BuildIdentityConfigFromOptions(opts...)
	if err != nil {
		t.Fatalf("BuildIdentityConfigFromOptions returned unexpected error %s", err)
	}
	if identityConfigOption == nil {
		t.Fatal("BuildIdentityConfigFromOptions call returned nil")
	}
}

func TestCreateCustomIdentityConfig(t *testing.T) {
	// try to build with partial implementations
	identityConfigOption, err := BuildIdentityConfigFromOptions(m1, m2, m3, m4)
	if err != nil {
		t.Fatalf("BuildIdentityConfigFromOptions returned unexpected error %s", err)
	}
	var ico *IdentityConfigOptions
	var ok bool
	if ico, ok = identityConfigOption.(*IdentityConfigOptions); !ok {
		t.Fatalf("BuildIdentityConfigFromOptions did not return an Options instance %T", identityConfigOption)
	}
	require.NotNil(t, ico, "build ConfigIdentityOption returned is nil")

	// test m1 implementation
	clnt := ico.Client()
	require.NotEmpty(t, clnt, "client returned must not be empty")

	// test m2 implementation
	caCfg, ok := ico.CAConfig("testCA")
	require.True(t, ok, "CAConfig failed")
	require.Equal(t, "test.url.com", caCfg.URL, "CAConfig did not return expected interface value")

	// test m3 implementation
	s, ok := ico.CAServerCerts("testCA")
	require.True(t, ok, "CAServerCerts failed")
	require.Equal(t, []byte("testCAservercert1"), s[0], "CAServerCerts did not return the right cert")
	require.Equal(t, []byte("testCAservercert2"), s[1], "CAServerCerts did not return the right cert")

	// test m4 implementation
	c, ok := ico.CAClientKey("testCA")
	require.True(t, ok, "CAClientKey failed")
	require.Equal(t, []byte("testCAclientkey"), c, "CAClientKey did not return the right cert")

	// verify if an interface was not passed as an option but was not nil, it should be nil (ie these implementations should not be populated in ico: m5, m6 and m7)
	require.Nil(t, ico.caClientCert, "caClientCert created with nil interface but got non nil one: %s. Expected nil interface", ico.caClientCert)
	require.Nil(t, ico.caKeyStorePath, "caKeyStorePath created with nil interface but got non nil one: %s. Expected nil interface", ico.caKeyStorePath)
	require.Nil(t, ico.credentialStorePath, "credentialStorePath created with nil interface but got non nil one: %s. Expected nil interface", ico.credentialStorePath)
	require.Nil(t, ico.tlsCACertPool, "tlsCACertPool created with nil interface but got non nil one: %s. Expected nil interface", ico.tlsCACertPool)
}

func TestCreateCustomIdentityConfigRemainingFunctions(t *testing.T) {
	// try to build with the remaining implementations not tested above
	identityConfigOption, err := BuildIdentityConfigFromOptions(m5, m6, m7, m8)
	if err != nil {
		t.Fatalf("BuildIdentityConfigFromOptions returned unexpected error %s", err)
	}
	var ico *IdentityConfigOptions
	var ok bool
	if ico, ok = identityConfigOption.(*IdentityConfigOptions); !ok {
		t.Fatalf("BuildIdentityConfigFromOptions did not return an Options instance %T", identityConfigOption)
	}
	require.NotNil(t, ico, "build ConfigIdentityOption returned is nil")

	// test m5 implementation
	c, ok := ico.CAClientCert("")
	require.True(t, ok, "CAClientCert failed")
	require.Equal(t, []byte("testCAclientcert"), c, "CAClientCert did not return expected interface value")

	// test m6 implementation
	s := ico.CAKeyStorePath()
	require.Equal(t, "test/store/path", s, "CAKeyStorePath did not return expected interface value")

	// test m7 implementation
	s = ico.CredentialStorePath()
	require.Equal(t, "test/cred/store/path", s, "CredentialStorePath did not return expected interface value")

	// test m8 implementation
	p := ico.TLSCACertPool()
	require.Equal(t, mockTLSCACertPoolInstance, p, "TLSCACertPool did not return expected interface value")

	// verify if an interface was not passed as an option but was not nil, it should be nil (ie these implementations should not be populated in ico: m1, m2, m3 and m4)
	require.Nil(t, ico.client, "client created with nil interface but got non nil one: %s. Expected nil interface", ico.client)
	require.Nil(t, ico.caConfig, "caConfig created with nil interface but got non nil one: %s. Expected nil interface", ico.caConfig)
	require.Nil(t, ico.caServerCerts, "caServerCerts created with nil interface but got non nil one: %s. Expected nil interface", ico.caServerCerts)
	require.Nil(t, ico.caClientKey, "caClientKey created with nil interface but got non nil one: %s. Expected nil interface", ico.caClientKey)

	// now try with a non related interface to test if an error returns
	var badType interface{}
	_, err = BuildIdentityConfigFromOptions(m4, m5, badType)
	require.Error(t, err, "BuildIdentityConfigFromOptions did not return error with badType")

}

func TestCreateCustomIdentityConfigWithSomeDefaultFunctions(t *testing.T) {
	// try to build with partial interfaces
	identityConfigOption, err := BuildIdentityConfigFromOptions(m1, m2, m3, m4)
	if err != nil {
		t.Fatalf("BuildIdentityConfigFromOptions returned unexpected error %s", err)
	}
	var ico *IdentityConfigOptions
	var ok bool
	if ico, ok = identityConfigOption.(*IdentityConfigOptions); !ok {
		t.Fatalf("BuildIdentityConfigFromOptions did not return an Options instance %T", identityConfigOption)
	}
	require.NotNil(t, ico, "build ConfigIdentityOption returned is nil")

	// now check if implementations that were not injected when building the config (ref first line in this function) are nil at this point
	// ie, verify these implementations should be nil: m5, m6, m7 and m8
	require.Nil(t, ico.caClientCert, "caClientCert should be nil but got a non-nil one: %s. Expected nil interface", ico.caClientCert)
	require.Nil(t, ico.caKeyStorePath, "caKeyStorePath should be nil but got non-nil one: %s. Expected nil interface", ico.caKeyStorePath)
	require.Nil(t, ico.credentialStorePath, "credentialStorePath should be nil but got non-nil one: %s. Expected nil interface", ico.credentialStorePath)
	require.Nil(t, ico.tlsCACertPool, "tlsCACertPool should be nil but got non-nil one: %s. Expected nil interface", ico.tlsCACertPool)

	// do the same test using IsIdentityConfigFullyOverridden() call
	require.False(t, IsIdentityConfigFullyOverridden(ico), "IsIdentityConfigFullyOverridden is supposed to return false with an Options instance not implementing all the interface functions")

	// now inject default interfaces (using m0 as default full implementation for the sake of this test) for the ones that were not overridden by options above
	identityConfigOptionWithSomeDefaults := UpdateMissingOptsWithDefaultConfig(ico, m0)

	// test implementations m1-m4 are still working

	// test m1 implementation
	clnt := identityConfigOptionWithSomeDefaults.Client()
	require.NotEmpty(t, clnt, "client returned must not be empty")

	// test m2 implementation
	caCfg, ok := identityConfigOptionWithSomeDefaults.CAConfig("testCA")
	require.True(t, ok, "CAConfig failed")
	require.Equal(t, "test.url.com", caCfg.URL, "CAConfig did not return expected interface value")

	// test m3 implementation
	s, ok := identityConfigOptionWithSomeDefaults.CAServerCerts("testCA")
	require.True(t, ok, "CAServerCerts failed")
	require.Equal(t, []byte("testCAservercert1"), s[0], "CAServerCerts did not return the right cert")
	require.Equal(t, []byte("testCAservercert2"), s[1], "CAServerCerts did not return the right cert")

	// test m4 implementation
	c, ok := identityConfigOptionWithSomeDefaults.CAClientKey("testCA")
	require.True(t, ok, "CAClientKey failed")
	require.Equal(t, []byte("testCAclientkey"), c, "CAClientKey did not return the right cert")

	if ico, ok = identityConfigOptionWithSomeDefaults.(*IdentityConfigOptions); !ok {
		t.Fatal("UpdateMissingOptsWithDefaultConfig() call did not return an implementation of IdentityConfigOptions")
	}

	// now check if implementations that were not injected when building the config (ref first line in this function) are defaulted with m0 this time
	// ie, verify these implementations should now be populated in ico: m5, m6, m7, m8
	require.NotNil(t, ico.caClientCert, "caClientCert should be populated with default interface but got nil one: %s. Expected default interface", ico.caClientCert)
	require.NotNil(t, ico.caKeyStorePath, "caKeyStorePath should be populated with default interface but got nil one: %s. Expected default interface", ico.caKeyStorePath)
	require.NotNil(t, ico.credentialStorePath, "credentialStorePath should be populated with default interface but got nil one: %s. Expected default interface", ico.credentialStorePath)
	require.NotNil(t, ico.tlsCACertPool, "tlsCACertPool should be populated with default interface but got nil one: %s. Expected default interface", ico.tlsCACertPool)

	// do the same test using IsIdentityConfigFullyOverridden() call
	require.True(t, IsIdentityConfigFullyOverridden(ico), "IsIdentityConfigFullyOverridden is supposed to return true since all the interface functions should be implemented")
}

type mockClient struct {
}

func (m *mockClient) Client() *msp.ClientConfig {
	return &msp.ClientConfig{
		CryptoConfig:    msp.CCType{Path: ""},
		CredentialStore: msp.CredentialStoreType{Path: "", CryptoStore: msp.CCType{Path: ""}},
		Logging:         logApi.LoggingType{Level: "INFO"},
		Organization:    "org1",
		TLSKey:          []byte(""),
		TLSCert:         []byte(""),
	}
}

type mockCaConfig struct{}

func (m *mockCaConfig) CAConfig(org string) (*msp.CAConfig, bool) {
	return &msp.CAConfig{
		ID:               "test.url.com",
		URL:              "test.url.com",
		Registrar:        msp.EnrollCredentials{EnrollSecret: "secret", EnrollID: ""},
		TLSCAClientKey:   []byte(""),
		TLSCAClientCert:  []byte(""),
		TLSCAServerCerts: [][]byte{[]byte("")},
	}, true
}

type mockCaServerCerts struct{}

func (m *mockCaServerCerts) CAServerCerts(org string) ([][]byte, bool) {
	return [][]byte{[]byte("testCAservercert1"), []byte("testCAservercert2")}, true
}

type mockCaClientKey struct{}

func (m *mockCaClientKey) CAClientKey(org string) ([]byte, bool) {
	return []byte("testCAclientkey"), true
}

type mockCaClientCert struct{}

func (m *mockCaClientCert) CAClientCert(org string) ([]byte, bool) {
	return []byte("testCAclientcert"), true
}

type mockCaKeyStorePath struct{}

func (m *mockCaKeyStorePath) CAKeyStorePath() string {
	return "test/store/path"
}

type mockCredentialStorePath struct{}

func (m *mockCredentialStorePath) CredentialStorePath() string {
	return "test/cred/store/path"
}

type mockTLSCACertPool struct{}

var mockTLSCACertPoolInstance = &mockfab.MockCertPool{}

func (m *mockTLSCACertPool) TLSCACertPool() commtls.CertPool {
	return mockTLSCACertPoolInstance
}
