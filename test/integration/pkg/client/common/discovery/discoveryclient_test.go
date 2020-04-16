// +build !prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"strings"
	"testing"
	"time"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	fabdiscovery "github.com/hyperledger/fabric-protos-go/discovery"
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const (
	peer0Org1 = "peer0.org1.example.com"

	peer0Org1URL = "peer0.org1.example.com:7051"
	peer1Org1URL = "peer1.org1.example.com:7151"
	peer0Org2URL = "peer0.org2.example.com:8051"
	peer1Org2URL = "peer1.org2.example.com:9051"
)

func TestDiscoveryClientPeers(t *testing.T) {
	orgsContext := setupOrgContext(t)
	err := ensureChannelCreatedAndPeersJoined(t, orgsContext)
	require.NoError(t, err)

	ctx, err := orgsContext[0].CtxProvider()
	require.NoError(t, err, "error getting channel context")

	client, err := discovery.New(ctx)
	require.NoError(t, err, "error creating discovery client")

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	req := discovery.NewRequest().OfChannel(orgChannelID).AddPeersQuery()

	peerCfg1, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), peer0Org1)
	require.NoErrorf(t, err, "error getting peer config for [%s]", peer0Org1)

	responsesCh, err := client.Send(reqCtx, req, peerCfg1.PeerConfig)
	require.NoError(t, err, "error calling discover service send")

	var responses []discovery.Response

	for resp := range responsesCh {
		responses = append(responses, resp)
	}
	require.NotEmpty(t, responses, "expecting one response but got none")

	resp := responses[0]
	chanResp := resp.ForChannel(orgChannelID)

	peers, err := chanResp.Peers()
	require.NoError(t, err, "error getting peers")
	require.NotEmpty(t, peers, "expecting at least one peer but got none")

	t.Logf("*** Peers for channel %s:", orgChannelID)
	for _, peer := range peers {
		aliveMsg := peer.AliveMessage.GetAliveMsg()
		if !assert.NotNil(t, aliveMsg, "got nil AliveMessage") {
			continue
		}
		if !assert.NotNil(t, aliveMsg.Membership, "got nil Membership") {
			continue
		}

		t.Logf("--- Endpoint: %s", aliveMsg.Membership.Endpoint)

		if !assert.NotNil(t, peer.StateInfoMessage, "got nil StateInfoMessage") {
			continue
		}

		stateInfo := peer.StateInfoMessage.GetStateInfo()
		if !assert.NotNil(t, stateInfo, "got nil stateInfo") {
			continue
		}

		if !assert.NotNil(t, stateInfo.Properties, "got nil stateInfo.Properties") {
			continue
		}

		t.Logf("--- Ledger Height: %d", stateInfo.Properties.LedgerHeight)
		t.Log("--- Chaincodes:")
		for _, cc := range stateInfo.Properties.Chaincodes {
			t.Logf("------ %s:%s", cc.Name, cc.Version)
		}
	}
}

func TestDiscoveryClientLocalPeers(t *testing.T) {
	sdk := mainSDK

	// By default, query for local peers (outside of a channel) requires admin privileges.
	// To bypass this restriction, set peer.discovery.orgMembersAllowedAccess=true in core.yaml.
	ctxProvider := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(org1Name))
	ctx, err := ctxProvider()
	require.NoError(t, err, "error getting channel context")

	client, err := discovery.New(ctx)
	require.NoError(t, err, "error creating discovery client")

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	req := discovery.NewRequest().AddLocalPeersQuery()

	peerCfg1, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), peer0Org1)
	require.NoErrorf(t, err, "error getting peer config for [%s]", peer0Org1)

	responsesCh, err := client.Send(reqCtx, req, peerCfg1.PeerConfig)
	require.NoError(t, err, "error calling discover service send")

	var responses []discovery.Response

	for resp := range responsesCh {
		responses = append(responses, resp)
	}

	require.NotEmpty(t, responses, "No responses")

	resp := responses[0]

	locResp := resp.ForLocal()

	peers, err := locResp.Peers()
	require.NoError(t, err, "error getting local peers")

	t.Log("*** Local Peers:")
	for _, peer := range peers {
		aliveMsg := peer.AliveMessage.GetAliveMsg()
		if !assert.NotNil(t, aliveMsg, "got nil AliveMessage") {
			continue
		}
		if !assert.NotNil(t, aliveMsg.Membership, "got nil Membership") {
			continue
		}

		t.Logf("--- Endpoint: %s", aliveMsg.Membership.Endpoint)

		assert.Nil(t, peer.StateInfoMessage, "expected nil StateInfoMessage for local peer")
	}
}

func TestDiscoveryClientEndorsers(t *testing.T) {
	orgsContext := setupOrgContext(t)
	err := ensureChannelCreatedAndPeersJoined(t, orgsContext)
	require.NoError(t, err)

	t.Run("Policy: Org1 Only", func(t *testing.T) {
		ccID := integration.GenerateExampleID(true)
		err = integration.InstallExampleChaincode(orgsContext, ccID)
		require.NoError(t, err)
		err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member')")
		require.NoError(t, err)

		testEndorsers(
			t, mainSDK,
			newInterest(newCCCall(ccID)),
			discclient.NoFilter,
			[]string{peer0Org1URL},
			[]string{peer1Org1URL},
		)
	})

	t.Run("Policy: Org2 Only", func(t *testing.T) {
		ccID := integration.GenerateExampleID(true)
		err = integration.InstallExampleChaincode(orgsContext, ccID)
		require.NoError(t, err)
		err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID, "OR('Org2MSP.member')")
		require.NoError(t, err)

		testEndorsers(
			t, mainSDK,
			newInterest(newCCCall(ccID)),
			discclient.NoFilter,
			[]string{peer0Org2URL},
			[]string{peer1Org2URL},
		)
	})

	t.Run("Policy: Org1 or Org2", func(t *testing.T) {
		ccID := integration.GenerateExampleID(true)
		err = integration.InstallExampleChaincode(orgsContext, ccID)
		require.NoError(t, err)
		err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')")
		require.NoError(t, err)

		require.NoError(t, err)
		testEndorsers(
			t, mainSDK,
			newInterest(newCCCall(ccID)),
			discclient.NoFilter,
			[]string{peer0Org1URL},
			[]string{peer1Org1URL},
			[]string{peer0Org2URL},
			[]string{peer1Org2URL},
		)
	})

	t.Run("Policy: Org1 and Org2", func(t *testing.T) {
		ccID := integration.GenerateExampleID(true)
		err = integration.InstallExampleChaincode(orgsContext, ccID)
		require.NoError(t, err)
		err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID, "AND('Org1MSP.member','Org2MSP.member')")
		require.NoError(t, err)

		testEndorsers(
			t, mainSDK,
			newInterest(newCCCall(ccID)),
			discclient.NoFilter,
			[]string{peer0Org1URL, peer0Org2URL},
			[]string{peer1Org1URL, peer0Org2URL},
			[]string{peer0Org1URL, peer1Org2URL},
			[]string{peer1Org1URL, peer1Org2URL},
		)
	})

	// Chaincode to Chaincode
	t.Run("Policy: CC1(Org1 Only) to CC2(Org2 Only)", func(t *testing.T) {

		ccID1 := integration.GenerateExampleID(true)
		err = integration.InstallExampleChaincode(orgsContext, ccID1)
		require.NoError(t, err)
		err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID1, "OR('Org1MSP.member')")
		require.NoError(t, err)

		ccID2 := integration.GenerateExampleID(true)
		err = integration.InstallExampleChaincode(orgsContext, ccID2)
		require.NoError(t, err)
		err = integration.InstantiateExampleChaincode(orgsContext, orgChannelID, ccID2, "OR('Org2MSP.member')")
		require.NoError(t, err)

		testEndorsers(
			t, mainSDK,
			newInterest(newCCCall(ccID1), newCCCall(ccID2)),
			discclient.NoFilter,
			[]string{peer0Org1URL, peer0Org2URL},
			[]string{peer1Org1URL, peer0Org2URL},
			[]string{peer0Org1URL, peer1Org2URL},
			[]string{peer1Org1URL, peer1Org2URL},
		)
	})
}

func testEndorsers(t *testing.T, sdk *fabsdk.FabricSDK, interest *fabdiscovery.ChaincodeInterest, filter discclient.Filter, expectedEndorserGroups ...[]string) {
	ctxProvider := sdk.Context(fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
	ctx, err := ctxProvider()
	require.NoError(t, err, "error getting channel context")

	client, err := discovery.New(ctx)
	require.NoError(t, err, "error creating discovery client")

	peerCfg1, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), peer0Org1)
	require.NoErrorf(t, err, "error getting peer config for [%s]", peer0Org1)

	chResponse, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			chanResp, err := sendEndorserQuery(t, ctx, client, interest, peerCfg1.PeerConfig)
			if err != nil && strings.Contains(err.Error(), "failed constructing descriptor for chaincodes") {
				// This error is a result of Gossip not being up-to-date with the instantiated chaincodes of all peers.
				// A retry should resolve the error.
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), err.Error(), nil)
			} else if err != nil {
				return nil, errors.WithMessage(err, "Got error from discovery")
			}

			return chanResp, nil
		},
	)

	require.NoError(t, err)
	chanResp := chResponse.(discclient.ChannelResponse)

	// Get endorsers a few times, since each time a different set may be returned
	for i := 0; i < 3; i++ {
		endorsers, err := chanResp.Endorsers(interest.Chaincodes, filter)
		require.NoError(t, err, "error getting endorsers")
		checkEndorsers(t, asURLs(t, endorsers), expectedEndorserGroups)
	}
}

func ensureChannelCreatedAndPeersJoined(t *testing.T, orgsContext []*integration.OrgContext) error {
	joined, err := integration.IsJoinedChannel(orgChannelID, orgsContext[0].ResMgmt, orgsContext[0].Peers[0])
	if err != nil {
		return err
	}

	if joined {
		return nil
	}

	// Create the channel and update anchor peers for all orgs
	if err := integration.CreateChannelAndUpdateAnchorPeers(t, mainSDK, orgChannelID, "orgchannel.tx", orgsContext); err != nil {
		return err
	}

	return integration.JoinPeersToChannel(orgChannelID, orgsContext)
}

func setupOrgContext(t *testing.T) []*integration.OrgContext {
	sdk := mainSDK

	org1AdminContext := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(org1Name))
	org1ResMgmt, err := resmgmt.New(org1AdminContext)
	require.NoError(t, err)

	org1MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org1Name))
	require.NoError(t, err)
	org1AdminUser, err := org1MspClient.GetSigningIdentity(adminUser)
	require.NoError(t, err)

	org2AdminContext := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(org2Name))
	org2ResMgmt, err := resmgmt.New(org2AdminContext)
	require.NoError(t, err)

	org2MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org2Name))
	require.NoError(t, err)
	org2AdminUser, err := org2MspClient.GetSigningIdentity(adminUser)
	require.NoError(t, err)

	// Ensure that Gossip has propagated it's view of local peers before invoking
	// install since some peers may be missed if we call InstallCC too early
	org1Peers, err := integration.DiscoverLocalPeers(org1AdminContext, 2)
	require.NoError(t, err)
	org2Peers, err := integration.DiscoverLocalPeers(org2AdminContext, 2)
	require.NoError(t, err)

	return []*integration.OrgContext{
		{
			OrgID:                org1Name,
			CtxProvider:          org1AdminContext,
			ResMgmt:              org1ResMgmt,
			Peers:                org1Peers,
			SigningIdentity:      org1AdminUser,
			AnchorPeerConfigFile: "orgchannelOrg1MSPanchors.tx",
		},
		{
			OrgID:                org2Name,
			CtxProvider:          org2AdminContext,
			ResMgmt:              org2ResMgmt,
			Peers:                org2Peers,
			SigningIdentity:      org2AdminUser,
			AnchorPeerConfigFile: "orgchannelOrg2MSPanchors.tx",
		},
	}
}

func sendEndorserQuery(t *testing.T, ctx contextAPI.Client, client discovery.Client, interest *fabdiscovery.ChaincodeInterest, peerConfig fab.PeerConfig) (discclient.ChannelResponse, error) {
	req, err := discovery.NewRequest().OfChannel(orgChannelID).AddEndorsersQuery(interest)
	require.NoError(t, err, "error adding endorsers query")

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	responsesCh, err := client.Send(reqCtx, req, peerConfig)
	require.NoError(t, err, "error calling discover service send")

	var responses []discovery.Response

	for resp := range responsesCh {
		responses = append(responses, resp)
	}

	require.NotEmpty(t, responses, "expecting one response but got none")

	chanResp := responses[0].ForChannel(orgChannelID)

	_, err = chanResp.Endorsers(interest.Chaincodes, discclient.NewFilter(discclient.NoPriorities, discclient.NoExclusion))
	if err != nil {
		return nil, err
	}
	return chanResp, nil
}

func checkEndorsers(t *testing.T, endorsers []string, expectedGroups [][]string) {
	for _, group := range expectedGroups {
		if containsAll(t, endorsers, group) {
			t.Logf("Found matching endorser group: %#v", group)
			return
		}
	}
	t.Fatalf("Unexpected endorser group: %#v - Expecting one of: %#v", endorsers, expectedGroups)
}

func containsAll(t *testing.T, endorsers []string, expectedEndorserGroup []string) bool {
	if len(endorsers) != len(expectedEndorserGroup) {
		return false
	}

	for _, endorser := range endorsers {
		t.Logf("Checking endpoint: %s ...", endorser)
		if !contains(expectedEndorserGroup, endorser) {
			return false
		}
	}
	return true
}

func contains(group []string, endorser string) bool {
	for _, e := range group {
		if e == endorser {
			return true
		}
	}
	return false
}

func newCCCall(ccID string, collections ...string) *fabdiscovery.ChaincodeCall {
	return &fabdiscovery.ChaincodeCall{
		Name:            ccID,
		CollectionNames: collections,
	}
}

func newInterest(ccCalls ...*fabdiscovery.ChaincodeCall) *fabdiscovery.ChaincodeInterest {
	return &fabdiscovery.ChaincodeInterest{Chaincodes: ccCalls}
}

func asURLs(t *testing.T, endorsers discclient.Endorsers) []string {
	var urls []string
	for _, endorser := range endorsers {
		aliveMsg := endorser.AliveMessage.GetAliveMsg()
		require.NotNil(t, aliveMsg, "got nil AliveMessage")
		require.NotNil(t, aliveMsg.Membership, "got nil Membership")
		urls = append(urls, aliveMsg.Membership.Endpoint)
	}
	return urls
}
