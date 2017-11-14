/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	api "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/spf13/viper"
)

var configImpl *Config
var org0 = "org0"
var org1 = "Org1"

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
	vConfig.SetConfigFile("../../test/fixtures/config/config_test.yaml")
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
	certfile, err := configImpl.CAClientCertFile(org1)

	if certfile == "" || err != nil {
		t.Fatalf("CA Cert file location read failed %s", err)
	}

	//Testing CA Key File Location
	keyFile, err := configImpl.CAClientKeyFile(org1)

	if keyFile == "" || err != nil {
		t.Fatal("CA Key file location read failed")
	}

	//Testing CA Server Cert Files
	sCertFiles, err := configImpl.CAServerCertFiles(org1)

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
	sampleConfig, err := InitConfig("../../test/fixtures/config/config_test.yaml")
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
	certfile, err := sampleConfig.CAClientCertFile("peerorg1")
	if certfile != "" || err == nil {
		t.Fatal("CA Cert file location read supposed to fail")
	}

	//Test CA client cert file failure scenario
	keyFile, err := sampleConfig.CAClientKeyFile("peerorg1")
	if keyFile != "" || err == nil {
		t.Fatal("CA Key file location read supposed to fail")
	}

	//Testing CA Server Cert Files failure scenario
	sCertFiles, err := sampleConfig.CAServerCertFiles("peerorg1")
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

	// Testing empty BCCSP Software provider
	sampleConfig.configViper.Set("client.BCCSP.security.default.provider", "")
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("BCCSP default provider set as empty should panic!")
			}
		}()
		sampleConfig.CSPConfig()
	}()

	// test empty network objects
	sampleConfig.configViper.Set("organizations", nil)
	_, err = sampleConfig.NetworkConfig()
	if err == nil {
		t.Fatalf("Organizations were empty, it should return an error")
	}
}

func TestTLSACAConfig(t *testing.T) {
	//Test TLSCA Cert Pool (Positive test case)
	certFile, _ := configImpl.CAClientCertFile(org1)
	_, err := configImpl.TLSCACertPool(certFile)
	if err != nil {
		t.Fatalf("TLS CA cert pool fetch failed, reason: %v", err)
	}

	//Test TLSCA Cert Pool (Negative test case)
	_, err = configImpl.TLSCACertPool("some random invalid path")
	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail")
	}

	keyFile, _ := configImpl.CAClientKeyFile(org1)
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
	} else if orderers[0].TLSCACerts.Pem == "" {
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
	} else if orderers[0].TLSCACerts.Pem == "" {
		t.Fatalf("Orderer %v must have at least a TlsCACerts.Path or TlsCACerts.Pem set", orderers[0])
	}
}

func TestCSPConfig(t *testing.T) {
	cspconfig := configImpl.CSPConfig()

	if cspconfig != nil && cspconfig.ProviderName == "SW" {
		if cspconfig.SwOpts.HashFamily != configImpl.SecurityAlgorithm() {
			t.Fatalf("Incorrect hashfamily found for cspconfig")
		}

		if cspconfig.SwOpts.SecLevel != configImpl.SecurityLevel() {
			t.Fatalf("Incorrect security level found for cspconfig")
		}

		if cspconfig.SwOpts.Ephemeral {
			t.Fatalf("Incorrect Ephemeral found for cspconfig")
		}

		if cspconfig.SwOpts.FileKeystore.KeyStorePath != configImpl.KeyStorePath() {
			t.Fatalf("Incorrect keystore path found for cspconfig")
		}
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
	} else if pc.TLSCACerts.Pem == "" {
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

func TestInitConfigSuccess(t *testing.T) {
	//Test init config
	//...Positive case
	_, err := InitConfig("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Failed to initialize config. Error: %s", err)
	}
}

func TestInitConfigWithCmdRoot(t *testing.T) {
	TestInitConfigSuccess(t)
	fileLoc := "../../test/fixtures/config/config_test.yaml"
	cmdRoot := "fabric_sdk"
	var logger = logging.NewLogger("fabric_sdk_go")
	logger.Infof("fileLoc is %s", fileLoc)

	logger.Infof("fileLoc right before calling InitConfigWithCmdRoot is %s", fileLoc)
	config, err := InitConfigWithCmdRoot(fileLoc, cmdRoot)
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

	InitConfig("../../test/fixtures/config/config_test.yaml")
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
	config, err := InitConfig("../../test/fixtures/config/config_test.yaml")
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

	config, err := InitConfigWithCmdRoot("../../test/fixtures/config/config_test.yaml", "test_root")
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
	configImpl, err = InitConfig("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func teardown() {
	// do any teadown activities here ..
	configImpl = nil
}

func crossCheckWithViperConfig(expected string, actual string, message string, t *testing.T) {
	expected = strings.Replace(expected, "$GOPATH", "", -1)
	if !strings.HasSuffix(actual, expected) {
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

func TestSetTLSCACertPool(t *testing.T) {
	configImpl.SetTLSCACertPool(nil)
	t.Log("TLSCACertRoot must be created. Nothing additional to verify..")
}
