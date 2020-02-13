// +build disabled

// TODO fix java integration tests

/*
 Copyright Mioto Yaku All Rights Reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/require"
)

// TestJavaChaincodeInstallInstantiateAndUpgrade tests install java chaincode,
// instantiate java chaincode upgrade java chaincode
func TestJavaChaincodeInstallInstantiateAndUpgrade(t *testing.T) {
	sdk := mainSDK

	orgsContext := setupMultiOrgContext(t, sdk)
	err := integration.EnsureChannelCreatedAndPeersJoined(t, sdk, orgChannelID, "orgchannel.tx", orgsContext)
	require.NoError(t, err)

	ccID := integration.GenerateExampleJavaID(false)

	err = integration.InstallExampleJavaChaincode(orgsContext, ccID)
	require.NoError(t, err)

	err = integration.InstantiateExampleJavaChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')")
	require.NoError(t, err)

	err = integration.UpgradeExampleJavaChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')")
	require.NoError(t, err)

	//prepare context
	org1ChannelClientContext := sdk.ChannelContext(orgChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	//get channel client
	chClient, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// test query
	testExampleCCQuery(t, chClient, "200", ccID, "b")
}

// testExampleCCQuery query examplechaincode
func testExampleCCQuery(t *testing.T, chClient *channel.Client, expected string, ccID, key string) {
	const (
		maxRetries = 10
		retrySleep = 500 * time.Millisecond
	)

	for r := 0; r < 10; r++ {
		response, err := chClient.Query(channel.Request{ChaincodeID: ccID, Fcn: "query", Args: [][]byte{[]byte(key)}},
			channel.WithRetry(retry.DefaultChannelOpts))
		if err == nil {
			actual := string(response.Payload)
			if actual == expected {
				return
			}

			t.Logf("On Attempt [%d / %d]: Response didn't match expected value [%s, %s]", r, maxRetries, actual, expected)
		} else {
			t.Logf("On Attempt [%d / %d]: failed to invoke example cc '%s' with Args:[%+v], error: %+v", r, maxRetries, ccID, integration.ExampleCCQueryArgs(key), err)
			if r < 9 {
				t.Logf("will retry in %v", retrySleep)
			}
		}

		time.Sleep(retrySleep)
	}

	t.Fatal("Exceeded max retries")
}
