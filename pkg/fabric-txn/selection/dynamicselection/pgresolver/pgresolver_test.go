/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package pgresolver

import (
	"testing"

	sdkApi "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	common "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

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

var p1 = peer("peer1")
var p2 = peer("peer2")
var p3 = peer("peer3")
var p4 = peer("peer4")
var p5 = peer("peer5")
var p6 = peer("peer6")
var p7 = peer("peer7")
var p8 = peer("peer8")
var p9 = peer("peer9")
var p10 = peer("peer10")
var p11 = peer("peer11")
var p12 = peer("peer12")

var peersByMSPID = map[string][]sdkApi.Peer{
	org1: peers(p1, p2),
	org2: peers(p3, p4),
	org3: peers(p5, p6, p7),
	org4: peers(p8, p9, p10),
	org5: peers(p11, p12),
}

var configImp = mocks.NewMockConfig()

const (
	o1 = iota
	o2
	o3
	o4
	o5
)

func TestPeerGroupResolverPolicyNoAvailablePeers(t *testing.T) {
	signedBy, identities, err := GetPolicies(org1)
	if err != nil {
		panic(err)
	}

	sigPolicyEnv := &common.SignaturePolicyEnvelope{
		Version: 0, Rule: signedBy[o1], Identities: identities,
	}

	expected := []PeerGroup{}

	testPeerGroupResolver(
		t, sigPolicyEnv,
		func(mspID string) []sdkApi.Peer {
			return nil
		},
		expected)
}

// 1 of [(2 of [1,2]),(2 of [1,3,4])]
func TestPeerGroupResolverPolicy1(t *testing.T) {
	signedBy, identities, err := GetPolicies(org1, org2, org3, org4)
	if err != nil {
		panic(err)
	}

	sigPolicyEnv := &common.SignaturePolicyEnvelope{
		Version: 0,
		Rule: NewNOutOfPolicy(1,
			NewNOutOfPolicy(2,
				signedBy[o1],
				signedBy[o2],
			),
			NewNOutOfPolicy(2,
				signedBy[o1],
				signedBy[o3],
				signedBy[o4],
			),
		),
		Identities: identities,
	}

	expected := []PeerGroup{
		// Org1 and Org2
		pg(p1, p3), pg(p1, p4), pg(p2, p3), pg(p2, p4),
		// Org1 and Org3
		pg(p1, p5), pg(p1, p6), pg(p1, p7), pg(p2, p5), pg(p2, p6), pg(p2, p7),
		// Org1 and Org4
		pg(p1, p8), pg(p1, p9), pg(p1, p10), pg(p2, p8), pg(p2, p9), pg(p2, p10),
		// Org3 and Org4
		pg(p5, p8), pg(p5, p9), pg(p5, p10), pg(p6, p8), pg(p6, p9), pg(p6, p10), pg(p7, p8), pg(p7, p9), pg(p7, p10),
	}

	testPeerGroupResolver(t, sigPolicyEnv, retrievePeersByMSPid, expected)
}

// 1 of [(2 of [1,2]),(3 of [3,4,5])]
func TestPeerGroupResolverPolicy2(t *testing.T) {
	signedBy, identities, err := GetPolicies(org1, org2, org3, org4, org5)
	if err != nil {
		panic(err)
	}

	sigPolicyEnv := &common.SignaturePolicyEnvelope{
		Version: 0,
		Rule: NewNOutOfPolicy(1,
			NewNOutOfPolicy(2,
				signedBy[o1],
				signedBy[o2],
			),
			NewNOutOfPolicy(3,
				signedBy[o3],
				signedBy[o4],
				signedBy[o5],
			),
		),
		Identities: identities,
	}

	expected := []PeerGroup{
		// Org1 and Org2
		pg(p1, p3), pg(p1, p4), pg(p2, p3), pg(p2, p4),
		// Org1 and Org2 and Org3
		pg(p5, p8, p11), pg(p5, p8, p12), pg(p5, p9, p11), pg(p5, p9, p12), pg(p5, p10, p11), pg(p5, p10, p12),
		pg(p6, p8, p11), pg(p6, p8, p12), pg(p6, p9, p11), pg(p6, p9, p12), pg(p6, p10, p11), pg(p6, p10, p12),
		pg(p7, p8, p11), pg(p7, p8, p12), pg(p7, p9, p11), pg(p7, p9, p12), pg(p7, p10, p11), pg(p7, p10, p12),
	}

	testPeerGroupResolver(t, sigPolicyEnv, retrievePeersByMSPid, expected)
}

// 2 of [(1 of [1,2]),(1 of [3,4,5])]
func TestPeerGroupResolverPolicy3(t *testing.T) {
	signedBy, identities, err := GetPolicies(org1, org2, org3, org4, org5)
	if err != nil {
		panic(err)
	}

	sigPolicyEnv := &common.SignaturePolicyEnvelope{
		Version: 0,
		Rule: NewNOutOfPolicy(2,
			NewNOutOfPolicy(1,
				signedBy[o1],
				signedBy[o2],
			),
			NewNOutOfPolicy(1,
				signedBy[o3],
				signedBy[o4],
				signedBy[o5],
			),
		),
		Identities: identities,
	}

	expected := []PeerGroup{
		// (Org1 or Org2) and (Org3 or Org4 or Org5)
		pg(p1, p5), pg(p1, p6), pg(p1, p7), pg(p1, p8), pg(p1, p9), pg(p1, p10), pg(p1, p11), pg(p1, p12),
		pg(p2, p5), pg(p2, p6), pg(p2, p7), pg(p2, p8), pg(p2, p9), pg(p2, p10), pg(p2, p11), pg(p2, p12),
		pg(p3, p5), pg(p3, p6), pg(p3, p7), pg(p3, p8), pg(p3, p9), pg(p3, p10), pg(p3, p11), pg(p3, p12),
		pg(p4, p5), pg(p4, p6), pg(p4, p7), pg(p4, p8), pg(p4, p9), pg(p4, p10), pg(p4, p11), pg(p4, p12),
	}
	testPeerGroupResolver(t, sigPolicyEnv, retrievePeersByMSPid, expected)
}

// 2 of [1,2,(2 of [3,4,5])]
func TestPeerGroupResolverPolicy4(t *testing.T) {
	signedBy, identities, err := GetPolicies(org1, org2, org3, org4)
	if err != nil {
		panic(err)
	}

	sigPolicyEnv := &common.SignaturePolicyEnvelope{
		Version: 0,
		Rule: NewNOutOfPolicy(1,
			signedBy[0],
			NewNOutOfPolicy(2,
				signedBy[1],
				NewNOutOfPolicy(1,
					signedBy[2],
					signedBy[3],
				),
			),
		),
		Identities: identities,
	}

	expected := []PeerGroup{
		// O1
		pg(p1),
		pg(p2),
		// O2 and O3
		pg(p3, p5), pg(p3, p6), pg(p3, p7),
		pg(p4, p5), pg(p4, p6), pg(p4, p7),
		// O2 and O4
		pg(p3, p8), pg(p3, p9), pg(p3, p10),
		pg(p4, p8), pg(p4, p9), pg(p4, p10),
	}
	testPeerGroupResolver(t, sigPolicyEnv, retrievePeersByMSPid, expected)
}

// 1 of [1,(2 of [2,(1 of [3,4])])]
func TestPeerGroupResolverPolicy5(t *testing.T) {
	signedBy, identities, err := GetPolicies(org1, org2, org3, org4, org5)
	if err != nil {
		panic(err)
	}

	sigPolicyEnv := &common.SignaturePolicyEnvelope{
		Version: 0,
		Rule: NewNOutOfPolicy(2,
			signedBy[o1],
			signedBy[o2],
			NewNOutOfPolicy(2,
				signedBy[o3],
				signedBy[o4],
				signedBy[o5],
			),
		),
		Identities: identities,
	}

	expected := []PeerGroup{
		// O1 and O2
		pg(p1, p3), pg(p1, p4),
		pg(p2, p3), pg(p2, p4),

		// O1 and (2 of (3,4,5))
		pg(p1, p5, p8), pg(p1, p5, p9), pg(p1, p5, p10),
		pg(p1, p6, p8), pg(p1, p6, p9), pg(p1, p6, p10),
		pg(p1, p7, p8), pg(p1, p7, p9), pg(p1, p7, p10),
		pg(p1, p5, p11), pg(p1, p5, p12),
		pg(p1, p6, p11), pg(p1, p6, p12),
		pg(p1, p7, p11), pg(p1, p7, p12),
		pg(p1, p6, p11), pg(p1, p6, p12),
		pg(p1, p6, p11), pg(p1, p6, p12),
		pg(p1, p7, p11), pg(p1, p7, p12),
		pg(p1, p8, p11), pg(p1, p8, p12),
		pg(p1, p9, p11), pg(p1, p9, p12),
		pg(p1, p10, p11), pg(p1, p10, p12),
		pg(p2, p5, p8), pg(p2, p5, p9), pg(p2, p5, p10),
		pg(p2, p6, p8), pg(p2, p6, p9), pg(p2, p6, p10),
		pg(p2, p7, p8), pg(p2, p7, p9), pg(p2, p7, p10),
		pg(p2, p5, p11), pg(p2, p5, p12),
		pg(p2, p6, p11), pg(p2, p6, p12),
		pg(p2, p7, p11), pg(p2, p7, p12),
		pg(p2, p6, p11), pg(p2, p6, p12),
		pg(p2, p6, p11), pg(p2, p6, p12),
		pg(p2, p7, p11), pg(p2, p7, p12),
		pg(p2, p8, p11), pg(p2, p8, p12),
		pg(p2, p9, p11), pg(p2, p9, p12),
		pg(p2, p10, p11), pg(p2, p10, p12),

		// O2 and (2 of (3,4,5))
		pg(p3, p5, p8), pg(p3, p5, p9), pg(p3, p5, p10),
		pg(p3, p6, p8), pg(p3, p6, p9), pg(p3, p6, p10),
		pg(p3, p7, p8), pg(p3, p7, p9), pg(p3, p7, p10),
		pg(p3, p5, p11), pg(p3, p5, p12),
		pg(p3, p6, p11), pg(p3, p6, p12),
		pg(p3, p7, p11), pg(p3, p7, p12),
		pg(p3, p6, p11), pg(p3, p6, p12),
		pg(p3, p6, p11), pg(p3, p6, p12),
		pg(p3, p7, p11), pg(p3, p7, p12),
		pg(p3, p8, p11), pg(p3, p8, p12),
		pg(p3, p9, p11), pg(p3, p9, p12),
		pg(p3, p10, p11), pg(p3, p10, p12),
		pg(p4, p5, p8), pg(p4, p5, p9), pg(p4, p5, p10),
		pg(p4, p6, p8), pg(p4, p6, p9), pg(p4, p6, p10),
		pg(p4, p7, p8), pg(p4, p7, p9), pg(p4, p7, p10),
		pg(p4, p5, p11), pg(p4, p5, p12),
		pg(p4, p6, p11), pg(p4, p6, p12),
		pg(p4, p7, p11), pg(p4, p7, p12),
		pg(p4, p6, p11), pg(p4, p6, p12),
		pg(p4, p6, p11), pg(p4, p6, p12),
		pg(p4, p7, p11), pg(p4, p7, p12),
		pg(p4, p8, p11), pg(p4, p8, p12),
		pg(p4, p9, p11), pg(p4, p9, p12),
		pg(p4, p10, p11), pg(p4, p10, p12),
	}

	testPeerGroupResolver(t, sigPolicyEnv, retrievePeersByMSPid, expected)
}

func testPeerGroupResolver(t *testing.T, sigPolicyEnv *common.SignaturePolicyEnvelope, peerRetriever PeerRetriever, expected []PeerGroup) {

	pgResolver, err := NewRoundRobinPeerGroupResolver(sigPolicyEnv, peerRetriever)
	if err != nil {
		t.Fatal(err)
	}
	verify(t, pgResolver, expected)
}

func peer(name string) sdkApi.Peer {
	mp := mocks.NewMockPeer(name, "localhost:7051")
	return mp
}

func peers(peers ...sdkApi.Peer) []sdkApi.Peer {
	return peers
}

func verify(t *testing.T, pgResolver PeerGroupResolver, expectedPeerGroups []PeerGroup) {
	for i := 0; i < len(expectedPeerGroups); i++ {
		peerGroup := pgResolver.Resolve()
		if !containsPeerGroup(expectedPeerGroups, peerGroup) {
			t.Fatalf("peer group %s is not one of the expected peer groups: %v", peerGroup, expectedPeerGroups)
		}
	}
}

func pg(peers ...sdkApi.Peer) PeerGroup {
	return NewPeerGroup(peers...)
}

func retrievePeersByMSPid(mspID string) []sdkApi.Peer {
	return peersByMSPID[mspID]
}

func containsPeerGroup(groups []PeerGroup, group PeerGroup) bool {
	for _, g := range groups {
		if containsAllPeers(group, g) {
			return true
		}
	}
	return false
}

func containsAllPeers(pg1 PeerGroup, pg2 PeerGroup) bool {
	if len(pg1.Peers()) != len(pg2.Peers()) {
		return false
	}
	for _, p1 := range pg1.Peers() {
		if !containsPeer(pg2.Peers(), p1) {
			return false
		}
	}
	return true
}

func containsPeer(pg []sdkApi.Peer, p sdkApi.Peer) bool {
	for _, peer := range pg {
		if peer.Name() == p.Name() {
			return true
		}
	}
	return false
}
