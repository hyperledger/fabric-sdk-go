/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	reqContext "context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	cpc "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

const (
	testTimeOut        = 20 * time.Second
	createAndSendError = "CreateAndSendTransaction failed"
	txError            = "MVCC_READ_CONFLICT"
	mockError          = "Mock fail"
)

func TestTransactionOptions(t *testing.T) {
	transient := make(map[string][]byte)
	transient["price"] = []byte("8500")

	c := mockChannelProvider("mychannel")

	gw := &Gateway{
		options: &gatewayOptions{
			Timeout: defaultTimeout,
		},
	}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")

	txn, err := contr.CreateTransaction(
		"txn1",
		WithTransient(transient),
		WithEndorsingPeers("peer1"),
		WithCollections("_implicit_org_org1"),
	)

	if err != nil {
		t.Fatalf("Failed to create transaction: %s", err)
	}

	data := txn.request.TransientMap["price"]
	if string(data) != "8500" {
		t.Fatalf("Incorrect transient data: %s", string(data))
	}

	endorsers := txn.endorsingPeers
	if endorsers[0] != "peer1" {
		t.Fatalf("Incorrect endorsing peer: %s", endorsers[0])
	}

	collections := txn.collections
	if collections[0] != "_implicit_org_org1" {
		t.Fatalf("Incorrect collection: %s", collections[0])
	}

	txn.Evaluate("arg1", "arg2")
	txn.Submit("arg1", "arg2")
}

func TestInitTransactionOptions(t *testing.T) {
	transient := make(map[string][]byte)
	transient["price"] = []byte("8500")

	c := mockChannelProvider("mychannel")

	gw := &Gateway{
		options: &gatewayOptions{
			Timeout: defaultTimeout,
		},
	}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")

	txn, err := contr.CreateTransaction(
		"txn1",
		WithTransient(transient),
		WithEndorsingPeers("peer1"),
		WithCollections("_implicit_org_org1"),
		WithInit(),
	)

	if err != nil {
		t.Fatalf("Failed to create transaction: %s", err)
	}

	data := txn.request.TransientMap["price"]
	if string(data) != "8500" {
		t.Fatalf("Incorrect transient data: %s", string(data))
	}

	endorsers := txn.endorsingPeers
	if endorsers[0] != "peer1" {
		t.Fatalf("Incorrect endorsing peer: %s", endorsers[0])
	}

	collections := txn.collections
	if collections[0] != "_implicit_org_org1" {
		t.Fatalf("Incorrect collection: %s", collections[0])
	}

	txn.Evaluate("arg1", "arg2")
	txn.Submit("arg1", "arg2")
}

func TestCommitEvent(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &Gateway{
		options: &gatewayOptions{
			Timeout: defaultTimeout,
		},
	}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")
	txn, err := contr.CreateTransaction("txn1")
	notifier := txn.RegisterCommitEvent()

	result, err := txn.Submit("arg1", "arg2")

	if err != nil {
		t.Fatalf("Failed to submit transaction: %s", err)
	}

	if string(result) != "abc" {
		t.Fatalf("Incorrect transaction result: %s", result)
	}

	var cEvent *fab.TxStatusEvent
	select {
	case cEvent = <-notifier:
		t.Logf("Received commit event: %#v\n", cEvent)
	case <-time.After(time.Second * 20):
		t.Fatal("Did NOT receive commit event\n")
	}

}

func TestSubmitHandlerTxCreateError(t *testing.T) {

	//Sample request
	request := invoke.Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, invoke.Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP",
		Status: 200, Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP",
		Status: 200, Payload: []byte("value1")}

	clientContext := setupChannelClientContext(nil, nil, []fab.Peer{mockPeer1, mockPeer2}, t)

	//Get commit handler
	commitHandler := &commitTxHandler{}
	//Perform action through handler
	commitHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), createAndSendError) {
		t.Fatal("Expected error: ", createAndSendError, ", Received error:", requestContext.Error.Error())
	}
}

func TestSubmitHandlerTxSendError(t *testing.T) {

	//Sample request
	request := invoke.Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, invoke.Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP",
		Status: 200, Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP",
		Status: 200, Payload: []byte("value1")}

	clientContext := setupChannelClientContext(nil, nil, []fab.Peer{mockPeer1, mockPeer2}, t)
	clientContext.Transactor = &mockTransactor{}

	//Get commit handler
	commitHandler := &commitTxHandler{}
	//Perform action through handler
	commitHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), mockError) {
		t.Fatal("Expected error: ", mockError, ", Received error:", requestContext.Error.Error())
	}
}

func TestSubmitHandlerCommitError(t *testing.T) {

	//Sample request
	request := invoke.Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, invoke.Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP",
		Status: 200, Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP",
		Status: 200, Payload: []byte("value1")}

	clientContext := setupChannelClientContext(nil, nil, []fab.Peer{mockPeer1, mockPeer2}, t)

	// add reponses to request context
	addProposalResponse(requestContext)
	clientContext.EventService.(*fcmocks.MockEventService).TxValidationCode = peer.TxValidationCode_MVCC_READ_CONFLICT

	//Get commit handler
	commitHandler := &commitTxHandler{}
	//Perform action through handler
	commitHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil {
		t.Fatal("Expected error, got none")
	}
	if !strings.Contains(requestContext.Error.Error(), txError) {
		t.Fatal("Expected error: ", txError, ", Received error:", requestContext.Error.Error())
	}

}

//prepareHandlerContexts prepares context objects for handlers
func prepareRequestContext(request invoke.Request, opts invoke.Opts, t *testing.T) *invoke.RequestContext {
	requestContext := &invoke.RequestContext{Request: request,
		Opts:     opts,
		Response: invoke.Response{},
		Ctx:      reqContext.Background(),
	}

	requestContext.Opts.Timeouts = make(map[fab.TimeoutType]time.Duration)
	requestContext.Opts.Timeouts[fab.Execute] = testTimeOut
	if opts.TargetFilter != nil {
		requestContext.SelectionFilter = func(peer fab.Peer) bool {
			return opts.TargetFilter.Accept(peer)
		}
	}
	if opts.TargetSorter != nil {
		requestContext.PeerSorter = func(peers []fab.Peer) []fab.Peer {
			return opts.TargetSorter.Sort(peers)
		}
	}

	return requestContext
}

func setupChannelClientContext(discErr error, selectionErr error, peers []fab.Peer, t *testing.T) *invoke.ClientContext {
	membership := fcmocks.NewMockMembership()

	ctx := setupTestContext()
	orderer := fcmocks.NewMockOrderer("", nil)
	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: "testChannel",
		Orderers:  []fab.Orderer{orderer},
	}

	return &invoke.ClientContext{
		Membership:   membership,
		Discovery:    txnmocks.NewMockDiscoveryService(discErr),
		Selection:    txnmocks.NewMockSelectionService(selectionErr, peers...),
		Transactor:   &transactor,
		EventService: fcmocks.NewMockEventService(),
	}

}

func setupTestContext() cpc.Client {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := fcmocks.NewMockContext(user)
	return ctx
}

func addProposalResponse(request *invoke.RequestContext) {
	r1 := &fab.TransactionProposalResponse{
		Endorser: "peer 1",
		Status:   http.StatusOK,
		ProposalResponse: &peer.ProposalResponse{Response: &peer.Response{
			Message: "test", Status: http.StatusOK, Payload: []byte("ResponsePayload")},
			Payload:     []byte("ProposalPayload1"),
			Endorsement: &peer.Endorsement{},
		}}
	p := &fab.TransactionProposal{
		Proposal: &peer.Proposal{},
	}
	request.Response = invoke.Response{
		Proposal:         p,
		Responses:        []*fab.TransactionProposalResponse{r1},
		TxValidationCode: peer.TxValidationCode_MVCC_READ_CONFLICT,
	}
}

type mockTransactor struct{}

func (t *mockTransactor) CreateTransaction(request fab.TransactionRequest) (*fab.Transaction, error) {
	return nil, nil
}

func (t *mockTransactor) SendTransaction(tx *fab.Transaction) (*fab.TransactionResponse, error) {
	return nil, errors.New("Mock fail")
}

func (t *mockTransactor) CreateTransactionHeader(opts ...fab.TxnHeaderOpt) (fab.TransactionHeader, error) {
	return nil, nil
}

func (t *mockTransactor) SendTransactionProposal(proposal *fab.TransactionProposal, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {
	return nil, nil
}
