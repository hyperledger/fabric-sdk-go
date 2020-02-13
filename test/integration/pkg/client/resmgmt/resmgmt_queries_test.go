/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestResMgmtClientQueries(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	sdk := mainSDK
	testSetup := mainTestSetup
	chaincodeID := mainChaincodeID

	//prepare contexts
	org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))

	// Resource management client
	client, err := resmgmt.New(org1AdminClientContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Our target for queries will be primary peer on this channel
	target := testSetup.Targets[0]

	testQueryConfigFromOrderer(t, testSetup.ChannelID, client)

	testInstalledChaincodes(t, chaincodeID, target, client)

	testInstantiatedChaincodes(t, testSetup.ChannelID, chaincodeID, target, client)

	testQueryChannels(t, testSetup.ChannelID, target, client)

	// TODO java and node integration tests need to be fixed.
	/*
	// test java chaincode installed and instantiated
	javaCCID := integration.GenerateExampleJavaID(false)

	testInstalledChaincodes(t, javaCCID, target, client)

	testInstantiatedChaincodes(t, orgChannelID, javaCCID, target, client)

	// test node chaincode installed and instantiated
	nodeCCID := integration.GenerateExampleNodeID(false)

	testInstalledChaincodes(t, nodeCCID, target, client)

	testInstantiatedChaincodes(t, orgChannelID, nodeCCID, target, client)

	*/
}

func testInstantiatedChaincodes(t *testing.T, channelID string, ccID string, target string, client *resmgmt.Client) {

	chaincodeQueryResponse, err := client.QueryInstantiatedChaincodes(channelID, resmgmt.WithTargetEndpoints(target), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatalf("QueryInstantiatedChaincodes return error: %s", err)
	}

	found := false
	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		t.Logf("**InstantiatedCC: %s", chaincode)
		if chaincode.Name == ccID {
			found = true
		}
	}

	if !found {
		t.Fatalf("QueryInstantiatedChaincodes failed to find instantiated %s chaincode", ccID)
	}
}

func testInstalledChaincodes(t *testing.T, ccID string, target string, client *resmgmt.Client) {

	chaincodeQueryResponse, err := client.QueryInstalledChaincodes(resmgmt.WithTargetEndpoints(target), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatalf("QueryInstalledChaincodes return error: %s", err)
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

func testQueryChannels(t *testing.T, channelID string, target string, client *resmgmt.Client) {

	channelQueryResponse, err := client.QueryChannels(resmgmt.WithTargetEndpoints(target), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatalf("QueryChannels return error: %s", err)
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

func testQueryConfigFromOrderer(t *testing.T, channelID string, client *resmgmt.Client) {
	expected := "orderer.example.com:7050"
	channelCfg, err := client.QueryConfigFromOrderer(channelID, resmgmt.WithOrdererEndpoint("orderer.example.com"), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatalf("QueryConfig return error: %s", err)
	}
	if !contains(channelCfg.Orderers(), expected) {
		t.Fatalf("Expected orderer %s, got %s", expected, channelCfg.Orderers())
	}
	block, err := client.QueryConfigBlockFromOrderer(channelID, resmgmt.WithOrdererEndpoint("orderer.example.com"))
	if err != nil {
		t.Fatalf("QueryConfigBlockFromOrderer return error: %s", err)
	}
	if block.Header.Number != channelCfg.BlockNumber() {
		t.Fatalf("QueryConfigBlockFromOrderer returned invalid block number: [%d, %d]", block.Header.Number, channelCfg.BlockNumber())
	}

	_, err = client.QueryConfigFromOrderer(channelID, resmgmt.WithOrdererEndpoint("non-existent"), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err == nil {
		t.Fatal("QueryConfig should have failed for invalid orderer")
	}

	_, err = client.QueryConfigBlockFromOrderer(channelID, resmgmt.WithOrdererEndpoint("non-existent"), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err == nil {
		t.Fatal("QueryConfigBlockFromOrderer should have failed for invalid orderer")
	}

}

func contains(list []string, value string) bool {
	for _, e := range list {
		if e == value {
			return true
		}
	}
	return false
}
