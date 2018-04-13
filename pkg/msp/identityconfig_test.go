/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"fmt"

	"os"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/stretchr/testify/assert"
)

const (
	configTestFilePath              = "../core/config/testdata/config_test.yaml"
	configEmbeddedUsersTestFilePath = "../core/config/testdata/config_test_embedded_pems.yaml"
	configPemTestFilePath           = "../core/config/testdata/config_test_pem.yaml"
	configType                      = "yaml"
)

func TestCAConfigFailsByNetworkConfig(t *testing.T) {

	//Tamper 'client.network' value and use a new config to avoid conflicting with other tests

	configBackend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatalf("Unexpected error reading config: %v", err)
	}

	backendMap := make(map[string]interface{})
	backendMap["client"], _ = configBackend.Lookup("client")
	backendMap["certificateAuthorities"], _ = configBackend.Lookup("certificateAuthorities")
	fmt.Println(configBackend.Lookup("certificateAuthorities"))
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
	certfile, err := sampleIdentityConfig.CAClientCertPath("peerorg1")
	fmt.Println(err)
	if certfile != "" || err == nil {
		t.Fatal("CA Cert file location read supposed to fail")
	}

	//Test CA client cert file failure scenario
	keyFile, err := sampleIdentityConfig.CAClientKeyPath("peerorg1")
	if keyFile != "" || err == nil {
		t.Fatal("CA Key file location read supposed to fail")
	}

	//Testing CA Server Cert Files failure scenario
	testCAServerCertFailureScenario(sampleIdentityConfig, t)

	//Testing CAConfig failure scenario
	testCAConfigFailureScenario(sampleIdentityConfig, t)

}

func testCAServerCertFailureScenario(sampleIdentityConfig *IdentityConfig, t *testing.T) {
	sCertFiles, err := sampleIdentityConfig.CAServerCertPaths("peerorg1")
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
	config, err := ConfigFromBackend(embeddedBackend)
	if err != nil {
		t.Fatalf("Failed to initialize identity config , reason: %v", err)
	}

	identityConfig := config.(*IdentityConfig)
	certPem, _ := identityConfig.CAClientCertPem(org1)
	certConfig := endpoint.TLSConfig{Pem: certPem}

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

	keyPem, _ := identityConfig.CAClientKeyPem(org1)

	keyConfig := endpoint.TLSConfig{Pem: keyPem}

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

	config, err := ConfigFromBackend(backend)
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
	checkCAServerCertPems("org1", idConfig, t)

	// get the client cert pem (embedded) for org1
	checkClientCertPem(idConfig, "org1", t)

	// get CA Server certs paths for org1
	checkCAServerCertsPath("org1", idConfig, t)

	// get the client cert path for org1
	checkClientCertPath(idConfig, "org1", t)

	// get the client key pem (embedded) for org1
	checkClientKeyPem(idConfig, "org1", t)

	// get the client key file path for org1
	checkClientKeyFilePath(idConfig, "org1", t)
}

func checkPeerPem(org string, idConfig *IdentityConfig, peer string, t *testing.T) {
	p0, err := idConfig.endpointConfig.PeerConfig(org, peer)
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

func checkCAServerCertPems(org string, idConfig *IdentityConfig, t *testing.T) {
	certs, err := idConfig.CAServerCertPems(org)
	if err != nil {
		t.Fatalf("Failed to load CAServerCertPems from config. Error: %s", err)
	}
	if len(certs) == 0 {
		t.Fatalf("Got empty PEM certs for CAServerCertPems")
	}
}

func checkClientCertPem(idConfig *IdentityConfig, org string, t *testing.T) {
	_, err := idConfig.CAClientCertPem(org)
	if err != nil {
		t.Fatalf("Failed to load CAClientCertPem from config. Error: %s", err)
	}
}

func checkCAServerCertsPath(org string, idConfig *IdentityConfig, t *testing.T) {
	certs, err := idConfig.CAServerCertPaths(org)
	if err != nil {
		t.Fatalf("Failed to load CAServerCertPaths from config. Error: %s", err)
	}
	if len(certs) == 0 {
		t.Fatalf("Got empty cert file paths for CAServerCertPaths")
	}
}

func checkClientCertPath(idConfig *IdentityConfig, org string, t *testing.T) {
	_, err := idConfig.CAClientCertPath(org)
	if err != nil {
		t.Fatalf("Failed to load CAClientCertPath from config. Error: %s", err)
	}
}

func checkClientKeyPem(idConfig *IdentityConfig, org string, t *testing.T) {
	_, err := idConfig.CAClientKeyPem(org)
	if err != nil {
		t.Fatalf("Failed to load CAClientKeyPem from config. Error: %s", err)
	}
}

func checkClientKeyFilePath(idConfig *IdentityConfig, org string, t *testing.T) {
	_, err := idConfig.CAClientKeyPath(org)
	if err != nil {
		t.Fatalf("Failed to load CAClientKeyPath from config. Error: %s", err)
	}
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

func TestCAConfig(t *testing.T) {
	//Test config
	backend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	config, err := ConfigFromBackend(backend)
	if err != nil {
		t.Fatal("Failed to get identity config")
	}
	identityConfig := config.(*IdentityConfig)
	//Test Crypto config path

	val, ok := backend.Lookup("client.cryptoconfig.path")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}

	assert.True(t, pathvar.Subst(val.(string)) == identityConfig.endpointConfig.CryptoConfigPath(), "Incorrect crypto config path", t)

	//Testing CA Client File Location
	certfile, err := identityConfig.CAClientCertPath(org1)

	if certfile == "" || err != nil {
		t.Fatalf("CA Cert file location read failed %s", err)
	}

	//Testing CA Key File Location
	keyFile, err := identityConfig.CAClientKeyPath(org1)

	if keyFile == "" || err != nil {
		t.Fatal("CA Key file location read failed")
	}

	//Testing CA Server Cert Files
	testCAServerCertFiles(identityConfig, t, org1)

	//Testing MSPID
	testMSPID(identityConfig, t, org1)

	//Testing CAConfig
	testCAConfig(identityConfig, t, org1)

	// Test CA KeyStore Path
	testCAKeyStorePath(backend, t, identityConfig)

	// test Client
	testClient(identityConfig, t)

	// testing empty OrgMSP
	testEmptyOrgMsp(identityConfig, t)
}

func testCAServerCertFiles(identityConfig *IdentityConfig, t *testing.T, org string) {
	sCertFiles, err := identityConfig.CAServerCertPaths(org)
	if len(sCertFiles) == 0 || err != nil {
		t.Fatal("Getting CA server cert files failed")
	}
}

func testMSPID(identityConfig *IdentityConfig, t *testing.T, org string) {
	mspID, err := identityConfig.endpointConfig.MSPID(org)
	if mspID != "Org1MSP" || err != nil {
		t.Fatal("Get MSP ID failed")
	}
}

func testCAConfig(identityConfig *IdentityConfig, t *testing.T, org string) {
	caConfig, err := identityConfig.CAConfig(org)
	if caConfig == nil || err != nil {
		t.Fatal("Get CA Config failed")
	}
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

func testClient(identityConfig *IdentityConfig, t *testing.T) {
	c, err := identityConfig.Client()
	if err != nil {
		t.Fatalf("Received error when fetching Client info, error is %s", err)
	}
	if c == nil {
		t.Fatal("Received empty client when fetching Client info")
	}
}

func testEmptyOrgMsp(identityConfig *IdentityConfig, t *testing.T) {
	_, err := identityConfig.endpointConfig.MSPID("dummyorg1")
	if err == nil {
		t.Fatal("Get MSP ID did not fail for dummyorg1")
	}
}
