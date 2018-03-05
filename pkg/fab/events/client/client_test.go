// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	mockconn "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

const (
	InitialState ConnectionState = -1
)

var (
	peer1 = fabmocks.NewMockPeer("peer1", "grpcs://peer1.example.com:7051")
	peer2 = fabmocks.NewMockPeer("peer2", "grpcs://peer2.example.com:7051")
)

func TestConnect(t *testing.T) {
	connectionProvider := clientmocks.NewProviderFactory().Provider(
		clientmocks.NewMockConnection(
			clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory)),
		),
	)

	discoveryService := clientmocks.NewDiscoveryService(peer1, peer2)
	eventClient, _, err := newClientWithMockConnAndOpts("mychannel", newMockContext(), connectionProvider, filteredClientProvider, discoveryService, []options.Opt{})
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if eventClient.ConnectionState() != Disconnected {
		t.Fatalf("expecting connection state %s but got %s", Disconnected, eventClient.ConnectionState())
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting: %s", err)
	}
	if err := eventClient.Connect(); err == nil {
		t.Fatalf("expecting error connecting since the client is already connected")
	} else {
		t.Logf("Got expected error: %s", err)
	}
	time.Sleep(500 * time.Millisecond)
	if eventClient.ConnectionState() != Connected {
		t.Fatalf("expecting connection state %s but got %s", Connected, eventClient.ConnectionState())
	}
	eventClient.Close()
	if eventClient.ConnectionState() != Disconnected {
		t.Fatalf("expecting connection state %s but got %s", Disconnected, eventClient.ConnectionState())
	}
	time.Sleep(2 * time.Second)
}

func TestFailConnect(t *testing.T) {
	eventClient, _, err := newClientWithMockConnAndOpts(
		"mychannel", newMockContext(),
		mockconn.NewProviderFactory().Provider(
			mockconn.NewMockConnection(
				mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory)),
			),
		),
		failAfterConnectClientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		[]options.Opt{},
	)
	if err != nil {
		t.Fatalf("error creating client: %s", err)
	}
	if err := eventClient.Connect(); err == nil {
		t.Fatalf("expecting error connecting client but got none")
	}
}

func TestCallsOnClosedClient(t *testing.T) {
	eventClient, _, err := newClientWithMockConn(
		"mychannel", newMockContext(),
		filteredClientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	eventClient.Close()

	// Make sure you can call Close again with no issues
	eventClient.Close()

	if err := eventClient.Connect(); err == nil {
		t.Fatalf("expecting error connecting to closed channel event client but got none")
	}

	if _, _, err := eventClient.RegisterConnectionEvent(); err == nil {
		t.Fatalf("expecting error registering for connection events on closed channel event client but got none")
	}

	if _, _, err := eventClient.RegisterFilteredBlockEvent(); err == nil {
		t.Fatalf("expecting error registering for block events on closed channel event client but got none")
	}

	if _, _, err := eventClient.RegisterChaincodeEvent("ccid", "event"); err == nil {
		t.Fatalf("expecting error registering for chaincode events on closed channel event client but got none")
	}

	if _, _, err := eventClient.RegisterTxStatusEvent("txid"); err == nil {
		t.Fatalf("expecting error registering for TX events on closed channel event client but got none")
	}

	// Make sure the client doesn't panic when calling unregister on disconnected client
	eventClient.Unregister(nil)
}

func TestInvalidUnregister(t *testing.T) {
	channelID := "mychannel"
	eventClient, _, err := newClientWithMockConn(
		channelID, newMockContext(),
		filteredClientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	// Make sure the client doesn't panic with invalid registration
	eventClient.Unregister("invalid registration")
}

func TestUnauthorizedBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, _, err := newClientWithMockConn(
		channelID, newMockContext(),
		filteredClientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	if _, _, err := eventClient.RegisterBlockEvent(); err == nil {
		t.Fatalf("expecting error registering for block events on a filtered client")
	}
}

func TestBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClientWithMockConn(
		channelID, newMockContext(),
		clientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	registration1, eventch1, err := eventClient.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer eventClient.Unregister(registration1)

	registration2, eventch2, err := eventClient.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer eventClient.Unregister(registration2)

	conn.Ledger().NewBlock(channelID,
		servicemocks.NewTransaction("txID", pb.TxValidationCode_VALID, cb.HeaderType_ENDORSER_TRANSACTION),
	)

	numExpected := 2
	numReceived := 0
	for {
		select {
		case _, ok := <-eventch1:
			if !ok {
				t.Fatalf("unexpected closed channel")
			}
			numReceived++
		case _, ok := <-eventch2:
			if !ok {
				t.Fatalf("unexpected closed channel")
			}
			numReceived++
		case <-time.After(2 * time.Second):
			if numReceived != numExpected {
				t.Fatalf("Expecting %d block events but received %d", numExpected, numReceived)
			}
			return
		}
	}
}

func TestFilteredBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClientWithMockConn(
		channelID, newMockContext(),
		filteredClientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	registration1, eventch1, err := eventClient.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("error registering for filtered block events: %s", err)
	}
	defer eventClient.Unregister(registration1)

	registration2, eventch2, err := eventClient.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("error registering for filtered block events: %s", err)
	}
	defer eventClient.Unregister(registration2)

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	conn.Ledger().NewFilteredBlock(
		channelID,
		servicemocks.NewFilteredTx(txID1, txCode1),
		servicemocks.NewFilteredTx(txID2, txCode2),
	)

	numExpected := 2
	numReceived := 0
	for {
		select {
		case fbevent, ok := <-eventch1:
			if !ok {
				t.Fatalf("unexpected closed channel")
			}
			if fbevent.FilteredBlock == nil {
				t.Fatalf("Expecting filtered block but got nil")
			}
			if fbevent.FilteredBlock.ChannelId != channelID {
				t.Fatalf("Expecting channel [%s] but got [%s]", channelID, fbevent.FilteredBlock.ChannelId)
			}
			numReceived++
		case fbevent, ok := <-eventch2:
			if !ok {
				t.Fatalf("unexpected closed channel")
			}
			if fbevent.FilteredBlock == nil {
				t.Fatalf("Expecting filtered block but got nil")
			}
			if fbevent.FilteredBlock.ChannelId != channelID {
				t.Fatalf("Expecting channel [%s] but got [%s]", channelID, fbevent.FilteredBlock.ChannelId)
			}
			numReceived++
		case <-time.After(2 * time.Second):
			if numReceived != numExpected {
				t.Fatalf("Expecting %d filtered block events but received %d", numExpected, numReceived)
			}
			return
		}
	}
}

func TestBlockAndFilteredBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClientWithMockConn(
		channelID, newMockContext(),
		clientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	fbreg, fbeventch, err := eventClient.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("error registering for filtered block events: %s", err)
	}
	defer eventClient.Unregister(fbreg)

	breg, beventch, err := eventClient.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer eventClient.Unregister(breg)

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	tx1 := &pb.FilteredTransaction{
		Txid:             txID1,
		TxValidationCode: txCode1,
	}

	tx2 := &pb.FilteredTransaction{
		Txid:             txID2,
		TxValidationCode: txCode2,
	}

	conn.Ledger().NewBlock(channelID,
		servicemocks.NewTransaction(txID1, txCode1, cb.HeaderType_ENDORSER_TRANSACTION),
		servicemocks.NewTransaction(txID2, txCode2, cb.HeaderType_ENDORSER_TRANSACTION),
	)

	numReceived := 0
	numExpected := 2

	for {
		select {
		case fbevent, ok := <-fbeventch:
			if !ok {
				t.Fatalf("unexpected closed channel")
			}
			numReceived++
			checkFilteredBlock(t, fbevent.FilteredBlock, channelID, tx1, tx2)

		case _, ok := <-beventch:
			if !ok {
				t.Fatalf("unexpected closed channel")
			}
			numReceived++

		case <-time.After(2 * time.Second):
			if numReceived != numExpected {
				t.Fatalf("Expecting %d events but received %d", numExpected, numReceived)
			}
			return
		}
	}
}

func TestTxStatusEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClientWithMockConn(
		channelID, newMockContext(),
		filteredClientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	if _, _, err := eventClient.RegisterTxStatusEvent(""); err == nil {
		t.Fatalf("expecting error registering for TxStatus event without a TX ID but got none")
	}
	reg1, _, err := eventClient.RegisterTxStatusEvent(txID1)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	_, _, err = eventClient.RegisterTxStatusEvent(txID1)
	if err == nil {
		t.Fatalf("expecting error registering multiple times for TxStatus events: %s", err)
	}
	eventClient.Unregister(reg1)

	reg1, eventch1, err := eventClient.RegisterTxStatusEvent(txID1)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	defer eventClient.Unregister(reg1)

	reg2, eventch2, err := eventClient.RegisterTxStatusEvent(txID2)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	defer eventClient.Unregister(reg2)

	conn.Ledger().NewFilteredBlock(
		channelID,
		servicemocks.NewFilteredTx(txID1, txCode1),
		servicemocks.NewFilteredTx(txID2, txCode2),
	)

	numExpected := 2
	numReceived := 0
	done := false
	for !done {
		select {
		case event, ok := <-eventch1:
			if !ok {
				t.Fatalf("unexpected closed channel")
			} else {
				checkTxStatusEvent(t, event, txID1, txCode1)
				numReceived++
			}
		case event, ok := <-eventch2:
			if !ok {
				t.Fatalf("unexpected closed channel")
			} else {
				checkTxStatusEvent(t, event, txID2, txCode2)
				numReceived++
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for [%d] TxStatus events. Only received [%d]", numExpected, numReceived)
		}

		if numReceived == numExpected {
			break
		}
	}
}

func TestCCEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClientWithMockConn(
		channelID, newMockContext(),
		filteredClientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	ccID1 := "mycc1"
	ccID2 := "mycc2"
	ccFilter1 := "event1"
	ccFilter2 := "event.*"
	event1 := "event1"
	event2 := "event2"
	event3 := "event3"

	if _, _, err := eventClient.RegisterChaincodeEvent("", ccFilter1); err == nil {
		t.Fatalf("expecting error registering for chaincode events without CC ID but got none")
	}
	if _, _, err := eventClient.RegisterChaincodeEvent(ccID1, ""); err == nil {
		t.Fatalf("expecting error registering for chaincode events without event filter but got none")
	}
	if _, _, err := eventClient.RegisterChaincodeEvent(ccID1, ".(xxx"); err == nil {
		t.Fatalf("expecting error registering for chaincode events with invalid (regular expression) event filter but got none")
	}
	reg1, _, err := eventClient.RegisterChaincodeEvent(ccID1, ccFilter1)
	if err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}
	_, _, err = eventClient.RegisterChaincodeEvent(ccID1, ccFilter1)
	if err == nil {
		t.Fatalf("expecting error registering multiple times for chaincode events: %s", err)
	}
	eventClient.Unregister(reg1)

	reg1, eventch1, err := eventClient.RegisterChaincodeEvent(ccID1, ccFilter1)
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer eventClient.Unregister(reg1)

	reg2, eventch2, err := eventClient.RegisterChaincodeEvent(ccID2, ccFilter2)
	if err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}
	defer eventClient.Unregister(reg2)

	conn.Ledger().NewFilteredBlock(
		channelID,
		servicemocks.NewFilteredTxWithCCEvent("txid1", ccID1, event1),
		servicemocks.NewFilteredTxWithCCEvent("txid2", ccID2, event2),
		servicemocks.NewFilteredTxWithCCEvent("txid3", ccID2, event3),
	)

	numExpected := 3
	numReceived := 0
	done := false
	for !done {
		select {
		case event, ok := <-eventch1:
			if !ok {
				t.Fatalf("unexpected closed channel")
			} else {
				checkCCEvent(t, event, ccID1, event1)
				numReceived++
			}
		case event, ok := <-eventch2:
			if !ok {
				t.Fatalf("unexpected closed channel")
			} else {
				checkCCEvent(t, event, ccID2, event2, event3)
				numReceived++
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for [%d] CC events. Only received [%d]", numExpected, numReceived)
		}

		if numReceived == numExpected {
			break
		}
	}
}

// TestReconnect tests the ability of the Channel Event Client to retry multiple
// times to connect, and reconnect after it has disconnected.
func TestReconnect(t *testing.T) {
	// (1) Connect
	//     -> should fail to connect on the first and second attempt but succeed on the third attempt
	t.Run("#1", func(t *testing.T) {
		t.Parallel()
		testConnect(t, 3, mockconn.ConnectedOutcome,
			mockconn.NewConnectResults(
				mockconn.NewConnectResult(mockconn.ThirdAttempt, mockconn.SucceedResult),
			),
		)
	})

	// (1) Connect
	//     -> should fail to connect on the first attempt and no further attempts are to be made
	t.Run("#2", func(t *testing.T) {
		t.Parallel()
		testConnect(t, 1, mockconn.ErrorOutcome,
			mockconn.NewConnectResults(),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect
	//     -> should fail to reconnect on the first and second attempt but succeed on the third attempt
	t.Run("#3", func(t *testing.T) {
		t.Parallel()
		testReconnect(t, true, 3, mockconn.ReconnectedOutcome,
			mockconn.NewConnectResults(
				mockconn.NewConnectResult(mockconn.FirstAttempt, mockconn.SucceedResult),
				mockconn.NewConnectResult(mockconn.FourthAttempt, mockconn.SucceedResult),
			),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect
	//     -> should fail to reconnect after two attempts and then close
	t.Run("#4", func(t *testing.T) {
		t.Parallel()
		testReconnect(t, true, 2, mockconn.ClosedOutcome,
			mockconn.NewConnectResults(
				mockconn.NewConnectResult(mockconn.FirstAttempt, mockconn.SucceedResult),
			),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect
	//     -> should fail and not attempt to reconnect and then close
	t.Run("#5", func(t *testing.T) {
		t.Parallel()
		testReconnect(t, false, 0, mockconn.ClosedOutcome,
			mockconn.NewConnectResults(
				mockconn.NewConnectResult(mockconn.FirstAttempt, mockconn.SucceedResult),
			),
		)
	})
}

// TestReconnectRegistration tests the ability of the Channel Event Client to
// re-establish the existing registrations after reconnecting.
func TestReconnectRegistration(t *testing.T) {
	// (1) Connect
	// (2) Register for block events
	// (3) Register for CC events
	// (4) Send a CONFIG_UPDATE block event
	//     -> should receive one block event
	// (5) Send a CC event
	//     -> should receive one CC event and one block event
	// (6) Disconnect
	//     -> should reconnect
	// (7) Save a CONFIG_UPDATE block event to the ledger
	// (8) Save a CC event to the ledger
	// (9) Should receive all events
	t.Run("#1", func(t *testing.T) {
		t.Parallel()
		testReconnectRegistration(
			t, mockconn.ExpectFiveBlocks, mockconn.ExpectThreeCC,
			mockconn.NewConnectResults(
				mockconn.NewConnectResult(mockconn.FirstAttempt, mockconn.SucceedResult),
				mockconn.NewConnectResult(mockconn.SecondAttempt, mockconn.SucceedResult)),
		)
	})
}

// TestConcurrentEvents ensures that the channel event client is thread-safe
func TestConcurrentEvents(t *testing.T) {
	numEvents := 1000

	// Expect double the block and filtered block events since we're also producing TxStatus events
	expectedBlockEvents := 2 * numEvents
	expectedFilteredBlockEvents := 2 * numEvents

	expectedCCEvents := numEvents
	expectedTxStatusEvents := numEvents

	channelID := "mychannel"
	ccID := "mycc1"
	ccFilter := "event.*"

	eventClient, conn, err := newClientWithMockConnAndOpts(
		channelID, newMockContext(),
		nil, clientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		[]options.Opt{
			esdispatcher.WithEventConsumerBufferSize(uint(numEvents) * 4),
		},
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}

	// First register for block, filtered block, and chaincode events ...
	breg, beventch, err := eventClient.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer eventClient.Unregister(breg)

	fbreg, fbeventch, err := eventClient.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("error registering for filtered block events: %s", err)
	}
	defer eventClient.Unregister(fbreg)

	ccreg, cceventch, err := eventClient.RegisterChaincodeEvent(ccID, ccFilter)
	if err != nil {
		t.Fatalf("error registering for chaincode events")
	}
	defer eventClient.Unregister(ccreg)

	blockTestErr := make(chan error)
	go listenBlockEvents(channelID, beventch, expectedBlockEvents, blockTestErr)

	fblockTestErr := make(chan error)
	go listenFilteredBlockEvents(channelID, fbeventch, expectedFilteredBlockEvents, fblockTestErr)

	ccTestErr := make(chan error)
	go listenChaincodeEvents(channelID, cceventch, expectedCCEvents, ccTestErr)

	txStatusTestErr := make(chan error)
	go txStatusTest(eventClient, conn.Ledger(), channelID, expectedTxStatusEvents, txStatusTestErr)

	// Produce some block events
	go func() {
		for i := 0; i < numEvents; i++ {
			txID := fmt.Sprintf("txid_tx_%d", i)
			conn.Ledger().NewBlock(channelID,
				servicemocks.NewTransaction(txID, pb.TxValidationCode_VALID, cb.HeaderType_CONFIG_UPDATE),
			)
		}
	}()

	blockTestDone := false
	fblockTestDone := false
	ccTestDone := false
	txStatusTestDone := false

	for {
		select {
		case err := <-blockTestErr:
			if err != nil {
				t.Fatalf("Block test returned error: %s", err)
			}
			blockTestDone = true
		case err := <-fblockTestErr:
			if err != nil {
				t.Fatalf("Filtered Block test returned error: %s", err)
			}
			fblockTestDone = true
		case err := <-ccTestErr:
			if err != nil {
				t.Fatalf("Chaincode test returned error: %s", err)
			}
			ccTestDone = true
		case err := <-txStatusTestErr:
			if err != nil {
				t.Fatalf("TxStatus test returned error: %s", err)
			}
			txStatusTestDone = true
		case <-time.After(10 * time.Second):
			if !blockTestDone {
				t.Fatalf("Timed out waiting for block test")
			}
			if !fblockTestDone {
				t.Fatalf("Timed out waiting for filtered block test")
			}
			if !ccTestDone {
				t.Fatalf("Timed out waiting for chaincode test")
			}
			if !txStatusTestDone {
				t.Fatalf("Timed out waiting for TxStatus test")
			}
		}
		if blockTestDone && fblockTestDone && ccTestDone && txStatusTestDone {
			fmt.Printf("All tests completed successfully\n")
			break
		}
	}
}

func listenBlockEvents(channelID string, eventch <-chan *fab.BlockEvent, expected int, errch chan<- error) {
	numReceived := 0

	for {
		select {
		case _, ok := <-eventch:
			if !ok {
				fmt.Printf("Block events channel was closed \n")
				return
			}
			numReceived++
		case <-time.After(5 * time.Second):
			if numReceived != expected {
				errch <- errors.Errorf("Expected [%d] events but received [%d]", expected, numReceived)
			} else {
				fmt.Printf("Received %d block events\n", numReceived)
				errch <- nil
			}
			return
		}
	}
}

func listenFilteredBlockEvents(channelID string, eventch <-chan *fab.FilteredBlockEvent, expected int, errch chan<- error) {
	numReceived := 0

	for {
		select {
		case fbevent, ok := <-eventch:
			if !ok {
				fmt.Printf("Filtered block events channel was closed \n")
				return
			}
			if fbevent.FilteredBlock == nil {
				errch <- errors.Errorf("Expecting filtered block but got nil")
				return
			}
			if fbevent.FilteredBlock.ChannelId != channelID {
				errch <- errors.Errorf("Expecting channel [%s] but got [%s]", channelID, fbevent.FilteredBlock.ChannelId)
				return
			}
			numReceived++
		case <-time.After(5 * time.Second):
			if numReceived != expected {
				errch <- errors.Errorf("Expected [%d] events but received [%d]", expected, numReceived)
			} else {
				fmt.Printf("Received %d filtered block events\n", numReceived)
				errch <- nil
			}
			return
		}
	}
}

func listenChaincodeEvents(channelID string, eventch <-chan *fab.CCEvent, expected int, errch chan<- error) {
	numReceived := 0

	for {
		select {
		case _, ok := <-eventch:
			if !ok {
				fmt.Printf("CC events channel was closed \n")
				return
			}
			numReceived++
		case <-time.After(5 * time.Second):
			if numReceived != expected {
				errch <- errors.Errorf("Expected [%d] events but received [%d]", expected, numReceived)
			} else {
				fmt.Printf("Received %d CC events\n", numReceived)
				errch <- nil
			}
			return
		}
	}
}

func txStatusTest(eventClient *Client, ledger servicemocks.Ledger, channelID string, expected int, errch chan<- error) {
	ccID := "mycc1"
	event1 := "event1"

	var wg sync.WaitGroup
	wg.Add(expected)

	var errs []error
	var mutex sync.Mutex
	var receivedEvents int

	for i := 0; i < expected; i++ {
		txID := fmt.Sprintf("txid_tx_%d", i)
		go func() {
			defer wg.Done()

			reg, txeventch, err := eventClient.RegisterTxStatusEvent(txID)
			if err != nil {
				mutex.Lock()
				errs = append(errs, errors.New("Error registering for TxStatus event"))
				mutex.Unlock()
				return
			}
			defer eventClient.Unregister(reg)

			ledger.NewBlock(channelID,
				servicemocks.NewTransactionWithCCEvent(txID, pb.TxValidationCode_VALID, ccID, event1),
			)

			select {
			case txStatus, ok := <-txeventch:
				mutex.Lock()
				if !ok {
					errs = append(errs, errors.New("unexpected closed channel"))
				} else {
					receivedEvents++
				}
				fmt.Printf("received TxStatus %#v\n", txStatus)
				mutex.Unlock()
			case <-time.After(5 * time.Second):
				mutex.Lock()
				errs = append(errs, errors.New("timed out waiting for TxStatus event"))
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		errch <- errors.Errorf("Received %d events and %d errors. First error %s", receivedEvents, len(errs), errs[0])
	} else {
		errch <- nil
	}
}

func testConnect(t *testing.T, maxConnectAttempts uint, expectedOutcome mockconn.Outcome, connAttemptResult mockconn.ConnectAttemptResults) {
	cp := mockconn.NewProviderFactory()

	eventClient, _, err := newClientWithMockConnAndOpts(
		"mychannel", newMockContext(),
		cp.FlakeyProvider(connAttemptResult, mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory))),
		clientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		[]options.Opt{
			esdispatcher.WithEventConsumerTimeout(time.Second),
			WithMaxConnectAttempts(maxConnectAttempts),
		},
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}

	var outcome mockconn.Outcome
	if err := eventClient.Connect(); err != nil {
		outcome = mockconn.ErrorOutcome
	} else {
		outcome = mockconn.ConnectedOutcome
		defer eventClient.Close()
	}

	if outcome != expectedOutcome {
		t.Fatalf("Expecting that the reconnection attempt would result in [%s] but got [%s]", expectedOutcome, outcome)
	}
}

func testReconnect(t *testing.T, reconnect bool, maxReconnectAttempts uint, expectedOutcome mockconn.Outcome, connAttemptResult mockconn.ConnectAttemptResults) {
	cp := mockconn.NewProviderFactory()

	connectch := make(chan *fab.ConnectionEvent)

	ledger := servicemocks.NewMockLedger(servicemocks.BlockEventFactory)

	eventClient, _, err := newClientWithMockConnAndOpts(
		"mychannel", newMockContext(),
		cp.FlakeyProvider(connAttemptResult, mockconn.WithLedger(ledger)),
		clientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		[]options.Opt{
			esdispatcher.WithEventConsumerTimeout(3 * time.Second),
			WithMaxConnectAttempts(1),
			WithReconnect(reconnect),
			WithReconnectInitialDelay(0),
			WithMaxReconnectAttempts(maxReconnectAttempts),
			WithTimeBetweenConnectAttempts(time.Millisecond),
			WithConnectionEvent(connectch),
			WithResponseTimeout(2 * time.Second),
		},
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	outcomech := make(chan mockconn.Outcome)
	go listenConnection(connectch, outcomech)

	// Test automatic reconnect handling
	cp.Connection().ProduceEvent(dispatcher.NewDisconnectedEvent(errors.New("testing reconnect handling")))

	var outcome mockconn.Outcome

	select {
	case outcome = <-outcomech:
	case <-time.After(5 * time.Second):
		outcome = mockconn.TimedOutOutcome
	}

	if outcome != expectedOutcome {
		t.Fatalf("Expecting that the reconnection attempt would result in [%s] but got [%s]", expectedOutcome, outcome)
	}
}

// testReconnectRegistration tests the scenario when an events client is registered to receive events and the connection to the
// event service is lost. After the connection is re-established, events should once again be received without the caller having to
// re-register for those events.
func testReconnectRegistration(t *testing.T, expectedBlockEvents mockconn.NumBlock, expectedCCEvents mockconn.NumChaincode, connectResults mockconn.ConnectAttemptResults) {
	channelID := "mychannel"
	ccID := "mycc"

	ledger := servicemocks.NewMockLedger(servicemocks.BlockEventFactory)

	cp := mockconn.NewProviderFactory()

	eventClient, _, err := newClientWithMockConnAndOpts(
		channelID, newMockContext(),
		cp.FlakeyProvider(connectResults, mockconn.WithLedger(ledger)),
		clientProvider,
		clientmocks.NewDiscoveryService(peer1, peer2),
		[]options.Opt{
			esdispatcher.WithEventConsumerTimeout(3 * time.Second),
			WithMaxConnectAttempts(1),
			WithMaxReconnectAttempts(1),
			WithTimeBetweenConnectAttempts(time.Millisecond),
		},
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	_, blockch, err := eventClient.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}

	_, ccch, err := eventClient.RegisterChaincodeEvent(ccID, ".*")
	if err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}

	numCh := make(chan mockconn.Received)
	go listenEvents(blockch, ccch, 10*time.Second, numCh, expectedBlockEvents, expectedCCEvents)

	time.Sleep(500 * time.Millisecond)

	numEvents := 0
	numCCEvents := 0

	// Produce a block event
	numEvents++
	ledger.NewBlock(channelID,
		servicemocks.NewTransaction("txID", pb.TxValidationCode_VALID, cb.HeaderType_CONFIG_UPDATE),
	)

	// Produce a chaincode event
	numEvents++
	numCCEvents++
	ledger.NewBlock(channelID,
		servicemocks.NewTransactionWithCCEvent("txID", pb.TxValidationCode_VALID, ccID, "event1"),
	)

	// Wait a while for the subscriber to receive the event
	time.Sleep(500 * time.Millisecond)

	// Simulate a connection error
	cp.Connection().ProduceEvent(dispatcher.NewDisconnectedEvent(errors.New("testing reconnect handling")))

	// Wait a while for the subscriber to reconnect
	time.Sleep(2 * time.Second)

	// Produce some more events while the client is disconnected
	for ; numCCEvents < int(expectedCCEvents); numCCEvents++ {
		numEvents++
		ledger.NewBlock(channelID,
			servicemocks.NewTransactionWithCCEvent("txID", pb.TxValidationCode_VALID, ccID, "event1"),
		)
	}
	for ; numEvents < int(expectedBlockEvents); numEvents++ {
		ledger.NewBlock(channelID,
			servicemocks.NewTransaction("txID", pb.TxValidationCode_VALID, cb.HeaderType_CONFIG_UPDATE),
		)
	}

	var eventsReceived mockconn.Received

	select {
	case received, ok := <-numCh:
		if !ok {
			t.Fatalf("connection closed prematurely")
		} else {
			eventsReceived = received
		}
	case <-time.After(20 * time.Second):
		t.Fatalf("timed out waiting for events")
	}

	if eventsReceived.NumBlock != expectedBlockEvents {
		t.Fatalf("Expecting to receive [%d] block events but received [%d]", expectedBlockEvents, eventsReceived.NumBlock)
	}
	if eventsReceived.NumChaincode != expectedCCEvents {
		t.Fatalf("Expecting to receive [%d] CC events but received [%d]", expectedCCEvents, eventsReceived.NumChaincode)
	}
}

func listenConnection(eventch chan *fab.ConnectionEvent, outcome chan mockconn.Outcome) {
	state := InitialState

	for {
		e, ok := <-eventch
		fmt.Printf("listenConnection - got event [%v] - ok=[%v]\n", e, ok)
		if !ok {
			fmt.Printf("listenConnection - Returning terminated outcome\n")
			outcome <- mockconn.ClosedOutcome
			break
		}
		if e.Connected {
			if state == Disconnected {
				fmt.Printf("listenConnection - Returning reconnected outcome\n")
				outcome <- mockconn.ReconnectedOutcome
			}
			state = Connected
		} else {
			state = Disconnected
		}
	}
}

func listenEvents(blockch <-chan *fab.BlockEvent, ccch <-chan *fab.CCEvent, waitDuration time.Duration, numEventsCh chan mockconn.Received, expectedBlockEvents mockconn.NumBlock, expectedCCEvents mockconn.NumChaincode) {
	var numBlockEventsReceived mockconn.NumBlock
	var numCCEventsReceived mockconn.NumChaincode

	for {
		select {
		case _, ok := <-blockch:
			if ok {
				numBlockEventsReceived++
			} else {
				// The channel was closed by the event client. Make a new channel so
				// that we don't get into a tight loop
				blockch = make(chan *fab.BlockEvent)
			}
		case _, ok := <-ccch:
			if ok {
				numCCEventsReceived++
			} else {
				// The channel was closed by the event client. Make a new channel so
				// that we don't get into a tight loop
				ccch = make(chan *fab.CCEvent)
			}
		case <-time.After(waitDuration):
			numEventsCh <- mockconn.NewReceived(numBlockEventsReceived, numCCEventsReceived)
			return
		}
		if numBlockEventsReceived >= expectedBlockEvents && numCCEventsReceived >= expectedCCEvents {
			numEventsCh <- mockconn.NewReceived(numBlockEventsReceived, numCCEventsReceived)
			return
		}
	}
}

type ClientProvider func(channelID string, context context.Client, connectionProvider api.ConnectionProvider, discoveryService fab.DiscoveryService, opts []options.Opt) (*Client, error)

var clientProvider = func(channelID string, context context.Client, connectionProvider api.ConnectionProvider, discoveryService fab.DiscoveryService, opts []options.Opt) (*Client, error) {
	return newClient(channelID, context, connectionProvider, discoveryService, opts, true,
		func() error {
			fmt.Printf("AfterConnect called")
			return nil
		},
		func() error {
			fmt.Printf("BeforeReconnect called")
			return nil
		})
}

var failAfterConnectClientProvider = func(channelID string, context context.Client, connectionProvider api.ConnectionProvider, discoveryService fab.DiscoveryService, opts []options.Opt) (*Client, error) {
	return newClient(channelID, context, connectionProvider, discoveryService, opts, true,
		func() error {
			return errors.New("simulated failure after connect")
		},
		nil)
}

var filteredClientProvider = func(channelID string, context context.Client, connectionProvider api.ConnectionProvider, discoveryService fab.DiscoveryService, opts []options.Opt) (*Client, error) {
	return newClient(channelID, context, connectionProvider, discoveryService, opts, false,
		func() error {
			fmt.Printf("AfterConnect called")
			return nil
		},
		func() error {
			fmt.Printf("BeforeReconnect called")
			return nil
		})
}

func newClient(channelID string, context context.Client, connectionProvider api.ConnectionProvider, discoveryService fab.DiscoveryService, opts []options.Opt, permitBlockEvents bool, afterConnect handler, beforeReconnect handler) (*Client, error) {
	client := New(
		permitBlockEvents,
		dispatcher.New(
			context, channelID,
			connectionProvider,
			discoveryService,
			opts...,
		),
		opts...,
	)
	client.SetAfterConnectHandler(afterConnect)
	client.SetBeforeReconnectHandler(beforeReconnect)

	if err := client.Start(); err != nil {
		return client, err
	}
	return client, nil
}

func newClientWithMockConn(channelID string, context context.Client, clientProvider ClientProvider, discoveryService fab.DiscoveryService, connOpts ...mockconn.Opt) (*Client, mockconn.Connection, error) {
	conn := mockconn.NewMockConnection(connOpts...)
	client, _, err := newClientWithMockConnAndOpts(channelID, context, mockconn.NewProviderFactory().Provider(conn), clientProvider, discoveryService, []options.Opt{})
	return client, conn, err
}

func newClientWithMockConnAndOpts(channelID string, context context.Client, connectionProvider api.ConnectionProvider, clientProvider ClientProvider, discoveryService fab.DiscoveryService, opts []options.Opt, connOpts ...mockconn.Opt) (*Client, mockconn.Connection, error) {
	var conn mockconn.Connection
	if connectionProvider == nil {
		conn = mockconn.NewMockConnection(connOpts...)
		connectionProvider = mockconn.NewProviderFactory().Provider(conn)
	}
	client, err := clientProvider(channelID, context, connectionProvider, discoveryService, opts)
	return client, conn, err
}

func checkFilteredBlock(t *testing.T, fblock *pb.FilteredBlock, expectedChannelID string, expectedFilteredTxs ...*pb.FilteredTransaction) {
	if fblock == nil {
		t.Fatalf("Expecting filtered block but got nil")
	}
	if fblock.ChannelId != expectedChannelID {
		t.Fatalf("Expecting channel [%s] but got [%s]", expectedChannelID, fblock.ChannelId)
	}
	if len(fblock.FilteredTransactions) != len(expectedFilteredTxs) {
		t.Fatalf("Expecting %d filtered transactions but got %d", len(expectedFilteredTxs), len(fblock.FilteredTransactions))
	}

	for _, expectedTx := range expectedFilteredTxs {
		found := false
		for _, tx := range fblock.FilteredTransactions {
			if tx.Txid == expectedTx.Txid {
				found = true
				if tx.TxValidationCode != expectedTx.TxValidationCode {
					t.Fatalf("Expecting TxValidationCode [%s] but got [%s] for TxID [%s]", expectedTx.TxValidationCode, tx.TxValidationCode, tx.Txid)
				}
				break
			}
		}
		if !found {
			t.Fatalf("No filtered transaction found for TxID [%s] not found", expectedTx.Txid)
		}
	}
}

func checkTxStatusEvent(t *testing.T, event *fab.TxStatusEvent, expectedTxID string, expectedCode pb.TxValidationCode) {
	if event.TxID != expectedTxID {
		t.Fatalf("expecting event for TxID [%s] but received event for TxID [%s]", expectedTxID, event.TxID)
	}
	if event.TxValidationCode != expectedCode {
		t.Fatalf("expecting TxValidationCode [%s] but received [%s]", expectedCode, event.TxValidationCode)
	}
}

func checkCCEvent(t *testing.T, event *fab.CCEvent, expectedCCID string, expectedEventNames ...string) {
	if event.ChaincodeID != expectedCCID {
		t.Fatalf("expecting event for CC [%s] but received event for CC [%s]", expectedCCID, event.ChaincodeID)
	}
	found := false
	for _, eventName := range expectedEventNames {
		if event.EventName == eventName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expecting one of [%v] but received [%s]", expectedEventNames, event.EventName)
	}
}

func newMockContext() context.Client {
	return fabmocks.NewMockContext(fabmocks.NewMockUser("user1"))
}
