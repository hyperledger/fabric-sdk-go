/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package invoke

import (
	reqContext "context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
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
	ccID1 := "test"
	ccID2 := "invokedcc"
	ccID3 := "lscc"
	ccID4 := "somescc"

	//Sample request
	request := Request{ChaincodeID: ccID1, Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	// Add a chaincode filter that will ignore ccID4 when examining the RWSet
	ccFilter := func(ccID string) bool {
		return ccID != ccID4
	}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, Opts{CCFilter: ccFilter}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}
	mockPeer1.SetRwSets(fcmocks.NewRwSet(ccID1), fcmocks.NewRwSet(ccID2), fcmocks.NewRwSet(ccID3), fcmocks.NewRwSet(ccID4))
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}
	mockPeer2.SetRwSets(mockPeer1.RwSets...)

	clientContext := setupChannelClientContext(nil, nil, []fab.Peer{mockPeer1, mockPeer2}, t)

	// Prepare mock event service
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

	t.Run("no targets, produces an error", func(t *testing.T) {
		clientContext := setupChannelClientContext(nil, nil, nil, t)
		requestContext := prepareRequestContext(request, Opts{Targets: nil}, t)
		handler := NewEndorsementHandler()

		handler.Handle(requestContext, clientContext)

		require.Error(t, requestContext.Error)
	})

	t.Run("with a target, runs without error", func(t *testing.T) {
		clientContext := setupChannelClientContext(nil, nil, nil, t)
		requestContext := prepareRequestContext(request, Opts{Targets: []fab.Peer{fcmocks.NewMockPeer("p2", "")}}, t)
		handler := NewEndorsementHandler()

		handler.Handle(requestContext, clientContext)

		require.NoError(t, requestContext.Error)
	})

	t.Run("calls opts provider", func(t *testing.T) {
		clientContext := setupChannelClientContext(nil, nil, nil, t)
		requestContext := prepareRequestContext(request, Opts{Targets: []fab.Peer{fcmocks.NewMockPeer("p2", "")}}, t)
		optsProviderCalled := false
		optsProvider := func() []fab.TxnHeaderOpt {
			optsProviderCalled = true
			return []fab.TxnHeaderOpt{
				fab.WithCreator([]byte("somecreator")),
				fab.WithNonce([]byte("somenonce")),
			}
		}
		handler := NewEndorsementHandlerWithOpts(nil, optsProvider)

		handler.Handle(requestContext, clientContext)
		require.NoError(t, requestContext.Error)

		require.True(t, optsProviderCalled, "expecting opts provider to be called")
	})

	t.Run("returns EndorserServerStatus error from Transactor", func(t *testing.T) {
		clientContext := setupChannelClientContext(nil, nil, nil, t)
		requestContext := prepareRequestContext(request, Opts{Targets: []fab.Peer{fcmocks.NewMockPeer("p2", "")}}, t)
		handler := NewEndorsementHandler()
		clientContext.Transactor.(*txnmocks.MockTransactor).Err = fmt.Errorf("error in simulation: failed to distribute private collection, txID 695560b")

		handler.Handle(requestContext, clientContext)

		s, ok := requestContext.Error.(*status.Status)
		require.True(t, ok)
		require.Equal(t, status.EndorserServerStatus, s.Group)
		require.Equal(t, status.PvtDataDisseminationFailed.ToInt32(), s.Code)
	})

	t.Run("returns error from simulation", func(t *testing.T) {
		clientContext := setupChannelClientContext(nil, nil, nil, t)
		errExpected := fmt.Errorf("error in simulation")
		requestContext := prepareRequestContext(request, Opts{Targets: []fab.Peer{fcmocks.NewMockPeer("p2", "")}}, t)
		handler := NewEndorsementHandler()
		clientContext.Transactor.(*txnmocks.MockTransactor).Err = errExpected

		handler.Handle(requestContext, clientContext)

		_, ok := requestContext.Error.(*status.Status)
		require.False(t, ok)
		require.EqualError(t, requestContext.Error, errExpected.Error())
	})

	t.Run("returns error deserializing proposal response payload", func(t *testing.T) {
		clientContext := setupChannelClientContext(nil, nil, nil, t)
		mockPeer := fcmocks.NewMockPeer("p2", "")
		mockPeer.ProposalResponsePayload = []byte("invalid serialized protobuf message")
		requestContext := prepareRequestContext(request, Opts{Targets: []fab.Peer{mockPeer}}, t)
		handler := NewEndorsementHandler()

		handler.Handle(requestContext, clientContext)

		require.Error(t, requestContext.Error)
		require.Contains(t, requestContext.Error.Error(), "failed to deserialize proposal response payload")
	})

	t.Run("returns error deserializing chaincode action", func(t *testing.T) {
		proposalResponsePayload := &pb.ProposalResponsePayload{
			Extension: []byte("invalid serialized protobuf message"),
		}
		proposalResponsePayloadBytes, err := proto.Marshal(proposalResponsePayload)
		require.NoError(t, err)

		clientContext := setupChannelClientContext(nil, nil, nil, t)
		mockPeer := fcmocks.NewMockPeer("p2", "")
		mockPeer.ProposalResponsePayload = proposalResponsePayloadBytes
		requestContext := prepareRequestContext(request, Opts{Targets: []fab.Peer{mockPeer}}, t)
		handler := NewEndorsementHandler()

		handler.Handle(requestContext, clientContext)

		require.Error(t, requestContext.Error)
		require.Contains(t, requestContext.Error.Error(), "failed to deserialize chaincode action")
	})
}

// Target filter
type filter struct {
	peer fab.Peer
}

func (f *filter) Accept(p fab.Peer) bool {
	return p.URL() == f.peer.URL()
}

// Target sorter
type sorter struct {
	preferredPeerIndex int
}

func (s *sorter) Sort(peers []fab.Peer) []fab.Peer {
	var sortedPeers []fab.Peer
	for i := s.preferredPeerIndex; i < len(peers); i++ {
		sortedPeers = append(sortedPeers, peers[i])
	}

	for i := 0; i < s.preferredPeerIndex; i++ {
		sortedPeers = append(sortedPeers, peers[i])
	}

	return sortedPeers
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
	peer3 := fcmocks.NewMockPeer("p3", "peer3:7051")
	discoveryPeers := []fab.Peer{peer1, peer2, peer3}

	handler := NewProposalProcessorHandler()
	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	t.Run("Basic", func(t *testing.T) {
		requestContext := prepareRequestContext(request, Opts{}, t)
		handler.Handle(requestContext, setupChannelClientContext(nil, nil, discoveryPeers, t))
		require.NoError(t, requestContext.Error)
		require.Equal(t, len(discoveryPeers), len(requestContext.Opts.Targets), "Unexpected number of proposal processors")
		assert.Falsef(t, requestContext.Opts.Targets[0] != peer1 || requestContext.Opts.Targets[1] != peer2, "Didn't get expected peers")
	})

	t.Run("Target Filter", func(t *testing.T) {
		requestContext := prepareRequestContext(request, Opts{TargetFilter: &filter{peer: peer2}}, t)
		handler.Handle(requestContext, setupChannelClientContext(nil, nil, discoveryPeers, t))
		require.NoError(t, requestContext.Error)
		require.Equal(t, 1, len(requestContext.Opts.Targets), "Unexpected number of proposal processors")
		assert.Equalf(t, peer2.URL(), requestContext.Opts.Targets[0].URL(), "Expecting [%s] but got [%s]", peer2.URL(), requestContext.Opts.Targets[0].URL())
	})

	t.Run("Target Sorter", func(t *testing.T) {
		for i := len(discoveryPeers) - 1; i >= 0; i-- {
			requestContext := prepareRequestContext(request, Opts{TargetSorter: &sorter{preferredPeerIndex: i}}, t)
			handler.Handle(requestContext, setupChannelClientContext(nil, nil, discoveryPeers, t))
			require.NoError(t, requestContext.Error)
			require.Equal(t, len(discoveryPeers), len(requestContext.Opts.Targets), "Unexpected number of proposal processors")
			assert.Equalf(t, discoveryPeers[i].URL(), requestContext.Opts.Targets[0].URL(), "Expecting [%s] to be the first target but got [%s]", discoveryPeers[i].URL(), requestContext.Opts.Targets[0].URL())
		}
	})
}

func TestNewInvocationChain(t *testing.T) {
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

	ccCalls := newInvocationChain(&RequestContext{Request: request})
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

	ccCalls = newInvocationChain(&RequestContext{Request: request})
	require.Truef(t, len(ccCalls) == 2, "expecting 2 CC calls")
	require.Equal(t, ccID1, ccCalls[0].ID)
	require.Equal(t, ccID2, ccCalls[1].ID)
	require.Truef(t, len(ccCalls[0].Collections) == 2, "expecting 2 collections for [%s]", ccID1)
	require.Truef(t, len(ccCalls[1].Collections) == 1, "expecting 1 collection for [%s]", ccID2)
}

func TestMergeInvocationChains(t *testing.T) {
	ccID1 := "cc1"
	ccID2 := "cc2"
	ccID3 := "cc3"
	col1 := "col1"
	col2 := "col2"
	col3 := "col3"

	ccCall1A := &fab.ChaincodeCall{ID: ccID1}
	ccCall1B := &fab.ChaincodeCall{ID: ccID2, Collections: []string{col1, col3}}

	ccCall2A := &fab.ChaincodeCall{ID: ccID1, Collections: []string{col1}}
	ccCall2B := &fab.ChaincodeCall{ID: ccID2, Collections: []string{col1, col2}}
	ccCall2C := &fab.ChaincodeCall{ID: ccID3}

	acceptAllFilter := func(ccID string) bool { return true }

	t.Run("No change to invocation chain", func(t *testing.T) {
		invocChain, changed := mergeInvocationChains([]*fab.ChaincodeCall{ccCall1A}, []*fab.ChaincodeCall{ccCall1A}, acceptAllFilter)
		assert.Falsef(t, changed, "Expecting invocation chain NOT to have changed")
		require.NotEmptyf(t, invocChain, "Invocation chain is empty")
		assert.Equalf(t, []*fab.ChaincodeCall{ccCall1A}, invocChain, "Expecting the invocation chain the be the same")
	})

	t.Run("Additional chaincodes and collections", func(t *testing.T) {
		invocChain, changed := mergeInvocationChains([]*fab.ChaincodeCall{ccCall1A, ccCall1B}, []*fab.ChaincodeCall{ccCall2A, ccCall2B, ccCall2C}, acceptAllFilter)
		assert.Truef(t, changed, "Expecting invocation chain to have changed")
		require.NotEmptyf(t, invocChain, "Invocation chain is empty")
		assert.Equalf(t, 3, len(invocChain), "Expecting 3 chaincode calls in the invocation chain")

		assertContainsAll := func(t *testing.T, expectedColls []string, colls []string, ccID string) {
			for _, coll := range expectedColls {
				assert.Containsf(t, colls, coll, ccID+" does not contain all collections")
			}
		}

		for _, ccCall := range invocChain {
			switch ccCall.ID {
			case ccID1:
				assertContainsAll(t, []string{col1}, ccCall.Collections, ccID1)
			case ccID2:
				assertContainsAll(t, []string{col1, col2, col3}, ccCall.Collections, ccID2)
			case ccID3:
				assertContainsAll(t, nil, ccCall.Collections, ccID3)
			}
		}
	})
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
	if opts.TargetSorter != nil {
		requestContext.PeerSorter = func(peers []fab.Peer) []fab.Peer {
			return opts.TargetSorter.Sort(peers)
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
