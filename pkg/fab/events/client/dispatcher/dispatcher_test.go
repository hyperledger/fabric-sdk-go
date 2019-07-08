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
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver/minblockheight"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver/preferorg"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		minblockheight.WithBlockHeightLagThreshold(5),
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

	// Wait for event that test is done
	err = <-errch
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestDisconnectIfBlockHeightLags(t *testing.T) {
	p1 := clientmocks.NewMockPeer("peer1", "grpcs://peer1.example.com:7051", 4)
	p2 := clientmocks.NewMockPeer("peer2", "grpcs://peer2.example.com:7051", 1)
	p3 := clientmocks.NewMockPeer("peer3", "grpcs://peer3.example.com:7051", 1)

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
		WithPeerResolver(minblockheight.NewResolver()),
		WithPeerMonitorPeriod(250*time.Millisecond),
		minblockheight.WithBlockHeightLagThreshold(2),
		minblockheight.WithReconnectBlockHeightThreshold(3),
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

	dispatcherEventch <- NewConnectedEvent()

	select {
	case e := <-connch:
		assert.Truef(t, e.Connected, "expecting connected event")
	case <-time.After(time.Second):
		t.Fatal("Expecting connected event but got none")
	}

	blockProducer := servicemocks.NewBlockProducer()
	dispatcherEventch <- esdispatcher.NewBlockEvent(blockProducer.NewBlock(channelID), sourceURL)
	dispatcherEventch <- esdispatcher.NewBlockEvent(blockProducer.NewBlock(channelID), sourceURL)
	dispatcherEventch <- esdispatcher.NewBlockEvent(blockProducer.NewBlock(channelID), sourceURL)
	dispatcherEventch <- esdispatcher.NewBlockEvent(blockProducer.NewBlock(channelID), sourceURL)
	dispatcherEventch <- esdispatcher.NewBlockEvent(blockProducer.NewBlock(channelID), sourceURL)

	p2.SetBlockHeight(15)

	select {
	case e := <-connch:
		assert.Falsef(t, e.Connected, "expecting disconnected event")
	case <-time.After(time.Second):
		t.Fatal("Expecting disconnected event but got none")
	}
}

// TestPreferLocalOrgConnection tests the scenario where an org wishes to connect to it's own peers
// if they are above the block height lag threshold but, if they fall below the threshold, the
// connection should be made to another org's peer. Once the local org's peers have caught up in
// block height, the connection to the other org's peer should be terminated.
func TestPreferLocalOrgConnection(t *testing.T) {
	channelID := "testchannel"
	org1MSP := "Org1MSP"
	org2MSP := "Org2MSP"

	p1O1 := clientmocks.NewMockStatefulPeer("p1_o1", "peer1.org1.com:7051", clientmocks.WithMSP(org1MSP), clientmocks.WithBlockHeight(4))
	p2O1 := clientmocks.NewMockStatefulPeer("p2_o1", "peer2.org1.com:7051", clientmocks.WithMSP(org1MSP), clientmocks.WithBlockHeight(3))
	p1O2 := clientmocks.NewMockStatefulPeer("p1_o2", "peer1.org2.com:7051", clientmocks.WithMSP(org2MSP), clientmocks.WithBlockHeight(10))
	p2O2 := clientmocks.NewMockStatefulPeer("p2_o2", "peer1.org2.com:7051", clientmocks.WithMSP(org2MSP), clientmocks.WithBlockHeight(12))

	dispatcher := New(
		fabmocks.NewMockContext(
			mspmocks.NewMockSigningIdentity("user1", org1MSP),
		),
		fabmocks.NewMockChannelCfg(channelID),
		clientmocks.NewDiscoveryService(p1O1, p2O1, p1O2, p2O2),
		clientmocks.NewProviderFactory().Provider(
			clientmocks.NewMockConnection(
				clientmocks.WithLedger(
					servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, sourceURL),
				),
			),
		),
		WithPeerResolver(preferorg.NewResolver()),
		WithPeerMonitorPeriod(250*time.Millisecond),
		minblockheight.WithBlockHeightLagThreshold(2),
		minblockheight.WithReconnectBlockHeightThreshold(3),
		WithLoadBalancePolicy(lbp.NewRandom()),
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

	dispatcherEventch <- NewConnectedEvent()

	select {
	case e := <-connch:
		assert.Truef(t, e.Connected, "expecting connected event")
	case <-time.After(time.Second):
		t.Fatal("Expecting connected event but got none")
	}

	// The initial connection should have been to an Org2 peer since their block heights are higher than Org1
	blockProducer := servicemocks.NewBlockProducer()
	dispatcherEventch <- esdispatcher.NewBlockEvent(blockProducer.NewBlock(channelID), sourceURL)
	dispatcherEventch <- esdispatcher.NewBlockEvent(blockProducer.NewBlock(channelID), sourceURL)
	dispatcherEventch <- esdispatcher.NewBlockEvent(blockProducer.NewBlock(channelID), sourceURL)
	dispatcherEventch <- esdispatcher.NewBlockEvent(blockProducer.NewBlock(channelID), sourceURL)
	dispatcherEventch <- esdispatcher.NewBlockEvent(blockProducer.NewBlock(channelID), sourceURL)

	p2O1.SetBlockHeight(15)

	select {
	case e := <-connch:
		assert.Falsef(t, e.Connected, "expecting disconnected event")
	case <-time.After(time.Second):
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

func TestOpts(t *testing.T) {
	channelID := "testchannel"

	config := &fabmocks.MockConfig{}
	context := fabmocks.NewMockContext(
		mspmocks.NewMockSigningIdentity("user1", "Org1MSP"),
	)
	context.SetEndpointConfig(config)

	t.Run("Default", func(t *testing.T) {
		config.SetCustomChannelConfig(channelID, &fab.ChannelEndpointConfig{
			Policies: fab.ChannelPolicies{
				EventService: fab.EventServicePolicy{},
			},
		})

		params := defaultParams(context, channelID)
		require.NotNil(t, params)
		assert.Equal(t, defaultPeerMonitorPeriod, params.peerMonitorPeriod)
		require.NotNil(t, params.peerResolverProvider)
	})

	t.Run("MinBlockStrategy", func(t *testing.T) {
		config.SetCustomChannelConfig(channelID, &fab.ChannelEndpointConfig{
			Policies: fab.ChannelPolicies{
				EventService: fab.EventServicePolicy{
					ResolverStrategy:  fab.MinBlockHeightStrategy,
					PeerMonitorPeriod: 7 * time.Second,
				},
			},
		})

		params := defaultParams(context, channelID)
		require.NotNil(t, params)
		assert.Equal(t, 7*time.Second, params.peerMonitorPeriod)
		require.NotNil(t, params.peerResolverProvider)
	})

	t.Run("PeerMonitor Off", func(t *testing.T) {
		config.SetCustomChannelConfig(channelID, &fab.ChannelEndpointConfig{
			Policies: fab.ChannelPolicies{
				EventService: fab.EventServicePolicy{
					ResolverStrategy: fab.PreferOrgStrategy,
					PeerMonitor:      fab.Disabled,
				},
			},
		})

		params := defaultParams(context, channelID)
		require.NotNil(t, params)
		assert.Equal(t, 0*time.Second, params.peerMonitorPeriod, "Expecting peer monitor to be disabled")
		require.NotNil(t, params.peerResolverProvider)
	})

	t.Run("Balanced Strategy", func(t *testing.T) {
		config.SetCustomChannelConfig(channelID, &fab.ChannelEndpointConfig{
			Policies: fab.ChannelPolicies{
				EventService: fab.EventServicePolicy{
					ResolverStrategy: fab.BalancedStrategy,
				},
			},
		})

		params := defaultParams(context, channelID)
		require.NotNil(t, params)
		assert.Equalf(t, 0*time.Second, params.peerMonitorPeriod, "Expecting peer monitor to be disabled for Balance strategy")
		require.NotNil(t, params.peerResolverProvider)
	})
}

func TestDisconnectedEvent(t *testing.T) {
	t.Run("Fatal", func(t *testing.T) {
		err := errors.New("injected error")
		evt := NewFatalDisconnectedEvent(err)
		require.NotNil(t, evt)
		require.NotNil(t, evt.Err)
		require.True(t, evt.Err.IsFatal())
	})

	t.Run("Non-fatal", func(t *testing.T) {
		err := errors.New("injected error")
		evt := NewDisconnectedEvent(err)
		require.NotNil(t, evt)
		require.NotNil(t, evt.Err)
		require.False(t, evt.Err.IsFatal())
	})
}
