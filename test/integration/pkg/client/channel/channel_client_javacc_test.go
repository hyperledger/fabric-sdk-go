/*
 Copyright Mioto Yaku All Rights Reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"testing"

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

	coll1 := "collection1"
	ccID := integration.GenerateExampleJavaID(true)
	collConfig, err := newCollectionConfig(coll1, "OR('Org1MSP.member','Org2MSP.member')", 0, 2, 1000)
	require.NoError(t, err)

	err = integration.InstallExampleJavaChaincode(orgsContext, ccID)
	require.NoError(t, err)

	err = integration.InstantiateExampleJavaChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')", collConfig)
	require.NoError(t, err)

	err = integration.UpgradeExampleJavaChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')", collConfig)
	require.NoError(t, err)
}
