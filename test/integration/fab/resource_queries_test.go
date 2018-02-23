/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/resource/api"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

func TestChannelQueries(t *testing.T) {
	chaincodeID := integration.GenerateRandomID()
	testSetup := initializeTests(t, chaincodeID)

	testQueryChannels(t, testSetup.Client, testSetup.Targets[0])

	testInstalledChaincodes(t, chaincodeID, testSetup.Client, testSetup.Targets[0])

}

func testQueryChannels(t *testing.T, client api.Resource, target fab.ProposalProcessor) {

	// Our target will be primary peer on this channel
	t.Logf("****QueryChannels for %s", target)
	channelQueryResponse, err := client.QueryChannels(target)
	if err != nil {
		t.Fatalf("QueryChannels return error: %v", err)
	}

	for _, channel := range channelQueryResponse.Channels {
		t.Logf("**Channel: %s", channel)
	}

}

func testInstalledChaincodes(t *testing.T, ccID string, client api.Resource, target fab.ProposalProcessor) {

	// Our target will be primary peer on this channel
	t.Logf("****QueryInstalledChaincodes for %s", target)

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
