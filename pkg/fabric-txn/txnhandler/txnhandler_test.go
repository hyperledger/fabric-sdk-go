/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txnhandler

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/mocks"
)

const (
	testTimeOut              = 20 * time.Second
	discoveryServiceError    = "Discovery service error"
	selectionServiceError    = "Selection service error"
	endorsementMisMatchError = "ProposalResponsePayloads do not match"

	filterTxError = "Filter Tx error"
)

func TestQueryHandlerSuccess(t *testing.T) {

	//Sample request
	request := chclient.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupChannelClientContext(nil, nil, []apifabclient.Peer{mockPeer1, mockPeer2}, t)

	//Get query handler
	queryHandler := NewQueryHandler()

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Error != nil {
		t.Fatal("Query handler failed", requestContext.Error)
	}
}

func TestExecuteTxHandlerSuccess(t *testing.T) {
	//Sample request
	request := chclient.Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupChannelClientContext(nil, nil, []apifabclient.Peer{mockPeer1, mockPeer2}, t)

	//Prepare mock eventhub
	mockEventHub := fcmocks.NewMockEventHub()
	clientContext.EventHub = mockEventHub

	go func() {
		select {
		case callback := <-mockEventHub.RegisteredTxCallbacks:
			callback("txid", 0, nil)
		case <-time.After(requestContext.Opts.Timeout):
			t.Fatal("Execute handler : time out not expected")
		}
	}()

	//Get query handler
	executeHandler := NewExecuteHandler()
	//Perform action through handler
	executeHandler.Handle(requestContext, clientContext)
	assert.Nil(t, requestContext.Error)
}

func TestQueryHandlerErrors(t *testing.T) {

	//Error Scenario 1
	request := chclient.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)
	clientContext := setupChannelClientContext(errors.New(discoveryServiceError), nil, nil, t)

	//Get query handler
	queryHandler := NewQueryHandler()

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), discoveryServiceError) {
		t.Fatal("Expected error: ", discoveryServiceError, ", Received error:", requestContext.Error.Error())
	}

	//Error Scenario 2
	clientContext = setupChannelClientContext(nil, errors.New(selectionServiceError), nil, t)

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), selectionServiceError) {
		t.Fatal("Expected error: ", selectionServiceError, ", Received error:", requestContext.Error.Error())
	}

	//Error Scenario 3 different payload return
	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200,
		Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200,
		Payload: []byte("value1")}

	clientContext = setupChannelClientContext(nil, nil, []apifabclient.Peer{mockPeer1, mockPeer2}, t)

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), endorsementMisMatchError) {
		t.Fatal("Expected error: ", endorsementMisMatchError, ", Received error:", requestContext.Error.Error())
	}
}

func TestExecuteTxHandlerErrors(t *testing.T) {

	//Sample request
	request := chclient.Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP",
		Status: 200, Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP",
		Status: 200, Payload: []byte("value1")}

	clientContext := setupChannelClientContext(nil, nil, []apifabclient.Peer{mockPeer1, mockPeer2}, t)

	//Get query handler
	executeHandler := NewExecuteHandler()
	//Perform action through handler
	executeHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), endorsementMisMatchError) {
		t.Fatal("Expected error: ", endorsementMisMatchError, ", Received error:", requestContext.Error.Error())
	}
}

func TestEndorsementHandler(t *testing.T) {
	request := chclient.Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	requestContext := prepareRequestContext(request, chclient.Opts{ProposalProcessors: []apifabclient.ProposalProcessor{fcmocks.NewMockPeer("p2", "")}}, t)
	clientContext := setupChannelClientContext(nil, nil, nil, t)

	handler := NewEndorsementHandler()
	handler.Handle(requestContext, clientContext)
	assert.Nil(t, requestContext.Error)
}

func TestProposalProcessorHandler(t *testing.T) {
	peer1 := fcmocks.NewMockPeer("p1", "")
	peer2 := fcmocks.NewMockPeer("p2", "")
	discoveryPeers := []apifabclient.Peer{peer1, peer2}

	//Get query handler
	handler := NewProposalProcessorHandler()

	request := chclient.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	selectionErr := errors.New("Some selection error")
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)
	handler.Handle(requestContext, setupChannelClientContext(nil, selectionErr, discoveryPeers, t))
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), selectionErr.Error()) {
		t.Fatal("Expected error: ", selectionErr, ", Received error:", requestContext.Error)
	}

	requestContext = prepareRequestContext(request, chclient.Opts{}, t)
	handler.Handle(requestContext, setupChannelClientContext(nil, nil, discoveryPeers, t))
	if requestContext.Error != nil {
		t.Fatalf("Got error: %s", requestContext.Error)
	}
	if len(requestContext.Opts.ProposalProcessors) != len(discoveryPeers) {
		t.Fatalf("Expecting %d proposal processors but got %d", len(discoveryPeers), len(requestContext.Opts.ProposalProcessors))
	}
	if requestContext.Opts.ProposalProcessors[0] != peer1 || requestContext.Opts.ProposalProcessors[1] != peer2 {
		t.Fatalf("Didn't get expected peers")
	}

	// Directly pass in the proposal processors. In this case it should use those directly
	requestContext = prepareRequestContext(request, chclient.Opts{ProposalProcessors: []apifabclient.ProposalProcessor{peer2}}, t)
	handler.Handle(requestContext, setupChannelClientContext(nil, nil, discoveryPeers, t))
	if requestContext.Error != nil {
		t.Fatalf("Got error: %s", requestContext.Error)
	}
	if len(requestContext.Opts.ProposalProcessors) != 1 {
		t.Fatalf("Expecting 1 proposal processor but got %d", len(requestContext.Opts.ProposalProcessors))
	}
	if requestContext.Opts.ProposalProcessors[0] != peer2 {
		t.Fatalf("Didn't get expected peers")
	}
}

//prepareHandlerContexts prepares context objects for handlers
func prepareRequestContext(request chclient.Request, opts chclient.Opts, t *testing.T) *chclient.RequestContext {

	var requestContext *chclient.RequestContext

	requestContext = &chclient.RequestContext{Request: request,
		Opts:     opts,
		Response: chclient.Response{},
	}

	requestContext.Opts.Timeout = testTimeOut

	return requestContext

}

func setupTestChannel() (*channel.Channel, error) {
	ctx := setupTestContext()
	return channel.New(ctx, fcmocks.NewMockChannelCfg("testChannel"))
}

func setupTestContext() apifabclient.Context {
	user := fcmocks.NewMockUser("test")
	ctx := fcmocks.NewMockContext(user)
	return ctx
}

func setupChannelClientContext(discErr error, selectionErr error, peers []apifabclient.Peer, t *testing.T) *chclient.ClientContext {

	testChannel, err := setupTestChannel()
	if err != nil {
		t.Fatalf("Failed to setup test channel: %s", err)
	}

	// Add mock msp to msp manager
	msps := make(map[string]msp.MSP)
	msps["Org1MSP"] = fcmocks.NewMockMSP(nil)

	testChannel.SetMSPManager(fcmocks.NewMockMSPManager(msps))

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

	return &chclient.ClientContext{
		Channel:   testChannel,
		Discovery: discoveryService,
		Selection: selectionService,
	}

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
