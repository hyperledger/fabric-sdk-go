/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"crypto/x509"
	"testing"

	"os"
	"strings"

	"encoding/pem"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	configTestFilePath               = "../core/config/testdata/config_test.yaml"
	configEmbeddedUsersTestFilePath  = "../core/config/testdata/config_test_embedded_pems.yaml"
	configPemTestFilePath            = "../core/config/testdata/config_test_pem.yaml"
	configTestEntityMatchersFilePath = "../core/config/testdata/config_test_entity_matchers.yaml"
	configType                       = "yaml"
)

func TestCAConfigFailsByNetworkConfig(t *testing.T) {

	//Tamper 'client.network' value and use a new config to avoid conflicting with other tests

	configBackends, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatalf("Unexpected error reading config: %s", err)
	}
	if len(configBackends) != 1 {
		t.Fatalf("expected 1 backend but got %d", len(configBackends))
	}
	configBackend := configBackends[0]

	backendMap := make(map[string]interface{})
	backendMap["client"], _ = configBackend.Lookup("client")
	backendMap["certificateAuthorities"], _ = configBackend.Lookup("certificateAuthorities")
	backendMap["entityMatchers"], _ = configBackend.Lookup("entityMatchers")
	backendMap["peers"], _ = configBackend.Lookup("peers")
	backendMap["organizations"], _ = configBackend.Lookup("organizations")
	backendMap["orderers"], _ = configBackend.Lookup("orderers")
	backendMap["channels"], _ = configBackend.Lookup("channels")

	customBackend := &mocks.MockConfigBackend{KeyValueMap: backendMap}

	identityCfg, err := ConfigFromBackend(customBackend)
	if err != nil {
		t.Fatalf("Unexpected error initializing endpoint config: %s", err)
	}

	sampleIdentityConfig := identityCfg.(*IdentityConfig)
	customBackend.KeyValueMap["certificateAuthorities"] = ""

	//Test CA client cert file failure scenario
	certfile, ok := sampleIdentityConfig.CAClientCert("peerorg1")
	if certfile != nil || ok {
		t.Fatal("CA Cert file location read supposed to fail")
	}

	//Test CA client cert file failure scenario
	keyFile, ok := sampleIdentityConfig.CAClientKey("peerorg1")
	if keyFile != nil || ok {
		t.Fatal("CA Key file location read supposed to fail")
	}

	//Testing CA Server Cert Files failure scenario
	testCAServerCertFailureScenario(sampleIdentityConfig, t)

	//Testing CAConfig failure scenario
	testCAConfigFailureScenario(sampleIdentityConfig, t)

}

func testCAServerCertFailureScenario(sampleIdentityConfig *IdentityConfig, t *testing.T) {
	sCertFiles, ok := sampleIdentityConfig.CAServerCerts("peerorg1")
	if len(sCertFiles) > 0 || ok {
		t.Fatal("Getting CA server cert files supposed to fail")
	}
}

func testCAConfigFailureScenario(sampleIdentityConfig *IdentityConfig, t *testing.T) {
	caConfig, ok := sampleIdentityConfig.CAConfig("peerorg1")
	if caConfig != nil || ok {
		t.Fatal("Get CA Config supposed to fail")
	}
}

func TestTLSCAConfigFromPems(t *testing.T) {
	embeddedBackend, err := config.FromFile(configEmbeddedUsersTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	//Test TLSCA Cert Pool (Positive test case)
	config, err := ConfigFromBackend(embeddedBackend...)
	assert.Nil(t, err, "Failed to initialize identity config , reason: %s", err)
	endpointConfig, err := fab.ConfigFromBackend(embeddedBackend...)
	assert.Nil(t, err, "Failed to initialize endpoint config , reason: %s", err)

	identityConfig := config.(*IdentityConfig)
	certPem, _ := identityConfig.CAClientCert(org1)
	certConfig := endpoint.TLSConfig{Pem: string(certPem)}

	err = certConfig.LoadBytes()
	assert.Nil(t, err, "TLS CA cert parse failed, reason: %s", err)

	cert, ok, err := certConfig.TLSCert()
	assert.Nil(t, err, "TLS CA cert parse failed, reason: %s", err)
	assert.True(t, ok, "TLS CA cert parse failed")

	_, err = endpointConfig.TLSCACertPool().Get(cert)
	assert.Nil(t, err, "TLS CA cert pool fetch failed, reason: %s", err)
	//Test TLSCA Cert Pool (Negative test case)

	badCertConfig := endpoint.TLSConfig{Pem: "some random invalid pem"}
	err = badCertConfig.LoadBytes()
	assert.Nil(t, err, "LoadBytes should not fail for bad pemgit g")

	badCert, ok, err := badCertConfig.TLSCert()
	assert.Nil(t, err, "TLS CA cert parse was supposed to fail")
	assert.False(t, ok, "TLS CA cert parse was supposed to fail")

	_, err = endpointConfig.TLSCACertPool().Get(badCert)
	assert.Nil(t, err, "TLSCACertPool failed %s", err)

	keyPem, ok := identityConfig.CAClientKey(org1)
	assert.True(t, ok, "CAClientKey supposed to succeed")

	keyConfig := endpoint.TLSConfig{Pem: string(keyPem)}

	_, ok, err = keyConfig.TLSCert()
	assert.Nil(t, err, "TLS CA cert pool was supposed to fail when provided with wrong cert file")
	assert.False(t, ok, "TLS CA cert pool was supposed to fail when provided with wrong cert file")

}

func TestInitConfigFromRawWithPem(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configPemTestFilePath)
	if err != nil {
		t.Fatalf("Failed to load sample bytes from File. Error: %s", err)
	}

	// test init config from bytes
	backend, err := config.FromRaw(cBytes, configType)()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}

	config, err := ConfigFromBackend(backend...)
	if err != nil {
		t.Fatalf("Failed to initialize identity config from bytes array. Error: %s", err)
	}
	endpointConfig, err := fab.ConfigFromBackend(backend...)
	if err != nil {
		t.Fatalf("Failed to initialize endpoint config from bytes array. Error: %s", err)
	}

	idConfig := config.(*IdentityConfig)

	o := endpointConfig.OrderersConfig()
	if len(o) == 0 {
		t.Fatal("orderer cannot be nil or empty")
	}

	oPem := `-----BEGIN CERTIFICATE-----
MIICNjCCAdygAwIBAgIRAILSPmMB3BzoLIQGsFxwZr8wCgYIKoZIzj0EAwIwbDEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xFDASBgNVBAoTC2V4YW1wbGUuY29tMRowGAYDVQQDExF0bHNjYS5l
eGFtcGxlLmNvbTAeFw0xNzA3MjgxNDI3MjBaFw0yNzA3MjYxNDI3MjBaMGwxCzAJ
BgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJh
bmNpc2NvMRQwEgYDVQQKEwtleGFtcGxlLmNvbTEaMBgGA1UEAxMRdGxzY2EuZXhh
bXBsZS5jb20wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQfgKb4db53odNzdMXn
P5FZTZTFztOO1yLvCHDofSNfTPq/guw+YYk7ZNmhlhj8JHFG6dTybc9Qb/HOh9hh
gYpXo18wXTAOBgNVHQ8BAf8EBAMCAaYwDwYDVR0lBAgwBgYEVR0lADAPBgNVHRMB
Af8EBTADAQH/MCkGA1UdDgQiBCBxaEP3nVHQx4r7tC+WO//vrPRM1t86SKN0s6XB
8LWbHTAKBggqhkjOPQQDAgNIADBFAiEA96HXwCsuMr7tti8lpcv1oVnXg0FlTxR/
SQtE5YgdxkUCIHReNWh/pluHTxeGu2jNCH1eh6o2ajSGeeizoapvdJbN
-----END CERTIFICATE-----`

	oCert, err := tlsCertByBytes([]byte(oPem))
	assert.Nil(t, err, "failed to cert from pem bytes")
	assert.Equal(t, oCert.RawSubject, o[0].TLSCACert.RawSubject, "certs supposed to match")

	pc, ok := endpointConfig.PeersConfig(org1)
	assert.True(t, ok)
	assert.NotEmpty(t, pc, "peers list cannot be nil or empty")

	peer0 := "peer0.org1.example.com"
	checkPeerPem(org1, endpointConfig, peer0, t)

	// get CA Server cert pems (embedded) for org1
	checkCAServerCerts("org1", idConfig, t)

	// get the client cert pem (embedded) for org1
	checkClientCert(idConfig, "org1", t)

	// get CA Server certs paths for org1
	checkCAServerCerts("org1", idConfig, t)

	// get the client cert path for org1
	checkClientCert(idConfig, "org1", t)

	// get the client key pem (embedded) for org1
	checkClientKey(idConfig, "org1", t)

	// get the client key file path for org1
	checkClientKey(idConfig, "org1", t)
}

func checkPeerPem(org string, endpointConfig fabImpl.EndpointConfig, peer string, t *testing.T) {
	p0, ok := endpointConfig.PeerConfig(peer)
	assert.True(t, ok)
	assert.NotNil(t, p0, "cannot be nil")

	pPem := `-----BEGIN CERTIFICATE-----
MIICSTCCAfCgAwIBAgIRAPQIzfkrCZjcpGwVhMSKd0AwCgYIKoZIzj0EAwIwdjEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHzAdBgNVBAMTFnRs
c2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQy
NzIwWjB2MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UE
BxMNU2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEfMB0G
A1UEAxMWdGxzY2Eub3JnMS5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABMOiG8UplWTs898zZ99+PhDHPbKjZIDHVG+zQXopw8SqNdX3NAmZUKUU
sJ8JZ3M49Jq4Ms8EHSEwQf0Ifx3ICHujXzBdMA4GA1UdDwEB/wQEAwIBpjAPBgNV
HSUECDAGBgRVHSUAMA8GA1UdEwEB/wQFMAMBAf8wKQYDVR0OBCIEID9qJz7xhZko
V842OVjxCYYQwCjPIY+5e9ORR+8pxVzcMAoGCCqGSM49BAMCA0cAMEQCIGZ+KTfS
eezqv0ml1VeQEmnAEt5sJ2RJA58+LegUYMd6AiAfEe6BKqdY03qFUgEYmtKG+3Dr
O94CDp7l2k7hMQI0zQ==
-----END CERTIFICATE-----`

	oCert, err := tlsCertByBytes([]byte(pPem))
	assert.Nil(t, err, "failed to cert from pem bytes")
	assert.Equal(t, oCert.RawSubject, p0.TLSCACert.RawSubject, "certs supposed to match")

}

func checkCAServerCerts(org string, idConfig *IdentityConfig, t *testing.T) {
	certs, ok := idConfig.CAServerCerts(org)
	assert.True(t, ok, "Failed to load CAServerCertPems from config.")
	assert.NotEmpty(t, certs, "Got empty PEM certs for CAServerCertPems")
}

func checkClientCert(idConfig *IdentityConfig, org string, t *testing.T) {
	cert, ok := idConfig.CAClientCert(org)
	assert.True(t, ok, "Failed to load CAClientCertPem from config.")
	assert.NotEmpty(t, cert, "Invalid cert")
}

func checkClientKey(idConfig *IdentityConfig, org string, t *testing.T) {
	key, ok := idConfig.CAClientKey(org)
	assert.True(t, ok, "Failed to load CAClientKeyPem from config.")
	assert.NotEmpty(t, key, "Invalid key")
}

func loadConfigBytesFromFile(t *testing.T, filePath string) ([]byte, error) {
	// read test config file into bytes array
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to read config file. Error: %s", err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("Failed to read config file stat. Error: %s", err)
	}
	s := fi.Size()
	cBytes := make([]byte, s)
	n, err := f.Read(cBytes)
	if err != nil {
		t.Fatalf("Failed to read test config for bytes array testing. Error: %s", err)
	}
	if n == 0 {
		t.Fatal("Failed to read test config for bytes array testing. Mock bytes array is empty")
	}
	return cBytes, err
}

func TestCAConfigCryptoFiles(t *testing.T) {
	//Test config
	backend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	config, err := ConfigFromBackend(backend...)
	if err != nil {
		t.Fatal("Failed to get identity config")
	}
	identityConfig := config.(*IdentityConfig)

	//Testing CA Client File Location
	certfile, ok := identityConfig.CAClientCert(org1)
	assert.True(t, ok, "CA Cert file location read failed ")
	assert.True(t, len(certfile) > 0)

	//Testing CA Key File Location
	keyFile, ok := identityConfig.CAClientKey(org1)
	assert.True(t, ok, "CA Key file location read failed ")
	assert.True(t, len(keyFile) > 0)

	//Testing CA Server Cert Files
	sCertFiles, ok := identityConfig.CAServerCerts(org1)
	assert.True(t, ok, "Getting CA server cert files failed")
	assert.True(t, len(sCertFiles) > 0)

}

func TestCAConfig(t *testing.T) {
	//Test config
	backend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	config, err := ConfigFromBackend(backend...)
	if err != nil {
		t.Fatal("Failed to get identity config")
	}

	identityConfig := config.(*IdentityConfig)
	//Test Crypto config path

	val, ok := backend[0].Lookup("client.cryptoconfig.path")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}

	//Testing CAConfig
	caConfig, ok := identityConfig.CAConfig(org1)
	assert.True(t, ok, "Get CA Config failed")
	assert.NotNil(t, caConfig, "Get CA Config failed")

	// Test CA KeyStore Path
	testCAKeyStorePath(backend[0], t, identityConfig)

	// test Client
	c := identityConfig.Client()
	assert.NotNil(t, c, "Received error when fetching Client info")

}

func testCAKeyStorePath(backend core.ConfigBackend, t *testing.T, identityConfig *IdentityConfig) {
	// Test User Store Path
	val, ok := backend.Lookup("client.credentialStore.path")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(string) != identityConfig.CredentialStorePath() {
		t.Fatal("Incorrect User Store path")
	}
	val, ok = backend.Lookup("client.credentialStore.cryptoStore.path")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(string) != identityConfig.CAKeyStorePath() {
		t.Fatal("Incorrect CA keystore path")
	}
}

func TestCACertAndKeys(t *testing.T) {

	backend, err := config.FromFile(configEmbeddedUsersTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}
	orgIDs := []string{"org1", "org2"}

	config, err := ConfigFromBackend(backend...)
	if err != nil {
		t.Fatal("Failed to get identity config")
	}
	identityConfig := config.(*IdentityConfig)

	for _, orgID := range orgIDs {
		val, ok := identityConfig.CAClientCert(orgID)
		assert.True(t, ok, "identityConfig.CAClientCert not supposed to return failure")
		assert.True(t, len(val) > 0, "identityConfig.CAClientCert supposed to return valid cert")

		val, ok = identityConfig.CAClientKey(orgID)
		assert.True(t, ok, "identityConfig.CAClientKey not supposed to return failure")
		assert.True(t, len(val) > 0, "identityConfig.CAClientKey supposed to return valid key")

		vals, ok := identityConfig.CAServerCerts(orgID)
		assert.True(t, ok, "identityConfig.CAClientKey not supposed to return failure")
		assert.True(t, len(vals) > 0, "identityConfig.CAClientKey supposed to return server certs")
		for _, v := range vals {
			assert.True(t, len(v) > 0, "identityConfig.CAClientKey supposed to return valid server cert")
		}
	}

}

func TestIdentityConfigWithMultipleBackends(t *testing.T) {

	sampleViper := newViper(configTestEntityMatchersFilePath)

	var backends []core.ConfigBackend
	backendMap := make(map[string]interface{})
	backendMap["client"] = sampleViper.Get("client")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["channels"] = sampleViper.Get("channels")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["certificateAuthorities"] = sampleViper.Get("certificateAuthorities")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["entityMatchers"] = sampleViper.Get("entityMatchers")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["organizations"] = sampleViper.Get("organizations")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["orderers"] = sampleViper.Get("orderers")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["peers"] = sampleViper.Get("peers")
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	//create endpointConfig with all 7 backends having 7 different entities
	identityConfig, err := ConfigFromBackend(backends...)

	assert.Nil(t, err, "ConfigFromBackend should have been successful for multiple backends")
	assert.NotNil(t, identityConfig, "Invalid identity config from multiple backends")

	//Client
	client := identityConfig.Client()
	assert.NotNil(t, client, "invalid client config")
	assert.Equal(t, client.Organization, "org1")

	//CA Config
	caConfig, ok := identityConfig.CAConfig("org1")
	assert.True(t, ok, "identityConfig.CAConfig(org1) should have been successful for multiple backends")
	assert.Equal(t, caConfig.URL, "https://ca.org1.example.com:7054")

}

func newViper(path string) *viper.Viper {
	myViper := viper.New()
	replacer := strings.NewReplacer(".", "_")
	myViper.SetEnvKeyReplacer(replacer)
	myViper.SetConfigFile(path)
	err := myViper.MergeInConfig()
	if err != nil {
		panic(err)
	}
	return myViper
}

func tlsCertByBytes(bytes []byte) (*x509.Certificate, error) {

	block, _ := pem.Decode(bytes)

	if block != nil {
		pub, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		return pub, nil
	}

	//no cert found and there is no error
	return nil, errors.New("empty byte")
}
