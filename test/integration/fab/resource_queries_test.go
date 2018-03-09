/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

func TestChannelQueries(t *testing.T) {
	chaincodeID := integration.GenerateRandomID()
	testSetup, sdk := initializeTests(t, chaincodeID)
	defer sdk.Close()

	// Low level resource
	client, err := getContext(sdk, "Admin", orgName)
	if err != nil {
		t.Fatalf("Failed to get resource: %s", err)
	}

	testQueryChannels(t, client, testSetup.Targets[0])

	testInstalledChaincodes(t, chaincodeID, client, testSetup.Targets[0])

}

func testQueryChannels(t *testing.T, client *context.Client, target fab.ProposalProcessor) {

	// Our target will be primary peer on this channel
	t.Logf("****QueryChannels for %s", target)
	channelQueryResponse, err := resource.QueryChannels(client, target)
	if err != nil {
		t.Fatalf("QueryChannels return error: %v", err)
	}

	for _, channel := range channelQueryResponse.Channels {
		t.Logf("**Channel: %s", channel)
	}

}

func testInstalledChaincodes(t *testing.T, ccID string, client *context.Client, target fab.ProposalProcessor) {

	// Our target will be primary peer on this channel
	t.Logf("****QueryInstalledChaincodes for %s", target)

	chaincodeQueryResponse, err := resource.QueryInstalledChaincodes(client, target)
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
