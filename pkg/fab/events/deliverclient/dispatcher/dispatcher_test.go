/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"testing"
	"time"

	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	delivermocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/blockfilter"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/pkg/errors"
)

var (
	peer1 = fabmocks.NewMockPeer("peer1", "grpcs://peer1.example.com:7051")
	peer2 = fabmocks.NewMockPeer("peer2", "grpcs://peer2.example.com:7051")
)

func TestSeek(t *testing.T) {
	channelID := "testchannel"

	dispatcher := New(
		newMockContext(), channelID,
		clientmocks.NewProviderFactory().Provider(
			delivermocks.NewConnection(
				clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory)),
			),
		),
		clientmocks.NewDiscoveryService(peer1, peer2),
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
	dispatcherEventch <- clientdisp.NewConnectEvent(errch)
	if err := <-errch; err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	dispatcherEventch <- NewSeekEvent(seek.InfoNewest(), errch)

	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("error from seek request: %s", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for deliver status response")
	}

	// Disconnect
	dispatcherEventch <- clientdisp.NewDisconnectEvent(errch)
	if err := <-errch; err != nil {
		t.Fatalf("Error disconnecting: %s", err)
	}

	// Disconnected
	dispatcherEventch <- clientdisp.NewDisconnectedEvent(errors.New("simulated disconnected"))

	time.Sleep(time.Second)

	// Stop the dispatcher
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func TestTimedOutSeek(t *testing.T) {
	channelID := "testchannel"

	dispatcher := New(
		newMockContext(), channelID,
		clientmocks.NewProviderFactory().Provider(
			delivermocks.NewConnection(
				clientmocks.WithResults(
					clientmocks.NewResult(delivermocks.Seek, clientmocks.NoOpResult),
				),
				clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory)),
			),
		),
		clientmocks.NewDiscoveryService(peer1, peer2),
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
	dispatcherEventch <- clientdisp.NewConnectEvent(errch)
	if err := <-errch; err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	dispatcherEventch <- NewSeekEvent(seek.InfoNewest(), errch)

	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("expecting timeout connecting due to no response from seek but got error: %s", err)
		} else {
			t.Fatalf("expecting timeout connecting due to no response from seek but got success")
		}
	case <-time.After(2 * time.Second):
		// Expecting timeout
	}

	// Stop
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func TestUnauthorized(t *testing.T) {
	channelID := "testchannel"

	dispatcher := New(
		newMockContext(), channelID,
		clientmocks.NewProviderFactory().Provider(
			delivermocks.NewConnection(
				clientmocks.WithResults(
					clientmocks.NewResult(delivermocks.Seek, delivermocks.ForbiddenResult),
				),
				clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory)),
			),
		),
		clientmocks.NewDiscoveryService(peer1, peer2),
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
	dispatcherEventch <- clientdisp.NewConnectEvent(errch)
	if err := <-errch; err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	dispatcherEventch <- NewSeekEvent(seek.InfoNewest(), errch)

	select {
	case err := <-errch:
		if err == nil {
			t.Fatalf("expecting error connecting due to insufficient permissions but got success")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for seek response")
	}

	// Stop
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func TestBlockEvents(t *testing.T) {
	channelID := "testchannel"
	ledger := servicemocks.NewMockLedger(servicemocks.BlockEventFactory)

	dispatcher := New(
		newMockContext(), channelID,
		clientmocks.NewProviderFactory().Provider(
			delivermocks.NewConnection(
				clientmocks.WithLedger(ledger),
			),
		),
		clientmocks.NewDiscoveryService(peer1, peer2),
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
	dispatcherEventch <- clientdisp.NewConnectEvent(errch)
	if err := <-errch; err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	// Register for block events
	eventch := make(chan *fab.BlockEvent, 10)
	regch := make(chan fab.Registration)
	dispatcherEventch <- esdispatcher.NewRegisterBlockEvent(blockfilter.AcceptAny, eventch, regch, errch)

	var reg fab.Registration
	select {
	case reg = <-regch:
	case err := <-errch:
		t.Fatalf("Error registering for block events: %s", err)
	}

	// Produce block - this should notify the connection
	ledger.NewBlock(channelID)

	select {
	case _, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for block event")
	}

	// Unregister block events
	dispatcherEventch <- esdispatcher.NewUnregisterEvent(reg)

	// Stop
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func TestFilteredBlockEvents(t *testing.T) {
	channelID := "testchannel"

	ledger := servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory)

	dispatcher := New(
		newMockContext(), channelID,
		clientmocks.NewProviderFactory().Provider(
			delivermocks.NewConnection(
				clientmocks.WithLedger(ledger),
			),
		),
		clientmocks.NewDiscoveryService(peer1, peer2),
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
	dispatcherEventch <- clientdisp.NewConnectEvent(errch)
	if err := <-errch; err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	// Register for filtered block events
	eventch := make(chan *fab.FilteredBlockEvent, 10)
	regch := make(chan fab.Registration)
	dispatcherEventch <- esdispatcher.NewRegisterFilteredBlockEvent(eventch, regch, errch)

	var reg fab.Registration
	select {
	case reg = <-regch:
	case err := <-errch:
		t.Fatalf("Error registering for filtered block events: %s", err)
	}

	// Produce filtered block - this should notify the connection
	ledger.NewFilteredBlock(channelID)

	select {
	case event, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
		if event.FilteredBlock.ChannelId != channelID {
			t.Fatalf("expecting channelID [%s] but got [%s]", channelID, event.FilteredBlock.ChannelId)
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("timed out waiting for filtered block event")
	}

	// Unregister filtered block events
	dispatcherEventch <- esdispatcher.NewUnregisterEvent(reg)

	// Stop
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func newMockContext() fabcontext.Context {
	return fabmocks.NewMockContext(fabmocks.NewMockUser("user1"))
}
