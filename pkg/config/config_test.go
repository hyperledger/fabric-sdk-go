/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"reflect"

	api "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/spf13/viper"
)

var configImpl *Config
var org0 = "org0"
var org1 = "Org1"
var configTestFilePath = "../../test/fixtures/config/config_test.yaml"
var configPemTestFilePath = "testdata/config_test_pem.yaml"
var configEmbeddedUsersTestFilePath = "../../test/fixtures/config/config_test_embedded_pems.yaml"
var configType = "yaml"

func TestDefaultConfig(t *testing.T) {
	vConfig := viper.New()
	vConfig.AddConfigPath(".")
	err := vConfig.ReadInConfig()
	if err != nil {
		t.Fatalf("Failed to load default config file")
	}
	//Test network name
	if vConfig.GetString("name") != "default-network" {
		t.Fatalf("Incorrect Network name from default config")
	}
}

func TestCAConfig(t *testing.T) {
	//Test config
	vConfig := viper.New()
	vConfig.SetConfigFile(configTestFilePath)
	vConfig.ReadInConfig()
	vc := vConfig.ConfigFileUsed()

	if vc == "" {
		t.Fatalf("Failed to load config file")
	}

	//Test network name
	if vConfig.GetString("name") != "global-trade-network" {
		t.Fatalf("Incorrect Network name")
	}

	//Test client app specific variable x-type
	if vConfig.GetString("x-type") != "hlfv1" {
		t.Fatalf("Incorrect Netwok x-type")
	}

	//Test network description
	if vConfig.GetString("description") != "The network to be in if you want to stay in the global trade business" {
		t.Fatalf("Incorrect Network description")
	}

	//Test network config version
	if vConfig.GetString("version") != "1.0.0" {
		t.Fatalf("Incorrect network version")
	}

	//Test client organization
	if vConfig.GetString("client.organization") != org1 {
		t.Fatalf("Incorrect Client organization")
	}

	//Test Crypto config path
	crossCheckWithViperConfig(configImpl.configViper.GetString("client.cryptoconfig.path"), configImpl.CryptoConfigPath(), "Incorrect crypto config path", t)

	//Testing CA Client File Location
	certfile, err := configImpl.CAClientCertPath(org1)

	if certfile == "" || err != nil {
		t.Fatalf("CA Cert file location read failed %s", err)
	}

	//Testing CA Key File Location
	keyFile, err := configImpl.CAClientKeyPath(org1)

	if keyFile == "" || err != nil {
		t.Fatal("CA Key file location read failed")
	}

	//Testing CA Server Cert Files
	sCertFiles, err := configImpl.CAServerCertPaths(org1)

	if sCertFiles == nil || len(sCertFiles) == 0 || err != nil {
		t.Fatal("Getting CA server cert files failed")
	}

	//Testing MSPID
	mspID, err := configImpl.MspID(org1)
	if mspID != "Org1MSP" || err != nil {
		t.Fatal("Get MSP ID failed")
	}

	//Testing CAConfig
	caConfig, err := configImpl.CAConfig(org1)
	if caConfig == nil || err != nil {
		t.Fatal("Get CA Config failed")
	}

	// Test CA KeyStore Path
	if vConfig.GetString("client.credentialStore.cryptoStore.path") != configImpl.CAKeyStorePath() {
		t.Fatalf("Incorrect CA keystore path")
	}

	// Test KeyStore Path
	if path.Join(vConfig.GetString("client.credentialStore.cryptoStore.path"), "keystore") != configImpl.KeyStorePath() {
		t.Fatalf("Incorrect keystore path ")
	}

	// Test BCCSP security is enabled
	if vConfig.GetBool("client.BCCSP.security.enabled") != configImpl.IsSecurityEnabled() {
		t.Fatalf("Incorrect BCCSP Security enabled flag")
	}

	// Test SecurityAlgorithm
	if vConfig.GetString("client.BCCSP.security.hashAlgorithm") != configImpl.SecurityAlgorithm() {
		t.Fatalf("Incorrect BCCSP Security Hash algorithm")
	}

	// Test Security Level
	if vConfig.GetInt("client.BCCSP.security.level") != configImpl.SecurityLevel() {
		t.Fatalf("Incorrect BCCSP Security Level")
	}

	// Test SecurityProvider provider
	if vConfig.GetString("client.BCCSP.security.default.provider") != configImpl.SecurityProvider() {
		t.Fatalf("Incorrect BCCSP SecurityProvider provider")
	}

	// Test Ephemeral flag
	if vConfig.GetBool("client.BCCSP.security.ephemeral") != configImpl.Ephemeral() {
		t.Fatalf("Incorrect BCCSP Ephemeral flag")
	}

	// Test SoftVerify flag
	if vConfig.GetBool("client.BCCSP.security.softVerify") != configImpl.SoftVerify() {
		t.Fatalf("Incorrect BCCSP Ephemeral flag")
	}

	// Test SecurityProviderPin
	if vConfig.GetString("client.BCCSP.security.pin") != configImpl.SecurityProviderPin() {
		t.Fatalf("Incorrect BCCSP SecurityProviderPin flag")
	}

	// Test SecurityProviderPin
	if vConfig.GetString("client.BCCSP.security.label") != configImpl.SecurityProviderLabel() {
		t.Fatalf("Incorrect BCCSP SecurityProviderPin flag")
	}

	// test Client
	c, err := configImpl.Client()
	if err != nil {
		t.Fatalf("Received error when fetching Client info, error is %s", err)
	}
	if c == nil {
		t.Fatal("Received empty client when fetching Client info")
	}

	// testing empty OrgMSP
	mspID, err = configImpl.MspID("dummyorg1")
	if err == nil {
		t.Fatal("Get MSP ID did not fail for dummyorg1")
	}
}

func TestCAConfigFailsByNetworkConfig(t *testing.T) {

	//Tamper 'client.network' value and use a new config to avoid conflicting with other tests
	sampleConfig, err := InitConfig(configTestFilePath)
	sampleConfig.configViper.Set("client", "INVALID")
	sampleConfig.configViper.Set("peers", "INVALID")
	sampleConfig.configViper.Set("organizations", "INVALID")
	sampleConfig.configViper.Set("orderers", "INVALID")
	sampleConfig.configViper.Set("channels", "INVALID")

	_, err = sampleConfig.NetworkConfig()
	if err == nil {
		t.Fatal("Network config load supposed to fail")
	}

	//Test CA client cert file failure scenario
	certfile, err := sampleConfig.CAClientCertPath("peerorg1")
	if certfile != "" || err == nil {
		t.Fatal("CA Cert file location read supposed to fail")
	}

	//Test CA client cert file failure scenario
	keyFile, err := sampleConfig.CAClientKeyPath("peerorg1")
	if keyFile != "" || err == nil {
		t.Fatal("CA Key file location read supposed to fail")
	}

	//Testing CA Server Cert Files failure scenario
	sCertFiles, err := sampleConfig.CAServerCertPaths("peerorg1")
	if len(sCertFiles) > 0 || err == nil {
		t.Fatal("Getting CA server cert files supposed to fail")
	}

	//Testing MSPID failure scenario
	mspID, err := sampleConfig.MspID("peerorg1")
	if mspID != "" || err == nil {
		t.Fatal("Get MSP ID supposed to fail")
	}

	//Testing CAConfig failure scenario
	caConfig, err := sampleConfig.CAConfig("peerorg1")
	if caConfig != nil || err == nil {
		t.Fatal("Get CA Config supposed to fail")
	}

	//Testing RandomOrdererConfig failure scenario
	oConfig, err := sampleConfig.RandomOrdererConfig()
	if oConfig != nil || err == nil {
		t.Fatal("Testing get RandomOrdererConfig supposed to fail")
	}

	//Testing RandomOrdererConfig failure scenario
	oConfig, err = sampleConfig.OrdererConfig("peerorg1")
	if oConfig != nil || err == nil {
		t.Fatal("Testing get OrdererConfig supposed to fail")
	}

	//Testing PeersConfig failure scenario
	pConfigs, err := sampleConfig.PeersConfig("peerorg1")
	if pConfigs != nil || err == nil {
		t.Fatal("Testing PeersConfig supposed to fail")
	}

	//Testing PeersConfig failure scenario
	pConfig, err := sampleConfig.PeerConfig("peerorg1", "peer1")
	if pConfig != nil || err == nil {
		t.Fatal("Testing PeerConfig supposed to fail")
	}

	//Testing ChannelConfig failure scenario
	chConfig, err := sampleConfig.ChannelConfig("invalid")
	if chConfig != nil || err == nil {
		t.Fatal("Testing ChannelConfig supposed to fail")
	}

	//Testing ChannelPeers failure scenario
	cpConfigs, err := sampleConfig.ChannelPeers("invalid")
	if cpConfigs != nil || err == nil {
		t.Fatal("Testing ChannelPeeers supposed to fail")
	}

	//Testing ChannelOrderers failure scenario
	coConfigs, err := sampleConfig.ChannelOrderers("invalid")
	if coConfigs != nil || err == nil {
		t.Fatal("Testing ChannelOrderers supposed to fail")
	}

	// test empty network objects
	sampleConfig.configViper.Set("organizations", nil)
	_, err = sampleConfig.NetworkConfig()
	if err == nil {
		t.Fatalf("Organizations were empty, it should return an error")
	}
}

func TestTLSACAConfig(t *testing.T) {
	//Test TLSCA Cert Pool (Positive test case)
	certFile, _ := configImpl.CAClientCertPath(org1)
	_, err := configImpl.TLSCACertPool(certFile)
	if err != nil {
		t.Fatalf("TLS CA cert pool fetch failed, reason: %v", err)
	}

	//Test TLSCA Cert Pool (Negative test case)
	_, err = configImpl.TLSCACertPool("some random invalid path")
	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail")
	}

	keyFile, _ := configImpl.CAClientKeyPath(org1)
	_, err = configImpl.TLSCACertPool(keyFile)
	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail when provided with wrong cert file")
	}
}

func TestTimeouts(t *testing.T) {
	configImpl.configViper.Set("client.peer.timeout.connection", "2s")
	configImpl.configViper.Set("client.eventService.timeout.connection", "2m")
	configImpl.configViper.Set("client.eventService.timeout.registrationResponse", "2h")
	configImpl.configViper.Set("client.orderer.timeout.connection", "2ms")
	configImpl.configViper.Set("client.peer.timeout.queryResponse", "7h")
	configImpl.configViper.Set("client.peer.timeout.executeTxResponse", "8h")
	configImpl.configViper.Set("client.orderer.timeout.response", "6s")

	t1 := configImpl.TimeoutOrDefault(api.Endorser)
	if t1 != time.Second*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.EventHub)
	if t1 != time.Minute*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.EventReg)
	if t1 != time.Hour*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.Query)
	if t1 != time.Hour*7 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.ExecuteTx)
	if t1 != time.Hour*8 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.OrdererConnection)
	if t1 != time.Millisecond*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.OrdererResponse)
	if t1 != time.Second*6 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}

	// Test default
	configImpl.configViper.Set("client.orderer.timeout.connection", "")
	t1 = configImpl.TimeoutOrDefault(api.OrdererConnection)
	if t1 != time.Second*5 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}

}

func TestOrdererConfig(t *testing.T) {
	oConfig, err := configImpl.RandomOrdererConfig()

	if oConfig == nil || err != nil {
		t.Fatal("Testing get RandomOrdererConfig failed")
	}

	oConfig, err = configImpl.OrdererConfig("invalid")

	if oConfig != nil || err != nil {
		t.Fatal("Testing non-existing OrdererConfig failed")
	}

	orderers, err := configImpl.OrderersConfig()
	if err != nil {
		t.Fatal(err)
	}

	if orderers[0].TLSCACerts.Path != "" {
		if !filepath.IsAbs(orderers[0].TLSCACerts.Path) {
			t.Fatal("Expected GOPATH relative path to be replaced")
		}
	} else if len(orderers[0].TLSCACerts.Pem) == 0 {
		t.Fatalf("Orderer %v must have at least a TlsCACerts.Path or TlsCACerts.Pem set", orderers[0])
	}
}

func TestChannelOrderers(t *testing.T) {
	orderers, err := configImpl.ChannelOrderers("mychannel")
	if orderers == nil || err != nil {
		t.Fatal("Testing ChannelOrderers failed")
	}

	if len(orderers) != 1 {
		t.Fatalf("Expecting one channel orderer got %d", len(orderers))
	}

	if orderers[0].TLSCACerts.Path != "" {
		if !filepath.IsAbs(orderers[0].TLSCACerts.Path) {
			t.Fatal("Expected GOPATH relative path to be replaced")
		}
	} else if len(orderers[0].TLSCACerts.Pem) == 0 {
		t.Fatalf("Orderer %v must have at least a TlsCACerts.Path or TlsCACerts.Pem set", orderers[0])
	}
}

func TestPeersConfig(t *testing.T) {
	pc, err := configImpl.PeersConfig(org0)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for _, value := range pc {
		if value.URL == "" {
			t.Fatalf("Url value for the host is empty")
		}
		if value.EventURL == "" {
			t.Fatalf("EventUrl value is empty")
		}
	}

	pc, err = configImpl.PeersConfig(org1)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for _, value := range pc {
		if value.URL == "" {
			t.Fatalf("Url value for the host is empty")
		}
		if value.EventURL == "" {
			t.Fatalf("EventUrl value is empty")
		}
	}
}

func TestPeerConfig(t *testing.T) {
	pc, err := configImpl.PeerConfig(org1, "peer0.org1.example.com")
	if err != nil {
		t.Fatalf(err.Error())
	}

	if pc.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}

	if pc.TLSCACerts.Path != "" {
		if !filepath.IsAbs(pc.TLSCACerts.Path) {
			t.Fatalf("Expected cert path to be absolute")
		}
	} else if len(pc.TLSCACerts.Pem) == 0 {
		t.Fatalf("Peer %s must have at least a TlsCACerts.Path or TlsCACerts.Pem set", "peer0")
	}
	if len(pc.GRPCOptions) == 0 || pc.GRPCOptions["ssl-target-name-override"] != "peer0.org1.example.com" {
		t.Fatalf("Peer %s must have grpcOptions set in config_test.yaml", "peer0")
	}
}

func TestPeerNotInOrgConfig(t *testing.T) {
	_, err := configImpl.PeerConfig(org1, "peer1.org0.example.com")
	if err == nil {
		t.Fatalf("Fetching peer config not for an unassigned org should fail")
	}
}

func TestInitConfigFromBytesSuccess(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configTestFilePath)

	// test init config from bytes
	_, err = InitConfigFromBytes(cBytes, configType)
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
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
	cBytes := make([]byte, s, s)
	n, err := f.Read(cBytes)
	if err != nil {
		t.Fatalf("Failed to read test config for bytes array testing. Error: %s", err)
	}
	if n == 0 {
		t.Fatalf("Failed to read test config for bytes array testing. Mock bytes array is empty")
	}
	return cBytes, err
}

func TestInitConfigFromBytesEmpty(t *testing.T) {
	// test init config from an empty bytes
	_, err := InitConfigFromBytes([]byte{}, configType)
	if err == nil {
		t.Fatalf("Expected to fail initialize config with empty bytes array.")
	}
	// test init config with an empty configType
	_, err = InitConfigFromBytes([]byte("test config"), "")
	if err == nil {
		t.Fatalf("Expected to fail initialize config with empty config type.")
	}
}

func TestInitConfigSuccess(t *testing.T) {
	//Test init config
	//...Positive case
	_, err := InitConfig(configTestFilePath)
	if err != nil {
		t.Fatalf("Failed to initialize config. Error: %s", err)
	}
}

func TestInitConfigWithCmdRoot(t *testing.T) {
	TestInitConfigSuccess(t)
	fileLoc := configTestFilePath
	cmdRoot := "fabric_sdk"
	var logger = logging.NewLogger("fabric_sdk_go")
	logger.Infof("fileLoc is %s", fileLoc)

	logger.Infof("fileLoc right before calling InitConfigWithCmdRoot is %s", fileLoc)
	config, err := initConfigWithCmdRoot(fileLoc, cmdRoot)
	if err != nil {
		t.Fatalf("Failed to initialize config with cmd root. Error: %s", err)
	}

	//Test if Viper is initialized after calling init config
	if config.configViper.GetString("client.BCCSP.security.hashAlgorithm") != configImpl.SecurityAlgorithm() {
		t.Fatal("Config initialized with incorrect viper configuration")
	}

}

func TestInitConfigPanic(t *testing.T) {

	os.Setenv("FABRIC_SDK_CLIENT_LOGGING_LEVEL", "INVALID")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Init config with cmdroot was supposed to panic")
		} else {
			//Setting it back during panic so as not to fail other tests
			os.Unsetenv("FABRIC_SDK_CLIENT_LOGGING_LEVEL")
		}
	}()

	InitConfig(configTestFilePath)
}

func TestInitConfigInvalidLocation(t *testing.T) {
	//...Negative case
	_, err := InitConfig("invalid file location")
	if err == nil {
		t.Fatalf("Config file initialization is supposed to fail. Error: %s", err)
	}
}

// Test case to create a new viper instance to prevent conflict with existing
// viper instances in applications that use the SDK
func TestMultipleVipers(t *testing.T) {
	viper.SetConfigFile("./test.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		t.Log(err.Error())
	}
	testValue1 := viper.GetString("test.testkey")
	// Read initial value from test.yaml
	if testValue1 != "testvalue" {
		t.Fatalf("Expected testValue before config initialization got: %s", testValue1)
	}
	// initialize go sdk
	config, err := InitConfig(configTestFilePath)
	if err != nil {
		t.Log(err.Error())
	}

	// Make sure initial value is unaffected
	testValue2 := viper.GetString("test.testkey")
	if testValue2 != "testvalue" {
		t.Fatalf("Expected testvalue after config initialization")
	}
	// Make sure Go SDK config is unaffected
	testValue3 := config.configViper.GetBool("client.BCCSP.security.softVerify")
	if testValue3 != true {
		t.Fatalf("Expected existing config value to remain unchanged")
	}
}

func TestEnvironmentVariablesDefaultCmdRoot(t *testing.T) {
	testValue := configImpl.configViper.GetString("env.test")
	if testValue != "" {
		t.Fatalf("Expected environment variable value to be empty but got: %s", testValue)
	}

	err := os.Setenv("FABRIC_SDK_ENV_TEST", "123")
	defer os.Unsetenv("FABRIC_SDK_ENV_TEST")

	if err != nil {
		t.Log(err.Error())
	}

	testValue = configImpl.configViper.GetString("env.test")
	if testValue != "123" {
		t.Fatalf("Expected environment variable value but got: %s", testValue)
	}
}

func TestEnvironmentVariablesSpecificCmdRoot(t *testing.T) {
	testValue := configImpl.configViper.GetString("env.test")
	if testValue != "" {
		t.Fatalf("Expected environment variable value to be empty but got: %s", testValue)
	}

	err := os.Setenv("TEST_ROOT_ENV_TEST", "456")
	defer os.Unsetenv("TEST_ROOT_ENV_TEST")

	if err != nil {
		t.Log(err.Error())
	}

	config, err := initConfigWithCmdRoot(configTestFilePath, "test_root")
	if err != nil {
		t.Log(err.Error())
	}

	testValue = config.configViper.GetString("env.test")
	if testValue != "456" {
		t.Fatalf("Expected environment variable value but got: %s", testValue)
	}
}

func TestNetworkConfig(t *testing.T) {
	conf, err := configImpl.NetworkConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(conf.Orderers) == 0 {
		t.Fatal("Expected orderers to be set")
	}
	if len(conf.Organizations) == 0 {
		t.Fatal("Expected atleast one organisation to be set")
	}
	// viper map keys are lowercase
	if len(conf.Organizations[strings.ToLower(org1)].Peers) == 0 {
		t.Fatalf("Expected org %s to be present in network configuration and peers to be set", org1)
	}
}

func TestMain(m *testing.M) {
	setUp(m)
	r := m.Run()
	teardown()
	os.Exit(r)
}

func setUp(m *testing.M) {
	// do any test setup here...
	var err error
	configImpl, err = InitConfig(configTestFilePath)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func teardown() {
	// do any teadown activities here ..
	configImpl = nil
}

func crossCheckWithViperConfig(expected string, actual string, message string, t *testing.T) {
	expected = substPathVars(expected)
	if actual != expected {
		t.Fatalf(message)
	}
}

func TestInterfaces(t *testing.T) {
	var apiConfig api.Config
	var config Config

	apiConfig = &config
	if apiConfig == nil {
		t.Fatalf("this shouldn't happen. apiConfig should not be nil.")
	}
}

func TestSystemCertPoolDisabled(t *testing.T) {

	// get a config file with pool disabled
	c, err := InitConfig(configTestFilePath)
	if err != nil {
		t.Fatal(err)
	}

	// cert pool should be empty
	if len(c.tlsCertPool.Subjects()) > 0 {
		t.Fatal("Expecting empty tls cert pool due to disabled system cert pool")
	}
}

func TestSystemCertPoolEnabled(t *testing.T) {

	// get a config file with pool enabled
	c, err := InitConfig(configPemTestFilePath)
	if err != nil {
		t.Fatal(err)
	}

	if len(c.tlsCertPool.Subjects()) == 0 {
		t.Fatal("System Cert Pool not loaded even though it is enabled")
	}

	// Org2 'mychannel' peer is missing cert + pem (it should not fail when systemCertPool enabled)
	_, err = c.ChannelPeers("mychannel")
	if err != nil {
		t.Fatalf("Should have skipped verifying ca cert + pem: %s", err)
	}

}

func TestSetTLSCACertPool(t *testing.T) {
	configImpl.SetTLSCACertPool(nil)
	t.Log("TLSCACertRoot must be created. Nothing additional to verify..")
}

func TestInitConfigFromBytesWithPem(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configPemTestFilePath)
	if err != nil {
		t.Fatalf("Failed to load sample bytes from File. Error: %s", err)
	}

	// test init config from bytes
	c, err := InitConfigFromBytes(cBytes, configType)
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}

	o, err := c.OrderersConfig()
	if err != nil {
		t.Fatalf("Failed to load orderers from config. Error: %s", err)
	}

	if o == nil || len(o) == 0 {
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

	pc, err := configImpl.PeersConfig(org1)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if pc == nil || len(pc) == 0 {
		t.Fatalf("peers list of %s cannot be nil or empty", org1)
	}
	peer0 := "peer0.org1.example.com"
	p0, err := c.PeerConfig(org1, peer0)
	if err != nil {
		t.Fatalf("Failed to load %s of %s from the config. Error: %s", peer0, org1, err)
	}
	if p0 == nil {
		t.Fatalf("%s of %s cannot be nil", peer0, org1)
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

	loadedPPem := strings.TrimSpace(p0.TLSCACerts.Pem) // viper's unmarshall adds a \n to the end of a string, hence the TrimeSpace
	if loadedPPem != pPem {
		t.Fatalf("%s Pem doesn't match. Expected \n'%s'\n, but got \n'%s'\n", peer0, pPem, loadedPPem)
	}

	// get CA Server cert pems (embedded) for org1
	certs, err := c.CAServerCertPems("org1")
	if err != nil {
		t.Fatalf("Failed to load CAServerCertPems from config. Error: %s", err)
	}
	if len(certs) == 0 {
		t.Fatalf("Got empty PEM certs for CAServerCertPems")
	}

	// get the client cert pem (embedded) for org1
	c.CAClientCertPem("org1")
	if err != nil {
		t.Fatalf("Failed to load CAClientCertPem from config. Error: %s", err)
	}

	// get CA Server certs paths for org1
	certs, err = c.CAServerCertPaths("org1")
	if err != nil {
		t.Fatalf("Failed to load CAServerCertPaths from config. Error: %s", err)
	}
	if len(certs) == 0 {
		t.Fatalf("Got empty cert file paths for CAServerCertPaths")
	}

	// get the client cert path for org1
	c.CAClientCertPath("org1")
	if err != nil {
		t.Fatalf("Failed to load CAClientCertPath from config. Error: %s", err)
	}

	// get the client key pem (embedded) for org1
	c.CAClientKeyPem("org1")
	if err != nil {
		t.Fatalf("Failed to load CAClientKeyPem from config. Error: %s", err)
	}

	// get the client key file path for org1
	c.CAClientKeyPath("org1")
	if err != nil {
		t.Fatalf("Failed to load CAClientKeyPath from config. Error: %s", err)
	}
}

func TestLoadConfigWithEmbeddedUsersWithPems(t *testing.T) {
	// get a config file with embedded users
	c, err := InitConfig(configEmbeddedUsersTestFilePath)
	if err != nil {
		t.Fatal(err)
	}

	conf, err := c.NetworkConfig()

	if err != nil {
		t.Fatal(err)
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUser")].Cert.Pem == "" {
		t.Fatal("Failed to parse the embedded cert for user EmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUser")].Key.Pem == "" {
		t.Fatal("Failed to parse the embedded key for user EmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("NonExistentEmbeddedUser")].Key.Pem != "" {
		t.Fatal("Mistakenly found an embedded key for user NonExistentEmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("NonExistentEmbeddedUser")].Cert.Pem != "" {
		t.Fatal("Mistakenly found an embedded cert for user NonExistentEmbeddedUser")
	}
}

func TestLoadConfigWithEmbeddedUsersWithPaths(t *testing.T) {
	// get a config file with embedded users
	c, err := InitConfig(configEmbeddedUsersTestFilePath)
	if err != nil {
		t.Fatal(err)
	}

	conf, err := c.NetworkConfig()

	if err != nil {
		t.Fatal(err)
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUserWithPaths")].Cert.Path == "" {
		t.Fatal("Failed to parse the embedded cert for user EmbeddedUserWithPaths")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUserWithPaths")].Key.Path == "" {
		t.Fatal("Failed to parse the embedded key for user EmbeddedUserWithPaths")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("NonExistentEmbeddedUser")].Key.Path != "" {
		t.Fatal("Mistakenly found an embedded key for user NonExistentEmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("NonExistentEmbeddedUser")].Cert.Path != "" {
		t.Fatal("Mistakenly found an embedded cert for user NonExistentEmbeddedUser")
	}
}

func TestInitConfigFromBytesWrongType(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configPemTestFilePath)
	if err != nil {
		t.Fatalf("Failed to load sample bytes from File. Error: %s", err)
	}

	// test init config with empty type
	c, err := InitConfigFromBytes(cBytes, "")
	if err == nil {
		t.Fatalf("Expected error when initializing config with wrong config type but got no error.")
	}

	// test init config with wrong type
	c, err = InitConfigFromBytes(cBytes, "json")
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}

	o, err := c.OrderersConfig()
	if len(o) > 0 {
		t.Fatalf("Expected to get an empty list of orderers for wrong config type")
	}

	np, err := c.NetworkPeers()
	if len(np) > 0 {
		t.Fatalf("Expected to get an empty list of peers for wrong config type")
	}
}

func TestTLSClientCertsFromFiles(t *testing.T) {
	configImpl.networkConfig.Client.TLSCerts.Client.Certfile = "../../test/fixtures/config/mutual_tls/client_sdk_go.pem"
	configImpl.networkConfig.Client.TLSCerts.Client.Keyfile = "../../test/fixtures/config/mutual_tls/client_sdk_go-key.pem"
	configImpl.networkConfig.Client.TLSCerts.Client.CertPem = ""
	configImpl.networkConfig.Client.TLSCerts.Client.KeyPem = ""

	certs, err := configImpl.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertsFromFilesIncorrectPaths(t *testing.T) {
	// incorrect paths to files
	configImpl.networkConfig.Client.TLSCerts.Client.Certfile = "/test/fixtures/config/mutual_tls/client_sdk_go.pem"
	configImpl.networkConfig.Client.TLSCerts.Client.Keyfile = "/test/fixtures/config/mutual_tls/client_sdk_go-key.pem"
	configImpl.networkConfig.Client.TLSCerts.Client.CertPem = ""
	configImpl.networkConfig.Client.TLSCerts.Client.KeyPem = ""

	_, err := configImpl.TLSClientCerts()
	if err == nil {
		t.Fatalf("Expected error but got no errors instead")
	}

	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Fatalf("Expected no such file or directory error")
	}
}

func TestTLSClientCertsFromPem(t *testing.T) {
	configImpl.networkConfig.Client.TLSCerts.Client.Certfile = ""
	configImpl.networkConfig.Client.TLSCerts.Client.Keyfile = ""

	configImpl.networkConfig.Client.TLSCerts.Client.CertPem = `-----BEGIN CERTIFICATE-----
MIIC5TCCAkagAwIBAgIUMYhiY5MS3jEmQ7Fz4X/e1Dx33J0wCgYIKoZIzj0EAwQw
gYwxCzAJBgNVBAYTAkNBMRAwDgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3Jv
bnRvMREwDwYDVQQKEwhsaW51eGN0bDEMMAoGA1UECxMDTGFiMTgwNgYDVQQDEy9s
aW51eGN0bCBFQ0MgUm9vdCBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAoTGFiKTAe
Fw0xNzEyMDEyMTEzMDBaFw0xODEyMDEyMTEzMDBaMGMxCzAJBgNVBAYTAkNBMRAw
DgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3JvbnRvMREwDwYDVQQKEwhsaW51
eGN0bDEMMAoGA1UECxMDTGFiMQ8wDQYDVQQDDAZzZGtfZ28wdjAQBgcqhkjOPQIB
BgUrgQQAIgNiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv2PotEEGuO6rMyaOu
pulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMhYZjcDi7YEOZ3Fs1h
xKmIxR+TTR2vf9KjgZAwgY0wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFDwS3xhpAWs81OVWvZt+iUNL
z26DMB8GA1UdIwQYMBaAFLRasbknomawJKuQGiyKs/RzTCujMBgGA1UdEQQRMA+C
DWZhYnJpY19zZGtfZ28wCgYIKoZIzj0EAwQDgYwAMIGIAkIAk1MxMogtMtNO0rM8
gw2rrxqbW67ulwmMQzp6EJbm/28T2pIoYWWyIwpzrquypI7BOuf8is5b7Jcgn9oz
7sdMTggCQgF7/8ZFl+wikAAPbciIL1I+LyCXKwXosdFL6KMT6/myYjsGNeeDeMbg
3YkZ9DhdH1tN4U/h+YulG/CkKOtUATtQxg==
-----END CERTIFICATE-----`

	configImpl.networkConfig.Client.TLSCerts.Client.KeyPem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

	certs, err := configImpl.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertFromPemAndKeyFromFile(t *testing.T) {
	configImpl.networkConfig.Client.TLSCerts.Client.Certfile = ""
	configImpl.networkConfig.Client.TLSCerts.Client.Keyfile = "../../test/fixtures/config/mutual_tls/client_sdk_go-key.pem"

	configImpl.networkConfig.Client.TLSCerts.Client.CertPem = `-----BEGIN CERTIFICATE-----
MIIC5TCCAkagAwIBAgIUMYhiY5MS3jEmQ7Fz4X/e1Dx33J0wCgYIKoZIzj0EAwQw
gYwxCzAJBgNVBAYTAkNBMRAwDgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3Jv
bnRvMREwDwYDVQQKEwhsaW51eGN0bDEMMAoGA1UECxMDTGFiMTgwNgYDVQQDEy9s
aW51eGN0bCBFQ0MgUm9vdCBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAoTGFiKTAe
Fw0xNzEyMDEyMTEzMDBaFw0xODEyMDEyMTEzMDBaMGMxCzAJBgNVBAYTAkNBMRAw
DgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3JvbnRvMREwDwYDVQQKEwhsaW51
eGN0bDEMMAoGA1UECxMDTGFiMQ8wDQYDVQQDDAZzZGtfZ28wdjAQBgcqhkjOPQIB
BgUrgQQAIgNiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv2PotEEGuO6rMyaOu
pulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMhYZjcDi7YEOZ3Fs1h
xKmIxR+TTR2vf9KjgZAwgY0wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFDwS3xhpAWs81OVWvZt+iUNL
z26DMB8GA1UdIwQYMBaAFLRasbknomawJKuQGiyKs/RzTCujMBgGA1UdEQQRMA+C
DWZhYnJpY19zZGtfZ28wCgYIKoZIzj0EAwQDgYwAMIGIAkIAk1MxMogtMtNO0rM8
gw2rrxqbW67ulwmMQzp6EJbm/28T2pIoYWWyIwpzrquypI7BOuf8is5b7Jcgn9oz
7sdMTggCQgF7/8ZFl+wikAAPbciIL1I+LyCXKwXosdFL6KMT6/myYjsGNeeDeMbg
3YkZ9DhdH1tN4U/h+YulG/CkKOtUATtQxg==
-----END CERTIFICATE-----`

	configImpl.networkConfig.Client.TLSCerts.Client.KeyPem = ""

	certs, err := configImpl.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertFromFileAndKeyFromPem(t *testing.T) {
	configImpl.networkConfig.Client.TLSCerts.Client.Certfile = "../../test/fixtures/config/mutual_tls/client_sdk_go.pem"
	configImpl.networkConfig.Client.TLSCerts.Client.Keyfile = ""

	configImpl.networkConfig.Client.TLSCerts.Client.CertPem = ""

	configImpl.networkConfig.Client.TLSCerts.Client.KeyPem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

	certs, err := configImpl.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertsPemBeforeFiles(t *testing.T) {
	// files have incorrect paths, but pems are loaded first
	configImpl.networkConfig.Client.TLSCerts.Client.Certfile = "/test/fixtures/config/mutual_tls/client_sdk_go.pem"
	configImpl.networkConfig.Client.TLSCerts.Client.Keyfile = "/test/fixtures/config/mutual_tls/client_sdk_go-key.pem"

	configImpl.networkConfig.Client.TLSCerts.Client.CertPem = `-----BEGIN CERTIFICATE-----
MIIC5TCCAkagAwIBAgIUMYhiY5MS3jEmQ7Fz4X/e1Dx33J0wCgYIKoZIzj0EAwQw
gYwxCzAJBgNVBAYTAkNBMRAwDgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3Jv
bnRvMREwDwYDVQQKEwhsaW51eGN0bDEMMAoGA1UECxMDTGFiMTgwNgYDVQQDEy9s
aW51eGN0bCBFQ0MgUm9vdCBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAoTGFiKTAe
Fw0xNzEyMDEyMTEzMDBaFw0xODEyMDEyMTEzMDBaMGMxCzAJBgNVBAYTAkNBMRAw
DgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3JvbnRvMREwDwYDVQQKEwhsaW51
eGN0bDEMMAoGA1UECxMDTGFiMQ8wDQYDVQQDDAZzZGtfZ28wdjAQBgcqhkjOPQIB
BgUrgQQAIgNiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv2PotEEGuO6rMyaOu
pulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMhYZjcDi7YEOZ3Fs1h
xKmIxR+TTR2vf9KjgZAwgY0wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFDwS3xhpAWs81OVWvZt+iUNL
z26DMB8GA1UdIwQYMBaAFLRasbknomawJKuQGiyKs/RzTCujMBgGA1UdEQQRMA+C
DWZhYnJpY19zZGtfZ28wCgYIKoZIzj0EAwQDgYwAMIGIAkIAk1MxMogtMtNO0rM8
gw2rrxqbW67ulwmMQzp6EJbm/28T2pIoYWWyIwpzrquypI7BOuf8is5b7Jcgn9oz
7sdMTggCQgF7/8ZFl+wikAAPbciIL1I+LyCXKwXosdFL6KMT6/myYjsGNeeDeMbg
3YkZ9DhdH1tN4U/h+YulG/CkKOtUATtQxg==
-----END CERTIFICATE-----`

	configImpl.networkConfig.Client.TLSCerts.Client.KeyPem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

	certs, err := configImpl.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertsNoCerts(t *testing.T) {
	configImpl.networkConfig.Client.TLSCerts.Client.Certfile = ""
	configImpl.networkConfig.Client.TLSCerts.Client.Keyfile = ""
	configImpl.networkConfig.Client.TLSCerts.Client.CertPem = ""
	configImpl.networkConfig.Client.TLSCerts.Client.KeyPem = ""

	certs, err := configImpl.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only empty tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if !reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is not equal to empty cert")
	}
}
