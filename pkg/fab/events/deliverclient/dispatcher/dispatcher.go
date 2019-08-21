/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	cb "github.com/hyperledger/fabric-protos-go/common"
	ab "github.com/hyperledger/fabric-protos-go/orderer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/connection"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

type dsConnection interface {
	api.Connection
	Send(seekInfo *ab.SeekInfo) error
}

// Dispatcher is responsible for handling all events, including connection and registration events originating from the client,
// and events originating from the channel event service. All events are processed in a single Go routine
// in order to avoid any race conditions and to ensure that events are processed in the order that they are received.
// This also avoids the need for synchronization.
type Dispatcher struct {
	*clientdisp.Dispatcher
}

// New returns a new deliver dispatcher
func New(context fabcontext.Client, chConfig fab.ChannelCfg, discoveryService fab.DiscoveryService, connectionProvider api.ConnectionProvider, opts ...options.Opt) *Dispatcher {
	return &Dispatcher{
		Dispatcher: clientdisp.New(context, chConfig, discoveryService, connectionProvider, opts...),
	}
}

// Start starts the dispatcher
func (ed *Dispatcher) Start() error {
	ed.registerHandlers()
	if err := ed.Dispatcher.Start(); err != nil {
		return errors.WithMessage(err, "error starting deliver event dispatcher")
	}
	return nil
}

func (ed *Dispatcher) connection() dsConnection {
	return ed.Dispatcher.Connection().(dsConnection)
}

func (ed *Dispatcher) handleSeekEvent(e esdispatcher.Event) {
	evt := e.(*SeekEvent)

	if ed.Connection() == nil {
		logger.Warn("Unable to register channel since no connection was established.")
		return
	}

	if err := ed.connection().Send(evt.SeekInfo); err != nil {
		evt.ErrCh <- errors.Wrapf(err, "error sending seek info for channel [%s]", ed.ChannelConfig().ID())
	} else {
		evt.ErrCh <- nil
	}
}

func (ed *Dispatcher) handleEvent(e esdispatcher.Event) {
	delevent := e.(*connection.Event)
	evt := delevent.Event.(*pb.DeliverResponse)
	switch response := evt.Type.(type) {
	case *pb.DeliverResponse_Status:
		ed.handleDeliverResponseStatus(response)
	case *pb.DeliverResponse_Block:
		ed.HandleBlock(response.Block, delevent.SourceURL)
	case *pb.DeliverResponse_FilteredBlock:
		ed.HandleFilteredBlock(response.FilteredBlock, delevent.SourceURL)
	default:
		logger.Errorf("handler not found for deliver response type %T", response)
	}
}

func (ed *Dispatcher) handleDeliverResponseStatus(evt *pb.DeliverResponse_Status) {
	logger.Debugf("Got deliver response status event: %#v", evt)

	if evt.Status == cb.Status_SUCCESS {
		return
	}

	logger.Warnf("Got deliver response status event: %#v. Disconnecting...", evt)

	errch := make(chan error, 1)
	ed.Dispatcher.HandleDisconnectEvent(&clientdisp.DisconnectEvent{
		Errch: errch,
	})
	err := <-errch
	if err != nil {
		logger.Warnf("Error disconnecting: %s", err)
	}

	ed.Dispatcher.HandleDisconnectedEvent(disconnectedEventFromStatus(evt.Status))
}

func (ed *Dispatcher) registerHandlers() {
	ed.RegisterHandler(&SeekEvent{}, ed.handleSeekEvent)
	ed.RegisterHandler(&connection.Event{}, ed.handleEvent)
}

func disconnectedEventFromStatus(status cb.Status) *clientdisp.DisconnectedEvent {
	err := errors.Errorf("got error status from deliver server: %s", status)

	if status == cb.Status_FORBIDDEN {
		return clientdisp.NewFatalDisconnectedEvent(err)
	}
	return clientdisp.NewDisconnectedEvent(err)
}
