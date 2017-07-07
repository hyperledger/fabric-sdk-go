/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"strings"
	"testing"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

func TestGenesisBlock(t *testing.T) {
	var peers []fab.Peer
	channel, _ := setupTestChannel()
	peer, _ := peer.NewPeer(testAddress, mocks.NewMockConfig())
	peers = append(peers, peer)
	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(mocks.NewSimpleMockError())
	txid, _ := channel.ClientContext().NewTxnID()
	badtxid := apitxn.TransactionID{
		ID: txid.ID,
	}

	genesisBlockReq := &fab.GenesisBlockRequest{
		TxnID: txid,
	}

	channel.AddOrderer(orderer)

	//Call get Genesis block
	_, err := channel.GenesisBlock(genesisBlockReq)

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
	_, err = channel.GenesisBlock(genesisBlockReq)

	//It should fail with timeout
	if err == nil || !strings.HasSuffix(err.Error(), "Timeout waiting for response from orderer") {
		t.Fatal("GenesisBlock Test supposed to fail with timeout error")
	}

	//Validation test
	genesisBlockReq = &fab.GenesisBlockRequest{}
	_, err = channel.GenesisBlock(genesisBlockReq)

	if err == nil || err.Error() != "GenesisBlock - error: Missing txId input parameter with the required transaction identifier" {
		t.Fatal("validation on missing txID input parameter is not working as expected")
	}

	genesisBlockReq = &fab.GenesisBlockRequest{
		TxnID: badtxid,
	}
	_, err = channel.GenesisBlock(genesisBlockReq)

	if err == nil || err.Error() != "GenesisBlock - error: Missing nonce input parameter with the required single use number" {
		t.Fatal("validation on missing nonce input parameter is not working as expected")
	}

	channel.RemoveOrderer(orderer)

	genesisBlockReq = &fab.GenesisBlockRequest{
		TxnID: txid,
	}

	_, err = channel.GenesisBlock(genesisBlockReq)

	if err == nil || err.Error() != "GenesisBlock - error: Missing orderer assigned to this channel for the GenesisBlock request" {
		t.Fatal("validation on no ordererds on channel is not working as expected")
	}

}
