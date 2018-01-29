/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

func TestChannelConfig(t *testing.T) {

	testSetup := integration.BaseSetupImpl{
		ConfigFile:      "../" + integration.ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(t); err != nil {
		t.Fatalf(err.Error())
	}

	// Create SDK setup for the integration tests
	sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	cs, err := sdk.NewClient(fabsdk.WithUser("User1")).ChannelService(testSetup.ChannelID)
	if err != nil {
		t.Fatalf("Failed to create new channel service: %s", err)
	}

	cfg, err := cs.ChannelConfig()
	if err != nil {
		t.Fatalf("Failed to create new channel config: %s", err)
	}

	response, err := cfg.Query()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := "orderer.example.com:7050"
	found := false
	for _, o := range response.Orderers {
		if o == expected {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("Expected orderer %s, got %s", expected, response.Orderers)
	}

}
