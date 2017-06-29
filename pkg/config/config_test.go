/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"fmt"
	"os"
	"strings"
	"testing"

	api "github.com/hyperledger/fabric-sdk-go/api"
	"github.com/spf13/viper"
)

var configImpl api.Config
var org1 = "peerorg1"

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

	//Test TLS config
	if vConfig.GetBool("client.tls.enabled") != configImpl.IsTLSEnabled() {
		t.Fatalf("Incorrect TLS config flag")
	}

	//Test Security enabled
	if vConfig.GetBool("client.security.enabled") != configImpl.IsSecurityEnabled() {
		t.Fatalf("Incorrect Security config flag")
	}

	//Test Tcert batch size
	if vConfig.GetInt("client.tcert.batch.size") != configImpl.TcertBatchSize() {
		t.Fatalf("Incorrect Tcert batch size")
	}

	//Test Security Algorithm
	if vConfig.GetString("client.security.hashAlgorithm") != configImpl.SecurityAlgorithm() {
		t.Fatalf("Incorrect security hash algorithm")
	}

	//Test Security level
	if vConfig.GetInt("client.security.level") != configImpl.SecurityLevel() {
		t.Fatalf("Incorrect Security Level")
	}

	//Test Crypto config path
	crossCheckWithViperConfig(myViper.GetString("client.cryptoconfig.path"), configImpl.CryptoConfigPath(), "Incorrect crypto config path", t)

}

func TestTLSACAConfig(t *testing.T) {

	//Test TLSCA Cert Pool (Negative test case
	_, err := configImpl.TLSCACertPool("some random invalid path")
	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail")
	}

	//Test TLSCA Cert Pool from roots(Negative test case
	samplebytes := [][]byte{
		[]byte(validRootCA),
	}
	_, err = configImpl.TLSCACertPoolFromRoots(samplebytes)
	if err != nil {
		t.Fatalf("TLS CA cert pool fetch failed, reason %v", err)
	}

	//Test TLSCA Cert Pool from roots(Negative test case
	samplebytes = [][]byte{
		[]byte("sample invalid string"),
	}
	_, err = configImpl.TLSCACertPoolFromRoots(samplebytes)
	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail")
	}

}

func TestCSPConfig(t *testing.T) {
	cspconfig := configImpl.CSPConfig()

	if cspconfig.ProviderName != "SW" {
		t.Fatalf("In correct provider name found for cspconfig")
	}

	if cspconfig.SwOpts.HashFamily != configImpl.SecurityAlgorithm() {
		t.Fatalf("In correct hashfamily found for cspconfig")
	}

	if cspconfig.SwOpts.SecLevel != configImpl.SecurityLevel() {
		t.Fatalf("In correct security level found for cspconfig")
	}

	if cspconfig.SwOpts.Ephemeral {
		t.Fatalf("In correct Ephemeral found for cspconfig")
	}

	if cspconfig.SwOpts.FileKeystore.KeyStorePath != configImpl.KeyStorePath() {
		t.Fatalf("In correct keystore path found for cspconfig")
	}
}

func TestGetPeersConfig(t *testing.T) {
	pc, err := configImpl.PeersConfig(org1)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for _, value := range pc {
		if value.Host == "" {
			t.Fatalf("Host value is empty")
		}
		if value.Port == 0 {
			t.Fatalf("Port value is empty")
		}
		if value.EventHost == "" {
			t.Fatalf("EventHost value is empty")
		}
		if value.EventPort == 0 {
			t.Fatalf("EventPort value is empty")
		}

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
	myViper := configImpl.FabricClientViper()

	if myViper.GetString("client.security.hashAlgorithm") != configImpl.SecurityAlgorithm() {
		t.Fatal("Config initialized with incorrect viper configuration")
	}

}

func TestInitConfigPanic(t *testing.T) {
	myViper := configImpl.FabricClientViper()
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
	testValue1 := viper.GetString("test.testKey")
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
	testValue2 := viper.GetString("test.testKey")
	if testValue2 != "testvalue" {
		t.Fatalf("Expected testvalue after config initialization")
	}
	// Make sure Go SDK config is unaffected
	testValue3 := myViper.GetBool("client.security.enabled")
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
