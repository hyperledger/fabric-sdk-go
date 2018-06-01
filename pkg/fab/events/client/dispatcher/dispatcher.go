/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"fmt"

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

	peer, err := ed.loadBalancePolicy.Choose(peers)
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
