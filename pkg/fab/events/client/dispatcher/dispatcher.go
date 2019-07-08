/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"sync"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

// Dispatcher is responsible for handling all events, including connection and registration events originating from the client,
// and events originating from the event server. All events are processed in a single Go routine
// in order to avoid any race conditions and to ensure that events are processed in the order that they are received.
// This avoids the need for synchronization.
type Dispatcher struct {
	*esdispatcher.Dispatcher
	params
	context                context.Client
	chConfig               fab.ChannelCfg
	connection             api.Connection
	connectionRegistration *ConnectionReg
	connectionProvider     api.ConnectionProvider
	discoveryService       fab.DiscoveryService
	peerResolver           peerresolver.Resolver
	peerMonitorDone        chan struct{}
	peer                   fab.Peer
	lock                   sync.RWMutex
}

// New creates a new dispatcher
func New(context context.Client, chConfig fab.ChannelCfg, discoveryService fab.DiscoveryService, connectionProvider api.ConnectionProvider, opts ...options.Opt) *Dispatcher {
	params := defaultParams(context, chConfig.ID())
	options.Apply(params, opts)

	dispatcher := &Dispatcher{
		Dispatcher:         esdispatcher.New(opts...),
		params:             *params,
		context:            context,
		chConfig:           chConfig,
		discoveryService:   discoveryService,
		connectionProvider: connectionProvider,
	}
	dispatcher.peerResolver = params.peerResolverProvider(dispatcher, context, chConfig.ID(), opts...)

	return dispatcher
}

// Start starts the dispatcher
func (ed *Dispatcher) Start() error {
	ed.registerHandlers()

	if err := ed.Dispatcher.Start(); err != nil {
		return errors.WithMessage(err, "error starting client event dispatcher")
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
	if ed.peerMonitorDone != nil {
		close(ed.peerMonitorDone)
		ed.peerMonitorDone = nil
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

	peer, err := ed.peerResolver.Resolve(peers)
	if err != nil {
		evt.ErrCh <- err
		return
	}

	conn, err := ed.connectionProvider(ed.context, ed.chConfig, peer)
	if err != nil {
		logger.Warnf("error creating connection: %s", err)
		evt.ErrCh <- errors.WithMessagef(err, "could not create client conn")
		return
	}

	ed.connection = conn
	ed.setConnectedPeer(peer)

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

	logger.Debug("Closing connection due to disconnect event...")

	ed.connection.Close()
	ed.connection = nil
	ed.setConnectedPeer(nil)

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

	if ed.peerMonitorPeriod > 0 {
		ed.peerMonitorDone = make(chan struct{})
		go ed.monitorPeer(ed.peerMonitorDone)
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

	if ed.peerMonitorDone != nil {
		close(ed.peerMonitorDone)
		ed.peerMonitorDone = nil
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

func (ed *Dispatcher) monitorPeer(done chan struct{}) {
	logger.Debugf("Starting peer monitor on channel [%s]", ed.chConfig.ID())

	ticker := time.NewTicker(ed.peerMonitorPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if ed.disconnected() {
				// Disconnected
				logger.Debugf("Client on channel [%s] has disconnected - stopping disconnect monitor", ed.chConfig.ID())
				return
			}
		case <-done:
			logger.Debugf("Stopping block height monitor on channel [%s]", ed.chConfig.ID())
			return
		}
	}
}

// disconnected checks if the currently connected peer should be disconnected
// Returns true if the client has been disconnected; false otherwise
func (ed *Dispatcher) disconnected() bool {
	connectedPeer := ed.ConnectedPeer()
	if connectedPeer == nil {
		logger.Debugf("Not connected yet")
		return false
	}

	logger.Debugf("Checking if event client should disconnect from peer [%s] on channel [%s]...", connectedPeer.URL(), ed.chConfig.ID())

	peers, err := ed.discoveryService.GetPeers()
	if err != nil {
		logger.Warnf("Error calling peer resolver: %s", err)
		return false
	}

	if !ed.peerResolver.ShouldDisconnect(peers, connectedPeer) {
		logger.Debugf("Event client will not disconnect from peer [%s] on channel [%s]...", connectedPeer.URL(), ed.chConfig.ID())
		return false
	}

	logger.Warnf("The peer resolver determined that the event client should be disconnected from connected peer [%s] on channel [%s]. Disconnecting ...", connectedPeer.URL(), ed.chConfig.ID())

	if err := ed.disconnect(); err != nil {
		logger.Warnf("Error disconnecting event client from peer [%s] on channel [%s]: %s", connectedPeer.URL(), ed.chConfig.ID(), err)
		return false
	}

	logger.Warnf("Successfully disconnected event client from peer [%s] on channel [%s]", connectedPeer.URL(), ed.chConfig.ID())
	return true
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
	eventch <- NewDisconnectedEvent(errors.New("event client was forced to disconnect"))
	return nil
}

func (ed *Dispatcher) setConnectedPeer(peer fab.Peer) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.peer = peer
}

// ConnectedPeer returns the connected peer
func (ed *Dispatcher) ConnectedPeer() fab.Peer {
	ed.lock.RLock()
	defer ed.lock.RUnlock()
	return ed.peer
}
