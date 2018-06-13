/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package invoke

import (
	reqContext "context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

const (
	testTimeOut              = 20 * time.Second
	selectionServiceError    = "Selection service error"
	endorsementMisMatchError = "ProposalResponsePayloads do not match"
)

func TestQueryHandlerSuccess(t *testing.T) {

	//Sample request
	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupChannelClientContext(nil, nil, []fab.Peer{mockPeer1, mockPeer2}, t)

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
	request := Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupChannelClientContext(nil, nil, []fab.Peer{mockPeer1, mockPeer2}, t)

	//Prepare mock eventhub
	mockEventService := fcmocks.NewMockEventService()
	clientContext.EventService = mockEventService

	go func() {
		select {
		case txStatusReg := <-mockEventService.TxStatusRegCh:
			txStatusReg.Eventch <- &fab.TxStatusEvent{TxID: txStatusReg.TxID, TxValidationCode: pb.TxValidationCode_VALID}
		case <-time.After(requestContext.Opts.Timeouts[fab.Execute]):
			panic("Execute handler : time out not expected")
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
	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, Opts{}, t)

	//Get query handler
	queryHandler := NewQueryHandler()

	//Error Scenario 1
	clientContext := setupChannelClientContext(nil, errors.New(selectionServiceError), nil, t)

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), selectionServiceError) {
		t.Fatal("Expected error: ", selectionServiceError, ", Received error:", requestContext.Error.Error())
	}

	//Error Scenario 2 different payload return
	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200,
		Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200,
		Payload: []byte("value1")}

	clientContext = setupChannelClientContext(nil, nil, []fab.Peer{mockPeer1, mockPeer2}, t)

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), endorsementMisMatchError) {
		t.Fatal("Expected error: ", endorsementMisMatchError, ", Received error:", requestContext.Error.Error())
	}
}

func TestExecuteTxHandlerErrors(t *testing.T) {

	//Sample request
	request := Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP",
		Status: 200, Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP",
		Status: 200, Payload: []byte("value1")}

	clientContext := setupChannelClientContext(nil, nil, []fab.Peer{mockPeer1, mockPeer2}, t)

	//Get query handler
	executeHandler := NewExecuteHandler()
	//Perform action through handler
	executeHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), endorsementMisMatchError) {
		t.Fatal("Expected error: ", endorsementMisMatchError, ", Received error:", requestContext.Error.Error())
	}
}

func TestEndorsementHandler(t *testing.T) {
	request := Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	clientContext := setupChannelClientContext(nil, nil, nil, t)
	requestContext := prepareRequestContext(request, Opts{Targets: nil}, t)

	handler := NewEndorsementHandler()
	handler.Handle(requestContext, clientContext)
	assert.NotNil(t, requestContext.Error)

	requestContext = prepareRequestContext(request, Opts{Targets: []fab.Peer{fcmocks.NewMockPeer("p2", "")}}, t)

	handler = NewEndorsementHandler()
	handler.Handle(requestContext, clientContext)
	assert.Nil(t, requestContext.Error)

}

// Target filter
type filter struct {
	peer fab.Peer
}

func (f *filter) Accept(p fab.Peer) bool {
	return p.URL() == f.peer.URL()
}

func TestResponseValidation(t *testing.T) {
	p1 := &fab.TransactionProposalResponse{
		Endorser: "peer 1",
		Status:   http.StatusOK,
		ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{
			Message: "test", Status: http.StatusOK, Payload: []byte("ResponsePayload")},
			Payload: []byte("ProposalPayload1"),
		}}
	p2 := &fab.TransactionProposalResponse{
		Endorser: "peer 1",
		Status:   http.StatusOK,
		ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{
			Message: "test", Status: http.StatusOK, Payload: []byte("ResponsePayload")},
			Payload: []byte("ProposalPayload2"),
		}}
	h := EndorsementValidationHandler{}
	err := h.validate([]*fab.TransactionProposalResponse{p1, p2})
	assert.NotNil(t, err, "expected error with different response payloads")
	s, ok := status.FromError(err)
	assert.True(t, ok, "expected status error")
	assert.EqualValues(t, int32(status.EndorsementMismatch), s.Code, "expected endorsement mismatch")
}

func TestProposalProcessorHandlerError(t *testing.T) {
	peer1 := fcmocks.NewMockPeer("p1", "peer1:7051")
	peer2 := fcmocks.NewMockPeer("p2", "peer2:7051")
	discoveryPeers := []fab.Peer{peer1, peer2}

	//Get query handler
	handler := NewProposalProcessorHandler()

	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	selectionErr := errors.New("Some selection error")
	requestContext := prepareRequestContext(request, Opts{}, t)
	handler.Handle(requestContext, setupChannelClientContext(nil, selectionErr, discoveryPeers, t))
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), selectionErr.Error()) {
		t.Fatal("Expected error: ", selectionErr, ", Received error:", requestContext.Error)
	}
}

func TestProposalProcessorHandlerPassDirectly(t *testing.T) {
	peer1 := fcmocks.NewMockPeer("p1", "peer1:7051")
	peer2 := fcmocks.NewMockPeer("p2", "peer2:7051")
	discoveryPeers := []fab.Peer{peer1, peer2}

	//Get query handler
	handler := NewProposalProcessorHandler()

	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	// Directly pass in the proposal processors. In this case it should use those directly
	requestContext := prepareRequestContext(request, Opts{Targets: []fab.Peer{peer2}}, t)
	handler.Handle(requestContext, setupChannelClientContext(nil, nil, discoveryPeers, t))
	if requestContext.Error != nil {
		t.Fatalf("Got error: %s", requestContext.Error)
	}
	if len(requestContext.Opts.Targets) != 1 {
		t.Fatalf("Expecting 1 proposal processor but got %d", len(requestContext.Opts.Targets))
	}
	if requestContext.Opts.Targets[0] != peer2 {
		t.Fatal("Didn't get expected peers")
	}
}

func TestProposalProcessorHandler(t *testing.T) {
	peer1 := fcmocks.NewMockPeer("p1", "peer1:7051")
	peer2 := fcmocks.NewMockPeer("p2", "peer2:7051")
	discoveryPeers := []fab.Peer{peer1, peer2}

	handler := NewProposalProcessorHandler()
	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	requestContext := prepareRequestContext(request, Opts{}, t)
	handler.Handle(requestContext, setupChannelClientContext(nil, nil, discoveryPeers, t))
	if requestContext.Error != nil {
		t.Fatalf("Got error: %s", requestContext.Error)
	}
	if len(requestContext.Opts.Targets) != len(discoveryPeers) {
		t.Fatalf("Expecting %d proposal processors but got %d", len(discoveryPeers), len(requestContext.Opts.Targets))
	}
	if requestContext.Opts.Targets[0] != peer1 || requestContext.Opts.Targets[1] != peer2 {
		t.Fatal("Didn't get expected peers")
	}

	requestContext = prepareRequestContext(request, Opts{TargetFilter: &filter{peer: peer2}}, t)
	handler.Handle(requestContext, setupChannelClientContext(nil, nil, discoveryPeers, t))
	if requestContext.Error != nil {
		t.Fatalf("Got error: %s", requestContext.Error)
	}
	if len(requestContext.Opts.Targets) != 1 {
		t.Fatalf("Expecting 1 proposal processor but got %d", len(requestContext.Opts.Targets))
	}
	if requestContext.Opts.Targets[0] != peer2 {
		t.Fatal("Didn't get expected peers")
	}
}

func TestNewChaincodeCalls(t *testing.T) {
	ccID1 := "cc1"
	ccID2 := "cc2"
	col1 := "col1"
	col2 := "col2"

	request := Request{
		ChaincodeID: ccID1,
		Fcn:         "invoke",
		Args:        [][]byte{[]byte("query"), []byte("b")},
		InvocationChain: []*fab.ChaincodeCall{
			{
				ID:          ccID2,
				Collections: []string{col1},
			},
		},
	}

	ccCalls := newChaincodeCalls(request)
	require.Truef(t, len(ccCalls) == 2, "expecting 2 CC calls")
	require.Equal(t, ccID1, ccCalls[0].ID)
	require.Equal(t, ccID2, ccCalls[1].ID)
	require.Emptyf(t, ccCalls[0].Collections, "expecting no collections for [%s]", ccID1)
	require.Truef(t, len(ccCalls[1].Collections) == 1, "expecting 1 collection for [%s]", ccID2)

	request = Request{
		ChaincodeID: ccID1,
		Fcn:         "invoke",
		Args:        [][]byte{[]byte("query"), []byte("b")},
		InvocationChain: []*fab.ChaincodeCall{
			{
				ID:          ccID1,
				Collections: []string{col1, col2},
			},
			{
				ID:          ccID2,
				Collections: []string{col1},
			},
		},
	}

	ccCalls = newChaincodeCalls(request)
	require.Truef(t, len(ccCalls) == 2, "expecting 2 CC calls")
	require.Equal(t, ccID1, ccCalls[0].ID)
	require.Equal(t, ccID2, ccCalls[1].ID)
	require.Truef(t, len(ccCalls[0].Collections) == 2, "expecting 2 collections for [%s]", ccID1)
	require.Truef(t, len(ccCalls[1].Collections) == 1, "expecting 1 collection for [%s]", ccID2)
}

//prepareHandlerContexts prepares context objects for handlers
func prepareRequestContext(request Request, opts Opts, t *testing.T) *RequestContext {
	requestContext := &RequestContext{Request: request,
		Opts:     opts,
		Response: Response{},
		Ctx:      reqContext.Background(),
	}

	requestContext.Opts.Timeouts = make(map[fab.TimeoutType]time.Duration)
	requestContext.Opts.Timeouts[fab.Execute] = testTimeOut
	if opts.TargetFilter != nil {
		requestContext.SelectionFilter = func(peer fab.Peer) bool {
			return opts.TargetFilter.Accept(peer)
		}
	}

	return requestContext
}

func setupChannelClientContext(discErr error, selectionErr error, peers []fab.Peer, t *testing.T) *ClientContext {
	membership := fcmocks.NewMockMembership()

	ctx := setupTestContext()
	orderer := fcmocks.NewMockOrderer("", nil)
	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: "testChannel",
		Orderers:  []fab.Orderer{orderer},
	}

	return &ClientContext{
		Membership: membership,
		Discovery:  txnmocks.NewMockDiscoveryService(discErr),
		Selection:  txnmocks.NewMockSelectionService(selectionErr, peers...),
		Transactor: &transactor,
	}

}

func setupTestContext() context.Client {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := fcmocks.NewMockContext(user)
	return ctx
}
