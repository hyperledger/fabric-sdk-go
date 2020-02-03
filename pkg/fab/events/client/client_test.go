// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	mockconn "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver/preferpeer"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	InitialState ConnectionState = -1
)

var (
	peer1 = clientmocks.NewMockPeer("peer1", "grpcs://peer1.example.com:7051", 100)
	peer2 = clientmocks.NewMockPeer("peer2", "grpcs://peer2.example.com:7051", 110)

	sourceURL = "localhost:9051"
)

func TestConnect(t *testing.T) {
	connectionProvider := clientmocks.NewProviderFactory().Provider(
		clientmocks.NewMockConnection(
			clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL)),
		),
	)

	eventClient, _, err := newClientWithMockConnAndOpts(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg("mychannel"),
		clientmocks.NewDiscoveryService(peer1, peer2),
		connectionProvider, filteredClientProvider, []options.Opt{},
	)
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
		t.Fatal("expecting error connecting since the client is already connected")
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
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg("mychannel"),
		clientmocks.NewDiscoveryService(peer1, peer2),
		mockconn.NewProviderFactory().Provider(
			mockconn.NewMockConnection(
				mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL)),
			),
		),
		failAfterConnectClientProvider, []options.Opt{},
	)
	if err != nil {
		t.Fatalf("error creating client: %s", err)
	}
	if err := eventClient.Connect(); err == nil {
		t.Fatal("expecting error connecting client but got none")
	}
}

func TestCallsOnClosedClient(t *testing.T) {
	eventClient, _, err := newClientWithMockConn(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg("mychannel"),
		clientmocks.NewDiscoveryService(peer1, peer2),
		filteredClientProvider,
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL)),
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
		t.Fatal("expecting error connecting to closed channel event client but got none")
	}

	if _, _, err := eventClient.RegisterFilteredBlockEvent(); err == nil {
		t.Fatal("expecting error registering for block events on closed channel event client but got none")
	}

	if _, _, err := eventClient.RegisterChaincodeEvent("ccid", "event"); err == nil {
		t.Fatal("expecting error registering for chaincode events on closed channel event client but got none")
	}

	if _, _, err := eventClient.RegisterTxStatusEvent("txid"); err == nil {
		t.Fatal("expecting error registering for TX events on closed channel event client but got none")
	}

	// Make sure the client doesn't panic when calling unregister on disconnected client
	eventClient.Unregister(nil)
}

func TestCloseIfIdle(t *testing.T) {
	channelID := "mychannel"
	eventClient, _, err := newClientWithMockConn(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		clientProvider,
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory, sourceURL)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err1 := eventClient.Connect(); err1 != nil {
		t.Fatalf("error connecting channel event client: %s", err1)
	}

	reg, _, err := eventClient.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}

	if eventClient.CloseIfIdle() {
		t.Fatal("expecting client to not close since there's an outstanding registration")
	}

	eventClient.Unregister(reg)

	if !eventClient.CloseIfIdle() {
		t.Fatal("expecting client to close since there are no outstanding registrations")
	}
}

func TestInvalidUnregister(t *testing.T) {
	channelID := "mychannel"
	eventClient, _, err := newClientWithMockConn(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		filteredClientProvider,
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL)),
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
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		filteredClientProvider,
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	if _, _, err := eventClient.RegisterBlockEvent(); err == nil {
		t.Fatal("expecting error registering for block events on a filtered client")
	}
}

func TestBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClientWithMockConn(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		clientProvider,
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory, sourceURL)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err1 := eventClient.Connect(); err1 != nil {
		t.Fatalf("error connecting channel event client: %s", err1)
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
	checkBlockEvent(t, channelID, conn, eventch1, eventch2)
}

func checkBlockEvent(t *testing.T, channelID string, conn mockconn.Connection, eventch1 <-chan *fab.BlockEvent, eventch2 <-chan *fab.BlockEvent) {
	conn.Ledger().NewBlock(channelID,
		servicemocks.NewTransaction("txID", pb.TxValidationCode_VALID, cb.HeaderType_ENDORSER_TRANSACTION),
	)

	numExpected := 2
	numReceived := 0
	for {
		select {
		case _, ok := <-eventch1:
			if !ok {
				t.Fatal("unexpected closed channel")
			}
			numReceived++
		case _, ok := <-eventch2:
			if !ok {
				t.Fatal("unexpected closed channel")
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
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		filteredClientProvider,
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err1 := eventClient.Connect(); err1 != nil {
		t.Fatalf("error connecting channel event client: %s", err1)
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
	checkFilteredBlockEvents(t, eventch1, eventch2, channelID)
}

func checkFilteredBlockEvents(t *testing.T, eventch1 <-chan *fab.FilteredBlockEvent, eventch2 <-chan *fab.FilteredBlockEvent, channelID string) {
	numExpected := 2
	numReceived := 0
	for {
		select {
		case fbevent, ok := <-eventch1:
			if !ok {
				t.Fatal("unexpected closed channel")
			}
			checkFbEvent(t, fbevent, channelID)
			numReceived++
		case fbevent, ok := <-eventch2:
			if !ok {
				t.Fatal("unexpected closed channel")
			}
			checkFbEvent(t, fbevent, channelID)
			numReceived++
		case <-time.After(2 * time.Second):
			if numReceived != numExpected {
				t.Fatalf("Expecting %d filtered block events but received %d", numExpected, numReceived)
			}
			return
		}
	}
}

func checkFbEvent(t *testing.T, fbevent *fab.FilteredBlockEvent, channelID string) {
	if fbevent.FilteredBlock == nil {
		t.Fatal("Expecting filtered block but got nil")
	}
	if fbevent.FilteredBlock.ChannelId != channelID {
		t.Fatalf("Expecting channel [%s] but got [%s]", channelID, fbevent.FilteredBlock.ChannelId)
	}
}

func TestBlockAndFilteredBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClientWithMockConn(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		clientProvider,
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory, sourceURL)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err1 := eventClient.Connect(); err1 != nil {
		t.Fatalf("error connecting channel event client: %s", err1)
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
	checkBlockAndFilteredBlockEvents(t, channelID, fbeventch, beventch, tx1, tx2)

}
func checkBlockAndFilteredBlockEvents(t *testing.T, channelID string, fbeventch <-chan *fab.FilteredBlockEvent, beventch <-chan *fab.BlockEvent, tx1 *pb.FilteredTransaction, tx2 *pb.FilteredTransaction) {
	numReceived := 0
	numExpected := 2

	for {
		select {
		case fbevent, ok := <-fbeventch:
			if !ok {
				t.Fatal("unexpected closed channel")
			}
			numReceived++
			checkFilteredBlock(t, fbevent.FilteredBlock, channelID, tx1, tx2)

		case _, ok := <-beventch:
			if !ok {
				t.Fatal("unexpected closed channel")
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
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		filteredClientProvider,
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err1 := eventClient.Connect(); err1 != nil {
		t.Fatalf("error connecting channel event client: %s", err1)
	}
	defer eventClient.Close()

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	if _, _, err1 := eventClient.RegisterTxStatusEvent(""); err1 == nil {
		t.Fatal("expecting error registering for TxStatus event without a TX ID but got none")
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
	checkTxStatusEvents(t, eventch1, eventch2, txID1, txID2, txCode1, txCode2)
}

func checkTxStatusEvents(t *testing.T, eventch1 <-chan *fab.TxStatusEvent, eventch2 <-chan *fab.TxStatusEvent, txID1, txID2 string, txCode1, txCode2 pb.TxValidationCode) {
	numExpected := 2
	numReceived := 0
	done := false
	for !done {
		select {
		case event, ok := <-eventch1:
			if !ok {
				t.Fatal("unexpected closed channel")
			} else {
				checkTxStatusEvent(t, event, txID1, txCode1)
				numReceived++
			}
		case event, ok := <-eventch2:
			if !ok {
				t.Fatal("unexpected closed channel")
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
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		filteredClientProvider,
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err1 := eventClient.Connect(); err1 != nil {
		t.Fatalf("error connecting channel event client: %s", err1)
	}
	defer eventClient.Close()

	ccID1 := "mycc1"
	ccID2 := "mycc2"
	ccFilter1 := "event1"
	ccFilter2 := "event.*"
	event1 := "event1"
	event2 := "event2"
	event3 := "event3"

	if _, _, err1 := eventClient.RegisterChaincodeEvent("", ccFilter1); err1 == nil {
		t.Fatal("expecting error registering for chaincode events without CC ID but got none")
	}
	if _, _, err1 := eventClient.RegisterChaincodeEvent(ccID1, ""); err1 == nil {
		t.Fatal("expecting error registering for chaincode events without event filter but got none")
	}
	if _, _, err1 := eventClient.RegisterChaincodeEvent(ccID1, ".(xxx"); err1 == nil {
		t.Fatal("expecting error registering for chaincode events with invalid (regular expression) event filter but got none")
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
	checkCCEvents(t, eventch1, eventch2, ccID1, ccID2, event1, event2, event3)

}

func checkCCEvents(t *testing.T, eventch1 <-chan *fab.CCEvent, eventch2 <-chan *fab.CCEvent, ccID1, ccID2, event1, event2, event3 string) {
	numExpected := 3
	numReceived := 0
	done := false
	for !done {
		select {
		case event, ok := <-eventch1:
			if !ok {
				t.Fatal("unexpected closed channel")
			} else {
				checkCCEvent(t, event, ccID1, nil, event1)
				numReceived++
			}
		case event, ok := <-eventch2:
			if !ok {
				t.Fatal("unexpected closed channel")
			} else {
				checkCCEvent(t, event, ccID2, nil, event2, event3)
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
				mockconn.NewConnectResult(mockconn.ThirdAttempt, clientmocks.ConnFactory),
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
		testReconnect(t, true, 3, mockconn.ReconnectedOutcome, newDisconnectedEvent(),
			mockconn.NewConnectResults(
				mockconn.NewConnectResult(mockconn.FirstAttempt, clientmocks.ConnFactory),
				mockconn.NewConnectResult(mockconn.FourthAttempt, clientmocks.ConnFactory),
			),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect
	//     -> should fail to reconnect after two attempts and then close
	t.Run("#4", func(t *testing.T) {
		t.Parallel()
		testReconnect(t, true, 2, mockconn.ClosedOutcome, newDisconnectedEvent(),
			mockconn.NewConnectResults(
				mockconn.NewConnectResult(mockconn.FirstAttempt, clientmocks.ConnFactory),
			),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect
	//     -> should fail and not attempt to reconnect and then close
	t.Run("#5", func(t *testing.T) {
		t.Parallel()
		testReconnect(t, false, 0, mockconn.ClosedOutcome, newDisconnectedEvent(),
			mockconn.NewConnectResults(
				mockconn.NewConnectResult(mockconn.FirstAttempt, clientmocks.ConnFactory),
			),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect with fatal error
	//     -> should fail and not attempt to reconnect and then close
	t.Run("#6", func(t *testing.T) {
		t.Parallel()
		testReconnect(t, true, 0, mockconn.ClosedOutcome, newFatalDisconnectedEvent(),
			mockconn.NewConnectResults(
				mockconn.NewConnectResult(mockconn.FirstAttempt, clientmocks.ConnFactory),
			),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect with non-fatal error
	//     -> will keep failing to reconnect and will retry forever
	// (3) After waiting a while, close the client
	//     -> the client should stop trying to reconnect and the client should close
	t.Run("#7", func(t *testing.T) {
		t.Parallel()

		closeCalled := false
		testReconnect(t, true, 0, mockconn.ClosedOutcome, newDisconnectedEvent(),
			mockconn.NewConnectResults(
				mockconn.NewConnectResult(mockconn.FirstAttempt, clientmocks.ConnFactory),
			),
			withTimeoutAction(func(c *Client) (outcome mockconn.Outcome, b bool) {
				if closeCalled {
					return mockconn.TimedOutOutcome, true
				}

				c.Close()
				closeCalled = true
				return "", false
			}),
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
				mockconn.NewConnectResult(mockconn.FirstAttempt, clientmocks.ConnFactory),
				mockconn.NewConnectResult(mockconn.SecondAttempt, clientmocks.ConnFactory)),
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
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg("mychannel"),
		clientmocks.NewDiscoveryService(peer1, peer2),
		nil, clientProvider,
		[]options.Opt{
			esdispatcher.WithEventConsumerBufferSize(uint(numEvents) * 4),
		},
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory, sourceURL)),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err1 := eventClient.Connect(); err1 != nil {
		t.Fatalf("error connecting channel event client: %s", err1)
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
		t.Fatal("error registering for chaincode events")
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

	checkConcurrentEvents(blockTestErr, t, fblockTestErr, ccTestErr, txStatusTestErr)
}

func checkConcurrentEvents(blockTestErr chan error, t *testing.T, fblockTestErr chan error, ccTestErr chan error, txStatusTestErr chan error) {
	var blockTestDone, fblockTestDone, ccTestDone, txStatusTestDone bool
	for {
		select {
		case err := <-blockTestErr:
			blockTestDone = checkBlockDone(err, t)
		case err := <-fblockTestErr:
			fblockTestDone = checkFblockDone(err, t)
		case err := <-ccTestErr:
			ccTestDone = checkCCDone(err, t)
		case err := <-txStatusTestErr:
			txStatusTestDone = checkTxStatusDone(err, t)
		case <-time.After(10 * time.Second):
			checkEventsAreDone(t, blockTestDone, fblockTestDone, ccTestDone, txStatusTestDone)
		}
		if checkIfAllEventsRecv(blockTestDone, fblockTestDone, ccTestDone, txStatusTestDone) {
			break
		}
	}
}

func checkIfAllEventsRecv(blockTestDone bool, fblockTestDone bool, ccTestDone bool, txStatusTestDone bool) bool {
	if blockTestDone && fblockTestDone && ccTestDone && txStatusTestDone {
		test.Logf("All tests completed successfully")
		return true
	}
	return false
}

func checkBlockDone(err error, t *testing.T) bool {
	if err != nil {
		t.Fatalf("Block test returned error: %s", err)
	}
	return true
}

func checkFblockDone(err error, t *testing.T) bool {
	if err != nil {
		t.Fatalf("Filtered Block test returned error: %s", err)
	}
	return true
}

func checkCCDone(err error, t *testing.T) bool {
	if err != nil {
		t.Fatalf("Chaincode test returned error: %s", err)
	}
	return true
}

func checkTxStatusDone(err error, t *testing.T) bool {
	if err != nil {
		t.Fatalf("TxStatus test returned error: %s", err)
	}
	return true
}

func checkEventsAreDone(t *testing.T, blockTestDone, fblockTestDone, ccTestDone, txStatusTestDone bool) {
	if !blockTestDone {
		t.Fatal("Timed out waiting for block test")
	}
	if !fblockTestDone {
		t.Fatal("Timed out waiting for filtered block test")
	}
	if !ccTestDone {
		t.Fatal("Timed out waiting for chaincode test")
	}
	if !txStatusTestDone {
		t.Fatal("Timed out waiting for TxStatus test")
	}
}

func listenBlockEvents(channelID string, eventch <-chan *fab.BlockEvent, expected int, errch chan<- error) {
	numReceived := 0

	for {
		select {
		case _, ok := <-eventch:
			if !ok {
				test.Logf("Block events channel was closed")
				return
			}
			numReceived++
		case <-time.After(5 * time.Second):
			if numReceived != expected {
				errch <- errors.Errorf("Expected [%d] events but received [%d]", expected, numReceived)
			} else {
				test.Logf("Received %d block events", numReceived)
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
				test.Logf("Filtered block events channel was closed")
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
				test.Logf("Received %d filtered block events", numReceived)
				errch <- nil
			}
			return
		}
	}
}

func listenChaincodeEvents(channelID string, eventch <-chan *fab.CCEvent, expected int, errch chan<- error) {
	numReceived := 0
	lastBlockNum := uint64(0)

	for {
		select {
		case event, ok := <-eventch:
			if !ok {
				test.Logf("CC events channel was closed")
				return
			}
			if event.BlockNumber > 0 && event.BlockNumber <= lastBlockNum {
				errch <- errors.Errorf("Expected block greater than [%d] but received [%d]", lastBlockNum, event.BlockNumber)
				return
			}
			numReceived++
		case <-time.After(5 * time.Second):
			if numReceived != expected {
				errch <- errors.Errorf("Expected [%d] events but received [%d]", expected, numReceived)
			} else {
				test.Logf("Received %d CC events", numReceived)
				errch <- nil
			}
			return
		}
	}
}

func txStatusTest(eventClient *Client, ledger servicemocks.Ledger, channelID string, expected int, errch chan<- error) {
	ccID := "mycc1"
	event1 := "event1"
	payload1 := []byte("payload1")

	var wg sync.WaitGroup
	wg.Add(expected)

	var errs []error
	var mutex sync.Mutex
	var receivedEvents int

	for i := 0; i < expected; i++ {
		txID := fmt.Sprintf("TxID_%d", i)
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

			block := ledger.NewBlock(channelID,
				servicemocks.NewTransactionWithCCEvent(txID, pb.TxValidationCode_VALID, ccID, event1, payload1),
			)

			select {
			case event, ok := <-txeventch:
				mutex.Lock()
				if !ok {
					errs = append(errs, errors.New("unexpected closed channel"))
				} else {
					receivedEvents++
				}
				if event.BlockNumber != block.Number() {
					errch <- errors.Errorf("Expected block number [%d] but received [%d]", block.Number(), event.BlockNumber)
					return
				}
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
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg("mychannel"),
		clientmocks.NewDiscoveryService(peer1, peer2),
		cp.FlakeyProvider(connAttemptResult, mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory, sourceURL))),
		clientProvider,
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

type timeoutAction = func(c *Client) (mockconn.Outcome, bool)

type reconnectOptions struct {
	timeoutAction timeoutAction
}

type reconnectOpt func(o *reconnectOptions)

func withTimeoutAction(a timeoutAction) reconnectOpt {
	return func(o *reconnectOptions) {
		o.timeoutAction = a
	}
}

func testReconnect(t *testing.T, reconnect bool,
	maxReconnectAttempts uint, expectedOutcome mockconn.Outcome, event esdispatcher.Event,
	connAttemptResult mockconn.ConnectAttemptResults, opts ...reconnectOpt) {

	reconOpts := &reconnectOptions{}
	reconOpts.timeoutAction = func(c *Client) (outcome mockconn.Outcome, b bool) {
		return mockconn.TimedOutOutcome, true
	}

	for _, opt := range opts {
		opt(reconOpts)
	}

	cp := mockconn.NewProviderFactory()

	connectch := make(chan *dispatcher.ConnectionEvent)

	ledger := servicemocks.NewMockLedger(servicemocks.BlockEventFactory, sourceURL)

	eventClient, _, err := newClientWithMockConnAndOpts(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg("mychannel"),
		clientmocks.NewDiscoveryService(peer1, peer2),
		cp.FlakeyProvider(connAttemptResult, mockconn.WithLedger(ledger)),
		clientProvider,
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
	cp.Connection().ProduceEvent(event)

	var outcome mockconn.Outcome

	stop := false
	for !stop {
		select {
		case outcome = <-outcomech:
			stop = true
		case <-time.After(5 * time.Second):
			outcome, stop = reconOpts.timeoutAction(eventClient)
		}
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

	ledger := servicemocks.NewMockLedger(servicemocks.BlockEventFactory, sourceURL)

	cp := mockconn.NewProviderFactory()

	eventClient, _, err := newClientWithMockConnAndOpts(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg("mychannel"),
		clientmocks.NewDiscoveryService(peer1, peer2),
		cp.FlakeyProvider(connectResults, mockconn.WithLedger(ledger)),
		clientProvider,
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
	if err1 := eventClient.Connect(); err1 != nil {
		t.Fatalf("error connecting channel event client: %s", err1)
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
		servicemocks.NewTransactionWithCCEvent("txID", pb.TxValidationCode_VALID, ccID, "event1", nil),
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
			servicemocks.NewTransactionWithCCEvent("txID", pb.TxValidationCode_VALID, ccID, "event1", nil),
		)
	}
	for ; numEvents < int(expectedBlockEvents); numEvents++ {
		ledger.NewBlock(channelID,
			servicemocks.NewTransaction("txID", pb.TxValidationCode_VALID, cb.HeaderType_CONFIG_UPDATE),
		)
	}

	checkEvents(numCh, t, expectedBlockEvents, expectedCCEvents)
}

func checkEvents(numCh chan mockconn.Received, t *testing.T, expectedBlockEvents mockconn.NumBlock, expectedCCEvents mockconn.NumChaincode) {
	var eventsReceived mockconn.Received
	select {
	case received, ok := <-numCh:
		if !ok {
			t.Fatal("connection closed prematurely")
		} else {
			eventsReceived = received
		}
	case <-time.After(20 * time.Second):
		t.Fatal("timed out waiting for events")
	}
	if eventsReceived.NumBlock != expectedBlockEvents {
		t.Fatalf("Expecting to receive [%d] block events but received [%d]", expectedBlockEvents, eventsReceived.NumBlock)
	}
	if eventsReceived.NumChaincode != expectedCCEvents {
		t.Fatalf("Expecting to receive [%d] CC events but received [%d]", expectedCCEvents, eventsReceived.NumChaincode)
	}
}

func listenConnection(eventch chan *dispatcher.ConnectionEvent, outcome chan mockconn.Outcome) {
	state := InitialState

	for {
		e, ok := <-eventch
		test.Logf("listenConnection - got event [%+v] - ok=[%t]", e, ok)
		if !ok {
			test.Logf("listenConnection - Returning terminated outcome")
			outcome <- mockconn.ClosedOutcome
			break
		}
		if e.Connected {
			if state == Disconnected {
				test.Logf("listenConnection - Returning reconnected outcome")
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

type ClientProvider func(context context.Client, chConfig fab.ChannelCfg, discoveryService fab.DiscoveryService, connectionProvider api.ConnectionProvider, opts []options.Opt) (*Client, error)

var clientProvider = func(context context.Client, chConfig fab.ChannelCfg, discoveryService fab.DiscoveryService, connectionProvider api.ConnectionProvider, opts []options.Opt) (*Client, error) {
	opts = append(opts, WithBlockEvents())
	return newClient(context, chConfig, discoveryService, connectionProvider, opts,
		func() error {
			test.Logf("AfterConnect called")
			return nil
		},
		func() error {
			test.Logf("BeforeReconnect called")
			return nil
		})
}

var failAfterConnectClientProvider = func(context context.Client, chConfig fab.ChannelCfg, discoveryService fab.DiscoveryService, connectionProvider api.ConnectionProvider, opts []options.Opt) (*Client, error) {
	opts = append(opts, WithBlockEvents())
	return newClient(context, chConfig, discoveryService, connectionProvider, opts,
		func() error {
			return errors.New("simulated failure after connect")
		},
		nil)
}

var filteredClientProvider = func(context context.Client, chConfig fab.ChannelCfg, discoveryService fab.DiscoveryService, connectionProvider api.ConnectionProvider, opts []options.Opt) (*Client, error) {
	return newClient(context, chConfig, discoveryService, connectionProvider, opts,
		func() error {
			test.Logf("AfterConnect called")
			return nil
		},
		func() error {
			test.Logf("BeforeReconnect called")
			return nil
		})
}

func newClient(context context.Client, chConfig fab.ChannelCfg, discoveryService fab.DiscoveryService, connectionProvider api.ConnectionProvider, opts []options.Opt, afterConnect handler, beforeReconnect handler) (*Client, error) {
	client := New(
		dispatcher.New(
			context, chConfig,
			discoveryService,
			connectionProvider,
			opts...,
		),
		opts...,
	)
	client.SetAfterConnectHandler(afterConnect)
	client.SetBeforeReconnectHandler(beforeReconnect)

	err := client.Start()
	return client, err
}

func newClientWithMockConn(context context.Client, chConfig fab.ChannelCfg, discovery fab.DiscoveryService, clientProvider ClientProvider, connOpts ...mockconn.Opt) (*Client, mockconn.Connection, error) {
	conn := mockconn.NewMockConnection(connOpts...)
	client, _, err := newClientWithMockConnAndOpts(context, chConfig, discovery, mockconn.NewProviderFactory().Provider(conn), clientProvider, []options.Opt{})
	return client, conn, err
}

func newClientWithMockConnAndOpts(context context.Client, chConfig fab.ChannelCfg, discovery fab.DiscoveryService, connectionProvider api.ConnectionProvider, clientProvider ClientProvider, opts []options.Opt, connOpts ...mockconn.Opt) (*Client, mockconn.Connection, error) {
	var conn mockconn.Connection
	if connectionProvider == nil {
		conn = mockconn.NewMockConnection(connOpts...)
		connectionProvider = mockconn.NewProviderFactory().Provider(conn)
	}
	client, err := clientProvider(context, chConfig, discovery, connectionProvider, opts)
	return client, conn, err
}

func checkFilteredBlock(t *testing.T, fblock *pb.FilteredBlock, expectedChannelID string, expectedFilteredTxs ...*pb.FilteredTransaction) {
	if fblock == nil {
		t.Fatal("Expecting filtered block but got nil")
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

func checkCCEvent(t *testing.T, event *fab.CCEvent, expectedCCID string, expectedPayload []byte, expectedEventNames ...string) {
	if event.ChaincodeID != expectedCCID {
		t.Fatalf("expecting event for CC [%s] but received event for CC [%s]", expectedCCID, event.ChaincodeID)
	}
	if !bytes.Equal(event.Payload, expectedPayload) {
		t.Fatalf("expecting payload [%s] but received payload [%s]", expectedPayload, event.Payload)
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

func TestDisconnectIfBlockHeightLags(t *testing.T) {
	p1 := clientmocks.NewMockPeer("peer1", "grpcs://peer1.example.com:7051", 4)
	p2 := clientmocks.NewMockPeer("peer2", "grpcs://peer2.example.com:7051", 1)
	p3 := clientmocks.NewMockPeer("peer3", "grpcs://peer3.example.com:7051", 1)

	connectch := make(chan *dispatcher.ConnectionEvent)

	conn := clientmocks.NewMockConnection(
		clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL)),
	)
	connectionProvider := clientmocks.NewProviderFactory().Provider(conn)

	channelID := "mychannel"

	ctx := fabmocks.NewMockContext(
		mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
	)

	ctx.SetEndpointConfig(clientmocks.NewMockConfig(channelID,
		fab.EventServicePolicy{
			ResolverStrategy:                 fab.MinBlockHeightStrategy,
			Balancer:                         fab.RoundRobin,
			BlockHeightLagThreshold:          2,
			ReconnectBlockHeightLagThreshold: 3,
			PeerMonitorPeriod:                250 * time.Millisecond,
		},
	))

	eventClient, _, err := newClientWithMockConnAndOpts(
		ctx,
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(p1, p2, p3),
		connectionProvider, filteredClientProvider,
		[]options.Opt{
			esdispatcher.WithEventConsumerTimeout(3 * time.Second),
			WithMaxConnectAttempts(1),
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

	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx1", pb.TxValidationCode_VALID))
	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx2", pb.TxValidationCode_VALID))
	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx3", pb.TxValidationCode_VALID))
	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx4", pb.TxValidationCode_VALID))
	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx5", pb.TxValidationCode_VALID))

	// Set the block height of another peer to be greater than the disconnect threshold
	// so that the event client can reconnect to another peer
	p2.SetBlockHeight(9)

	select {
	case outcome := <-outcomech:
		assert.Equal(t, mockconn.ReconnectedOutcome, outcome)
	case <-time.After(3 * time.Second):
		t.Fatal("Timed out waiting for reconnect")
	}
}

// TestPreferLocalOrgConnection tests the scenario where an org wishes to connect to it's own peers
// if they are above the block height lag threshold but, if they fall below the threshold, the
// connection should be made to another org's peer. Once the local org's peers have caught up in
// block height, the connection to the local peer should be re-established.
func TestPreferLocalOrgConnection(t *testing.T) {
	channelID := "testchannel"
	org1MSP := "Org1MSP"
	org2MSP := "Org2MSP"
	blockHeightLagThreshold := 2

	p1O1 := clientmocks.NewMockStatefulPeer("p1_o1", "peer1.org1.com:7051", clientmocks.WithMSP(org1MSP), clientmocks.WithBlockHeight(4))
	p2O1 := clientmocks.NewMockStatefulPeer("p2_o1", "peer2.org1.com:7051", clientmocks.WithMSP(org1MSP), clientmocks.WithBlockHeight(3))
	p1O2 := clientmocks.NewMockStatefulPeer("p1_o2", "peer1.org2.com:7051", clientmocks.WithMSP(org2MSP), clientmocks.WithBlockHeight(10))
	p2O2 := clientmocks.NewMockStatefulPeer("p2_o2", "peer1.org2.com:7051", clientmocks.WithMSP(org2MSP), clientmocks.WithBlockHeight(11))

	connectch := make(chan *dispatcher.ConnectionEvent)

	conn := clientmocks.NewMockConnection(
		clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL)),
	)
	connectionProvider := clientmocks.NewProviderFactory().Provider(conn)

	ctx := fabmocks.NewMockContext(
		mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
	)
	ctx.SetEndpointConfig(clientmocks.NewMockConfig(channelID,
		fab.EventServicePolicy{
			ResolverStrategy:                 fab.PreferOrgStrategy,
			Balancer:                         fab.RoundRobin,
			BlockHeightLagThreshold:          blockHeightLagThreshold,
			ReconnectBlockHeightLagThreshold: 3,
			PeerMonitorPeriod:                250 * time.Millisecond,
		},
	))

	eventClient, _, err := newClientWithMockConnAndOpts(
		ctx,
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(p1O1, p2O1, p1O2, p2O2),
		connectionProvider, filteredClientProvider,
		[]options.Opt{
			esdispatcher.WithEventConsumerTimeout(3 * time.Second),
			WithMaxConnectAttempts(1),
			WithTimeBetweenConnectAttempts(time.Millisecond),
			WithConnectionEvent(connectch),
			WithResponseTimeout(2 * time.Second),
		},
	)
	require.NoErrorf(t, err, "error creating channel event client")
	err = eventClient.Connect()
	require.NoErrorf(t, err, "errorconnecting channel event client")
	defer eventClient.Close()

	connectedPeer := eventClient.Dispatcher().(*dispatcher.Dispatcher).ConnectedPeer()
	assert.Equal(t, org2MSP, connectedPeer.MSPID())

	outcomech := make(chan mockconn.Outcome)
	go listenConnection(connectch, outcomech)

	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx1", pb.TxValidationCode_VALID))
	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx2", pb.TxValidationCode_VALID))
	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx3", pb.TxValidationCode_VALID))
	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx4", pb.TxValidationCode_VALID))
	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx5", pb.TxValidationCode_VALID))

	// Set the block height of the local peer to be greater than the disconnect threshold
	// so that the event client can reconnect to the local peer
	p2O1.SetBlockHeight(9)

	select {
	case outcome := <-outcomech:
		assert.Equal(t, mockconn.ReconnectedOutcome, outcome)
		connectedPeer := eventClient.Dispatcher().(*dispatcher.Dispatcher).ConnectedPeer()
		assert.Equal(t, org1MSP, connectedPeer.MSPID())
	case <-time.After(3 * time.Second):
		t.Fatal("Timed out waiting for reconnect")
	}
}

// TestPreferLocalPeersConnection tests the scenario where an org wishes to connect to one of a list of preferred peers
// if they are above the block height lag threshold but, if they fall below the threshold, the
// connection should be made to another peer. Once the preferred peers have caught up in
// block height, the connection to one of the preferred peers should be re-established.
func TestPreferLocalPeersConnection(t *testing.T) {
	channelID := "testchannel"
	org1MSP := "Org1MSP"
	org2MSP := "Org2MSP"
	blockHeightLagThreshold := 2

	p1O1 := clientmocks.NewMockStatefulPeer("p1_o1", "peer1.org1.com:7051", clientmocks.WithMSP(org1MSP), clientmocks.WithBlockHeight(4))
	p2O1 := clientmocks.NewMockStatefulPeer("p2_o1", "peer2.org1.com:7051", clientmocks.WithMSP(org1MSP), clientmocks.WithBlockHeight(3))
	p1O2 := clientmocks.NewMockStatefulPeer("p1_o2", "peer1.org2.com:7051", clientmocks.WithMSP(org2MSP), clientmocks.WithBlockHeight(10))
	p2O2 := clientmocks.NewMockStatefulPeer("p2_o2", "peer1.org2.com:7051", clientmocks.WithMSP(org2MSP), clientmocks.WithBlockHeight(11))

	connectch := make(chan *dispatcher.ConnectionEvent)

	conn := clientmocks.NewMockConnection(
		clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL)),
	)
	connectionProvider := clientmocks.NewProviderFactory().Provider(conn)

	ctx := fabmocks.NewMockContext(
		mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
	)

	ctx.SetEndpointConfig(clientmocks.NewMockConfig(channelID,
		fab.EventServicePolicy{
			ResolverStrategy:                 fab.PreferOrgStrategy,
			Balancer:                         fab.RoundRobin,
			BlockHeightLagThreshold:          blockHeightLagThreshold,
			ReconnectBlockHeightLagThreshold: 3,
			PeerMonitorPeriod:                250 * time.Millisecond,
		},
	))

	eventClient, _, err := newClientWithMockConnAndOpts(
		ctx,
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(p1O1, p2O1, p1O2, p2O2),
		connectionProvider, filteredClientProvider,
		[]options.Opt{
			esdispatcher.WithEventConsumerTimeout(3 * time.Second),
			WithMaxConnectAttempts(1),
			WithTimeBetweenConnectAttempts(time.Millisecond),
			WithConnectionEvent(connectch),
			WithResponseTimeout(2 * time.Second),
			dispatcher.WithPeerResolver(preferpeer.NewResolver(p1O1.URL(), p2O1.URL())),
		},
	)
	require.NoErrorf(t, err, "error creating channel event client")
	err = eventClient.Connect()
	require.NoErrorf(t, err, "errorconnecting channel event client")
	defer eventClient.Close()

	connectedPeer := eventClient.Dispatcher().(*dispatcher.Dispatcher).ConnectedPeer()
	assert.Equal(t, org2MSP, connectedPeer.MSPID())

	outcomech := make(chan mockconn.Outcome)
	go listenConnection(connectch, outcomech)

	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx1", pb.TxValidationCode_VALID))
	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx2", pb.TxValidationCode_VALID))
	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx3", pb.TxValidationCode_VALID))
	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx4", pb.TxValidationCode_VALID))
	conn.Ledger().NewFilteredBlock(channelID, servicemocks.NewFilteredTx("tx5", pb.TxValidationCode_VALID))

	// Set the block height of the local peer to be greater than the disconnect threshold
	// so that the event client can reconnect to the local peer
	p2O1.SetBlockHeight(9)

	select {
	case outcome := <-outcomech:
		assert.Equal(t, mockconn.ReconnectedOutcome, outcome)
		connectedPeer := eventClient.Dispatcher().(*dispatcher.Dispatcher).ConnectedPeer()
		assert.Equal(t, org1MSP, connectedPeer.MSPID())
	case <-time.After(3 * time.Second):
		t.Fatal("Timed out waiting for reconnect")
	}
}

func TestTransferRegistrations(t *testing.T) {
	// Tests the scenario where all event registrations are transferred to another event client.
	t.Run("Transfer", func(t *testing.T) {
		testTransferRegistrations(t, func(client *Client) (fab.EventSnapshot, error) {
			return client.TransferRegistrations(false)
		})
	})

	// Tests the scenario where one event client is stopped and all
	// of the event registrations are transferred to another event client.
	t.Run("TransferAndClose", func(t *testing.T) {
		testTransferRegistrations(t, func(client *Client) (fab.EventSnapshot, error) {
			return client.TransferRegistrations(true)
		})
	})
}

type transferFunc func(client *Client) (fab.EventSnapshot, error)

// TestTransferRegistrations tests the scenario where one event client is stopped and all
// of the event registrations are transferred to another event client.
func testTransferRegistrations(t *testing.T, transferFunc transferFunc) {
	channelID := "mychannel"
	eventClient1, conn1, err := newClientWithMockConn(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		clientProvider,
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory, sourceURL)),
	)
	require.NoErrorf(t, err, "error creating channel event client")

	err = eventClient1.Connect()
	require.NoErrorf(t, err, "error connecting channel event client")

	breg, beventch, err := eventClient1.RegisterBlockEvent()
	require.NoErrorf(t, err, "error registering for block events")

	conn1.Ledger().NewBlock(channelID,
		servicemocks.NewTransaction("txID", pb.TxValidationCode_VALID, cb.HeaderType_ENDORSER_TRANSACTION),
	)

	select {
	case <-beventch:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for block event")
	}

	snapshot, err := transferFunc(eventClient1)
	require.NoErrorf(t, err, "error transferring snapshot")

	eventClient2, conn2, err := newClientWithMockConnAndOpts(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg("mychannel"),
		clientmocks.NewDiscoveryService(peer1, peer2),
		nil, clientProvider,
		[]options.Opt{
			esdispatcher.WithSnapshot(snapshot),
		},
		mockconn.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory, sourceURL)),
	)
	require.NoErrorf(t, err, "error creating channel event client")

	err = eventClient2.Connect()
	require.NoErrorf(t, err, "error connecting channel event client")

	conn2.Ledger().NewBlock(channelID,
		servicemocks.NewTransaction("txID", pb.TxValidationCode_VALID, cb.HeaderType_ENDORSER_TRANSACTION),
	)

	select {
	case <-beventch:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for block event")
	}

	eventClient2.Unregister(breg)
}

func newDisconnectedEvent() esdispatcher.Event {
	return dispatcher.NewDisconnectedEvent(errors.New("testing reconnect handling"))
}

func newFatalDisconnectedEvent() esdispatcher.Event {
	return dispatcher.NewFatalDisconnectedEvent(errors.New("testing reconnect handling"))
}
