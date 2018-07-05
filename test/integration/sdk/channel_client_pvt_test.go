// +build !prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

// TestPrivateData tests selection of endorsers in the case where the chaincode policy contains a different
// set of MSPs than that of the collection policy. The chaincode policy is defined as (Org1MSP OR Org2MSP) and the
// collection policy is defined as (Org2MSP).
func TestPrivateData(t *testing.T) {
	sdk := mainSDK

	orgsContext := setupMultiOrgContext(t, sdk)
	err := integration.EnsureChannelCreatedAndPeersJoined(t, sdk, orgChannelID, "orgchannel.tx", orgsContext)
	require.NoError(t, err)

	ccVersion := "v0"
	ccPath := "github.com/example_pvt_cc"
	ccPkg, err := packager.NewCCPackage(ccPath, "../../fixtures/testdata")
	require.NoError(t, err)

	coll1 := "collection1"
	ccID := integration.GenerateRandomID()
	collConfig, err := newCollectionConfig(coll1, "OR('Org2MSP.member')", 0, 2, 1000)
	require.NoError(t, err)
	err = integration.InstallAndInstantiateChaincode(orgChannelID, ccPkg, ccPath, ccID, ccVersion, "OR('Org1MSP.member','Org2MSP.member')", orgsContext, collConfig)
	require.NoError(t, err)

	ctxProvider := sdk.ChannelContext(orgChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	chClient, err := channel.New(ctxProvider)
	require.NoError(t, err)

	t.Run("Specified Invocation Chain", func(t *testing.T) {
		response, err := chClient.Execute(
			channel.Request{
				ChaincodeID: ccID,
				Fcn:         "putprivate",
				Args:        [][]byte{[]byte(coll1), []byte("key"), []byte("value")},
				InvocationChain: []*fab.ChaincodeCall{
					{ID: ccID, Collections: []string{coll1}},
				},
			},
			channel.WithRetry(retry.DefaultChannelOpts),
		)
		require.NoError(t, err)
		t.Logf("Got %d response(s)", len(response.Responses))
		require.NotEmptyf(t, response.Responses, "expecting at least one response")
	})

	t.Run("Auto-detect Invocation Chain", func(t *testing.T) {
		response, err := chClient.Execute(
			channel.Request{
				ChaincodeID: ccID,
				Fcn:         "putprivate",
				Args:        [][]byte{[]byte(coll1), []byte("key"), []byte("value")},
			},
			channel.WithRetry(retry.DefaultChannelOpts),
		)
		require.NoError(t, err)
		t.Logf("Got %d response(s)", len(response.Responses))
		require.NotEmptyf(t, response.Responses, "expecting at least one response")
	})
}

// TestPrivateDataWithOrgDown tests selection of endorsers in the case where a chaincode endorsement can succeed with
// none of the peers of a private collection's org being available. The chaincode policy is defined as (Org1MSP OR Org2MSP)
// and the collection policy is defined as (Org2MSP).
func TestPrivateDataWithOrgDown(t *testing.T) {
	sdk := mainSDK

	orgsContext := setupMultiOrgContext(t, sdk)

	// Just join peers in org1 for now
	err := integration.EnsureChannelCreatedAndPeersJoined(t, sdk, orgChannelID, "orgchannel.tx", orgsContext)
	require.NoError(t, err)

	ccVersion := "v0"
	ccPath := "github.com/example_pvt_cc"
	ccPkg, err := packager.NewCCPackage(ccPath, "../../fixtures/testdata")
	require.NoError(t, err)

	coll1 := "collection1"
	ccID := integration.GenerateRandomID()
	collConfig, err := newCollectionConfig(coll1, "OR('Org3MSP.member')", 0, 2, 1000)
	require.NoError(t, err)
	err = integration.InstallAndInstantiateChaincode(orgChannelID, ccPkg, ccPath, ccID, ccVersion, "OR('Org1MSP.member','Org2MSP.member','Org3MSP.member')", orgsContext, collConfig)
	require.NoError(t, err)

	ctxProvider := sdk.ChannelContext(orgChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	chClient, err := channel.New(ctxProvider)
	require.NoError(t, err)

	t.Run("Specified Invocation Chain", func(t *testing.T) {
		_, err := chClient.Execute(
			channel.Request{
				ChaincodeID: ccID,
				Fcn:         "putprivate",
				Args:        [][]byte{[]byte(coll1), []byte("key"), []byte("value")},
				InvocationChain: []*fab.ChaincodeCall{
					{ID: ccID, Collections: []string{coll1}},
				},
			},
			channel.WithRetry(retry.DefaultChannelOpts),
		)
		require.Errorf(t, err, "expecting error due to all Org2MSP peers down")
	})

	t.Run("Automatic Invocation Chain", func(t *testing.T) {
		response, err := chClient.Execute(
			channel.Request{
				ChaincodeID: ccID,
				Fcn:         "putprivate",
				Args:        [][]byte{[]byte(coll1), []byte("key"), []byte("value")},
			},
			channel.WithRetry(retry.DefaultChannelOpts),
		)
		require.NoError(t, err)
		t.Logf("Got %d response(s)", len(response.Responses))
		require.NotEmptyf(t, response.Responses, "expecting at least one response")
	})
}
