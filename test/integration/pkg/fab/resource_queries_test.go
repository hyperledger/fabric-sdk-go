/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	reqContext "context"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/stretchr/testify/require"
)

func TestChannelQueries(t *testing.T) {
	// Using shared SDK instance to increase test speed.
	sdk := mainSDK
	testSetup := mainTestSetup
	chaincodeID := mainChaincodeID

	//chaincodeID := integration.GenerateRandomID()

	// Low level resource
	reqCtx, cancel, err := getContext(sdk, "Admin", org1Name)
	if err != nil {
		t.Fatalf("Failed to get resource: %s", err)
	}
	defer cancel()

	peers, err := getProposalProcessors(sdk, "Admin", testSetup.OrgID, testSetup.Targets[:1])
	require.Nil(t, err, "creating peers failed")

	testQueryChannels(t, reqCtx, peers[0])

	testInstalledChaincodes(t, reqCtx, chaincodeID, peers[0])

}

func testQueryChannels(t *testing.T, reqCtx reqContext.Context, target fab.ProposalProcessor) {

	// Our target will be primary peer on this channel
	t.Logf("****QueryChannels for %s", target)
	channelQueryResponse, err := resource.QueryChannels(reqCtx, target, resource.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatalf("QueryChannels return error: %s", err)
	}

	for _, channel := range channelQueryResponse.Channels {
		t.Logf("**Channel: %s", channel)
	}

}

func testInstalledChaincodes(t *testing.T, reqCtx reqContext.Context, ccID string, target fab.ProposalProcessor) {

	// Our target will be primary peer on this channel
	t.Logf("****QueryInstalledChaincodes for %s", target)

	chaincodeQueryResponse, err := resource.QueryInstalledChaincodes(reqCtx, target, resource.WithRetry(retry.DefaultResMgmtOpts))
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
