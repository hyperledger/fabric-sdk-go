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
	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
)

var configImpl api.Config
var org0 = "org0"
var org1 = "org1"
var bccspProviderType string

var securityLevel = 256

const (
	providerTypeSW = "SW"
)

var validRootCA = `-----BEGIN CERTIFICATE-----
MIICYjCCAgmgAwIBAgIUB3CTDOU47sUC5K4kn/Caqnh114YwCgYIKoZIzj0EAwIw
fzELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xHzAdBgNVBAoTFkludGVybmV0IFdpZGdldHMsIEluYy4xDDAK
BgNVBAsTA1dXVzEUMBIGA1UEAxMLZXhhbXBsZS5jb20wHhcNMTYxMDEyMTkzMTAw
WhcNMjExMDExMTkzMTAwWjB/MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZv
cm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEfMB0GA1UEChMWSW50ZXJuZXQg
V2lkZ2V0cywgSW5jLjEMMAoGA1UECxMDV1dXMRQwEgYDVQQDEwtleGFtcGxlLmNv
bTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABKIH5b2JaSmqiQXHyqC+cmknICcF
i5AddVjsQizDV6uZ4v6s+PWiJyzfA/rTtMvYAPq/yeEHpBUB1j053mxnpMujYzBh
MA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQXZ0I9
qp6CP8TFHZ9bw5nRtZxIEDAfBgNVHSMEGDAWgBQXZ0I9qp6CP8TFHZ9bw5nRtZxI
EDAKBggqhkjOPQQDAgNHADBEAiAHp5Rbp9Em1G/UmKn8WsCbqDfWecVbZPQj3RK4
oG5kQQIgQAe4OOKYhJdh3f7URaKfGTf492/nmRmtK+ySKjpHSrU=
-----END CERTIFICATE-----`

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
	if vConfig.GetString("client.organization") != "Org1" {
		t.Fatalf("Incorrect Client organization")
	}

	//Test Crypto config path
	crossCheckWithViperConfig(myViper.GetString("client.cryptoconfig.path"), configImpl.CryptoConfigPath(), "Incorrect crypto config path", t)

	//Testing CA Client File Location
	certfile, err := configImpl.CAClientCertFile("org1")

	if certfile == "" || err != nil {
		t.Fatalf("CA Cert file location read failed %s", err)
	}

	//Testing CA Key File Location
	keyFile, err := configImpl.CAClientKeyFile("org1")

	if keyFile == "" || err != nil {
		t.Fatal("CA Key file location read failed")
	}

	//Testing CA Server Cert Files
	sCertFiles, err := configImpl.CAServerCertFiles("org1")

	if sCertFiles == nil || len(sCertFiles) == 0 || err != nil {
		t.Fatal("Getting CA server cert files failed")
	}

	//Testing MSPID
	mspID, err := configImpl.MspID("org1")
	if mspID != "Org1MSP" || err != nil {
		t.Fatal("Get MSP ID failed")
	}

	//Testing CAConfig
	caConfig, err := configImpl.CAConfig("org1")
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
	clientNetworkName := myViper.Get("client")
	peers := myViper.Get("peers")
	organizations := myViper.Get("organizations")
	orderers := myViper.Get("orderers")
	channels := myViper.Get("channels")
	bccspSwProvider := myViper.GetString("client.BCCSP.security.default.provider")
	myViper.Set("client", "INVALID")
	myViper.Set("peers", "INVALID")
	myViper.Set("organizations", "INVALID")
	myViper.Set("orderers", "INVALID")
	myViper.Set("channels", "INVALID")
	//...

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

	// Testing empty BCCSP Software provider
	myViper.Set("client.BCCSP.security.default.provider", "")
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("BCCSP default provider set as empty should panic!")
			}
		}()
		sampleConfig.CSPConfig()
	}()

	// test empty network objects
	myViper.Set("organizations", nil)
	_, err = sampleConfig.NetworkConfig()
	if err == nil {
		t.Fatalf("Organizations were empty, it should return an error")
	}

	//Set it back to valid one, otherwise other tests may fail
	myViper.Set("client.network", clientNetworkName)
	myViper.Set("peers", peers)
	myViper.Set("organizations", organizations)
	myViper.Set("orderers", orderers)
	myViper.Set("channels", channels)
	myViper.Set("client.BCCSP.security.default.provider", bccspSwProvider)
}

func TestTLSACAConfig(t *testing.T) {
	//Test TLSCA Cert Pool (Positive test case)
	certFile, _ := configImpl.CAClientCertFile("org1")
	_, err := configImpl.TLSCACertPool(certFile)
	if err != nil {
		t.Fatalf("TLS CA cert pool fetch failed, reason: %v", err)
	}

	//Test TLSCA Cert Pool (Negative test case)
	_, err = configImpl.TLSCACertPool("some random invalid path")
	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail")
	}

	keyFile, _ := configImpl.CAClientKeyFile("org1")
	_, err = configImpl.TLSCACertPool(keyFile)
	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail when provided with wrong cert file")
	}
}

func TestTimeouts(t *testing.T) {
	myViper.Set("client.connection.timeout.peer.endorser", "2s")
	myViper.Set("client.connection.timeout.peer.eventhub", "2m")
	myViper.Set("client.connection.timeout.peer.eventreg", "2h")
	myViper.Set("client.connection.timeout.orderer", "2ms")

	t1 := configImpl.TimeoutOrDefault(api.Endorser)
	if t1 != time.Second*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t2 := configImpl.TimeoutOrDefault(api.EventHub)
	if t2 != time.Minute*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t2)
	}
	t3 := configImpl.TimeoutOrDefault(api.EventReg)
	if t3 != time.Hour*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t3)
	}
	t4 := configImpl.TimeoutOrDefault(api.Orderer)
	if t4 != time.Millisecond*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t4)
	}
	// Test default
	myViper.Set("client.connection.timeout.orderer", "")
	t5 := configImpl.TimeoutOrDefault(api.Orderer)
	if t5 != time.Second*5 {
		t.Fatalf("Timeout not read correctly. Got: %s", t5)
	}
}

func TestOrdererConfig(t *testing.T) {
	oConfig, err := configImpl.RandomOrdererConfig()

	if oConfig == nil || err != nil {
		t.Fatal("Testing get RandomOrdererConfig failed")
	}

	oConfig, err = configImpl.OrdererConfig("peerorg1")

	if oConfig == nil || err != nil {
		t.Fatal("Testing get OrdererConfig failed")
	}

	orderers, err := configImpl.OrderersConfig()
	if err != nil {
		t.Fatal(err)
	}

	if orderers[0].TlsCACerts.Path != "" {
		if !filepath.IsAbs(orderers[0].TlsCACerts.Path) {
			t.Fatal("Expected GOPATH relative path to be replaced")
		}
	} else if orderers[0].TlsCACerts.Pem == "" {
		t.Fatalf("Orderer %s must have at least a TlsCACerts.Path or TlsCACerts.Pem set", orderers[0])
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
		if value.Url == "" {
			t.Fatalf("Url value for the host is empty")
		}
		if value.EventUrl == "" {
			t.Fatalf("EventUrl value is empty")
		}
	}

	pc, err = configImpl.PeersConfig(org1)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for _, value := range pc {
		if value.Url == "" {
			t.Fatalf("Url value for the host is empty")
		}
		if value.EventUrl == "" {
			t.Fatalf("EventUrl value is empty")
		}
	}
}

func TestPeerConfig(t *testing.T) {
	pc, err := configImpl.PeerConfig(org1, "peer0.org1.example.com")
	if err != nil {
		t.Fatalf(err.Error())
	}

	if pc.Url == "" {
		t.Fatalf("Url value for the host is empty")
	}

	if pc.TlsCACerts.Path != "" {
		if !filepath.IsAbs(pc.TlsCACerts.Path) {
			t.Fatalf("Expected cert path to be absolute")
		}
	} else if pc.TlsCACerts.Pem == "" {
		t.Fatalf("Peer %s must have at least a TlsCACerts.Path or TlsCACerts.Pem set", "peer0")
	}
	if len(pc.GrpcOptions) == 0 || pc.GrpcOptions["ssl-target-name-override"] != "peer0.org1.example.com" {
		t.Fatalf("Peer %s must have grpcOptions set in config_test.yaml", "peer0")
	}
}

func TestPeerNotInOrgConfig(t *testing.T) {
	_, err := configImpl.PeerConfig(org1, "peer1.org0.example.com")
	if err == nil {
		t.Fatalf("Fetching peer config not for an unassigned org should fail")
	}
}

func TestInitConfig(t *testing.T) {
	//Test init config
	//...Positive case
	_, err := InitConfig("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatal("Failed to initialize config")
	}
	//...Negative case
	_, err = InitConfig("invalid file location")
	if err == nil {
		t.Fatal("Config file initialization is supposed to fail")
	}

	//Test init config with cmd root
	cmdRoot := "fabric_sdk"
	_, err = InitConfigWithCmdRoot("../../test/fixtures/config/config_test.yaml", cmdRoot)
	if err != nil {
		t.Fatal("Failed to initialize config with cmd root")
	}

	//Test if Viper is initialized after calling init config
	if myViper.GetString("client.BCCSP.security.hashAlgorithm") != configImpl.SecurityAlgorithm() {
		t.Fatal("Config initialized with incorrect viper configuration")
	}

}

func TestInitConfigPanic(t *testing.T) {
	existingLogLevel := myViper.Get("client.logging.level")
	myViper.Set("client.logging.level", "INVALID")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Init config with cmdroot was supposed to panic")
		} else {
			//Setting it back during panic so as not to fail other tests
			myViper.Set("client.logging.level", existingLogLevel)
		}

	}()
	InitConfigWithCmdRoot("../../test/fixtures/config/config_test.yaml", "fabric-sdk")
}

// Test case to create a new viper instance to prevent conflict with existing
// viper instances in applications that use the SDK
func TestMultipleVipers(t *testing.T) {
	viper.SetConfigFile("./test.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err.Error())
	}
	testValue1 := viper.GetString("test.testkey")
	// Read initial value from test.yaml
	if testValue1 != "testvalue" {
		t.Fatalf("Expected testValue before config initialization got: %s", testValue1)
	}
	// initialize go sdk
	_, err = InitConfig("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}

	// Make sure initial value is unaffected
	testValue2 := viper.GetString("test.testkey")
	if testValue2 != "testvalue" {
		t.Fatalf("Expected testvalue after config initialization")
	}
	// Make sure Go SDK config is unaffected
	testValue3 := myViper.GetBool("client.BCCSP.security.softVerify")
	if testValue3 != true {
		t.Fatalf("Expected existing config value to remain unchanged")
	}
}

func TestEnvironmentVariablesDefaultCmdRoot(t *testing.T) {
	testValue := myViper.GetString("env.test")
	if testValue != "" {
		t.Fatalf("Expected environment variable value to be empty but got: %s", testValue)
	}

	err := os.Setenv("FABRIC_SDK_ENV_TEST", "123")
	defer os.Unsetenv("FABRIC_SDK_ENV_TEST")

	if err != nil {
		fmt.Println(err.Error())
	}

	_, err = InitConfig("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}

	testValue = myViper.GetString("env.test")
	if testValue != "123" {
		t.Fatalf("Expected environment variable value but got: %s", testValue)
	}
}

func TestEnvironmentVariablesSpecificCmdRoot(t *testing.T) {
	testValue := myViper.GetString("env.test")
	if testValue != "" {
		t.Fatalf("Expected environment variable value to be empty but got: %s", testValue)
	}

	err := os.Setenv("TEST_ROOT_ENV_TEST", "456")
	defer os.Unsetenv("TEST_ROOT_ENV_TEST")

	if err != nil {
		fmt.Println(err.Error())
	}

	_, err = InitConfigWithCmdRoot("../../test/fixtures/config/config_test.yaml", "test_root")
	if err != nil {
		fmt.Println(err.Error())
	}

	testValue = myViper.GetString("env.test")
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
	if len(conf.Organizations[org1].Peers) == 0 {
		t.Fatalf("Expected org %s to be present in network configuration and peers to be set", org1)
	}
}

func TestMain(m *testing.M) {
	var err error
	configImpl, err = InitConfig("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}

	os.Exit(m.Run())
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

// Test go-logging concurrency fix: If this test fails, the concurrency fix
// must be applied to go-logging. See: https://gerrit.hyperledger.org/r/12491
// for fix details.
func TestGoLoggingConcurrencyFix(t *testing.T) {
	logger := logging.MustGetLogger("concurrencytest")
	go func() {
		for i := 0; i < 100; i++ {
			logging.SetLevel(logging.Level(logging.DEBUG), "concurrencytest")
		}
	}()
	for i := 0; i < 100; i++ {
		logger.Info("testing")
	}
}

func TestSetTLSCACertPool(t *testing.T) {
	configImpl.SetTLSCACertPool(nil)
	t.Log("TLSCACertRoot must be created. Nothing additional to verify..")
}
