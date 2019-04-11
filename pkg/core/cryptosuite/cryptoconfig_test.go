/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"path/filepath"
	"testing"

	"os"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/stretchr/testify/assert"
)

const configTestFile = "config_test.yaml"
const configEmptyTestFile = "viper-test.yaml"

func TestEmptyTestFile(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", configEmptyTestFile)
	backend, err := config.FromFile(configPath)()
	assert.Nil(t, err, "Failed to read from empty config")

	cryptoConfig := ConfigFromBackend(backend[0]).(*Config)

	// Test for defaults
	assert.Equal(t, true, cryptoConfig.IsSecurityEnabled())
	assert.Equal(t, "SHA2", cryptoConfig.SecurityAlgorithm())
	assert.Equal(t, 256, cryptoConfig.SecurityLevel())
	// Note that we transform to lower case in SecurityProvider()
	assert.Equal(t, "sw", cryptoConfig.SecurityProvider())
	assert.Equal(t, true, cryptoConfig.SoftVerify())
}

func TestCAConfigKeyStorePath(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", configTestFile)
	backend, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend...)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test KeyStore Path
	val, ok := customBackend.Lookup("client.credentialStore.cryptoStore.path")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}

	if filepath.Join(val.(string), "keystore") != cryptoConfig.KeyStorePath() {
		t.Fatal("Incorrect keystore path")
	}
}

func TestCAConfigBCCSPSecurityEnabled(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", configTestFile)
	backend, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend...)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test BCCSP security is enabled
	val, ok := customBackend.Lookup("client.BCCSP.security.enabled")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(bool) != cryptoConfig.IsSecurityEnabled() {
		t.Fatal("Incorrect BCCSP Security enabled flag")
	}
}

func TestCAConfigSecurityAlgorithm(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", configTestFile)
	backend, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend...)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test SecurityAlgorithm
	val, ok := customBackend.Lookup("client.BCCSP.security.hashAlgorithm")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(string) != cryptoConfig.SecurityAlgorithm() {
		t.Fatal("Incorrect BCCSP Security Hash algorithm")
	}
}

func TestCAConfigSecurityLevel(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", configTestFile)
	backend, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend...)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test Security Level
	val, ok := customBackend.Lookup("client.BCCSP.security.level")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(int) != cryptoConfig.SecurityLevel() {
		t.Fatal("Incorrect BCCSP Security Level")
	}
}

func TestCAConfigSecurityProvider(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", configTestFile)
	backend, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend...)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test SecurityProvider provider
	val, ok := customBackend.Lookup("client.BCCSP.security.default.provider")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if !strings.EqualFold(val.(string), cryptoConfig.SecurityProvider()) {
		t.Fatalf("Incorrect BCCSP SecurityProvider provider : %s", cryptoConfig.SecurityProvider())
	}
}

func TestCAConfigSoftVerifyFlag(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", configTestFile)
	backend, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend...)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test SoftVerify flag
	val, ok := customBackend.Lookup("client.BCCSP.security.softVerify")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(bool) != cryptoConfig.SoftVerify() {
		t.Fatal("Incorrect BCCSP Ephemeral flag")
	}
}

func TestCAConfigSecurityProviderPin(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", configTestFile)
	backend, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend...)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test SecurityProviderPin
	val, ok := customBackend.Lookup("client.BCCSP.security.pin")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(string) != cryptoConfig.SecurityProviderPin() {
		t.Fatal("Incorrect BCCSP SecurityProviderPin flag")
	}
}

func TestCAConfigSecurityProviderLabel(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", configTestFile)
	backend, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend...)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test SecurityProviderLabel
	val, ok := customBackend.Lookup("client.BCCSP.security.label")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(string) != cryptoConfig.SecurityProviderLabel() {
		t.Fatal("Incorrect BCCSP SecurityProviderPin flag")
	}
}

func TestCAConfigSecurityProviderCase(t *testing.T) {

	// we expect the following values
	const expectedPkcs11Value = "pkcs11"
	const expectedSwValue = "sw"

	// map key represents what we will input
	providerTestValues := map[string]string{
		// all upper case
		"SW":     expectedSwValue,
		"PKCS11": expectedPkcs11Value,
		// all lower case
		"sw":     expectedSwValue,
		"pkcs11": expectedPkcs11Value,
		// mixed case
		"Sw":     expectedSwValue,
		"Pkcs11": expectedPkcs11Value,
	}

	for inputValue, expectedValue := range providerTestValues {

		// set the input value, overriding what's in file
		os.Setenv("FABRIC_SDK_CLIENT_BCCSP_SECURITY_DEFAULT_PROVIDER", inputValue)

		configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", configTestFile)
		backend, err := config.FromFile(configPath)()
		if err != nil {
			t.Fatal("Failed to get config backend")
		}

		customBackend := getCustomBackend(backend...)

		cryptoConfig := ConfigFromBackend(customBackend).(*Config)

		// expected values should be uppercase
		if expectedValue != cryptoConfig.SecurityProvider() {
			t.Fatalf(
				"Incorrect BCCSP SecurityProvider - input:%s actual:%s, expected:%s",
				inputValue,
				cryptoConfig.SecurityProvider(),
				expectedValue,
			)
		}

	}
}

func TestCryptoConfigWithMultipleBackends(t *testing.T) {
	var backends []core.ConfigBackend
	backendMap := make(map[string]interface{})
	backendMap["client.BCCSP.security.enabled"] = true
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["client.BCCSP.security.hashAlgorithm"] = "SHA2"
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["client.BCCSP.security.default.provider"] = "PKCS11"
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["client.BCCSP.security.level"] = 2
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["client.BCCSP.security.pin"] = "1234"
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["client.credentialStore.cryptoStore.path"] = "/tmp"
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	backendMap = make(map[string]interface{})
	backendMap["client.BCCSP.security.label"] = "TESTLABEL"
	backends = append(backends, &mocks.MockConfigBackend{KeyValueMap: backendMap})

	cryptoConfig := ConfigFromBackend(backends...)

	assert.Equal(t, cryptoConfig.IsSecurityEnabled(), true)
	assert.Equal(t, cryptoConfig.SecurityAlgorithm(), "SHA2")
	assert.Equal(t, cryptoConfig.SecurityProvider(), "pkcs11")
	assert.Equal(t, cryptoConfig.SecurityLevel(), 2)
	assert.Equal(t, cryptoConfig.SecurityProviderPin(), "1234")
	assert.Equal(t, cryptoConfig.KeyStorePath(), "/tmp/keystore")
	assert.Equal(t, cryptoConfig.SecurityProviderLabel(), "TESTLABEL")
}

//getCustomBackend returns custom backend to override config values and to avoid using new config file for test scenarios
func getCustomBackend(configBackend ...core.ConfigBackend) *mocks.MockConfigBackend {
	backendMap := make(map[string]interface{})
	backendMap["client.BCCSP.security.enabled"], _ = configBackend[0].Lookup("client.BCCSP.security.enabled")
	backendMap["client.BCCSP.security.hashAlgorithm"], _ = configBackend[0].Lookup("client.BCCSP.security.hashAlgorithm")
	backendMap["client.BCCSP.security.default.provider"], _ = configBackend[0].Lookup("client.BCCSP.security.default.provider")
	backendMap["client.BCCSP.security.ephemeral"], _ = configBackend[0].Lookup("client.BCCSP.security.ephemeral")
	backendMap["client.BCCSP.security.softVerify"], _ = configBackend[0].Lookup("client.BCCSP.security.softVerify")
	backendMap["client.BCCSP.security.level"] = 2
	backendMap["client.BCCSP.security.pin"] = "1234"
	backendMap["client.credentialStore.cryptoStore.path"] = "/tmp"
	backendMap["client.BCCSP.security.label"] = "TESTLABEL"
	return &mocks.MockConfigBackend{KeyValueMap: backendMap}
}
