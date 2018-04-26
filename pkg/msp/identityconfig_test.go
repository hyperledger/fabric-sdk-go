/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"os"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
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
		t.Fatalf("Unexpected error reading config: %v", err)
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
		t.Fatalf("Unexpected error initializing endpoint config: %v", err)
	}

	sampleIdentityConfig := identityCfg.(*IdentityConfig)
	sampleIdentityConfig.endpointConfig.ResetNetworkConfig()

	customBackend.KeyValueMap["channels"] = "INVALID"
	_, err = sampleIdentityConfig.networkConfig()
	if err == nil {
		t.Fatal("Network config load supposed to fail")
	}

	customBackend.KeyValueMap["channels"], _ = configBackend.Lookup("channels")
	customBackend.KeyValueMap["certificateAuthorities"] = ""

	//Test CA client cert file failure scenario
	certfile, err := sampleIdentityConfig.CAClientCert("peerorg1")
	if certfile != nil || err == nil {
		t.Fatal("CA Cert file location read supposed to fail")
	}

	//Test CA client cert file failure scenario
	keyFile, err := sampleIdentityConfig.CAClientKey("peerorg1")
	if keyFile != nil || err == nil {
		t.Fatal("CA Key file location read supposed to fail")
	}

	//Testing CA Server Cert Files failure scenario
	testCAServerCertFailureScenario(sampleIdentityConfig, t)

	//Testing CAConfig failure scenario
	testCAConfigFailureScenario(sampleIdentityConfig, t)

}

func testCAServerCertFailureScenario(sampleIdentityConfig *IdentityConfig, t *testing.T) {
	sCertFiles, err := sampleIdentityConfig.CAServerCerts("peerorg1")
	if len(sCertFiles) > 0 || err == nil {
		t.Fatal("Getting CA server cert files supposed to fail")
	}
}

func testCAConfigFailureScenario(sampleIdentityConfig *IdentityConfig, t *testing.T) {
	caConfig, err := sampleIdentityConfig.CAConfig("peerorg1")
	if caConfig != nil || err == nil {
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
	if err != nil {
		t.Fatalf("Failed to initialize identity config , reason: %v", err)
	}

	identityConfig := config.(*IdentityConfig)
	certPem, _ := identityConfig.CAClientCert(org1)
	certConfig := endpoint.TLSConfig{Pem: string(certPem)}

	cert, err := certConfig.TLSCert()

	if err != nil {
		t.Fatalf("TLS CA cert parse failed, reason: %v", err)
	}

	_, err = identityConfig.endpointConfig.TLSCACertPool(cert)

	if err != nil {
		t.Fatalf("TLS CA cert pool fetch failed, reason: %v", err)
	}
	//Test TLSCA Cert Pool (Negative test case)

	badCertConfig := endpoint.TLSConfig{Pem: "some random invalid pem"}

	badCert, err := badCertConfig.TLSCert()

	if err == nil {
		t.Fatalf("TLS CA cert parse was supposed to fail")
	}

	_, err = identityConfig.endpointConfig.TLSCACertPool(badCert)
	if err != nil {
		t.Fatalf("TLSCACertPool failed %v", err)
	}

	keyPem, _ := identityConfig.CAClientKey(org1)

	keyConfig := endpoint.TLSConfig{Pem: string(keyPem)}

	_, err = keyConfig.TLSCert()
	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail when provided with wrong cert file")
	}

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
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}

	idConfig := config.(*IdentityConfig)

	o, err := idConfig.endpointConfig.OrderersConfig()
	if err != nil {
		t.Fatalf("Failed to load orderers from config. Error: %s", err)
	}

	if len(o) == 0 {
		t.Fatalf("orderer cannot be nil or empty")
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
	loadedOPem := strings.TrimSpace(o[0].TLSCACerts.Pem) // viper's unmarshall adds a \n to the end of a string, hence the TrimeSpace
	if loadedOPem != oPem {
		t.Fatalf("Orderer Pem doesn't match. Expected \n'%s'\n, but got \n'%s'\n", oPem, loadedOPem)
	}

	pc, err := idConfig.endpointConfig.PeersConfig(org1)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(pc) == 0 {
		t.Fatalf("peers list of %s cannot be nil or empty", org1)
	}
	peer0 := "peer0.org1.example.com"
	checkPeerPem(org1, idConfig, peer0, t)

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

func checkPeerPem(org string, idConfig *IdentityConfig, peer string, t *testing.T) {
	p0, err := idConfig.endpointConfig.PeerConfig(peer)
	if err != nil {
		t.Fatalf("Failed to load %s of %s from the config. Error: %s", peer, org, err)
	}
	if p0 == nil {
		t.Fatalf("%s of %s cannot be nil", peer, org)
	}
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
	loadedPPem := strings.TrimSpace(p0.TLSCACerts.Pem)
	// viper's unmarshall adds a \n to the end of a string, hence the TrimeSpace
	if loadedPPem != pPem {
		t.Fatalf("%s Pem doesn't match. Expected \n'%s'\n, but got \n'%s'\n", peer, pPem, loadedPPem)
	}
}

func checkCAServerCerts(org string, idConfig *IdentityConfig, t *testing.T) {
	certs, err := idConfig.CAServerCerts(org)
	if err != nil {
		t.Fatalf("Failed to load CAServerCertPems from config. Error: %s", err)
	}
	if len(certs) == 0 {
		t.Fatalf("Got empty PEM certs for CAServerCertPems")
	}
}

func checkClientCert(idConfig *IdentityConfig, org string, t *testing.T) {
	cert, err := idConfig.CAClientCert(org)
	if err != nil {
		t.Fatalf("Failed to load CAClientCertPem from config. Error: %s", err)
	}
	assert.True(t, len(cert) > 0, "Invalid cert")
}

func checkClientKey(idConfig *IdentityConfig, org string, t *testing.T) {
	key, err := idConfig.CAClientKey(org)
	if err != nil {
		t.Fatalf("Failed to load CAClientKeyPem from config. Error: %s", err)
	}
	assert.True(t, len(key) > 0, "Invalid key")
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
		t.Fatalf("Failed to read test config for bytes array testing. Mock bytes array is empty")
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
	certfile, err := identityConfig.CAClientCert(org1)
	assert.Nil(t, err, "CA Cert file location read failed ")
	assert.True(t, len(certfile) > 0)

	//Testing CA Key File Location
	keyFile, err := identityConfig.CAClientKey(org1)
	assert.Nil(t, err, "CA Key file location read failed ")
	assert.True(t, len(keyFile) > 0)

	//Testing CA Server Cert Files
	sCertFiles, err := identityConfig.CAServerCerts(org1)
	assert.Nil(t, err, "Getting CA server cert files failed")
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

	assert.True(t, pathvar.Subst(val.(string)) == identityConfig.endpointConfig.CryptoConfigPath(), "Incorrect crypto config path", t)

	//Testing MSPID
	mspID, err := identityConfig.endpointConfig.MSPID(org1)
	assert.Nil(t, err, "Get MSP ID failed")
	assert.True(t, mspID == "Org1MSP", "Get MSP ID failed")

	// testing empty OrgMSP
	_, err = identityConfig.endpointConfig.MSPID("dummyorg1")
	assert.NotNil(t, err, "Get MSP ID did not fail for dummyorg1")
	assert.True(t, err.Error() == "MSP ID is empty for org: dummyorg1", "Get MSP ID did not fail for dummyorg1")

	//Testing CAConfig
	caConfig, err := identityConfig.CAConfig(org1)
	assert.Nil(t, err, "Get CA Config failed")
	assert.NotNil(t, caConfig, "Get CA Config failed")

	// Test CA KeyStore Path
	testCAKeyStorePath(backend[0], t, identityConfig)

	// test Client
	c, err := identityConfig.Client()
	assert.Nil(t, err, "Received error when fetching Client info")
	assert.NotNil(t, c, "Received error when fetching Client info")

}

func testCAKeyStorePath(backend core.ConfigBackend, t *testing.T, identityConfig *IdentityConfig) {
	// Test User Store Path
	val, ok := backend.Lookup("client.credentialStore.path")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(string) != identityConfig.CredentialStorePath() {
		t.Fatalf("Incorrect User Store path")
	}
	val, ok = backend.Lookup("client.credentialStore.cryptoStore.path")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(string) != identityConfig.CAKeyStorePath() {
		t.Fatalf("Incorrect CA keystore path")
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
		val, err := identityConfig.CAClientCert(orgID)
		assert.Nil(t, err, "identityConfig.CAClientCert not supposed to return error")
		assert.True(t, len(val) > 0, "identityConfig.CAClientCert supposed to return valid cert")

		val, err = identityConfig.CAClientKey(orgID)
		assert.Nil(t, err, "identityConfig.CAClientKey not supposed to return error")
		assert.True(t, len(val) > 0, "identityConfig.CAClientKey supposed to return valid key")

		vals, err := identityConfig.CAServerCerts(orgID)
		assert.Nil(t, err, "identityConfig.CAClientKey not supposed to return error")
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
	client, err := identityConfig.Client()
	assert.Nil(t, err, "identityConfig.Client() should have been successful for multiple backends")
	assert.Equal(t, client.Organization, "org1")

	//CA Config
	caConfig, err := identityConfig.CAConfig("org1")
	assert.Nil(t, err, "identityConfig.CAConfig(org1) should have been successful for multiple backends")
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
