/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
)

const configTestFilePath = "../config/testdata/config_test.yaml"

func TestCAConfigKeyStorePath(t *testing.T) {
	backend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test KeyStore Path
	val, ok := customBackend.Lookup("client.credentialStore.cryptoStore.path")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}

	if path.Join(val.(string), "keystore") != cryptoConfig.KeyStorePath() {
		t.Fatalf("Incorrect keystore path ")
	}
}

func TestCAConfigBCCSPSecurityEnabled(t *testing.T) {
	backend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test BCCSP security is enabled
	val, ok := customBackend.Lookup("client.BCCSP.security.enabled")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(bool) != cryptoConfig.IsSecurityEnabled() {
		t.Fatalf("Incorrect BCCSP Security enabled flag")
	}
}

func TestCAConfigSecurityAlgorithm(t *testing.T) {
	backend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test SecurityAlgorithm
	val, ok := customBackend.Lookup("client.BCCSP.security.hashAlgorithm")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(string) != cryptoConfig.SecurityAlgorithm() {
		t.Fatalf("Incorrect BCCSP Security Hash algorithm")
	}
}

func TestCAConfigSecurityLevel(t *testing.T) {
	backend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test Security Level
	val, ok := customBackend.Lookup("client.BCCSP.security.level")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(int) != cryptoConfig.SecurityLevel() {
		t.Fatalf("Incorrect BCCSP Security Level")
	}
}

func TestCAConfigSecurityProvider(t *testing.T) {
	backend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test SecurityProvider provider
	val, ok := customBackend.Lookup("client.BCCSP.security.default.provider")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(string) != cryptoConfig.SecurityProvider() {
		t.Fatalf("Incorrect BCCSP SecurityProvider provider")
	}
}

func TestCAConfigEphemeralFlag(t *testing.T) {
	backend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test Ephemeral flag
	val, ok := customBackend.Lookup("client.BCCSP.security.ephemeral")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(bool) != cryptoConfig.Ephemeral() {
		t.Fatalf("Incorrect BCCSP Ephemeral flag")
	}
}

func TestCAConfigSoftVerifyFlag(t *testing.T) {
	backend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test SoftVerify flag
	val, ok := customBackend.Lookup("client.BCCSP.security.softVerify")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(bool) != cryptoConfig.SoftVerify() {
		t.Fatalf("Incorrect BCCSP Ephemeral flag")
	}
}

func TestCAConfigSecurityProviderPin(t *testing.T) {
	backend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test SecurityProviderPin
	val, ok := customBackend.Lookup("client.BCCSP.security.pin")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(string) != cryptoConfig.SecurityProviderPin() {
		t.Fatalf("Incorrect BCCSP SecurityProviderPin flag")
	}
}

func TestCAConfigSecurityProviderLabel(t *testing.T) {
	backend, err := config.FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal("Failed to get config backend")
	}

	customBackend := getCustomBackend(backend)

	cryptoConfig := ConfigFromBackend(customBackend).(*Config)

	// Test SecurityProviderLabel
	val, ok := customBackend.Lookup("client.BCCSP.security.label")
	if !ok || val == nil {
		t.Fatal("expected valid value")
	}
	if val.(string) != cryptoConfig.SecurityProviderLabel() {
		t.Fatalf("Incorrect BCCSP SecurityProviderPin flag")
	}
}

//getCustomBackend returns custom backend to override config values and to avoid using new config file for test scenarios
func getCustomBackend(configBackend core.ConfigBackend) *mocks.MockConfigBackend {
	backendMap := make(map[string]interface{})
	backendMap["client.BCCSP.security.enabled"], _ = configBackend.Lookup("client.BCCSP.security.enabled")
	backendMap["client.BCCSP.security.hashAlgorithm"], _ = configBackend.Lookup("client.BCCSP.security.hashAlgorithm")
	backendMap["client.BCCSP.security.default.provider"], _ = configBackend.Lookup("client.BCCSP.security.default.provider")
	backendMap["client.BCCSP.security.ephemeral"], _ = configBackend.Lookup("client.BCCSP.security.ephemeral")
	backendMap["client.BCCSP.security.softVerify"], _ = configBackend.Lookup("client.BCCSP.security.softVerify")
	backendMap["client.BCCSP.security.level"] = 2
	backendMap["client.BCCSP.security.pin"] = "1234"
	backendMap["client.credentialStore.cryptoStore.path"] = "/tmp"
	backendMap["client.BCCSP.security.label"] = "TESTLABEL"
	return &mocks.MockConfigBackend{KeyValueMap: backendMap}
}
