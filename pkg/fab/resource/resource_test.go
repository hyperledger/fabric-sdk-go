/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/pkg/errors"
)

const testAddress = "127.0.0.1:0"

func TestSignChannelConfig(t *testing.T) {
	ctx := setupContext()

	configTx, err := ioutil.ReadFile(filepath.Join(metadata.GetProjectPath(), metadata.ChannelConfigPath, "mychannel.tx"))
	if err != nil {
		t.Fatalf(err.Error())
	}

	_, err = SignChannelConfig(ctx, nil, nil)
	if err == nil {
		t.Fatal("Expected 'channel configuration required")
	}

	_, err = SignChannelConfig(ctx, configTx, nil)
	if err != nil {
		t.Fatalf("Expected 'channel configuration required %s", err)
	}
}

func TestCreateChannel(t *testing.T) {
	ctx := setupContext()

	configTx, err := ioutil.ReadFile(filepath.Join(metadata.GetProjectPath(), metadata.ChannelConfigPath, "mychannel.tx"))
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Setup mock orderer
	verifyBroadcast := make(chan *fab.SignedEnvelope)
	orderer := mocks.NewMockOrderer(fmt.Sprintf("0.0.0.0:1234"), verifyBroadcast)

	// Create channel without envelope
	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	_, err = CreateChannel(reqCtx, CreateChannelRequest{
		Orderer: orderer,
		Name:    "mychannel",
	})
	if err == nil {
		t.Fatal("Expected error creating channel without envelope")
	}

	// Create channel without orderer
	_, err = CreateChannel(reqCtx, CreateChannelRequest{
		Envelope: configTx,
		Name:     "mychannel",
	})
	if err == nil {
		t.Fatal("Expected error creating channel without orderer")
	}

	// Create channel without name
	_, err = CreateChannel(reqCtx, CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
	})
	if err == nil {
		t.Fatal("Expected error creating channel without name")
	}

	// Test with valid cofiguration
	request := CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
		Name:     "mychannel",
	}
	_, err = CreateChannel(reqCtx, request)
	if err != nil {
		t.Fatalf("Did not expect error from create channel. Got error: %s", err)
	}
	select {
	case b := <-verifyBroadcast:
		logger.Debugf("Verified broadcast: %v", b)
	case <-time.After(time.Second):
		t.Fatal("Expected broadcast")
	}
}

func TestJoinChannel(t *testing.T) {
	var peers []fab.ProposalProcessor

	srv := mocks.MockEndorserServer{}
	addr := srv.Start(testAddress)
	defer srv.Stop()

	peer, _ := peer.New(mocks.NewMockEndpointConfig(), peer.WithURL("grpc://"+addr), peer.WithInsecure())
	peers = append(peers, peer)

	ctx := setupContext()

	genesisBlock := mocks.NewSimpleMockBlock()

	request := JoinChannelRequest{}
	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	err := JoinChannel(reqCtx, request, peers)
	if err == nil {
		t.Fatal("Should not have been able to join channel because of missing GenesisBlock parameter")
	}

	// Test join channel with valid arguments
	request = JoinChannelRequest{
		GenesisBlock: genesisBlock,
	}
	err = JoinChannel(reqCtx, request, peers)
	if err != nil {
		t.Fatalf("Did not expect error from join channel. Got: %s", err)
	}

	// Test failed proposal error handling
	srv.ProposalError = errors.New("Test Error")
	request = JoinChannelRequest{}
	err = JoinChannel(reqCtx, request, peers)
	if err == nil {
		t.Fatal("Expected error")
	}
}

func setupContext() context.Client {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)
	return ctx
}

func TestQueryByChaincode(t *testing.T) {
	ctx := setupContext()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "peer1.example.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Status: 200}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	resp, err := queryChaincodeWithTarget(reqCtx, request, &peer, options{})
	if err != nil {
		t.Fatalf("Failed to query: %s", err)
	}
	expectedResp := []byte("A")

	if !bytes.Equal(resp, expectedResp) {
		t.Fatalf("Unexpected transaction proposal response: %v (expected %v)", resp, expectedResp)
	}
}

func TestQueryByChaincodeBadStatus(t *testing.T) {
	ctx := setupContext()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Status: 99}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	_, err := queryChaincodeWithTarget(reqCtx, request, &peer, options{})
	if err == nil {
		t.Fatal("expected failure due to bad status")
	}
}

func TestQueryByChaincodeError(t *testing.T) {
	ctx := setupContext()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Error: errors.New("error")}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}

	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	_, err := queryChaincodeWithTarget(reqCtx, request, &peer, options{})
	if err == nil {
		t.Fatal("expected failure due to error")
	}
}

func TestGenesisBlockOrdererErr(t *testing.T) {
	const channelName = "testchannel"
	ctx := setupContext()

	orderer := mocks.NewMockOrderer("", nil)
	orderer.EnqueueForSendDeliver(mocks.NewSimpleMockError())
	orderer.CloseQueue()

	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	_, err := GenesisBlockFromOrderer(reqCtx, channelName, orderer)

	if err == nil {
		t.Fatal("GenesisBlock Test supposed to fail with error")
	}
}

func TestGenesisBlock(t *testing.T) {
	const channelName = "testchannel"
	ctx := setupContext()

	orderer := mocks.NewMockOrderer("", nil)
	orderer.EnqueueForSendDeliver(
		mocks.NewSimpleMockBlock(),
		common.Status_SUCCESS,
	)
	orderer.CloseQueue()
	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	_, err := GenesisBlockFromOrderer(reqCtx, channelName, orderer)

	if err != nil {
		t.Fatalf("GenesisBlock failed: %s", err)
	}
}

func TestGenesisBlockWithRetry(t *testing.T) {
	const channelName = "testchannel"
	ctx := setupContext()

	orderer := mocks.NewMockOrderer("", nil)
	orderer.EnqueueForSendDeliver(
		status.New(status.OrdererServerStatus, int32(common.Status_SERVICE_UNAVAILABLE), "service unavailable", []interface{}{}),
	)
	orderer.EnqueueForSendDeliver(
		mocks.NewSimpleMockBlock(),
		common.Status_SUCCESS,
	)
	orderer.CloseQueue()

	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	block, err := GenesisBlockFromOrderer(reqCtx, channelName, orderer, WithRetry(retry.DefaultResMgmtOpts))

	if err != nil {
		t.Fatalf("GenesisBlock failed: %s", err)
	}
	t.Logf("Block [%#v]", block)
}

/*
// TODO - make a much shorter timeout for this test.
func TestGenesisBlockOrdererTimeout(t *testing.T) {
	const channelName = "testchannel"

	ctx := setupContext()
	orderer := mockcore.NewMockOrderer("", nil)

	_, err := GenesisBlockFromOrderer(ctx, channelName, orderer)

	//It should fail with timeout
	if err == nil || !strings.HasSuffix(err.Error(), "timeout waiting for response from orderer") {
		t.Fatal("GenesisBlock Test supposed to fail with timeout error")
	}
}*/

func TestGenesisBlockOrderer(t *testing.T) {
	const channelName = "testchannel"
	ctx := setupContext()

	orderer := mocks.NewMockOrderer("", nil)
	orderer.EnqueueForSendDeliver(mocks.NewSimpleMockError())
	orderer.CloseQueue()

	//Call get Genesis block
	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	_, err := GenesisBlockFromOrderer(reqCtx, channelName, orderer)

	//Expecting error
	if err == nil {
		t.Fatal("GenesisBlock Test supposed to fail with error")
	}
}
