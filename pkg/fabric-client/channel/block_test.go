/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"strings"
	"testing"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

func TestGenesisBlock(t *testing.T) {
	var peers []fab.Peer
	channel, _ := setupTestChannel()
	peer, _ := peer.New(mocks.NewMockConfig(), peer.WithURL(testAddress))
	peers = append(peers, peer)
	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(mocks.NewSimpleMockError())

	channel.AddOrderer(orderer)

	//Call get Genesis block
	_, err := channel.GenesisBlock()

	//Expecting error
	if err == nil {
		t.Fatal("GenesisBlock Test supposed to fail with error")
	}

	//Remove existing orderer
	channel.RemoveOrderer(orderer)

	//Create new orderer
	orderer = mocks.NewMockOrderer("", nil)

	channel.AddOrderer(orderer)

	//Call get Genesis block
	_, err = channel.GenesisBlock()

	//It should fail with timeout
	if err == nil || !strings.HasSuffix(err.Error(), "timeout waiting for response from orderer") {
		t.Fatal("GenesisBlock Test supposed to fail with timeout error")
	}

	//Validation test
	channel.RemoveOrderer(orderer)
	_, err = channel.GenesisBlock()

	if err == nil || !strings.Contains(err.Error(), "missing orderer assigned to this channel for the GenesisBlock request") {
		t.Fatal("validation on no ordererds on channel is not working as expected")
	}

}
