/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/endpoint"
	ehmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/blockfilter"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	"github.com/pkg/errors"
)

var (
	endpoint1 = newMockEventEndpoint("grpcs://peer1.example.com:7053")
	endpoint2 = newMockEventEndpoint("grpcs://peer2.example.com:7053")

	sourceURL = "localhost:9051"
)

func TestRegisterInterests(t *testing.T) {
	channelID := "testchannel"
	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(endpoint1, endpoint2),
		clientmocks.NewProviderFactory().Provider(
			ehmocks.NewConnection(
				clientmocks.WithLedger(servicemocks.NewMockLedger(ehmocks.BlockEventFactory, sourceURL)),
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
	errch, dispatcherEventch := connectToDispatcher(dispatcherEventch, t)

	// Register interests
	dispatcherEventch = registerFilteredBlockEvent(dispatcherEventch, errch, t)

	// Unregister interests
	dispatcherEventch <- NewUnregisterInterestsEvent(
		[]*pb.Interest{
			{
				EventType: pb.EventType_FILTEREDBLOCK,
			},
		},
		errch)

	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("error unregistering interests: %s", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for unregister interests response")
	}

	// Disconnect
	dispatcherEventch <- clientdisp.NewDisconnectEvent(errch)
	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("Error disconnecting: %s", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for connection response")
	}

	// Disconnected
	dispatcherEventch <- clientdisp.NewDisconnectedEvent(errors.New("simulating disconnected"))

	time.Sleep(time.Second)

	// Stop the dispatcher
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func registerFilteredBlockEvent(dispatcherEventch chan<- interface{}, errch chan error, t *testing.T) chan<- interface{} {
	dispatcherEventch <- NewRegisterInterestsEvent(
		[]*pb.Interest{
			{
				EventType: pb.EventType_FILTEREDBLOCK,
			},
		},
		errch)
	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("error registering interests: %s", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for register interests response")
	}
	return dispatcherEventch
}

func checkFailedRegisterInterest(dispatcherEventch chan<- interface{}, errch chan error, t *testing.T) chan<- interface{} {
	dispatcherEventch <- NewRegisterInterestsEvent(
		[]*pb.Interest{
			{
				EventType: pb.EventType_FILTEREDBLOCK,
			},
		},
		errch)
	select {
	case err := <-errch:
		if err == nil {
			t.Fatalf("expecting error registering interests but got none")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for register interests response")
	}
	return dispatcherEventch
}

func TestRegisterInterestsInvalid(t *testing.T) {
	channelID := "testchannel"
	dispatcher := newDispatcher(channelID)
	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	// Connect
	errch, dispatcherEventch := connectToDispatcher(dispatcherEventch, t)

	// Register interests
	dispatcherEventch = checkFailedRegisterInterest(dispatcherEventch, errch, t)

	// Unregister interests
	dispatcherEventch <- NewUnregisterInterestsEvent(
		[]*pb.Interest{
			{
				EventType: pb.EventType_FILTEREDBLOCK,
			},
		},
		errch)

	select {
	case err := <-errch:
		if err == nil {
			t.Fatal("expecting error unregistering interests but got none")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for unregister interests response")
	}

	// Disconnect
	dispatcherEventch <- clientdisp.NewDisconnectEvent(errch)
	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("Error disconnecting: %s", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for connection response")
	}

	// Disconnected
	dispatcherEventch <- clientdisp.NewDisconnectedEvent(errors.New("simulating disconnected"))

	time.Sleep(time.Second)

	// Stop the dispatcher
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func connectToDispatcher(dispatcherEventch chan<- interface{}, t *testing.T) (chan error, chan<- interface{}) {
	errch := make(chan error)
	dispatcherEventch <- clientdisp.NewConnectEvent(errch)
	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("Error connecting: %s", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for connection response")
	}
	return errch, dispatcherEventch
}

func newDispatcher(channelID string) *Dispatcher {
	return New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(endpoint1, endpoint2),
		clientmocks.NewProviderFactory().Provider(
			ehmocks.NewConnection(
				clientmocks.WithLedger(servicemocks.NewMockLedger(ehmocks.BlockEventFactory, sourceURL)),
				clientmocks.WithResults(
					clientmocks.NewResult(ehmocks.RegInterests, clientmocks.FailResult),
					clientmocks.NewResult(ehmocks.UnregInterests, clientmocks.FailResult),
				),
			),
		),
	)
}

func TestTimedOutRegister(t *testing.T) {
	channelID := "testchannel"
	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(endpoint1, endpoint2),
		clientmocks.NewProviderFactory().Provider(
			ehmocks.NewConnection(
				clientmocks.WithResults(
					clientmocks.NewResult(ehmocks.RegInterests, clientmocks.NoOpResult),
				),
				clientmocks.WithLedger(servicemocks.NewMockLedger(ehmocks.BlockEventFactory, sourceURL)),
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

	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("Error connecting: %s", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for connection response")
	}

	// Register interests
	dispatcherEventch <- NewRegisterInterestsEvent(
		[]*pb.Interest{
			{
				EventType: pb.EventType_FILTEREDBLOCK,
			},
		},
		errch)

	select {
	case err := <-errch:
		if err == nil {
			t.Fatal("expecting error due to no response from register interests but got none")
		}
	case <-time.After(2 * time.Second):

	}

}

func TestBlockEvents(t *testing.T) {
	channelID := "testchannel"
	ledger := servicemocks.NewMockLedger(ehmocks.BlockEventFactory, sourceURL)
	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(endpoint1, endpoint2),
		clientmocks.NewProviderFactory().Provider(
			ehmocks.NewConnection(
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

	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("Error connecting: %s", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for connection response")
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

	checkBlockEvent(eventch, t)

	// Unregister block events
	dispatcherEventch <- esdispatcher.NewUnregisterEvent(reg)

	// Stop
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func checkBlockEvent(eventch chan *fab.BlockEvent, t *testing.T) {
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
	ledger := servicemocks.NewMockLedger(ehmocks.FilteredBlockEventFactory, sourceURL)
	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(endpoint1, endpoint2),
		clientmocks.NewProviderFactory().Provider(
			ehmocks.NewConnection(
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

	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("Error connecting: %s", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for connection response")
	}

	// Register for filtered block events
	eventch := make(chan *fab.FilteredBlockEvent, 10)
	regch := make(chan fab.Registration)

	dispatcherEventch <- esdispatcher.NewRegisterFilteredBlockEvent(eventch, regch, errch)

	var reg fab.Registration
	select {
	case reg = <-regch:
	case err := <-errch:
		t.Fatalf("Error registering for block events: %s", err)
	}

	// Produce filtered block - this should notify the connection
	ledger.NewFilteredBlock(channelID)

	checkFilteredBlockEvent(eventch, t, channelID)

	// Unregister filtered block events
	dispatcherEventch <- esdispatcher.NewUnregisterEvent(reg)

	// Stop
	stopResp := make(chan error)
	dispatcherEventch <- esdispatcher.NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func checkFilteredBlockEvent(eventch chan *fab.FilteredBlockEvent, t *testing.T, channelID string) {
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

func newMockEventEndpoint(url string) api.EventEndpoint {
	return &endpoint.EventEndpoint{
		EvtURL: url,
	}
}
