/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"fmt"
	"math"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

// Dispatcher is responsible for handling all events, including connection and registration events originating from the client,
// and events originating from the event server. All events are processed in a single Go routine
// in order to avoid any race conditions and to ensure that events are processed in the order that they are received.
// This avoids the need for synchronization.
type Dispatcher struct {
	esdispatcher.Dispatcher
	params
	context                context.Client
	chConfig               fab.ChannelCfg
	connection             api.Connection
	connectionRegistration *ConnectionReg
	connectionProvider     api.ConnectionProvider
	discoveryService       fab.DiscoveryService
	ticker                 *time.Ticker
}

// New creates a new dispatcher
func New(context context.Client, chConfig fab.ChannelCfg, discoveryService fab.DiscoveryService, connectionProvider api.ConnectionProvider, opts ...options.Opt) *Dispatcher {
	params := defaultParams()
	options.Apply(params, opts)

	return &Dispatcher{
		Dispatcher:         *esdispatcher.New(opts...),
		params:             *params,
		context:            context,
		chConfig:           chConfig,
		discoveryService:   discoveryService,
		connectionProvider: connectionProvider,
	}
}

// Start starts the dispatcher
func (ed *Dispatcher) Start() error {
	ed.registerHandlers()

	if err := ed.Dispatcher.Start(); err != nil {
		return errors.WithMessage(err, "error starting client event dispatcher")
	}
	if ed.reconnectBlockHeightLagThreshold > 0 {
		ed.ticker = time.NewTicker(ed.blockHeightMonitorPeriod)
		go ed.monitorBlockHeight()
	}
	return nil
}

// ChannelConfig returns the channel configuration
func (ed *Dispatcher) ChannelConfig() fab.ChannelCfg {
	return ed.chConfig
}

// Connection returns the connection to the event server
func (ed *Dispatcher) Connection() api.Connection {
	return ed.connection
}

// HandleStopEvent handles a Stop event by clearing all registrations
// and stopping the listener
func (ed *Dispatcher) HandleStopEvent(e esdispatcher.Event) {
	// Remove all registrations and close the associated event channels
	// so that the client is notified that the registration has been removed
	ed.clearConnectionRegistration()
	if ed.ticker != nil {
		ed.ticker.Stop()
	}

	ed.Dispatcher.HandleStopEvent(e)
}

// HandleConnectEvent initiates a connection to the event server
func (ed *Dispatcher) HandleConnectEvent(e esdispatcher.Event) {
	evt := e.(*ConnectEvent)

	if ed.connection != nil {
		// Already connected. No error.
		evt.ErrCh <- nil
		return
	}

	eventch, err := ed.EventCh()
	if err != nil {
		evt.ErrCh <- err
		return
	}

	peers, err := ed.discoveryService.GetPeers()
	if err != nil {
		evt.ErrCh <- err
		return
	}

	if len(peers) == 0 {
		evt.ErrCh <- errors.New("no peers to connect to")
		return
	}

	peer, err := ed.loadBalancePolicy.Choose(ed.filterByBlockHeght(peers))
	if err != nil {
		evt.ErrCh <- err
		return
	}

	conn, err := ed.connectionProvider(ed.context, ed.chConfig, peer)
	if err != nil {
		logger.Warnf("error creating connection: %s", err)
		evt.ErrCh <- errors.WithMessage(err, fmt.Sprintf("could not create client conn"))
		return
	}

	ed.connection = conn

	go ed.connection.Receive(eventch)

	evt.ErrCh <- nil
}

// HandleDisconnectEvent disconnects from the event server
func (ed *Dispatcher) HandleDisconnectEvent(e esdispatcher.Event) {
	evt := e.(*DisconnectEvent)

	if ed.connection == nil {
		evt.Errch <- errors.New("connection already closed")
		return
	}

	logger.Debug("Closing connection...")

	ed.connection.Close()
	ed.connection = nil

	evt.Errch <- nil
}

// HandleRegisterConnectionEvent registers a connection listener
func (ed *Dispatcher) HandleRegisterConnectionEvent(e esdispatcher.Event) {
	evt := e.(*RegisterConnectionEvent)

	if ed.connectionRegistration != nil {
		evt.ErrCh <- errors.New("registration already exists for connection event")
		return
	}

	ed.connectionRegistration = evt.Reg
	evt.RegCh <- evt.Reg
}

// HandleConnectedEvent sends a 'connected' event to any registered listener
func (ed *Dispatcher) HandleConnectedEvent(e esdispatcher.Event) {
	evt := e.(*ConnectedEvent)

	logger.Debugf("Handling connected event: %+v", evt)

	if ed.connectionRegistration != nil && ed.connectionRegistration.Eventch != nil {
		select {
		case ed.connectionRegistration.Eventch <- NewConnectionEvent(true, nil):
		default:
			logger.Warn("Unable to send to connection event channel.")
		}
	}
}

// HandleDisconnectedEvent sends a 'disconnected' event to any registered listener
func (ed *Dispatcher) HandleDisconnectedEvent(e esdispatcher.Event) {
	evt := e.(*DisconnectedEvent)

	logger.Debugf("Disconnecting from event server: %s", evt.Err)

	if ed.connection != nil {
		ed.connection.Close()
		ed.connection = nil
	}

	if ed.connectionRegistration != nil {
		logger.Debugf("Disconnected from event server: %s", evt.Err)
		select {
		case ed.connectionRegistration.Eventch <- NewConnectionEvent(false, evt.Err):
		default:
			logger.Warn("Unable to send to connection event channel.")
		}
	} else {
		logger.Warnf("Disconnected from event server: %s", evt.Err)
	}
}

func (ed *Dispatcher) registerHandlers() {
	// Override existing handlers
	ed.RegisterHandler(&esdispatcher.StopEvent{}, ed.HandleStopEvent)

	// Register new handlers
	ed.RegisterHandler(&ConnectEvent{}, ed.HandleConnectEvent)
	ed.RegisterHandler(&DisconnectEvent{}, ed.HandleDisconnectEvent)
	ed.RegisterHandler(&ConnectedEvent{}, ed.HandleConnectedEvent)
	ed.RegisterHandler(&DisconnectedEvent{}, ed.HandleDisconnectedEvent)
	ed.RegisterHandler(&RegisterConnectionEvent{}, ed.HandleRegisterConnectionEvent)
}

func (ed *Dispatcher) clearConnectionRegistration() {
	if ed.connectionRegistration != nil {
		logger.Debug("Closing connection registration event channel.")
		close(ed.connectionRegistration.Eventch)
		ed.connectionRegistration = nil
	}
}

func (ed *Dispatcher) filterByBlockHeght(peers []fab.Peer) []fab.Peer {
	if ed.blockHeightLagThreshold < 0 || len(peers) == 1 {
		logger.Debugf("Returning all peers")
		return peers
	}

	maxHeight := getMaxBlockHeight(peers)
	logger.Debugf("Max block height of peers: %d", maxHeight)

	if maxHeight <= uint64(ed.blockHeightLagThreshold) {
		logger.Debugf("Max block height of peers is %d and lag threshold is %d so returning all peers", maxHeight, ed.blockHeightLagThreshold)
		return peers
	}

	cutoffHeight := maxHeight - uint64(ed.blockHeightLagThreshold)

	logger.Debugf("Choosing peers whose block heights are greater than the cutoff height %d ...", cutoffHeight)

	var retPeers []fab.Peer
	for _, p := range peers {
		peerState, ok := p.(fab.PeerState)
		if !ok {
			logger.Debugf("Accepting peer [%s] since it does not have state (may be a local peer)", p.URL())
			retPeers = append(retPeers, p)
		} else if peerState.BlockHeight() >= cutoffHeight {
			logger.Debugf("Accepting peer [%s] at block height %d which is greater than or equal to the cutoff %d", p.URL(), peerState.BlockHeight(), cutoffHeight)
			retPeers = append(retPeers, p)
		} else {
			logger.Debugf("Rejecting peer [%s] at block height %d which is less than the cutoff %d", p.URL(), peerState.BlockHeight(), cutoffHeight)
		}
	}
	return retPeers
}

func getMaxBlockHeight(peers []fab.Peer) uint64 {
	var maxHeight uint64
	for _, peer := range peers {
		peerState, ok := peer.(fab.PeerState)
		if ok {
			blockHeight := peerState.BlockHeight()
			if blockHeight > maxHeight {
				maxHeight = blockHeight
			}
		}
	}
	return maxHeight
}

func (ed *Dispatcher) monitorBlockHeight() {
	logger.Debugf("Starting block height monitor. Lag threshold: %d", ed.reconnectBlockHeightLagThreshold)
	for {
		if _, ok := <-ed.ticker.C; !ok {
			logger.Debugf("Stopping block height monitor")
			return
		}
		ed.checkBlockHeight()
	}
}

func (ed *Dispatcher) checkBlockHeight() {
	logger.Debugf("Checking block heights...")

	lastBlockReceived := ed.LastBlockNum()
	if lastBlockReceived == math.MaxUint64 {
		logger.Debugf("No blocks have been received yet")
		return
	}

	peers, err := ed.discoveryService.GetPeers()
	if err != nil {
		logger.Warnf("Error checking block height on peers: %s", err)
		return
	}

	connectedPeerBlockHeight := lastBlockReceived + 1
	maxHeight := getMaxBlockHeight(peers)

	logger.Debugf("Block height from blocks received: %d, Max block height from Discovery: %d", connectedPeerBlockHeight, maxHeight)

	cutoffHeight := maxHeight - uint64(ed.reconnectBlockHeightLagThreshold)

	if connectedPeerBlockHeight >= cutoffHeight {
		logger.Debugf("Block height from blocks received is %d which is greater than or equal to the cutoff %d", connectedPeerBlockHeight, cutoffHeight)
	} else {
		logger.Infof("Block height from blocks received is %d which is less than the cutoff %d. Disconnecting from the peer...", connectedPeerBlockHeight, cutoffHeight)
		if err := ed.disconnect(); err != nil {
			logger.Warnf("Error disconnecting event client: %s", err)
		} else {
			logger.Info("Successfully disconnected event client")
		}
	}
}

func (ed *Dispatcher) disconnect() error {
	eventch, err := ed.EventCh()
	if err != nil {
		return errors.WithMessage(err, "unable to get event dispatcher channel")
	}

	errch := make(chan error)
	eventch <- NewDisconnectEvent(errch)
	err = <-errch
	if err != nil {
		return err
	}

	// Send a DisconnectedEvent. This will trigger a reconnect.
	eventch <- NewDisconnectedEvent(nil)
	return nil
}
