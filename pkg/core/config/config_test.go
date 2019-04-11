/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var configBackend core.ConfigBackend

const (
	configFile = "config_test.yaml"
	configType = "yaml"
)

var (
	configTestFilePath = filepath.Join("testdata", configFile)
	defaultConfigPath  = filepath.Join("testdata", "template")
)

func TestFromRawSuccess(t *testing.T) {
	// get a config byte for testing
	configPath := filepath.Join("testdata", configFile)
	cBytes, _ := loadConfigBytesFromFile(t, configPath)

	// test init config from bytes
	_, err := FromRaw(cBytes, configType)()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromReaderSuccess(t *testing.T) {
	// get a config byte for testing
	configPath := filepath.Join("testdata", configFile)
	cBytes, _ := loadConfigBytesFromFile(t, configPath)
	buf := bytes.NewBuffer(cBytes)

	// test init config from bytes
	_, err := FromReader(buf, configType)()
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

func TestFromFileEmptyFilename(t *testing.T) {
	_, err := FromFile("")()
	if err == nil {
		t.Fatal("Expected error when passing empty string to FromFile")
	}
}

func TestInitConfigSuccess(t *testing.T) {
	//Test init config
	//...Positive case
	_, err := FromFile(configTestFilePath)()
	if err != nil {
		t.Fatalf("Failed to initialize config. Error: %s", err)
	}
}

func TestInitConfigWithCmdRoot(t *testing.T) {
	TestInitConfigSuccess(t)
	fileLoc := configTestFilePath
	cmdRoot := "fabric_sdk"
	var logger = logging.NewLogger("fabsdk/core")
	logger.Infof("fileLoc is %s", fileLoc)

	logger.Infof("fileLoc right before calling InitConfigWithCmdRoot is %s", fileLoc)

	configBackend1, err := FromFile(fileLoc, WithEnvPrefix(cmdRoot))()
	if err != nil {
		t.Fatalf("Failed to initialize config backend with cmd root. Error: %s", err)
	}

	if len(configBackend1) == 0 {
		t.Fatal("invalid backend")
	}

	config := cryptosuite.ConfigFromBackend(configBackend1...)

	secAlg, ok := configBackend1[0].Lookup("client.BCCSP.security.hashAlgorithm")
	if !ok {
		t.Fatal("supposed to get valid value")
	}
	//Test if Viper is initialized after calling init config
	if secAlg != config.SecurityAlgorithm() {
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

	_, err := FromFile(configTestFilePath)()
	assert.Nil(t, err, "not supposed to get error")
}

func TestInitConfigInvalidLocation(t *testing.T) {
	//...Negative case
	_, err := FromFile("invalid file location")()
	if err == nil {
		t.Fatalf("Config file initialization is supposed to fail. Error: %s", err)
	}
}

// Test case to create a new viper instance to prevent conflict with existing
// viper instances in applications that use the SDK
func TestMultipleVipers(t *testing.T) {
	configPath := filepath.Join("testdata", "viper-test.yaml")
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		t.Log(err)
	}
	testValue1 := viper.GetString("test.testkey")
	// Read initial value from test.yaml
	if testValue1 != "testvalue" {
		t.Fatalf("Expected testValue before config initialization got: %s", testValue1)
	}
	// initialize go sdk
	configBackend1, err := FromFile(configTestFilePath)()
	if err != nil {
		t.Log(err)
	}

	if len(configBackend1) == 0 {
		t.Fatal("invalid backend")
	}

	// Make sure initial value is unaffected
	testValue2 := viper.GetString("test.testkey")
	if testValue2 != "testvalue" {
		t.Fatal("Expected testvalue after config initialization")
	}
	// Make sure Go SDK config is unaffected
	testValue3, ok := configBackend1[0].Lookup("client.BCCSP.security.softVerify")
	if !ok {
		t.Fatal("Expected valid value")
	}
	if testValue3 != true {
		t.Fatal("Expected existing config value to remain unchanged")
	}
}

func TestEnvironmentVariablesDefaultCmdRoot(t *testing.T) {
	testValue, _ := configBackend.Lookup("env.test")
	if testValue != nil {
		t.Fatalf("Expected environment variable value to be empty but got: %s", testValue)
	}

	err := os.Setenv("FABRIC_SDK_ENV_TEST", "123")
	defer os.Unsetenv("FABRIC_SDK_ENV_TEST")

	if err != nil {
		t.Log(err)
	}

	testValue, ok := configBackend.Lookup("env.test")
	if testValue != "123" || !ok {
		t.Fatalf("Expected environment variable value but got: %s", testValue)
	}
}

func TestEnvironmentVariablesSpecificCmdRoot(t *testing.T) {
	testValue, _ := configBackend.Lookup("env.test")
	if testValue != nil {
		t.Fatalf("Expected environment variable value to be empty but got: %s", testValue)
	}

	err := os.Setenv("TEST_ROOT_ENV_TEST", "456")
	defer os.Unsetenv("TEST_ROOT_ENV_TEST")

	if err != nil {
		t.Log(err)
	}

	configBackend1, err := FromFile(configTestFilePath, WithEnvPrefix("test_root"))()
	if err != nil {
		t.Log(err)
	}

	if len(configBackend1) == 0 {
		t.Fatal("invalid backend")
	}

	value, ok := configBackend1[0].Lookup("env.test")
	if value != "456" || !ok {
		t.Fatalf("Expected environment variable value but got: %s", testValue)
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
	cfgBackend, err := FromFile(configTestFilePath)()
	if err != nil {
		test.Logf(err.Error())
	}
	if len(cfgBackend) != 1 {
		panic("invalid backend found")
	}
	configBackend = cfgBackend[0]
}

func teardown() {
	// do any teadown activities here ..
	configBackend = nil
}

func TestNewGoodOpt(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, configFile)
	_, err := FromFile(configPath, goodOpt())()
	if err != nil {
		t.Fatalf("Expected no error from New, but got %s", err)
	}

	cBytes, err := loadConfigBytesFromFile(t, configTestFilePath)
	if err != nil || len(cBytes) == 0 {
		t.Fatal("Unexpected error from loadConfigBytesFromFile")
	}

	buf := bytes.NewBuffer(cBytes)

	_, err = FromReader(buf, configType, goodOpt())()
	if err != nil {
		t.Fatalf("Unexpected error from FromReader: %s", err)
	}

	_, err = FromRaw(cBytes, configType, goodOpt())()
	if err != nil {
		t.Fatalf("Unexpected error from FromRaw %s", err)
	}

	err = os.Setenv("FABRIC_SDK_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	/*
		_, err = FromDefaultPath(goodOpt())
		if err != nil {
			t.Fatalf("Unexpected error from FromRaw: %s", err)
		}
	*/
}

func goodOpt() Option {
	return func(opts *options) error {
		return nil
	}
}

func TestNewBadOpt(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, configFile)
	_, err := FromFile(configPath, badOpt())()
	if err == nil {
		t.Fatal("Expected error from FromFile")
	}

	cBytes, err := loadConfigBytesFromFile(t, configTestFilePath)
	if err != nil || len(cBytes) == 0 {
		t.Fatal("Unexpected error from loadConfigBytesFromFile")
	}

	buf := bytes.NewBuffer(cBytes)

	_, err = FromReader(buf, configType, badOpt())()
	if err == nil {
		t.Fatal("Expected error from FromReader")
	}

	_, err = FromRaw(cBytes, configType, badOpt())()
	if err == nil {
		t.Fatal("Expected error from FromRaw")
	}

	err = os.Setenv("FABRIC_SDK_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	/*
		_, err = FromDefaultPath(badOpt())
		if err == nil {
			t.Fatalf("Expected error from FromRaw")
		}
	*/
}

func badOpt() Option {
	return func(opts *options) error {
		return errors.New("Bad Opt")
	}
}

func TestConfigBackend_Lookup(t *testing.T) {

	checkConfigStringKey(t, "name", "global-trade-network", true)

	checkConfigStringKey(t, "description", "", false)

	checkConfigStringKey(t, "x-type", "h1fv1", true)

	checkConfigMapKey(t, "channels.mychannel.peers", 1)

	checkConfigMapKey(t, "organizations", 3)
}

func checkConfigStringKey(t *testing.T, configKey string, expectedValue string, checkNotEquals bool) {
	value, ok := configBackend.Lookup(configKey)
	if !ok {
		t.Fatalf("can't lookup key %s in the config", configKey)
	}
	v := value.(string)
	if v != expectedValue && checkNotEquals {
		t.Fatalf("Expected %s to be '%s' but got '%s'", configKey, expectedValue, v)
	}
}

func checkConfigMapKey(t *testing.T, configKey string, expectedNumItems int) {
	value, ok := configBackend.Lookup(configKey)
	if !ok {
		t.Fatalf("can't lookup key %s in the config", configKey)
	}
	v := value.(map[string]interface{})
	if len(v) != expectedNumItems {
		t.Fatalf("Expected only %d %s but got %d", expectedNumItems, configKey, len(v))
	}
}

/*
func TestDefaultConfigFromFile(t *testing.T) {
	c, err := FromFile(configEmptyTestFilePath, WithTemplatePath(defaultConfigPath))

	if err != nil {
		t.Fatalf("Unexpected error from FromFile: %s", err)
	}

	n, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("Failed to load default network config: %s", err)
	}

	if n.Name != "default-network" {
		t.Fatalf("Default network was not loaded. Network name loaded is: %s", n.Name)
	}

	if n.Description != "hello" {
		t.Fatalf("Incorrect Network name from default config. Got %s", n.Description)
	}
}

func TestDefaultConfigFromRaw(t *testing.T) {
	cBytes, err := loadConfigBytesFromFile(t, configEmptyTestFilePath)
	c, err := FromRaw(cBytes, configType, WithTemplatePath(defaultConfigPath))

	if err != nil {
		t.Fatalf("Unexpected error from FromFile: %s", err)
	}

	n, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("Failed to load default network config: %s", err)
	}

	if n.Name != "default-network" {
		t.Fatalf("Default network was not loaded. Network name loaded is: %s", n.Name)
	}

	if n.Description != "hello" {
		t.Fatalf("Incorrect Network name from default config. Got %s", n.Description)
	}
}
*/

/*
func TestFromDefaultPathSuccess(t *testing.T) {
	err := os.Setenv("FABRIC_SDK_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	// test init config from bytes
	_, err = FromDefaultPath()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromDefaultPathCustomPrefixSuccess(t *testing.T) {
	err := os.Setenv("FABRIC_SDK2_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK2_CONFIG_PATH")

	// test init config from bytes
	_, err = FromDefaultPath(WithEnvPrefix("FABRIC_SDK2"))
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromDefaultPathCustomPathSuccess(t *testing.T) {
	err := os.Setenv("FABRIC_SDK2_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK2_CONFIG_PATH")

	// test init config from bytes
	_, err = FromDefaultPath(WithTemplatePath(defaultConfigPath))
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromDefaultPathEmptyFailure(t *testing.T) {
	os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	// test init config from bytes
	_, err := FromDefaultPath()
	if err == nil {
		t.Fatal("Expected failure from unset FABRIC_SDK_CONFIG_PATH")
	}
}

func TestFromDefaultPathFailure(t *testing.T) {
	err := os.Setenv("FABRIC_SDK_CONFIG_PATH", defaultConfigPath+"/bad")
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	// test init config from bytes
	_, err = FromDefaultPath()
	if err == nil {
		t.Fatal("Expected failure from bad default path")
	}
}
*/
