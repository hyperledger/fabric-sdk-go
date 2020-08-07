// +build !prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

// TestPrivateDataPutAndGet tests put and get for private data
func TestPrivateDataPutAndGet(t *testing.T) {
	sdk := mainSDK

	orgsContext := setupMultiOrgContext(t, sdk)
	err := integration.EnsureChannelCreatedAndPeersJoined(t, sdk, orgChannelID, "orgchannel.tx", orgsContext)
	require.NoError(t, err)

	coll1 := "collection1"
	ccID := integration.GenerateExamplePvtID(true)
	collConfig, err := newCollectionConfig(coll1, "OR('Org1MSP.member','Org2MSP.member')", 0, 2, 1000)
	require.NoError(t, err)

	if metadata.CCMode == "lscc" {
		err = integration.InstallExamplePvtChaincode(orgsContext, ccID)
		require.NoError(t, err)
		err = integration.InstantiateExamplePvtChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')", collConfig)
		require.NoError(t, err)
	} else {
		err := integration.InstantiatePvtExampleChaincodeLc(sdk, orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')", collConfig)
		require.NoError(t, err)
	}

	ctxProvider := sdk.ChannelContext(orgChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	chClient, err := channel.New(ctxProvider)
	require.NoError(t, err)

	key1 := "key1"
	key2 := "key2"
	key3 := "key3"
	value1 := "pvtValue1"
	value2 := "pvtValue2"
	value3 := "pvtValue3"

	response, err := chClient.Query(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "getprivate",
			Args:        [][]byte{[]byte(coll1), []byte(key1)},
		},
		channel.WithRetry(retry.DefaultChannelOpts),
	)
	require.NoError(t, err)
	t.Logf("Got response payload: [%s]", string(response.Payload))
	require.Nil(t, response.Payload)

	response, err = chClient.Query(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "getprivatebyrange",
			Args:        [][]byte{[]byte(coll1), []byte(key1), []byte(key3)},
		},
		channel.WithRetry(retry.DefaultChannelOpts),
	)
	require.NoError(t, err)
	t.Logf("Got response payload: [%s]", string(response.Payload))
	require.Empty(t, string(response.Payload))

	response, err = chClient.Execute(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "putprivate",
			Args:        [][]byte{[]byte(coll1), []byte(key1), []byte(value1)},
		},
		channel.WithRetry(retry.DefaultChannelOpts),
	)
	require.NoError(t, err)
	require.NotEmptyf(t, response.Responses, "expecting at least one response")

	response, err = chClient.Execute(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "putprivate",
			Args:        [][]byte{[]byte(coll1), []byte(key2), []byte(value2)},
		},
		channel.WithRetry(retry.DefaultChannelOpts),
	)
	require.NoError(t, err)
	require.NotEmptyf(t, response.Responses, "expecting at least one response")

	response, err = chClient.Execute(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "putprivate",
			Args:        [][]byte{[]byte(coll1), []byte(key3), []byte(value3)},
		},
		channel.WithRetry(retry.TestRetryOpts),
	)
	require.NoError(t, err)
	require.NotEmptyf(t, response.Responses, "expecting at least one response")

	response, err = chClient.Query(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "getprivate",
			Args:        [][]byte{[]byte(coll1), []byte(key1)},
		},
		channel.WithRetry(retry.TestRetryOpts),
	)
	require.NoError(t, err)
	t.Logf("Got response payload: %s", string(response.Payload))
	require.Equal(t, value1, string(response.Payload))

	response, err = chClient.Query(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "getprivatebyrange",
			Args:        [][]byte{[]byte(coll1), []byte(key1), []byte(key3)},
		},
		channel.WithRetry(retry.DefaultChannelOpts),
	)
	require.NoError(t, err)
	t.Logf("Got response payload: [%s]", string(response.Payload))
	require.NotEmpty(t, string(response.Payload))
}

// TestPrivateData tests selection of endorsers in the case where the chaincode policy contains a different
// set of MSPs than that of the collection policy. The chaincode policy is defined as (Org1MSP OR Org2MSP) and the
// collection policy is defined as (Org2MSP).
func TestPrivateData(t *testing.T) {
	sdk := mainSDK

	orgsContext := setupMultiOrgContext(t, sdk)
	err := integration.EnsureChannelCreatedAndPeersJoined(t, sdk, orgChannelID, "orgchannel.tx", orgsContext)
	require.NoError(t, err)

	coll1 := "collection1"
	ccID := integration.GenerateExamplePvtID(true)
	collConfig, err := newCollectionConfig(coll1, "OR('Org2MSP.member')", 0, 2, 1000)
	require.NoError(t, err)

	if metadata.CCMode == "lscc" {
		err = integration.InstallExamplePvtChaincode(orgsContext, ccID)
		require.NoError(t, err)
		err = integration.InstantiateExamplePvtChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')", collConfig)
		require.NoError(t, err)
	} else {
		err := integration.InstantiatePvtExampleChaincodeLc(sdk, orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')", collConfig)
		require.NoError(t, err)
	}

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
			channel.WithRetry(retry.TestRetryOpts),
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
			channel.WithRetry(retry.TestRetryOpts),
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
	// 'ApproveChaincodeDefinitionForMyOrg' failed: error validating chaincode definition: collection-name: collection1 -- collection member 'Org3MSP' is not part of the channel
	if metadata.CCMode != "lscc" {
		t.Skip("this test is only valid for legacy chaincode")
	}
	sdk := mainSDK

	orgsContext := setupMultiOrgContext(t, sdk)

	// Just join peers in org1 for now
	err := integration.EnsureChannelCreatedAndPeersJoined(t, sdk, orgChannelID, "orgchannel.tx", orgsContext)
	require.NoError(t, err)

	coll1 := "collection1"
	ccID := integration.GenerateExamplePvtID(true)
	collConfig, err := newCollectionConfig(coll1, "OR('Org3MSP.member')", 0, 2, 1000)
	require.NoError(t, err)

	err = integration.InstallExamplePvtChaincode(orgsContext, ccID)
	require.NoError(t, err)
	err = integration.InstantiateExamplePvtChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member','Org3MSP.member')", collConfig)
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
			channel.WithRetry(retry.TestRetryOpts),
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
			channel.WithRetry(retry.TestRetryOpts),
		)
		require.NoError(t, err)
		t.Logf("Got %d response(s)", len(response.Responses))
		require.NotEmptyf(t, response.Responses, "expecting at least one response")
	})

}

// Data in a private data collection must be left untouched if the client receives an MVCC_READ_CONFLICT error.
// We test this by submitting two cumulative changes to a private data collection, ensuring that the MVCC_READ_CONFLICT error
// is reproduced, then asserting that only one of the changes was applied.
func TestChannelClientRollsBackPvtDataIfMvccReadConflict(t *testing.T) {
	// 'ApproveChaincodeDefinitionForMyOrg' failed: error validating chaincode definition: collection-name: collection1 -- collection member 'Org3MSP' is not part of the channel
	if metadata.CCMode == "lscc" {
		orgsContext := setupMultiOrgContext(t, mainSDK)
		require.NoError(t, integration.EnsureChannelCreatedAndPeersJoined(t, mainSDK, orgChannelID, "orgchannel.tx", orgsContext))
		// private data collection used for test
		const coll = "collection1"
		// collection key used for test
		const key = "collection_key"
		ccID := integration.GenerateExamplePvtID(true)
		collConfig, err := newCollectionConfig(coll, "OR('Org1MSP.member','Org2MSP.member','Org3MSP.member')", 0, 2, 1000)
		require.NoError(t, err)

		require.NoError(t, integration.InstallExamplePvtChaincode(orgsContext, ccID))
		require.NoError(t, integration.InstantiateExamplePvtChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member','Org3MSP.member')", collConfig))

		ctxProvider := mainSDK.ChannelContext(orgChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
		chClient, err := channel.New(ctxProvider)
		require.NoError(t, err)

		var errMtx sync.Mutex
		errs := multi.Errors{}
		var wg sync.WaitGroup

		// test function; invokes a CC function that mutates the private data collection
		changePvtData := func(amount int) {
			defer wg.Done()
			_, err := chClient.Execute(
				channel.Request{
					ChaincodeID: ccID,
					Fcn:         "addToInt",
					Args:        [][]byte{[]byte(coll), []byte(key), []byte(strconv.Itoa(amount))},
				},
			)
			if err != nil {
				errMtx.Lock()
				errs = append(errs, err)
				errMtx.Unlock()
				return
			}
		}

		// expected value at the end of the test
		const expected = 10

		wg.Add(2)
		go changePvtData(expected)
		go changePvtData(expected)
		wg.Wait()

		// ensure the MVCC_READ_CONFLICT was reproduced
		require.Truef(t, len(errs) > 0 && strings.Contains(errs[0].Error(), "MVCC_READ_CONFLICT"), "could not reproduce MVCC_READ_CONFLICT")

		// read current value of private data collection
		//resp, err := chClient.Query(
		//	channel.Request{
		//		ChaincodeID: ccID,
		//		Fcn:         "getprivate",
		//		Args:        [][]byte{[]byte(coll), []byte(key)},
		//	},
		//	channel.WithRetry(retry.TestRetryOpts),
		//)
		resp, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
			func() (interface{}, error) {
				b, e := chClient.Query(
					channel.Request{
						ChaincodeID: ccID,
						Fcn:         "getprivate",
						Args:        [][]byte{[]byte(coll), []byte(key)},
					},
					channel.WithRetry(retry.TestRetryOpts),
				)
				if e != nil || strings.TrimSpace(string(b.Payload)) == "" {
					return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("getprivate data returned error: %v", e), nil)
				}
				return b, e
			},
		)
		require.NoErrorf(t, err, "error attempting to read private data")
		require.NotEmptyf(t, strings.TrimSpace(string(resp.(channel.Response).Payload)), "reading private data returned empty response")

		actual, err := strconv.Atoi(string(resp.(channel.Response).Payload))
		require.NoError(t, err)

		assert.Truef(t, actual == expected, "Private data not rolled back during MVCC_READ_CONFLICT")
	}
}

func newCollectionConfig(colName, policy string, reqPeerCount, maxPeerCount int32, blockToLive uint64) (*pb.CollectionConfig, error) {
	p, err := policydsl.FromString(policy)
	if err != nil {
		return nil, err
	}
	cpc := &pb.CollectionPolicyConfig{
		Payload: &pb.CollectionPolicyConfig_SignaturePolicy{
			SignaturePolicy: p,
		},
	}
	return &pb.CollectionConfig{
		Payload: &pb.CollectionConfig_StaticCollectionConfig{
			StaticCollectionConfig: &pb.StaticCollectionConfig{
				Name:              colName,
				MemberOrgsPolicy:  cpc,
				RequiredPeerCount: reqPeerCount,
				MaximumPeerCount:  maxPeerCount,
				BlockToLive:       blockToLive,
			},
		},
	}, nil
}

// TestPrivateDataReconcilePutAndGet tests put and get for private data with reconciliation of missing eligible data on some peers (org2's peers)
// the idea to test private data reconciliation is to set a test collection with a policy of 1 member org, put/get private data
// then update the collection config with a new policy of 2 member orgs, private data should be reconciled on peers of the newly added org
func TestPrivateDataReconcilePutAndGet(t *testing.T) {
	sdk := mainSDK
	singleOrgPolicy := "AND('Org1MSP.member')"
	multiOrgsPolicy := "OR('Org1MSP.member','Org2MSP.member')"
	coll1 := "collectionx"
	ccID := integration.GenerateExamplePvtID(true)
	orgsContext := setupMultiOrgContext(t, sdk)

	// instantiate and install CC on all peers using collection policy for org1 only then put/get some pvt data
	runPvtDataPreReconcilePutAndGet(t, sdk, orgsContext, singleOrgPolicy, ccID, coll1)
	// now verify pvt data is not available on org2 peers
	verifyPvtDataPreReconcileGet(t, sdk, ccID, coll1)
	// upgrade CC to include org2 in collection policy then verify pvt data is available on org2's peers
	runPvtDataPostReconcileGet(t, sdk, orgsContext, multiOrgsPolicy, ccID, coll1)
}

func runPvtDataPreReconcilePutAndGet(t *testing.T, sdk *fabsdk.FabricSDK, orgsContext []*integration.OrgContext, policy, ccID, coll1 string) {
	err := integration.EnsureChannelCreatedAndPeersJoined(t, sdk, orgChannelID, "orgchannel.tx", orgsContext)
	require.NoError(t, err)

	collConfig, err := newCollectionConfig(coll1, policy, 0, 2, 1000)
	require.NoError(t, err)

	if metadata.CCMode == "lscc" {
		err = integration.InstallExamplePvtChaincode(orgsContext, ccID)
		require.NoError(t, err)
		err = integration.InstantiateExamplePvtChaincode(orgsContext, orgChannelID, ccID, policy, collConfig)
		require.NoError(t, err)
	} else {
		err := integration.InstantiatePvtExampleChaincodeLc(sdk, orgsContext, orgChannelID, ccID, policy, collConfig)
		require.NoError(t, err)
	}

	ctxProvider := sdk.ChannelContext(orgChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	chClient, err := channel.New(ctxProvider)
	require.NoError(t, err)

	key1 := "key1"
	key2 := "key2"
	key3 := "key3"
	value1 := "pvtValue1"
	value2 := "pvtValue2"
	value3 := "pvtValue3"
	re, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			re, err := chClient.Query(
				channel.Request{
					ChaincodeID: ccID,
					Fcn:         "getprivate",
					Args:        [][]byte{[]byte(coll1), []byte(key1)},
				},
			)
			if err != nil {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("query returned : %v", re), nil)
			}
			return re, nil
		},
	)
	require.NoError(t, err)
	response := re.(channel.Response)
	t.Logf("Got response payload: [%s]", string(response.Payload))
	require.Nil(t, response.Payload)

	response, err = chClient.Query(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "getprivatebyrange",
			Args:        [][]byte{[]byte(coll1), []byte(key1), []byte(key3)},
		},
		channel.WithRetry(retry.DefaultChannelOpts),
	)
	require.NoError(t, err)
	t.Logf("Got response payload: [%s]", string(response.Payload))
	require.Empty(t, string(response.Payload))

	response, err = chClient.Execute(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "putprivate",
			Args:        [][]byte{[]byte(coll1), []byte(key1), []byte(value1)},
		},
		channel.WithRetry(retry.DefaultChannelOpts),
	)
	require.NoError(t, err)
	require.NotEmptyf(t, response.Responses, "expecting at least one response")

	response, err = chClient.Execute(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "putprivate",
			Args:        [][]byte{[]byte(coll1), []byte(key2), []byte(value2)},
		},
		channel.WithRetry(retry.DefaultChannelOpts),
	)
	require.NoError(t, err)
	require.NotEmptyf(t, response.Responses, "expecting at least one response")

	response, err = chClient.Execute(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "putprivate",
			Args:        [][]byte{[]byte(coll1), []byte(key3), []byte(value3)},
		},
		channel.WithRetry(retry.TestRetryOpts),
	)
	require.NoError(t, err)
	require.NotEmptyf(t, response.Responses, "expecting at least one response")

	response, err = chClient.Query(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "getprivate",
			Args:        [][]byte{[]byte(coll1), []byte(key1)},
		},
		channel.WithRetry(retry.TestRetryOpts),
	)
	require.NoError(t, err)
	t.Logf("Got response payload for getprivate: %s", string(response.Payload))
	require.Equal(t, value1, string(response.Payload))

	response, err = chClient.Query(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "getprivatebyrange",
			Args:        [][]byte{[]byte(coll1), []byte(key1), []byte(key3)},
		},
		channel.WithRetry(retry.DefaultChannelOpts),
	)
	require.NoError(t, err)
	t.Logf("Got response payload for getprivatebyrange: [%s]", string(response.Payload))
	require.NotEmpty(t, string(response.Payload))
}

func verifyPvtDataPreReconcileGet(t *testing.T, sdk *fabsdk.FabricSDK, ccID, coll1 string) {
	// create ctxProvider for org2 to query org2 peers for pvt data (should be empty)
	ctxProvider := sdk.ChannelContext(orgChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org2Name))

	// org2 peers are the only targets to test pre reconciliation as they should not have the pvt data as per the collection policy (singleOrgPolicy)
	org2TargetOpts := channel.WithTargetEndpoints("peer0.org2.example.com", "peer1.org2.example.com")
	chClient, err := channel.New(ctxProvider)
	require.NoError(t, err)

	key1 := "key1"
	key3 := "key3"

	response, err := chClient.Query(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "getprivate",
			Args:        [][]byte{[]byte(coll1), []byte(key1)},
		},
		channel.WithRetry(retry.TestRetryOpts),
		org2TargetOpts, // query org2 peers to ensure they don't have pvt data
	)
	require.Error(t, err)
	t.Logf("Got response payload for getprivate: %s", string(response.Payload))
	require.Empty(t, response.Payload)

	response, err = chClient.Query(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "getprivatebyrange",
			Args:        [][]byte{[]byte(coll1), []byte(key1), []byte(key3)},
		},
		channel.WithRetry(retry.DefaultChannelOpts),
		org2TargetOpts, // query org2 peers to ensure they don't have pvt data
	)

	// for some reason, the peer throws an error for getprivate cc invoke(Failed to handle GET_STATE. error: private data matching public hash version is not available.),
	// but not for getprivatebyrange cc invoke (it only returns empty payload)
	require.NoError(t, err)
	t.Logf("Got response payload for getprivatebyrange: [%s]", string(response.Payload))
	require.Empty(t, string(response.Payload))

}

func runPvtDataPostReconcileGet(t *testing.T, sdk *fabsdk.FabricSDK, orgsContext []*integration.OrgContext, policy, ccID, coll1 string) {
	collConfig, err := newCollectionConfig(coll1, policy, 0, 2, 1000)
	require.NoError(t, err)

	// org2 peers are the only targets to test post reconciliation as they should have the pvt data after cc upgrade as per the new collection policy (multiOrgsPolicy)
	org2TargetOpts := channel.WithTargetEndpoints("peer0.org2.example.com", "peer1.org2.example.com")
	if metadata.CCMode == "lscc" {
		err = integration.UpgradeExamplePvtChaincode(orgsContext, orgChannelID, ccID, policy, collConfig)
		require.NoError(t, err)
	} else {
		err = integration.UpgradeExamplePvtChaincodeLc(sdk, orgsContext, orgChannelID, ccID, policy, collConfig)
		require.NoError(t, err)
	}

	// wait for pvt data reconciliation occurs on peers of org2
	time.Sleep(2 * time.Second)

	// create ctxProvider for org2 to query org2 peers for pvt data (should be not empty/reconciled)
	ctxProvider := sdk.ChannelContext(orgChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org2Name))

	chClient, err := channel.New(ctxProvider)
	require.NoError(t, err)

	key1 := "key1"
	key3 := "key3"
	value1 := "pvtValue1"

	response, err := chClient.Query(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "getprivate",
			Args:        [][]byte{[]byte(coll1), []byte(key1)},
		},
		channel.WithRetry(retry.TestRetryOpts),
		org2TargetOpts,
	)
	require.NoError(t, err)
	t.Logf("Got response payload for getprivate: %s", string(response.Payload))
	require.Equal(t, value1, string(response.Payload))

	response, err = chClient.Query(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "getprivatebyrange",
			Args:        [][]byte{[]byte(coll1), []byte(key1), []byte(key3)},
		},
		channel.WithRetry(retry.DefaultChannelOpts),
		org2TargetOpts,
	)
	require.NoError(t, err)
	t.Logf("Got response payload for getprivatebyrange: [%s]", string(response.Payload))
	require.NotEmpty(t, string(response.Payload))
}
