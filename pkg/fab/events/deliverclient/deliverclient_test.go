/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deliverclient

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	delivermocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	fabclientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

const (
	initialState client.ConnectionState = -1
)

var (
	peer1 = fabclientmocks.NewMockPeer("peer1", "peer1.example.com:7051")
	peer2 = fabclientmocks.NewMockPeer("peer2", "peer2.example.com:7051")

	sourceURL = "localhost:9051"
)

func TestOptionsInNewClient(t *testing.T) {
	channelID := "mychannel"
	client, err := New(
		newMockContext(),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		client.WithBlockEvents(),
	)
	if err != nil {
		t.Fatalf("error creating deliver client: %s", err)
	}
	client.Close()
}

func TestClientConnect(t *testing.T) {
	channelID := "mychannel"
	eventClient, err := New(
		newMockContext(),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		client.WithBlockEvents(),
		withConnectionProvider(
			clientmocks.NewProviderFactory().Provider(
				delivermocks.NewConnection(
					clientmocks.WithLedger(servicemocks.NewMockLedger(delivermocks.BlockEventFactory, sourceURL)),
				),
			),
		),
		WithSeekType(seek.FromBlock),
		WithBlockNum(0),
		client.WithResponseTimeout(3*time.Second),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if eventClient.ConnectionState() != client.Disconnected {
		t.Fatalf("expecting connection state %s but got %s", client.Disconnected, eventClient.ConnectionState())
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting: %s", err)
	}
	time.Sleep(500 * time.Millisecond)
	if eventClient.ConnectionState() != client.Connected {
		t.Fatalf("expecting connection state %s but got %s", client.Connected, eventClient.ConnectionState())
	}
	eventClient.Close()
	if eventClient.ConnectionState() != client.Disconnected {
		t.Fatalf("expecting connection state %s but got %s", client.Disconnected, eventClient.ConnectionState())
	}
	time.Sleep(2 * time.Second)
}

// TestReconnect tests the ability of the Channel Event Client to retry multiple
// times to connect, and reconnect after it has disconnected.
func TestReconnect(t *testing.T) {
	// (1) Connect
	//     -> should fail to connect on the first and second attempt but succeed on the third attempt
	t.Run("#1", func(t *testing.T) {
		t.Parallel()
		testConnect(t, 3, clientmocks.ConnectedOutcome,
			clientmocks.NewConnectResults(
				clientmocks.NewConnectResult(clientmocks.ThirdAttempt, clientmocks.SucceedResult),
			),
		)
	})

	// (1) Connect
	//     -> should fail to connect on the first attempt and no further attempts are to be made
	t.Run("#2", func(t *testing.T) {
		t.Parallel()
		testConnect(t, 1, clientmocks.ErrorOutcome,
			clientmocks.NewConnectResults(),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect
	//     -> should fail to reconnect on the first and second attempt but succeed on the third attempt
	t.Run("#3", func(t *testing.T) {
		t.Parallel()
		testReconnect(t, true, 3, clientmocks.ReconnectedOutcome,
			clientmocks.NewConnectResults(
				clientmocks.NewConnectResult(clientmocks.FirstAttempt, clientmocks.SucceedResult),
				clientmocks.NewConnectResult(clientmocks.FourthAttempt, clientmocks.SucceedResult),
			),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect
	//     -> should fail to reconnect after two attempts and then cleanly disconnect
	t.Run("#4", func(t *testing.T) {
		t.Parallel()
		testReconnect(t, true, 2, clientmocks.ClosedOutcome,
			clientmocks.NewConnectResults(
				clientmocks.NewConnectResult(clientmocks.FirstAttempt, clientmocks.SucceedResult),
			),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect
	//     -> should fail and not attempt to reconnect and then cleanly disconnect
	t.Run("#5", func(t *testing.T) {
		t.Parallel()
		testReconnect(t, false, 0, clientmocks.ClosedOutcome,
			clientmocks.NewConnectResults(
				clientmocks.NewConnectResult(clientmocks.FirstAttempt, clientmocks.SucceedResult),
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
	// (7) Save some more blocks to the ledger
	// (8) Should reconnect and receive all events that were
	//     saved to the ledger while the client was disconnected
	t.Run("#1", func(t *testing.T) {
		t.Parallel()
		testReconnectRegistration(
			t,
			clientmocks.NewConnectResults(
				clientmocks.NewConnectResult(clientmocks.FirstAttempt, clientmocks.SucceedResult),
				clientmocks.NewConnectResult(clientmocks.SecondAttempt, clientmocks.SucceedResult)),
		)
	})
}

func testConnect(t *testing.T, maxConnectAttempts uint, expectedOutcome clientmocks.Outcome, connAttemptResult clientmocks.ConnectAttemptResults) {
	cp := clientmocks.NewProviderFactory()

	channelID := "mychannel"
	eventClient, err := New(
		newMockContext(),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		client.WithBlockEvents(),
		withConnectionProvider(
			cp.FlakeyProvider(
				connAttemptResult,
				clientmocks.WithLedger(servicemocks.NewMockLedger(delivermocks.BlockEventFactory, sourceURL)),
				clientmocks.WithFactory(func(opts ...clientmocks.Opt) clientmocks.Connection {
					return delivermocks.NewConnection(opts...)
				}),
			),
		),
		esdispatcher.WithEventConsumerTimeout(time.Second),
		client.WithMaxConnectAttempts(maxConnectAttempts),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}

	var outcome clientmocks.Outcome
	if err := eventClient.Connect(); err != nil {
		outcome = clientmocks.ErrorOutcome
	} else {
		outcome = clientmocks.ConnectedOutcome
		defer eventClient.Close()
	}

	if outcome != expectedOutcome {
		t.Fatalf("Expecting that the reconnection attempt would result in [%s] but got [%s]", expectedOutcome, outcome)
	}
}

func testReconnect(t *testing.T, reconnect bool, maxReconnectAttempts uint, expectedOutcome clientmocks.Outcome, connAttemptResult clientmocks.ConnectAttemptResults) {
	cp := clientmocks.NewProviderFactory()

	connectch := make(chan *clientdisp.ConnectionEvent)
	ledger := servicemocks.NewMockLedger(delivermocks.BlockEventFactory, sourceURL)

	channelID := "mychannel"
	eventClient, err := New(
		newMockContext(),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		client.WithBlockEvents(),
		withConnectionProvider(
			cp.FlakeyProvider(
				connAttemptResult,
				clientmocks.WithLedger(ledger),
				clientmocks.WithFactory(func(opts ...clientmocks.Opt) clientmocks.Connection {
					return delivermocks.NewConnection(opts...)
				}),
			),
		),
		esdispatcher.WithEventConsumerTimeout(3*time.Second),
		client.WithReconnect(reconnect),
		client.WithReconnectInitialDelay(0),
		client.WithMaxConnectAttempts(1),
		client.WithMaxReconnectAttempts(maxReconnectAttempts),
		client.WithTimeBetweenConnectAttempts(time.Millisecond),
		client.WithConnectionEvent(connectch),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	outcomech := make(chan clientmocks.Outcome)
	go listenConnection(connectch, outcomech)

	// Test automatic reconnect handling
	cp.Connection().ProduceEvent(clientdisp.NewDisconnectedEvent(errors.New("testing reconnect handling")))

	var outcome clientmocks.Outcome

	select {
	case outcome = <-outcomech:
	case <-time.After(5 * time.Second):
		outcome = clientmocks.TimedOutOutcome
	}

	if outcome != expectedOutcome {
		t.Fatalf("Expecting that the reconnection attempt would result in [%s] but got [%s]", expectedOutcome, outcome)
	}
}

// testReconnectRegistration tests the scenario when an events client is registered to receive events and the connection to the
// event service is lost. After the connection is re-established, events should once again be received without the caller having to
// re-register for those events.
func testReconnectRegistration(t *testing.T, connectResults clientmocks.ConnectAttemptResults) {
	var expectedBlockEvents clientmocks.NumBlock = 6
	var expectedCCEvents clientmocks.NumChaincode = 3

	channelID := "mychannel"
	ccID := "mycc"

	ledger := servicemocks.NewMockLedger(delivermocks.BlockEventFactory, sourceURL)

	// Add 2 blocks to the ledger befor the client has connected
	ledger.NewBlock(channelID,
		servicemocks.NewTransaction("txID", pb.TxValidationCode_VALID, cb.HeaderType_CONFIG_UPDATE),
	)
	ledger.NewBlock(channelID,
		servicemocks.NewTransactionWithCCEvent("txID", pb.TxValidationCode_VALID, ccID, "event1", nil),
	)

	cp := clientmocks.NewProviderFactory()

	eventClient, err := New(
		newMockContext(),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		client.WithBlockEvents(),
		withConnectionProvider(
			cp.FlakeyProvider(
				connectResults,
				clientmocks.WithLedger(ledger),
				clientmocks.WithFactory(func(opts ...clientmocks.Opt) clientmocks.Connection {
					return delivermocks.NewConnection(opts...)
				}),
			),
		),
		esdispatcher.WithEventConsumerTimeout(3*time.Second),
		client.WithReconnect(true),
		client.WithReconnectInitialDelay(5*time.Second), // Wait some time before trying to reconnect
		client.WithMaxConnectAttempts(1),
		client.WithMaxReconnectAttempts(1),
		client.WithTimeBetweenConnectAttempts(time.Millisecond),
		WithSeekType(seek.Oldest), // Ask for all block after having connected
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}

	_, blockch, err := eventClient.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}

	_, ccch, err := eventClient.RegisterChaincodeEvent(ccID, ".*")
	if err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}

	numCh := make(chan clientmocks.Received)
	go listenEvents(blockch, ccch, 20*time.Second, numCh, expectedBlockEvents, expectedCCEvents)

	time.Sleep(500 * time.Millisecond)

	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	time.Sleep(500 * time.Millisecond)

	// Produce a block event
	ledger.NewBlock(channelID,
		servicemocks.NewTransaction("txID", pb.TxValidationCode_VALID, cb.HeaderType_CONFIG_UPDATE),
	)

	// Produce a chaincode event
	ledger.NewBlock(channelID,
		servicemocks.NewTransactionWithCCEvent("txID", pb.TxValidationCode_VALID, ccID, "event1", nil),
	)

	// Wait a while for the subscriber to receive the events
	time.Sleep(1 * time.Second)

	// Simulate a connection error
	cp.Connection().ProduceEvent(clientdisp.NewDisconnectedEvent(errors.New("testing reconnect handling")))

	time.Sleep(1 * time.Second)

	// Produce some more events while the client is disconnected
	ledger.NewBlock(channelID,
		servicemocks.NewTransactionWithCCEvent("txID", pb.TxValidationCode_VALID, ccID, "event1", nil),
	)
	ledger.NewBlock(channelID,
		servicemocks.NewTransaction("txID", pb.TxValidationCode_VALID, cb.HeaderType_CONFIG_UPDATE),
	)

	var eventsReceived clientmocks.Received

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

func listenConnection(eventch chan *clientdisp.ConnectionEvent, outcome chan clientmocks.Outcome) {
	state := initialState

	for {
		e, ok := <-eventch
		if !ok {
			outcome <- clientmocks.ClosedOutcome
			break
		}
		if e.Connected {
			if state == client.Disconnected {
				outcome <- clientmocks.ReconnectedOutcome
			}
			state = client.Connected
		} else {
			state = client.Disconnected
		}
	}
}

func listenEvents(blockch <-chan *fab.BlockEvent, ccch <-chan *fab.CCEvent, waitDuration time.Duration, numEventsCh chan clientmocks.Received, expectedBlockEvents clientmocks.NumBlock, expectedCCEvents clientmocks.NumChaincode) {
	var numBlockEventsReceived clientmocks.NumBlock
	var numCCEventsReceived clientmocks.NumChaincode

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
			numEventsCh <- clientmocks.NewReceived(numBlockEventsReceived, numCCEventsReceived)
			return
		}
		if numBlockEventsReceived >= expectedBlockEvents && numCCEventsReceived >= expectedCCEvents {
			numEventsCh <- clientmocks.NewReceived(numBlockEventsReceived, numCCEventsReceived)
			return
		}
	}
}

type mockConfig struct {
	fab.EndpointConfig
}

func newMockConfig() *mockConfig {
	return &mockConfig{
		EndpointConfig: fabmocks.NewMockEndpointConfig(),
	}
}

func newMockContext() *fabmocks.MockContext {
	ctx := fabmocks.NewMockContext(
		mspmocks.NewMockSigningIdentity("user1", "test1"),
	)
	ctx.SetEndpointConfig(newMockConfig())
	return ctx
}

// withConnectionProvider is used only for testing
func withConnectionProvider(connProvider api.ConnectionProvider) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(connectionProviderSetter); ok {
			setter.SetConnectionProvider(connProvider)
		}
	}
}

// connectionProviderSetter is only used in unit tests
type connectionProviderSetter interface {
	SetConnectionProvider(value api.ConnectionProvider)
}
