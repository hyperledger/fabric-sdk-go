/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	delivermocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/blockfilter"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var (
	peer1 = fabmocks.NewMockPeer("peer1", "grpcs://peer1.example.com:7051")
	peer2 = fabmocks.NewMockPeer("peer2", "grpcs://peer2.example.com:7051")

	sourceURL = "localhost:9051"
)

func TestSeek(t *testing.T) {
	channelID := "testchannel"

	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		clientmocks.NewProviderFactory().Provider(
			delivermocks.NewConnection(
				clientmocks.WithLedger(servicemocks.NewMockLedger(delivermocks.BlockEventFactory, sourceURL)),
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
		t.Fatal("timeout waiting for deliver status response")
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

func TestUnauthorized(t *testing.T) {
	channelID := "testchannel"

	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		clientmocks.NewProviderFactory().Provider(
			delivermocks.NewConnection(
				clientmocks.WithResults(
					clientmocks.NewResult(delivermocks.Connect, delivermocks.ForbiddenResult),
				),
				clientmocks.WithLedger(servicemocks.NewMockLedger(delivermocks.BlockEventFactory, sourceURL)),
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

	// Register connection event
	errch := make(chan error)
	regch := make(chan fab.Registration)
	conneventch := make(chan *clientdisp.ConnectionEvent, 5)
	dispatcherEventch <- clientdisp.NewRegisterConnectionEvent(conneventch, regch, errch)

	checkErrorFromReg(errch, t, regch)

	// Connect
	dispatcherEventch <- clientdisp.NewConnectEvent(errch)
	if err := <-errch; err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	for {
		select {
		case event := <-conneventch:
			if event.Connected {
				t.Log("Got connected event")
			} else {
				t.Logf("Got disconnected event with error [%s]", event.Err)
				return
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for disconnected event")
		}
	}
}

func checkErrorFromReg(errch chan error, t *testing.T, regch chan fab.Registration) {
	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("Error registering for connection events: %s", err)
		}
	case <-regch:
	}
}

func TestBlockEvents(t *testing.T) {
	channelID := "testchannel"
	ledger := servicemocks.NewMockLedger(delivermocks.BlockEventFactory, sourceURL)

	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		clientmocks.NewProviderFactory().Provider(
			delivermocks.NewConnection(
				clientmocks.WithLedger(ledger),
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

	checkBlockEvents(eventch, t)

	lastBlockReceived := dispatcher.LastBlockNum()
	assert.Equal(t, uint64(0), lastBlockReceived)

	ledger.NewBlock(channelID)

	checkBlockEvents(eventch, t)

	lastBlockReceived = dispatcher.LastBlockNum()
	assert.Equal(t, uint64(1), lastBlockReceived)

	// Unregister block events
	dispatcherEventch <- esdispatcher.NewUnregisterEvent(reg)

	// Stop
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func checkBlockEvents(eventch chan *fab.BlockEvent, t *testing.T) {
	select {
	case event, ok := <-eventch:
		if !ok {
			t.Fatal("unexpected closed channel")
		}
		if event.SourceURL != sourceURL {
			t.Fatalf("expecting source URL [%s] but got [%s]", sourceURL, event.SourceURL)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for block event")
	}
}

func TestFilteredBlockEvents(t *testing.T) {
	channelID := "testchannel"
	ledger := servicemocks.NewMockLedger(delivermocks.FilteredBlockEventFactory, sourceURL)

	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(peer1, peer2),
		clientmocks.NewProviderFactory().Provider(
			delivermocks.NewConnection(
				clientmocks.WithLedger(ledger),
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

	checkFilteredBlockEvents(eventch, t, channelID)

	// Unregister filtered block events
	dispatcherEventch <- esdispatcher.NewUnregisterEvent(reg)

	// Stop
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func checkFilteredBlockEvents(eventch chan *fab.FilteredBlockEvent, t *testing.T, channelID string) {
	select {
	case event, ok := <-eventch:
		if !ok {
			t.Fatal("unexpected closed channel")
		}
		if event.FilteredBlock.ChannelId != channelID {
			t.Fatalf("expecting channelID [%s] but got [%s]", channelID, event.FilteredBlock.ChannelId)
		}
		if event.SourceURL != sourceURL {
			t.Fatalf("expecting source URL [%s] but got [%s]", sourceURL, event.SourceURL)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for filtered block event")
	}
}
