/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package fab

import (
	"os"
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

func TestDefaultConfig(t *testing.T) {
	testSetup := &integration.BaseSetupImpl{
		ConfigFile:      "../../../pkg/config/config.yaml", // explicitly set default config.yaml as setup() sets config_test.yaml for all tests
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	c, err := testSetup.InitConfig()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}
	n, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("Failed to load default network config: %v", err)
	}

	if n.Name != "default-network" {
		t.Fatalf("Default network was not loaded. Network name loaded is: %s", n.Name)
	}
}

func TestDefaultConfigFromEnvVariable(t *testing.T) {
	testSetup := &integration.BaseSetupImpl{
		ConfigFile:      "../../../pkg/config/config.yaml", // explicitly set default config.yaml as Setup test sets config_test.yaml for all tests
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}
	// set env variable
	os.Setenv("DEFAULT_SDK_CONFIG_PATH", "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/pkg/config")
	c, err := testSetup.InitConfig()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}
	n, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("Failed to load default network config: %v", err)
	}

	if n.Name != "default-network" {
		t.Fatalf("Default network was not loaded. Network name loaded is: %s", n.Name)
	}
}
