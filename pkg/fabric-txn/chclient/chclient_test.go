/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chclient

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/retry"

	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/status"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

const (
	testAddress = "127.0.0.1:47882"
)

func TestTxProposalResponseFilter(t *testing.T) {
	testErrorResponse := "internal error"
	// failed if status not 200
	testPeer1 := fcmocks.NewMockPeer("Peer1", "http://peer1.com")
	testPeer2 := fcmocks.NewMockPeer("Peer2", "http://peer2.com")
	testPeer2.Status = 500
	testPeer2.ResponseMessage = testErrorResponse

	peers := []apifabclient.Peer{testPeer1, testPeer2}
	chClient := setupChannelClient(peers, t)

	response := chClient.Query(apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if response.Error == nil {
		t.Fatalf("Should have failed for not success status")
	}
	statusError, ok := status.FromError(response.Error)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, common.Status_INTERNAL_SERVER_ERROR, status.ToPeerStatusCode(statusError.Code))
	assert.Equal(t, status.EndorserServerStatus, statusError.Group)
	assert.Equal(t, testErrorResponse, statusError.Message, "Expected response message from server")

	testPeer2.Payload = []byte("wrongPayload")
	testPeer2.Status = 200
	peers = []apifabclient.Peer{testPeer1, testPeer2}
	chClient = setupChannelClient(peers, t)
	response = chClient.Query(apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if response.Error == nil {
		t.Fatalf("Should have failed for not success status")
	}
	statusError, ok = status.FromError(response.Error)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, status.EndorsementMismatch, status.ToSDKStatusCode(statusError.Code))
	assert.Equal(t, status.EndorserClientStatus, statusError.Group)
	assert.Equal(t, "ProposalResponsePayloads do not match", statusError.Message, "Expected response message from server")
}

func TestQuery(t *testing.T) {

	chClient := setupChannelClient(nil, t)

	response := chClient.Query(apitxn.Request{})
	if response.Error == nil {
		t.Fatalf("Should have failed for empty query request")
	}

	response = chClient.Query(apitxn.Request{Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if response.Error == nil {
		t.Fatalf("Should have failed for empty chaincode ID")
	}

	response = chClient.Query(apitxn.Request{ChaincodeID: "testCC", Args: [][]byte{[]byte("query"), []byte("b")}})
	if response.Error == nil {
		t.Fatalf("Should have failed for empty function")
	}

	response = chClient.Query(apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if response.Error != nil {
		t.Fatalf("Failed to invoke test cc: %s", response.Error)
	}

	if response.Payload != nil {
		t.Fatalf("Expecting nil, got %s", response.Payload)
	}

}

func TestQueryDiscoveryError(t *testing.T) {
	chClient := setupChannelClientWithError(errors.New("Test Error"), nil, nil, t)

	response := chClient.Query(apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if response.Error == nil {
		t.Fatalf("Should have failed to query with error in discovery.GetPeers()")
	}
}

func TestQuerySelectionError(t *testing.T) {
	chClient := setupChannelClientWithError(nil, errors.New("Test Error"), nil, t)

	response := chClient.Query(apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if response.Error == nil {
		t.Fatalf("Should have failed to query with error in selection.GetEndorsersFor ...")
	}
}

func TestQueryWithOptSync(t *testing.T) {
	chClient := setupChannelClient(nil, t)

	response := chClient.Query(apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if response.Error != nil {
		t.Fatalf("Failed to invoke test cc: %s", response.Error)
	}

	if response.Payload != nil {
		t.Fatalf("Expecting nil, got %s", response.Payload)
	}
}

// TestQueryWithOptAsync demonstrates an example of an asynchronous query call
func TestQueryWithOptAsync(t *testing.T) {
	chClient := setupChannelClient(nil, t)

	notifier := make(chan apitxn.Response)
	go func() {
		notifier <- chClient.Query(apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	}()
	response := <-notifier
	if response.Error != nil {
		t.Fatalf("Failed to invoke test cc: %s", response.Error)
	}
	if response.Payload != nil {
		t.Fatalf("Expecting nil, got %s", response.Payload)
	}
}

func TestQueryWithOptTarget(t *testing.T) {
	chClient := setupChannelClient(nil, t)

	testPeer := fcmocks.NewMockPeer("Peer1", "http://peer1.com")

	peers := []apifabclient.Peer{testPeer}

	targets := peer.PeersToTxnProcessors(peers)

	response := chClient.Query(apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke",
		Args: [][]byte{[]byte("query"), []byte("b")}}, apitxn.WithProposalProcessor(targets...))
	if response.Error != nil {
		t.Fatalf("Failed to invoke test cc: %s", response.Error)
	}

	if response.Payload != nil {
		t.Fatalf("Expecting nil, got %s", response.Payload)
	}
}

func TestExecuteTx(t *testing.T) {
	chClient := setupChannelClient(nil, t)

	response := chClient.Execute(apitxn.Request{})
	if response.Error == nil {
		t.Fatalf("Should have failed for empty invoke request")
	}

	response = chClient.Execute(apitxn.Request{Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
	if response.Error == nil {
		t.Fatalf("Should have failed for empty chaincode ID")
	}

	response = chClient.Execute(apitxn.Request{ChaincodeID: "testCC", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
	if response.Error == nil {
		t.Fatalf("Should have failed for empty function")
	}

	// TODO: Test Valid Scenario with mocks
}

func TestExecuteTxDiscoveryError(t *testing.T) {
	chClient := setupChannelClientWithError(errors.New("Test Error"), nil, nil, t)

	response := chClient.Execute(apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke",
		Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
	if response.Error == nil {
		t.Fatalf("Should have failed to execute tx with error in discovery.GetPeers()")
	}
}

func TestExecuteTxSelectionError(t *testing.T) {
	chClient := setupChannelClientWithError(nil, errors.New("Test Error"), nil, t)

	response := chClient.Execute(apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke",
		Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
	if response.Error == nil {
		t.Fatalf("Should have failed to execute tx with error in selection.GetEndorserrsFor ...")
	}
}

// TestRPCErrorPropagation tests if status errors are wrapped and propagated from
// the lower level APIs to the high level channel client API
// This ensures that the status is not swallowed by calling error.Error()
func TestRPCStatusErrorPropagation(t *testing.T) {
	testErrMessage := "Test RPC Error"
	testStatus := status.New(status.EndorserClientStatus, status.ConnectionFailed.ToInt32(), testErrMessage, nil)

	testPeer1 := fcmocks.NewMockPeer("Peer1", "http://peer1.com")
	testPeer1.Error = testStatus
	chClient := setupChannelClient([]apifabclient.Peer{testPeer1}, t)

	response := chClient.Query(apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if response.Error == nil {
		t.Fatalf("Should have failed for not success status")
	}
	statusError, ok := status.FromError(response.Error)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, status.ConnectionFailed, status.ToSDKStatusCode(statusError.Code))
	assert.Equal(t, status.EndorserClientStatus, statusError.Group)
	assert.Equal(t, testErrMessage, statusError.Message, "Expected response message from server")
}

// TestOrdererStatusError ensures that status errors are propagated through
// the code execution paths from the low-level orderer broadcast APIs
func TestOrdererStatusError(t *testing.T) {
	testErrorMessage := "test error"

	testPeer1 := fcmocks.NewMockPeer("Peer1", "http://peer1.com")
	peers := []apifabclient.Peer{testPeer1}
	testOrderer1 := fcmocks.NewMockOrderer("", make(chan *apifabclient.SignedEnvelope))
	orderers := []apifabclient.Orderer{testOrderer1}
	chClient := setupChannelClientWithNodes(peers, orderers, t)
	chClient.eventHub = fcmocks.NewMockEventHub()

	mockOrderer, ok := testOrderer1.(fcmocks.MockOrderer)
	assert.True(t, ok, "Expected object to be mock orderer")
	mockOrderer.EnqueueSendBroadcastError(status.New(status.OrdererClientStatus,
		status.ConnectionFailed.ToInt32(), testErrorMessage, nil))

	response := chClient.Execute(apitxn.Request{ChaincodeID: "test", Fcn: "invoke",
		Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
	statusError, ok := status.FromError(response.Error)
	assert.True(t, ok, "Expected status error got %+v", response.Error)
	assert.EqualValues(t, status.ConnectionFailed, status.ToSDKStatusCode(statusError.Code))
	assert.Equal(t, status.OrdererClientStatus, statusError.Group)
	assert.Equal(t, testErrorMessage, statusError.Message, "Expected response message from server")

	chClient.Close()
}

func TestTransactionValidationError(t *testing.T) {
	validationCode := pb.TxValidationCode_BAD_RWSET
	mockEventHub := fcmocks.NewMockEventHub()
	testPeer1 := fcmocks.NewMockPeer("Peer1", "http://peer1.com")
	peers := []apifabclient.Peer{testPeer1}

	go func() {
		select {
		case callback := <-mockEventHub.RegisteredTxCallbacks:
			callback("txid", validationCode,
				status.New(status.EventServerStatus, int32(validationCode), "test", nil))
		case <-time.After(time.Second * 5):
			t.Fatal("Timed out waiting for execute Tx to register event callback")
		}
	}()

	chClient := setupChannelClient(peers, t)
	chClient.eventHub = mockEventHub
	response := chClient.Execute(apitxn.Request{ChaincodeID: "test", Fcn: "invoke",
		Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
	assert.Nil(t, response.Payload, "Expected nil result on failed execute operation")
	assert.NotNil(t, response.Error, "expected error")
	statusError, ok := status.FromError(response.Error)
	assert.True(t, ok, "Expected status error got %+v", response.Error)
	assert.EqualValues(t, validationCode, status.ToTransactionValidationCode(statusError.Code))
}

func TestExecuteTxWithRetries(t *testing.T) {
	testStatus := status.New(status.EndorserClientStatus, status.ConnectionFailed.ToInt32(), "test", nil)

	testPeer1 := fcmocks.NewMockPeer("Peer1", "http://peer1.com")
	testPeer1.Error = testStatus
	chClient := setupChannelClient([]apifabclient.Peer{testPeer1}, t)
	retryOpts := retry.DefaultOpts
	retryOpts.RetryableCodes = retry.ChannelClientRetryableCodes

	response := chClient.Query(apitxn.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}},
		apitxn.WithRetry(retryOpts))
	if response.Error == nil {
		t.Fatalf("Should have failed for not success status")
	}
	assert.Equal(t, retry.DefaultOpts.Attempts, testPeer1.ProcessProposalCalls-1, "Expected peer to be called (retry attempts + 1) times")
}

func setupTestChannel() (*channel.Channel, error) {
	ctx := setupTestContext()
	return channel.New(ctx, "testChannel")
}

func setupTestContext() apifabclient.Context {
	user := fcmocks.NewMockUser("test")
	ctx := fcmocks.NewMockContext(user)
	return ctx
}

func setupTestDiscovery(discErr error, peers []apifabclient.Peer) (apifabclient.DiscoveryService, error) {

	mockDiscovery, err := txnmocks.NewMockDiscoveryProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockDiscoveryProvider failed")
	}

	return mockDiscovery.NewDiscoveryService("mychannel")
}

func setupTestSelection(discErr error, peers []apifabclient.Peer) (apifabclient.SelectionService, error) {

	mockSelection, err := txnmocks.NewMockSelectionProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockSelectinProvider failed")
	}

	return mockSelection.NewSelectionService("mychannel")
}

func setupChannelClient(peers []apifabclient.Peer, t *testing.T) *ChannelClient {

	return setupChannelClientWithError(nil, nil, peers, t)
}

func setupChannelClientWithError(discErr error, selectionErr error, peers []apifabclient.Peer, t *testing.T) *ChannelClient {

	fabCtx := setupTestContext()

	testChannel, err := setupTestChannel()
	if err != nil {
		t.Fatalf("Failed to setup test channel: %s", err)
	}

	orderer := fcmocks.NewMockOrderer("", nil)
	testChannel.AddOrderer(orderer)

	discoveryService, err := setupTestDiscovery(discErr, nil)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	selectionService, err := setupTestSelection(selectionErr, peers)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	ctx := Context{
		ProviderContext:  fabCtx,
		DiscoveryService: discoveryService,
		SelectionService: selectionService,
		Channel:          testChannel,
	}
	ch, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	return ch
}

func setupChannelClientWithNodes(peers []apifabclient.Peer,
	orderers []apifabclient.Orderer, t *testing.T) *ChannelClient {

	fabCtx := setupTestContext()
	testChannel, err := setupTestChannel()
	assert.Nil(t, err, "Failed to setup test channel")

	for _, orderer := range orderers {
		err = testChannel.AddOrderer(orderer)
		assert.Nil(t, err, "Failed to add orderer %+v", orderer)
	}

	discoveryService, err := setupTestDiscovery(nil, nil)
	assert.Nil(t, err, "Failed to setup discovery service")

	selectionService, err := setupTestSelection(nil, peers)
	assert.Nil(t, err, "Failed to setup discovery service")

	ctx := Context{
		ProviderContext:  fabCtx,
		DiscoveryService: discoveryService,
		SelectionService: selectionService,
		Channel:          testChannel,
	}
	ch, err := New(ctx)
	assert.Nil(t, err, "Failed to create new channel client")

	return ch
}
