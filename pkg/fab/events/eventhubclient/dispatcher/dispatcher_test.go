/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"

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
)

func TestRegisterInterests(t *testing.T) {
	channelID := "testchannel"
	dispatcher := New(
		newMockContext(), channelID,
		clientmocks.NewProviderFactory().Provider(
			ehmocks.NewConnection(
				clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory)),
			),
		),
		clientmocks.CreateDiscoveryService(endpoint1, endpoint2),
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
		err = errors.New("timeout waiting for connection response")
	}

	// Register interests
	dispatcherEventch <- NewRegisterInterestsEvent(
		[]*pb.Interest{
			&pb.Interest{
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
		err = errors.New("timeout waiting for register interests response")
	}

	// Unregister interests
	dispatcherEventch <- NewUnregisterInterestsEvent(
		[]*pb.Interest{
			&pb.Interest{
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
		err = errors.New("timeout waiting for unregister interests response")
	}

	// Disconnect
	dispatcherEventch <- clientdisp.NewDisconnectEvent(errch)
	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("Error disconnecting: %s", err)
		}
	case <-time.After(2 * time.Second):
		err = errors.New("timeout waiting for connection response")
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

func TestRegisterInterestsInvalid(t *testing.T) {
	channelID := "testchannel"
	dispatcher := New(
		newMockContext(), channelID,
		clientmocks.NewProviderFactory().Provider(
			ehmocks.NewConnection(
				clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory)),
				clientmocks.WithResults(
					clientmocks.NewResult(ehmocks.RegInterests, clientmocks.FailResult),
					clientmocks.NewResult(ehmocks.UnregInterests, clientmocks.FailResult),
				),
			),
		),
		clientmocks.CreateDiscoveryService(endpoint1, endpoint2),
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
		err = errors.New("timeout waiting for connection response")
	}

	// Register interests
	dispatcherEventch <- NewRegisterInterestsEvent(
		[]*pb.Interest{
			&pb.Interest{
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
		err = errors.New("timeout waiting for register interests response")
	}

	// Unregister interests
	dispatcherEventch <- NewUnregisterInterestsEvent(
		[]*pb.Interest{
			&pb.Interest{
				EventType: pb.EventType_FILTEREDBLOCK,
			},
		},
		errch)

	select {
	case err := <-errch:
		if err == nil {
			t.Fatalf("expecting error unregistering interests but got none")
		}
	case <-time.After(2 * time.Second):
		err = errors.New("timeout waiting for unregister interests response")
	}

	// Disconnect
	dispatcherEventch <- clientdisp.NewDisconnectEvent(errch)
	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("Error disconnecting: %s", err)
		}
	case <-time.After(2 * time.Second):
		err = errors.New("timeout waiting for connection response")
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

func TestTimedOutRegister(t *testing.T) {
	channelID := "testchannel"
	dispatcher := New(
		newMockContext(), channelID,
		clientmocks.NewProviderFactory().Provider(
			ehmocks.NewConnection(
				clientmocks.WithResults(
					clientmocks.NewResult(ehmocks.RegInterests, clientmocks.NoOpResult),
				),
				clientmocks.WithLedger(servicemocks.NewMockLedger(servicemocks.BlockEventFactory)),
			),
		),
		clientmocks.CreateDiscoveryService(endpoint1, endpoint2),
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
		err = errors.New("timeout waiting for connection response")
	}

	// Register interests
	dispatcherEventch <- NewRegisterInterestsEvent(
		[]*pb.Interest{
			&pb.Interest{
				EventType: pb.EventType_FILTEREDBLOCK,
			},
		},
		errch)

	select {
	case err := <-errch:
		if err == nil {
			t.Fatalf("expecting error due to no response from register interests but got none")
		}
	case <-time.After(2 * time.Second):
		err = errors.New("timeout waiting for register interests response")
	}

}

func TestBlockEvents(t *testing.T) {
	channelID := "testchannel"
	ledger := servicemocks.NewMockLedger(servicemocks.BlockEventFactory)
	dispatcher := New(
		newMockContext(), channelID,
		clientmocks.NewProviderFactory().Provider(
			ehmocks.NewConnection(
				clientmocks.WithLedger(ledger),
			),
		),
		clientmocks.CreateDiscoveryService(endpoint1, endpoint2),
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
		err = errors.New("timeout waiting for connection response")
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
			ehmocks.NewConnection(
				clientmocks.WithLedger(ledger),
			),
		),
		clientmocks.CreateDiscoveryService(endpoint1, endpoint2),
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
		err = errors.New("timeout waiting for connection response")
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

func newPeerConfig(peerURL string) *core.PeerConfig {
	return &core.PeerConfig{
		URL:         peerURL,
		GRPCOptions: make(map[string]interface{}),
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
