/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"reflect"
	"testing"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/stretchr/testify/require"
)

const (
	collCfgName              = "collection1"
	collCfgBlockToLive       = 1000
	collCfgRequiredPeerCount = 0
	collCfgMaximumPeerCount  = 2
	collCfgPolicy            = "OR('Org1MSP.member','Org2MSP.member')"
)

func TestQueryCollectionsConfig(t *testing.T) {
	sdk := mainSDK

	orgsContext := setupMultiOrgContext(t, sdk)
	err := integration.EnsureChannelCreatedAndPeersJoined(t, sdk, orgChannelID, "orgchannel.tx", orgsContext)
	require.NoError(t, err)

	ccID := integration.GenerateExamplePvtID(true)
	collConfig, err := newCollectionConfig(collCfgName, collCfgPolicy, collCfgRequiredPeerCount, collCfgMaximumPeerCount, collCfgBlockToLive)
	require.NoError(t, err)

	err = integration.InstallExamplePvtChaincode(orgsContext, ccID)
	require.NoError(t, err)
	err = integration.InstantiateExamplePvtChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')", collConfig)
	require.NoError(t, err)

	org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))
	client, err := resmgmt.New(org1AdminClientContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	resp, err := client.QueryCollectionsConfig(orgChannelID, ccID)
	if err != nil {
		t.Fatalf("QueryCollectionsConfig return error: %s", err)
	}
	if len(resp.Config) != 1 {
		t.Fatalf("The number of collection config is incorrect, expected 1, got %d", len(resp.Config))
	}

	conf := resp.Config[0]
	switch cconf := conf.Payload.(type) {
	case *pb.CollectionConfig_StaticCollectionConfig:
		checkStaticCollectionConfig(t, cconf.StaticCollectionConfig)
	default:
		t.Fatalf("The CollectionConfig.Payload's type is incorrect, expected `CollectionConfig_StaticCollectionConfig`, got %+v", reflect.TypeOf(conf.Payload))
	}
}

func checkStaticCollectionConfig(t *testing.T, collConf *pb.StaticCollectionConfig) {
	if collConf.Name != collCfgName {
		t.Fatalf("CollectionConfig'name is incorrect, expected collection1, got %s", collConf.Name)
	}
	if collConf.BlockToLive != collCfgBlockToLive {
		t.Fatalf("The property of BlockToLive is incorrect, expected 1000, got %d", collConf.BlockToLive)
	}
	if collConf.RequiredPeerCount != collCfgRequiredPeerCount {
		t.Fatalf("The property of RequiredPeerCount is incorrect, expected 0, got %d", collConf.RequiredPeerCount)
	}
	if collConf.MaximumPeerCount != collCfgMaximumPeerCount {
		t.Fatalf("The property of MaximumPeerCount is incorrect, expected 2, got %d", collConf.MaximumPeerCount)
	}
	if collConf.MemberOrgsPolicy.GetSignaturePolicy() == nil {
		t.Fatalf("The property of MemberOrgsPolicy must be SignaturePolicy")
	}
}

func newCollectionConfig(colName, policy string, reqPeerCount, maxPeerCount int32, blockToLive uint64) (*pb.CollectionConfig, error) {
	p, err := cauthdsl.FromString(policy)
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
