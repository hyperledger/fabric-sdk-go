/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package filter

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

const channelID = "mychannel"

func TestInvalidOpt(t *testing.T) {

	channel, err := mocks.NewMockChannel(channelID)
	if err != nil {
		t.Fatalf("Failed to create mock channel: %s", err)
	}

	ef := NewEndpointFilter(channel, 10)

	peer := mocks.NewMockPeer("Peer1", "example.com")
	if !ef.Accept(peer) {
		t.Fatal("Should have accepted peer")
	}

}

func TestChaincodeQueryFilter(t *testing.T) {

	channel, err := mocks.NewMockChannel(channelID)
	if err != nil {
		t.Fatalf("Failed to create mock channel: %s", err)
	}

	ef := NewEndpointFilter(channel, ChaincodeQuery)

	if !ef.Accept(mocks.NewMockPeer("Peer1", "non-configured.com")) {
		t.Fatal("Should have accepted peer that is not configured")
	}

	// Configured peer
	peer := mocks.NewMockPeer("Peer1", "example.com")
	if !ef.Accept(peer) {
		t.Fatal("Should have accepted peer")
	}

	channel, err = mocks.NewMockChannel("noEndpoints")
	ef = NewEndpointFilter(channel, ChaincodeQuery)
	if err != nil {
		t.Fatalf("Failed to create mock channel: %s", err)
	}
	if ef.Accept(peer) {
		t.Fatal("Should NOT have accepted peer since peers chaincode query option is configured to false")
	}

	channel, err = mocks.NewMockChannel("noChannelPeers")
	ef = NewEndpointFilter(channel, ChaincodeQuery)
	if err != nil {
		t.Fatalf("Failed to create mock channel: %s", err)
	}
	if !ef.Accept(peer) {
		t.Fatal("Should have accepted peer since no peers configured")
	}

}

func TestLedgerQueryFilter(t *testing.T) {

	channel, err := mocks.NewMockChannel(channelID)
	if err != nil {
		t.Fatalf("Failed to create mock channel: %s", err)
	}

	ef := NewEndpointFilter(channel, LedgerQuery)

	peer := mocks.NewMockPeer("Peer1", "example.com")
	if !ef.Accept(peer) {
		t.Fatal("Should have accepted peer")
	}

}

func TestEndorsingPeerFilter(t *testing.T) {

	channel, err := mocks.NewMockChannel(channelID)
	if err != nil {
		t.Fatalf("Failed to create mock channel: %s", err)
	}

	ef := NewEndpointFilter(channel, EndorsingPeer)

	peer := mocks.NewMockPeer("Peer1", "example.com")
	if !ef.Accept(peer) {
		t.Fatal("Should have accepted peer")
	}

}

func TestEventSourceFilter(t *testing.T) {

	channel, err := mocks.NewMockChannel(channelID)
	if err != nil {
		t.Fatalf("Failed to create mock channel: %s", err)
	}

	ef := NewEndpointFilter(channel, EventSource)

	peer := mocks.NewMockPeer("Peer1", "example.com")
	if !ef.Accept(peer) {
		t.Fatal("Should have accepted peer")
	}

}
