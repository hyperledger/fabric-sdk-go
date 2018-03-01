/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

func TestResMgmtClientQueries(t *testing.T) {

	testSetup := integration.BaseSetupImpl{
		ConfigFile:      "../" + integration.ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	ccID := integration.GenerateRandomID()
	if err := integration.InstallAndInstantiateExampleCC(testSetup.SDK, fabsdk.WithUser("Admin"), testSetup.OrgID, ccID); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	// Create SDK setup for the integration tests
	sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	// Resource management client
	client, err := sdk.NewClient(fabsdk.WithUser("Admin")).ResourceMgmt()
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Our target for queries will be primary peer on this channel
	target := testSetup.Targets[0]

	testInstalledChaincodes(t, ccID, target, client)

	testQueryChannels(t, testSetup.ChannelID, target, client)

}

func testInstalledChaincodes(t *testing.T, ccID string, target fab.ProposalProcessor, client *resmgmt.Client) {

	chaincodeQueryResponse, err := client.QueryInstalledChaincodes(target)
	if err != nil {
		t.Fatalf("QueryInstalledChaincodes return error: %v", err)
	}

	found := false
	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		t.Logf("**InstalledCC: %s", chaincode)
		if chaincode.Name == ccID {
			found = true
		}
	}

	if !found {
		t.Fatalf("QueryInstalledChaincodes failed to find installed %s chaincode", ccID)
	}
}

func testQueryChannels(t *testing.T, channelID string, target fab.ProposalProcessor, client *resmgmt.Client) {

	channelQueryResponse, err := client.QueryChannels(target)
	if err != nil {
		t.Fatalf("QueryChannels return error: %v", err)
	}

	found := false
	for _, channel := range channelQueryResponse.Channels {
		t.Logf("**Channel: %s", channel)
		if channel.ChannelId == channelID {
			found = true
		}
	}

	if !found {
		t.Fatalf("QueryChannels failed, peer did not join '%s' channel", channelID)
	}

}
