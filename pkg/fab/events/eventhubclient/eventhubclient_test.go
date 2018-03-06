// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package eventhubclient

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/endpoint"
	ehclientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient/mocks"
	ehmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient/mocks"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/options"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

const (
	initialState client.ConnectionState = -1
)

var (
	defaultOpts = []options.Opt{}

	endpoint1 = newMockEventEndpoint("grpcs://peer1.example.com:7053")
	endpoint2 = newMockEventEndpoint("grpcs://peer2.example.com:7053")
)

func TestOptionsInNewClient(t *testing.T) {
	if _, err := New(newMockContext(), "", clientmocks.NewDiscoveryService(endpoint1, endpoint2)); err == nil {
		t.Fatalf("expecting error with no channel ID but got none")
	}

	client, err := New(newMockContext(), "mychannel", clientmocks.NewDiscoveryService(endpoint1, endpoint2),
		WithBlockEvents(),
	)
	if err != nil {
		t.Fatalf("error creating new event hub client: %s", err)
	}
	client.Close()
}

func TestClientConnect(t *testing.T) {
	eventClient, err := New(
		newMockContext(), "mychannel",
		clientmocks.NewDiscoveryService(endpoint1, endpoint2),
		withConnectionProviderAndInterests(
			clientmocks.NewProviderFactory().Provider(
				ehclientmocks.NewConnection(
					clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory)),
				)),
			filteredBlockInterests, false,
		),
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

func TestTimeoutClientConnect(t *testing.T) {
	eventClient, err := New(
		newMockContext(), "mychannel",
		clientmocks.NewDiscoveryService(endpoint1, endpoint2),
		withConnectionProviderAndInterests(
			clientmocks.NewProviderFactory().Provider(
				ehclientmocks.NewConnection(
					clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory)),
					clientmocks.WithResults(
						clientmocks.NewResult(ehmocks.RegInterests, clientmocks.NoOpResult),
					),
				)),
			filteredBlockInterests, false,
		),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err == nil {
		t.Fatalf("expecting error connecting due to timeout registering interests")
	}
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
	// (7) Save a CONFIG_UPDATE block event to the ledger
	// (8) Save a CC event to the ledger
	// (9) Should reconnect and receive all events that were
	//     saved to the ledger while the client was disconnected
	t.Run("#1", func(t *testing.T) {
		t.Parallel()
		testReconnectRegistration(
			t, clientmocks.ExpectFiveBlocks, clientmocks.ExpectThreeCC,
			clientmocks.NewConnectResults(
				clientmocks.NewConnectResult(clientmocks.FirstAttempt, clientmocks.SucceedResult),
				clientmocks.NewConnectResult(clientmocks.SecondAttempt, clientmocks.SucceedResult)),
		)
	})
}

func testConnect(t *testing.T, maxConnectAttempts uint, expectedOutcome clientmocks.Outcome, connAttemptResult clientmocks.ConnectAttemptResults) {
	eventClient, err := New(
		newMockContext(), "mychannel",
		clientmocks.NewDiscoveryService(endpoint1, endpoint2),
		withConnectionProviderAndInterests(
			clientmocks.NewProviderFactory().FlakeyProvider(
				connAttemptResult,
				clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory)),
				clientmocks.WithFactory(func(opts ...clientmocks.Opt) clientmocks.Connection {
					return ehclientmocks.NewConnection(opts...)
				}),
			),
			blockInterests, true,
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

	connectch := make(chan *fab.ConnectionEvent)
	ledger := servicemocks.NewMockLedger(servicemocks.BlockEventFactory)

	eventClient, err := New(
		newMockContext(), "mychannel",
		clientmocks.NewDiscoveryService(endpoint1, endpoint2),
		withConnectionProviderAndInterests(
			cp.FlakeyProvider(
				connAttemptResult,
				clientmocks.WithLedger(ledger),
				clientmocks.WithFactory(func(opts ...clientmocks.Opt) clientmocks.Connection {
					return ehclientmocks.NewConnection(opts...)
				}),
			),
			blockInterests, true,
		),
		esdispatcher.WithEventConsumerTimeout(3*time.Second),
		client.WithReconnect(reconnect),
		client.WithReconnectInitialDelay(0),
		client.WithMaxConnectAttempts(1),
		client.WithMaxReconnectAttempts(maxReconnectAttempts),
		client.WithTimeBetweenConnectAttempts(time.Millisecond),
		client.WithConnectionEvent(connectch),
		client.WithResponseTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Close()

	outcomech := make(chan clientmocks.Outcome)
	go listenConnection(t, connectch, outcomech)

	// Test automatic reconnect handling
	cp.Connection().ProduceEvent(dispatcher.NewDisconnectedEvent(errors.New("testing reconnect handling")))

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
func testReconnectRegistration(t *testing.T, expectedBlockEvents clientmocks.NumBlock, expectedCCEvents clientmocks.NumChaincode, connectResults clientmocks.ConnectAttemptResults) {
	channelID := "mychannel"
	ccID := "mycc"

	ledger := servicemocks.NewMockLedger(servicemocks.BlockEventFactory)
	cp := clientmocks.NewProviderFactory()

	eventClient, err := New(
		newMockContext(), channelID,
		clientmocks.NewDiscoveryService(endpoint1, endpoint2),
		withConnectionProviderAndInterests(
			cp.FlakeyProvider(
				connectResults,
				clientmocks.WithLedger(ledger),
				clientmocks.WithFactory(func(opts ...clientmocks.Opt) clientmocks.Connection {
					return ehclientmocks.NewConnection(opts...)
				}),
			),
			blockInterests, true,
		),
		esdispatcher.WithEventConsumerTimeout(3*time.Second),
		client.WithReconnectInitialDelay(0),
		client.WithMaxConnectAttempts(1),
		client.WithMaxReconnectAttempts(1),
		client.WithTimeBetweenConnectAttempts(time.Millisecond),
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

	numCh := make(chan clientmocks.Received)
	go listenEvents(blockch, ccch, 20*time.Second, numCh, expectedBlockEvents, expectedCCEvents)

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

	// Wait for the client to reconnect
	time.Sleep(2 * time.Second)

	// Produce some more events after the client has reconnected
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

	var eventsReceived clientmocks.Received

	select {
	case received, ok := <-numCh:
		if !ok {
			t.Fatalf("connection closed prematurely")
		} else {
			eventsReceived = received
		}
	case <-time.After(30 * time.Second):
		t.Fatalf("timed out waiting for events")
	}

	if eventsReceived.NumBlock != expectedBlockEvents {
		t.Fatalf("Expecting to receive [%d] block events but received [%d]", expectedBlockEvents, eventsReceived.NumBlock)
	}
	if eventsReceived.NumChaincode != expectedCCEvents {
		t.Fatalf("Expecting to receive [%d] CC events but received [%d]", expectedCCEvents, eventsReceived.NumChaincode)
	}
}

func listenConnection(t *testing.T, eventch chan *fab.ConnectionEvent, outcome chan clientmocks.Outcome) {
	state := initialState

	for {
		e, ok := <-eventch
		t.Logf("Got event [%v] - ok=[%v]", e, ok)
		if !ok {
			t.Logf("Returning terminated outcome")
			outcome <- clientmocks.ClosedOutcome
			break
		}
		if e.Connected {
			if state == client.Disconnected {
				t.Logf("Returning reconnected outcome")
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

func newMockContext() context.Client {
	return fabmocks.NewMockContext(fabmocks.NewMockUser("user1"))
}

func newMockEventEndpoint(url string) api.EventEndpoint {
	return &endpoint.EventEndpoint{
		EvtURL: url,
	}
}
