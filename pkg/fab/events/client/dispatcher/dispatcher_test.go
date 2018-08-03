/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/lbp"

	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var (
	peer1 = clientmocks.NewMockPeer("peer1", "grpcs://peer1.example.com:7051", 100)
	peer2 = clientmocks.NewMockPeer("peer2", "grpcs://peer2.example.com:7051", 110)
	peer3 = clientmocks.NewMockPeer("peer3", "grpcs://peer3.example.com:7051", 111)

	sourceURL = "localhost:9051"
)

func TestConnect(t *testing.T) {
	channelID := "testchannel"

	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2, peer3),
		clientmocks.NewProviderFactory().Provider(
			clientmocks.NewMockConnection(
				clientmocks.WithLedger(
					servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL),
				),
			),
		),
		WithLoadBalancePolicy(lbp.NewRandom()),
		WithBlockHeightLagThreshold(5),
	)

	if dispatcher.ChannelConfig().ID() != channelID {
		t.Fatalf("Expecting channel ID [%s] but got [%s]", channelID, dispatcher.ChannelConfig().ID())
	}

	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	// Connect
	errch := make(chan error)
	dispatcherEventch <- NewConnectEvent(errch)
	err = <-errch
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	// Connect again
	dispatcherEventch <- NewConnectEvent(errch)
	err = <-errch
	if err != nil {
		t.Fatalf("Error connecting again. Connect can be sent multiple times without error but got error: %s", err)
	}

	if dispatcher.Connection() == nil {
		t.Fatal("Got nil connection")
	}

	testConn(dispatcherEventch, errch, t, dispatcher)
}

func testConn(dispatcherEventch chan<- interface{}, errch chan error, t *testing.T, dispatcher *Dispatcher) {
	// Disconnect
	dispatcherEventch <- NewDisconnectEvent(errch)
	err := <-errch
	if err != nil {
		t.Fatalf("Error disconnecting: %s", err)
	}
	if dispatcher.Connection() != nil {
		t.Fatal("Expecting nil connection")
	}
	// Disconnect again
	dispatcherEventch <- NewDisconnectEvent(errch)
	err = <-errch
	if err == nil {
		t.Fatal("Expecting error disconnecting since the connection should already be closed")
	}
	time.Sleep(time.Second)
	// Stop the dispatcher
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func TestConnectNoPeers(t *testing.T) {
	channelID := "testchannel"

	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID), // Add no peers to discovery service
		clientmocks.NewDiscoveryService(),
		clientmocks.NewProviderFactory().Provider(
			clientmocks.NewMockConnection(
				clientmocks.WithLedger(
					servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL),
				),
			),
		),
	)

	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	// Connect
	errch := make(chan error)
	dispatcherEventch <- NewConnectEvent(errch)
	err = <-errch
	if err == nil {
		t.Fatal("Expecting error connecting with no peers but got none")
	}

	// Stop the dispatcher
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func TestConnectionEvent(t *testing.T) {
	channelID := "testchannel"

	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2, peer3),
		clientmocks.NewProviderFactory().Provider(
			clientmocks.NewMockConnection(
				clientmocks.WithLedger(
					servicemocks.NewMockLedger(servicemocks.BlockEventFactory, sourceURL),
				),
			),
		),
	)
	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}
	expectedDisconnectErr := "simulated disconnect error"
	// Register connection event
	connch := make(chan *ConnectionEvent, 10)
	errch := make(chan error)
	state := ""
	go checkEvent(connch, errch, state, expectedDisconnectErr)

	// Register for connection events
	regerrch := make(chan error)
	regch := make(chan fab.Registration)
	dispatcherEventch <- NewRegisterConnectionEvent(connch, regch, regerrch)

	select {
	case <-regch:
		// No need get the registration to unregister since we're relying on the
		// connch channel being closed when the dispatcher is stopped.
	case err1 := <-regerrch:
		t.Fatalf("Error registering for connection events: %s", err1)
	}

	// Connect
	dispatcherEventch <- NewConnectedEvent()
	time.Sleep(500 * time.Millisecond)

	// Disconnect
	dispatcherEventch <- NewDisconnectedEvent(errors.New(expectedDisconnectErr))
	time.Sleep(500 * time.Millisecond)

	// Stop (should close the event channel)
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err1 := <-stopResp; err1 != nil {
		t.Fatalf("Error stopping dispatcher: %s", err1)
	}

	err = <-errch
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestFilterByBlockHeight(t *testing.T) {
	dispatcher := &Dispatcher{}

	dispatcher.blockHeightLagThreshold = -1
	filteredPeers := dispatcher.filterByBlockHeght([]fab.Peer{peer1, peer2, peer3})
	assert.Equal(t, 3, len(filteredPeers))

	dispatcher.blockHeightLagThreshold = 0
	filteredPeers = dispatcher.filterByBlockHeght([]fab.Peer{peer1, peer2, peer3})
	assert.Equal(t, 1, len(filteredPeers))

	dispatcher.blockHeightLagThreshold = 5
	filteredPeers = dispatcher.filterByBlockHeght([]fab.Peer{peer1, peer2, peer3})
	assert.Equal(t, 2, len(filteredPeers))

	dispatcher.blockHeightLagThreshold = 20
	filteredPeers = dispatcher.filterByBlockHeght([]fab.Peer{peer1, peer2, peer3})
	assert.Equal(t, 3, len(filteredPeers))
}

func TestDisconnectIfBlockHeightLags(t *testing.T) {
	p1 := clientmocks.NewMockPeer("peer1", "grpcs://peer1.example.com:7051", 10)
	p2 := clientmocks.NewMockPeer("peer2", "grpcs://peer2.example.com:7051", 8)
	p3 := clientmocks.NewMockPeer("peer3", "grpcs://peer3.example.com:7051", 8)

	channelID := "testchannel"

	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(p1, p2, p3),
		clientmocks.NewProviderFactory().Provider(
			clientmocks.NewMockConnection(
				clientmocks.WithLedger(
					servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL),
				),
			),
		),
		WithBlockHeightLagThreshold(5),
		WithReconnectBlockHeightThreshold(10),
		WithBlockHeightMonitorPeriod(250*time.Millisecond),
	)

	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	// Register for connection events
	regerrch := make(chan error)
	regch := make(chan fab.Registration)
	connch := make(chan *ConnectionEvent, 10)
	dispatcherEventch <- NewRegisterConnectionEvent(connch, regch, regerrch)

	select {
	case <-regch:
		// No need get the registration to unregister since we're relying on the
		// connch channel being closed when the dispatcher is stopped.
	case err := <-regerrch:
		t.Fatalf("Error registering for connection events: %s", err)
	}

	// Connect
	errch := make(chan error)
	dispatcherEventch <- NewConnectEvent(errch)
	err = <-errch
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	dispatcherEventch <- esdispatcher.NewBlockEvent(servicemocks.NewBlockProducer().NewBlock(channelID), sourceURL)

	time.Sleep(time.Second)
	p2.SetBlockHeight(15)
	time.Sleep(time.Second)

	select {
	case e := <-connch:
		assert.Falsef(t, e.Connected, "expecting disconnected event")
	default:
		t.Fatal("Expecting disconnected event but got none")
	}
}

func checkEvent(connch chan *ConnectionEvent, errch chan error, state, expectedDisconnectErr string) {
	for {
		select {
		case event, ok := <-connch:
			if !ok {
				disconnect(state, errch)
				return
			}
			if event.Connected {
				if state != "" {
					errch <- errors.New("unexpected connected event")
					return
				}
				state = "connected"
			} else {
				if state != "connected" {
					errch <- errors.New("unexpected disconnected event")
					return
				}
				if event.Err == nil || event.Err.Error() != expectedDisconnectErr {
					errch <- errors.Errorf("unexpected disconnect error [%s] but got [%s]", expectedDisconnectErr, event.Err.Error())
					return
				}
				state = "disconnected"
			}
		case <-time.After(5 * time.Second):
			errch <- errors.New("timed out waiting for connection event")
			return
		}
	}
}

func disconnect(state string, errch chan error) {
	if state != "disconnected" {
		errch <- errors.New("unexpected closed channel")
	} else {
		errch <- nil
	}
}
