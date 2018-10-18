// +build !prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
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

	err = integration.InstallExamplePvtChaincode(orgsContext, ccID)
	require.NoError(t, err)
	err = integration.InstantiateExamplePvtChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')", collConfig)
	require.NoError(t, err)

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

	err = integration.InstallExamplePvtChaincode(orgsContext, ccID)
	require.NoError(t, err)
	err = integration.InstantiateExamplePvtChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')", collConfig)
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
	resp, err := chClient.Query(
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "getprivate",
			Args:        [][]byte{[]byte(coll), []byte(key)},
		},
		channel.WithRetry(retry.TestRetryOpts),
	)
	require.NoErrorf(t, err, "error attempting to read private data")
	require.NotEmptyf(t, resp.Payload, "reading private data returned empty response")

	actual, err := strconv.Atoi(string(resp.Payload))
	require.NoError(t, err)

	assert.Truef(t, actual == expected, "Private data not rolled back during MVCC_READ_CONFLICT")
}

func newCollectionConfig(colName, policy string, reqPeerCount, maxPeerCount int32, blockToLive uint64) (*cb.CollectionConfig, error) {
	p, err := cauthdsl.FromString(policy)
	if err != nil {
		return nil, err
	}
	cpc := &cb.CollectionPolicyConfig{
		Payload: &cb.CollectionPolicyConfig_SignaturePolicy{
			SignaturePolicy: p,
		},
	}
	return &cb.CollectionConfig{
		Payload: &cb.CollectionConfig_StaticCollectionConfig{
			StaticCollectionConfig: &cb.StaticCollectionConfig{
				Name:              colName,
				MemberOrgsPolicy:  cpc,
				RequiredPeerCount: reqPeerCount,
				MaximumPeerCount:  maxPeerCount,
				BlockToLive:       blockToLive,
			},
		},
	}, nil
}
