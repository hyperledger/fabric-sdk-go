/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

const (
	pollRetries = 5
)

// TestOrgsEndToEnd creates a channel with two organisations, installs chaincode
// on each of them, and finally invokes a transaction on an org2 peer and queries
// the result from an org1 peer
func TestOrgsEndToEnd(t *testing.T) {

	// Create SDK setup for the integration tests
	sdkOptions := fabapi.Options{
		ConfigFile: "../" + integration.ConfigTestFile,
	}

	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	// Bootstrap network
	initializeFabricClient(t)
	loadOrgUsers(t)
	loadOrgPeers(t)
	loadOrderer(t)
	createTestChannel(t, sdk)
	joinTestChannel(t, sdk)
	installAndInstantiate(t)

	t.Logf("peer0 is %+v, peer1 is %+v", orgTestPeer0, orgTestPeer1)

	// Org1 user connects to 'orgchannel'
	chClientOrg1User, err := sdk.NewChannelClientWithOpts("orgchannel", "User1", &fabapi.ChannelClientOpts{OrgName: "Org1"})
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org1 user: %s", err)
	}

	// Org2 user connects to 'orgchannel'
	chClientOrg2User, err := sdk.NewChannelClientWithOpts("orgchannel", "User1", &fabapi.ChannelClientOpts{OrgName: "Org2"})
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org2 user: %s", err)
	}

	// Org1 user queries initial value on both peers
	initialValue, err := chClientOrg1User.Query(apitxn.QueryRequest{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}

	// Org2 user moves funds on org2 peer
	txOpts := apitxn.ExecuteTxOpts{ProposalProcessors: []apitxn.ProposalProcessor{orgTestPeer1}}
	_, err = chClientOrg2User.ExecuteTxWithOpts(apitxn.ExecuteTxRequest{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, txOpts)
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	// Assert that funds have changed value on org1 peer
	initialInt, _ := strconv.Atoi(string(initialValue))
	var finalInt int
	for i := 0; i < pollRetries; i++ {
		// Query final value on org1 peer
		queryOpts := apitxn.QueryOpts{ProposalProcessors: []apitxn.ProposalProcessor{orgTestPeer0}}
		finalValue, err := chClientOrg1User.QueryWithOpts(apitxn.QueryRequest{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCQueryArgs()}, queryOpts)
		if err != nil {
			t.Fatalf("Failed to query funds after transaction: %s", err)
		}
		// If value has not propogated sleep with exponential backoff
		finalInt, _ = strconv.Atoi(string(finalValue))
		if initialInt+1 != finalInt {
			backoffFactor := math.Pow(2, float64(i))
			time.Sleep(time.Millisecond * 50 * time.Duration(backoffFactor))
		} else {
			break
		}
	}
	if initialInt+1 != finalInt {
		t.Fatalf("Org2 'move funds' transaction result was not propagated to Org1. Expected %d, got: %d",
			(initialInt + 1), finalInt)
	}
}
