// +build disabled

// TODO fix node integration tests

/*
 Copyright Mioto Yaku All Rights Reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

// TestNodeChaincodeInstallInstantiateAndUpgrade tests install node chaincode,
// instantiate node chaincode upgrade node chaincode
func TestNodeChaincodeInstallInstantiateAndUpgrade(t *testing.T) {
	sdk := mainSDK

	orgsContext := setupMultiOrgContext(t, sdk)
	err := integration.EnsureChannelCreatedAndPeersJoined(t, sdk, orgChannelID, "orgchannel.tx", orgsContext)
	require.NoError(t, err)

	ccID := integration.GenerateExampleNodeID(false)

	err = integration.InstallExampleNodeChaincode(orgsContext, ccID)
	require.NoError(t, err)

	err = integration.InstantiateExampleNodeChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')")
	require.NoError(t, err)

	err = integration.UpgradeExampleNodeChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')")
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
