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
)

var (
	peer1 = fabmocks.NewMockPeer("peer1", "grpcs://peer1.example.com:7051")
	peer2 = fabmocks.NewMockPeer("peer2", "grpcs://peer2.example.com:7051")

	sourceURL = "localhost:9051"
)

func TestConnect(t *testing.T) {
	channelID := "testchannel"

	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		clientmocks.NewProviderFactory().Provider(
			clientmocks.NewMockConnection(
				clientmocks.WithLedger(
					servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL),
				),
			),
		),
		WithLoadBalancePolicy(lbp.NewRandom()),
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
		clientmocks.NewDiscoveryService(peer1, peer2),
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
