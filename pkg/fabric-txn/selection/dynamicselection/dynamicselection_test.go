/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicselection

import (
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context/defprovider"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/selection/dynamicselection/pgresolver"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

var testConfig *config.Config

const (
	org1  = "Org1MSP"
	org2  = "Org2MSP"
	org3  = "Org3MSP"
	org4  = "Org4MSP"
	org5  = "Org5MSP"
	org6  = "Org6MSP"
	org7  = "Org7MSP"
	org8  = "Org8MSP"
	org9  = "Org9MSP"
	org10 = "Org10MSP"
)

const (
	channel1 = "channel1"
	channel2 = "channel2"
)

const (
	cc1 = "cc1"
	cc2 = "cc2"
	cc3 = "cc3"
)

const (
	o1 = iota
	o2
	o3
	o4
	o5
)

var p1 = peer("peer1", org1)
var p2 = peer("peer2", org1)
var p3 = peer("peer3", org2)
var p4 = peer("peer4", org2)
var p5 = peer("peer5", org3)
var p6 = peer("peer6", org3)
var p7 = peer("peer7", org3)
var p8 = peer("peer8", org4)
var p9 = peer("peer9", org4)
var p10 = peer("peer10", org4)
var p11 = peer("peer11", org5)
var p12 = peer("peer12", org5)

func TestGetEndorsersForChaincodeOneCC(t *testing.T) {

	channelPeers := []apifabclient.Peer{p1, p2, p3, p4, p5, p6, p7, p8}

	service := newMockSelectionService(
		newMockCCDataProvider(channel1).
			add(cc1, getPolicy1()),
		pgresolver.NewRoundRobinLBP())

	// Channel1(Policy(cc1)) = Org1
	expected := []pgresolver.PeerGroup{
		// Org1
		pg(p1), pg(p2),
	}
	verify(t, service, expected, channel1, channelPeers, cc1)
}

func TestGetEndorsersForChaincodeTwoCCs(t *testing.T) {

	service := newMockSelectionService(
		newMockCCDataProvider(channel1).
			add(cc1, getPolicy1()).
			add(cc2, getPolicy2()),
		pgresolver.NewRoundRobinLBP())

	// Channel1(Policy(cc1) and Policy(cc2)) = Org1 and (1 of [(2 of [Org1,Org2]),(2 of [Org1,Org3,Org4])])
	expected := []pgresolver.PeerGroup{
		// Org1 and Org2
		pg(p1, p3), pg(p1, p4), pg(p2, p3), pg(p2, p4),
		// Org1 and Org3
		pg(p1, p5), pg(p1, p6), pg(p1, p7), pg(p2, p5), pg(p2, p6), pg(p2, p7),
		// Org1 and Org4
		pg(p1, p8), pg(p1, p9), pg(p1, p10), pg(p2, p8), pg(p2, p9), pg(p2, p10),
		// Org1 and Org3 and Org4
		pg(p1, p5, p8), pg(p1, p5, p9), pg(p1, p5, p10), pg(p1, p6, p8), pg(p1, p6, p9), pg(p1, p6, p10), pg(p1, p7, p8), pg(p1, p7, p9), pg(p1, p7, p10),
		pg(p2, p5, p8), pg(p2, p5, p9), pg(p2, p5, p10), pg(p2, p6, p8), pg(p2, p6, p9), pg(p2, p6, p10), pg(p2, p7, p8), pg(p2, p7, p9), pg(p2, p7, p10),
	}
	channelPeers := []apifabclient.Peer{p1, p2, p3, p4, p5, p6, p7, p8}
	verify(t, service, expected, channel1, channelPeers, cc1, cc2)
}

func TestGetEndorsersForChaincodeTwoCCsTwoChannels(t *testing.T) {

	service := newMockSelectionService(
		newMockCCDataProvider(channel1).
			add(cc1, getPolicy1()).
			add(cc2, getPolicy2()),
		pgresolver.NewRoundRobinLBP(),
	)

	// Channel1(Policy(cc1) and Policy(cc2)) = Org1 and (1 of [(2 of [Org1,Org2]),(2 of [Org1,Org3,Org4])])
	expected := []pgresolver.PeerGroup{
		// Org1 and Org2
		pg(p1, p3), pg(p1, p4), pg(p2, p3), pg(p2, p4),
		// Org1 and Org3
		pg(p1, p5), pg(p1, p6), pg(p1, p7), pg(p2, p5), pg(p2, p6), pg(p2, p7),
		// Org1 and Org4
		pg(p1, p8), pg(p1, p9), pg(p1, p10), pg(p2, p8), pg(p2, p9), pg(p2, p10),
		// Org1 and Org3 and Org4
		pg(p1, p5, p8), pg(p1, p5, p9), pg(p1, p5, p10), pg(p1, p6, p8), pg(p1, p6, p9), pg(p1, p6, p10), pg(p1, p7, p8), pg(p1, p7, p9), pg(p1, p7, p10),
		pg(p2, p5, p8), pg(p2, p5, p9), pg(p2, p5, p10), pg(p2, p6, p8), pg(p2, p6, p9), pg(p2, p6, p10), pg(p2, p7, p8), pg(p2, p7, p9), pg(p2, p7, p10),
	}

	channel1Peers := []apifabclient.Peer{p1, p2, p3, p4, p5, p6, p7, p8}
	verify(t, service, expected, channel1, channel1Peers, cc1, cc2)

	service = newMockSelectionService(
		newMockCCDataProvider(channel2).
			add(cc1, getPolicy3()).
			add(cc2, getPolicy2()),
		pgresolver.NewRoundRobinLBP(),
	)

	// Channel2(Policy(cc1) and Policy(cc2)) = Org5 and (1 of [(2 of [Org1,Org2]),(2 of [Org1,Org3,Org4])])
	expected = []pgresolver.PeerGroup{
		// Org5 and Org2
		pg(p11, p1, p3), pg(p11, p1, p4), pg(p11, p2, p3), pg(p11, p2, p4),
		pg(p12, p1, p3), pg(p12, p1, p4), pg(p12, p2, p3), pg(p12, p2, p4),
		// Org5 and Org3
		pg(p11, p1, p5), pg(p11, p1, p6), pg(p11, p1, p7), pg(p11, p2, p5), pg(p11, p2, p6), pg(p11, p2, p7),
		pg(p12, p1, p5), pg(p12, p1, p6), pg(p12, p1, p7), pg(p12, p2, p5), pg(p12, p2, p6), pg(p12, p2, p7),
		// Org5 and Org4
		pg(p11, p1, p8), pg(p11, p1, p9), pg(p11, p1, p10), pg(p11, p2, p8), pg(p11, p2, p9), pg(p11, p2, p10),
		pg(p12, p1, p8), pg(p12, p1, p9), pg(p12, p1, p10), pg(p12, p2, p8), pg(p12, p2, p9), pg(p12, p2, p10),
		// Org5 and Org3 and Org4
		pg(p11, p5, p8), pg(p11, p5, p9), pg(p11, p5, p10), pg(p11, p6, p8), pg(p11, p6, p9), pg(p11, p6, p10), pg(p11, p7, p8), pg(p11, p7, p9), pg(p11, p7, p10),
		pg(p12, p5, p8), pg(p12, p5, p9), pg(p12, p5, p10), pg(p12, p6, p8), pg(p12, p6, p9), pg(p12, p6, p10), pg(p12, p7, p8), pg(p12, p7, p9), pg(p12, p7, p10),
	}

	channel2Peers := []apifabclient.Peer{p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12}
	verify(t, service, expected, channel2, channel2Peers, cc1, cc2)
}

func verify(t *testing.T, service apifabclient.SelectionService, expectedPeerGroups []pgresolver.PeerGroup, channelID string, channelPeers []apifabclient.Peer, chaincodeIDs ...string) {
	// Set the log level to WARNING since the following spits out too much info in DEBUG
	module := "pg-resolver"
	level := logging.GetLevel(module)
	logging.SetLevel(module, apilogging.WARNING)
	defer logging.SetLevel(module, level)

	for i := 0; i < len(expectedPeerGroups); i++ {
		peers, err := service.GetEndorsersForChaincode(channelPeers, chaincodeIDs...)
		if err != nil {
			t.Fatalf("error getting endorsers: %s", err)
		}
		if !containsPeerGroup(expectedPeerGroups, peers) {
			t.Fatalf("peer group %s is not one of the expected peer groups: %v", toString(peers), expectedPeerGroups)
		}
	}

}

func containsPeerGroup(groups []pgresolver.PeerGroup, peers []apifabclient.Peer) bool {
	for _, g := range groups {
		if containsAllPeers(peers, g) {
			return true
		}
	}
	return false
}

func containsAllPeers(peers []apifabclient.Peer, pg pgresolver.PeerGroup) bool {
	if len(peers) != len(pg.Peers()) {
		return false
	}
	for _, peer := range peers {
		if !containsPeer(pg.Peers(), peer) {
			return false
		}
	}
	return true
}

func containsPeer(peers []apifabclient.Peer, peer apifabclient.Peer) bool {
	for _, p := range peers {
		if p.URL() == peer.URL() {
			return true
		}
	}
	return false
}

func pg(peers ...apifabclient.Peer) pgresolver.PeerGroup {
	return pgresolver.NewPeerGroup(peers...)
}

func peer(name string, mspID string) apifabclient.Peer {

	mp := mocks.NewMockPeer(name, name+":7051")
	mp.SetMSPID(mspID)
	return mp
}

func newMockSelectionService(ccPolicyProvider CCPolicyProvider, lbp pgresolver.LoadBalancePolicy) apifabclient.SelectionService {
	return &selectionService{
		ccPolicyProvider: ccPolicyProvider,
		pgLBP:            lbp,
		pgResolvers:      make(map[string]pgresolver.PeerGroupResolver),
	}
}

type mockCCDataProvider struct {
	channelID string
	ccData    map[string]*ccprovider.ChaincodeData
}

func newMockCCDataProvider(channelID string) *mockCCDataProvider {
	return &mockCCDataProvider{channelID: channelID, ccData: make(map[string]*ccprovider.ChaincodeData)}
}

func (p *mockCCDataProvider) GetChaincodePolicy(chaincodeID string) (*common.SignaturePolicyEnvelope, error) {
	return unmarshalPolicy(p.ccData[newResolverKey(p.channelID, chaincodeID).String()].Policy)
}

func (p *mockCCDataProvider) add(chaincodeID string, policy *ccprovider.ChaincodeData) *mockCCDataProvider {
	p.ccData[newResolverKey(p.channelID, chaincodeID).String()] = policy
	return p
}

// Policy: Org1
func getPolicy1() *ccprovider.ChaincodeData {
	signedBy, identities, err := pgresolver.GetPolicies(org1)
	if err != nil {
		panic(err)
	}

	return newCCData(&common.SignaturePolicyEnvelope{
		Version:    0,
		Rule:       signedBy[o1],
		Identities: identities,
	})
}

// Policy: 1 of [(2 of [Org1, Org2]),(2 of [Org1, Org3, Org4])]
func getPolicy2() *ccprovider.ChaincodeData {
	signedBy, identities, err := pgresolver.GetPolicies(org1, org2, org3, org4)
	if err != nil {
		panic(err)
	}

	return newCCData(&common.SignaturePolicyEnvelope{
		Version: 0,
		Rule: pgresolver.NewNOutOfPolicy(1,
			pgresolver.NewNOutOfPolicy(2,
				signedBy[o1],
				signedBy[o2],
			),
			pgresolver.NewNOutOfPolicy(2,
				signedBy[o1],
				signedBy[o3],
				signedBy[o4],
			),
		),
		Identities: identities,
	})
}

// Policy: Org5
func getPolicy3() *ccprovider.ChaincodeData {
	signedBy, identities, err := pgresolver.GetPolicies(org1, org2, org3, org4, org5)
	if err != nil {
		panic(err)
	}

	return newCCData(&common.SignaturePolicyEnvelope{
		Version:    0,
		Rule:       signedBy[o5],
		Identities: identities,
	})
}

func newCCData(sigPolicyEnv *common.SignaturePolicyEnvelope) *ccprovider.ChaincodeData {
	policyBytes, err := proto.Marshal(sigPolicyEnv)
	if err != nil {
		panic(err)
	}

	return &ccprovider.ChaincodeData{Policy: policyBytes}
}

func toString(peers []apifabclient.Peer) string {
	str := "["
	for i, p := range peers {
		str += p.URL()
		if i+1 < len(peers) {
			str += ","
		}
	}
	str += "]"
	return str
}

func TestDynamicSelection(t *testing.T) {

	config, err := config.InitConfig("../../../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	mychannelUser := ChannelUser{ChannelID: "mychannel", UserName: "User1", OrgName: "Org1"}

	selectionProvider, err := NewSelectionProvider(config, []ChannelUser{mychannelUser}, nil)
	if err != nil {
		t.Fatalf("Failed to setup selection provider: %s", err)
	}

	selectionService, err := selectionProvider.NewSelectionService("")
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	selectionService, err = selectionProvider.NewSelectionService("mychannel")
	if err == nil {
		t.Fatalf("Should have failed since sdk not provided")
	}

	// Create SDK setup for channel client with dynamic selection
	sdkOptions := fabapi.Options{
		ConfigFile:      "../../../../test/fixtures/config/config_test.yaml",
		ProviderFactory: &DynamicSelectionProviderFactory{ChannelUsers: []ChannelUser{mychannelUser}},
	}

	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	selectionProvider.Initialize(sdk)

	selectionService, err = selectionProvider.NewSelectionService("mychannel")
	if err != nil {
		t.Fatalf("Failed to create new selection service for channel: %s", err)
	}

	if selectionProvider.lbp == nil {
		t.Fatalf("Default load balancing policy is nil")
	}

	if got, want := reflect.TypeOf(selectionProvider.lbp), reflect.TypeOf(pgresolver.NewRandomLBP()); got != want {
		t.Fatalf("Default load balancing policy is wrong type. Want %v, Got %v", want, got)
	}

	_, err = selectionService.GetEndorsersForChaincode(nil)
	if err == nil {
		t.Fatalf("Should have failed for no chaincode IDs provided")
	}

	_, err = selectionService.GetEndorsersForChaincode(nil, "")
	if err == nil {
		t.Fatalf("Should have failed since no channel peers are provided")
	}

	chPeers := []apifabclient.Peer{peer("p0", "Org1")}
	_, err = selectionService.GetEndorsersForChaincode(chPeers, "")
	if err == nil {
		t.Fatalf("Should have failed since empty cc ID provided")
	}

	_, err = selectionService.GetEndorsersForChaincode(chPeers, "abc")
	if err == nil {
		t.Fatalf("Should have failed for non-existent cc ID")
	}

	// Test custom load balancer
	selectionProvider, err = NewSelectionProvider(config, []ChannelUser{mychannelUser}, newCustomLBP())
	if err != nil {
		t.Fatalf("Failed to setup selection provider: %s", err)
	}

	if selectionProvider.lbp == nil {
		t.Fatalf("Failed to set load balancing policy")
	}

	// Check correct load balancer policy
	if got, want := reflect.TypeOf(selectionProvider.lbp), reflect.TypeOf(&customLBP{}); got != want {
		t.Fatalf("Failed to set load balancing policy. Want %v, Got %v", want, got)
	}

}

// DynamicSelectionProviderFactory is configured with dynamic (endorser) selection provider
type DynamicSelectionProviderFactory struct {
	defprovider.DefaultProviderFactory
	ChannelUsers []ChannelUser
}

// NewSelectionProvider returns a new implementation of dynamic selection provider
func (f *DynamicSelectionProviderFactory) NewSelectionProvider(config apiconfig.Config) (apifabclient.SelectionProvider, error) {
	return NewSelectionProvider(config, f.ChannelUsers, nil)
}

type customLBP struct {
}

// newCustomLBP returns a test load-balance policy
func newCustomLBP() pgresolver.LoadBalancePolicy {
	return &customLBP{}
}

func (lbp *customLBP) Choose(peerGroups []pgresolver.PeerGroup) pgresolver.PeerGroup {
	if len(peerGroups) == 0 {
		return pgresolver.NewPeerGroup()
	}
	return peerGroups[0]
}
