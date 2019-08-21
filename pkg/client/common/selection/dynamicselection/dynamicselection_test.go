/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicselection

import (
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection/pgresolver"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-protos-go/common"
)

const (
	org1 = "Org1MSP"
	org2 = "Org2MSP"
	org3 = "Org3MSP"
	org4 = "Org4MSP"
	org5 = "Org5MSP"
)

const (
	channel1 = "channel1"
	channel2 = "channel2"
)

const (
	cc1 = "cc1"
	cc2 = "cc2"
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

	channelPeers := []fab.Peer{p1, p2, p3, p4, p5, p6, p7, p8}

	service, err := newMockSelectionService(
		newMockCCDataProvider(channel1).
			add(cc1, getPolicy1()),
		pgresolver.NewRoundRobinLBP(),
		newMockDiscoveryService(channelPeers...),
	)
	if err != nil {
		t.Fatalf("got error creating selection service: %s", err)
	}
	// Channel1(Policy(cc1)) = Org1
	expected := []pgresolver.PeerGroup{
		// Org1
		pg(p1), pg(p2),
	}
	verify(t, service, expected, channel1, nil, cc1)
}

func TestGetEndorsersWithPeerFilter(t *testing.T) {

	channelPeers := []fab.Peer{p1, p2, p3, p4, p5, p6, p7, p8}

	service, err := newMockSelectionService(
		newMockCCDataProvider(channel1).
			add(cc1, getPolicy1()),
		pgresolver.NewRoundRobinLBP(),
		newMockDiscoveryService(channelPeers...),
	)
	if err != nil {
		t.Fatalf("got error creating selection service: %s", err)
	}

	// Channel1(Policy(cc1)) = Org1
	expected := []pgresolver.PeerGroup{
		// Org1
		pg(p1),
	}
	opts := []coptions.Opt{
		options.WithPeerFilter(func(peer fab.Peer) bool {
			return peer.URL() == p1.URL()
		}),
	}
	verify(t, service, expected, channel1, opts, cc1)

	// Channel1(Policy(cc1)) = Org1
	expected = []pgresolver.PeerGroup{
		// Org1
		pg(p2),
	}
	opts = []coptions.Opt{
		options.WithPeerFilter(func(peer fab.Peer) bool {
			return peer.URL() == p2.URL()
		}),
	}
	verify(t, service, expected, channel1, opts, cc1)
}

func TestGetEndorsersForChaincodeTwoCCs(t *testing.T) {
	channelPeers := []fab.Peer{p1, p2, p3, p4, p5, p6, p7, p8}

	service, err := newMockSelectionService(
		newMockCCDataProvider(channel1).
			add(cc1, getPolicy1()).
			add(cc2, getPolicy2()),
		pgresolver.NewRoundRobinLBP(),
		newMockDiscoveryService(channelPeers...),
	)
	if err != nil {
		t.Fatalf("got error creating selection service: %s", err)
	}

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
	verify(t, service, expected, channel1, nil, cc1, cc2)
}

func TestGetEndorsersForChaincodeTwoCCsTwoChannels(t *testing.T) {
	channel1Peers := []fab.Peer{p1, p2, p3, p4, p5, p6, p7, p8}

	service, err := newMockSelectionService(
		newMockCCDataProvider(channel1).
			add(cc1, getPolicy1()).
			add(cc2, getPolicy2()),
		pgresolver.NewRoundRobinLBP(),
		newMockDiscoveryService(channel1Peers...),
	)
	if err != nil {
		t.Fatalf("got error creating selection service: %s", err)
	}

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

	verify(t, service, expected, channel1, nil, cc1, cc2)

	channel2Peers := []fab.Peer{p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12}
	service, err = newMockSelectionService(
		newMockCCDataProvider(channel2).
			add(cc1, getPolicy3()).
			add(cc2, getPolicy2()),
		pgresolver.NewRoundRobinLBP(),
		newMockDiscoveryService(channel2Peers...),
	)
	if err != nil {
		t.Fatalf("got error creating selection service: %s", err)
	}

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

	verify(t, service, expected, channel2, nil, cc1, cc2)
}

func verify(t *testing.T, service fab.SelectionService, expectedPeerGroups []pgresolver.PeerGroup, channelID string, getEndorsersOpts []coptions.Opt, chaincodeIDs ...string) {
	// Set the log level to WARNING since the following spits out too much info in DEBUG
	module := "pg-resolver"
	level := logging.GetLevel(module)
	logging.SetLevel(module, logging.WARNING)
	defer logging.SetLevel(module, level)

	var chaincodes []*fab.ChaincodeCall
	for _, ccID := range chaincodeIDs {
		chaincodes = append(chaincodes, &fab.ChaincodeCall{ID: ccID})
	}

	for i := 0; i < len(expectedPeerGroups); i++ {
		peers, err := service.GetEndorsersForChaincode(chaincodes, getEndorsersOpts...)
		if err != nil {
			t.Fatalf("error getting endorsers: %s", err)
		}
		if !containsPeerGroup(expectedPeerGroups, peers) {
			t.Fatalf("peer group %s is not one of the expected peer groups: %+v", toString(peers), expectedPeerGroups)
		}
	}

}

func containsPeerGroup(groups []pgresolver.PeerGroup, peers []fab.Peer) bool {
	for _, g := range groups {
		if containsAllPeers(peers, g) {
			return true
		}
	}
	return false
}

func containsAllPeers(peers []fab.Peer, pg pgresolver.PeerGroup) bool {
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

func containsPeer(peers []fab.Peer, peer fab.Peer) bool {
	for _, p := range peers {
		if p.URL() == peer.URL() {
			return true
		}
	}
	return false
}

func pg(peers ...fab.Peer) pgresolver.PeerGroup {
	return pgresolver.NewPeerGroup(peers...)
}

func peer(name string, mspID string) fab.Peer {

	mp := mocks.NewMockPeer(name, name+":7051")
	mp.SetMSPID(mspID)
	return mp
}

func newMockSelectionService(ccPolicyProvider CCPolicyProvider, lbp pgresolver.LoadBalancePolicy, discoveryService fab.DiscoveryService) (fab.SelectionService, error) {
	context := mocks.NewMockContext(
		mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
	)
	service, err := newService(context, "testchannel", discoveryService,
		func() (CCPolicyProvider, error) {
			return ccPolicyProvider, nil
		},
		WithCacheTimeout(5*time.Second),
		WithLoadBalancePolicy(lbp),
	)
	if err != nil {
		return nil, err
	}
	service.discoveryService = discoveryService
	return service, nil
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

func toString(peers []fab.Peer) string {
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

type mockDiscoveryService struct {
	peers []fab.Peer
}

func newMockDiscoveryService(peers ...fab.Peer) fab.DiscoveryService {
	return &mockDiscoveryService{peers: peers}
}

func (s *mockDiscoveryService) GetPeers() ([]fab.Peer, error) {
	return s.peers, nil
}
