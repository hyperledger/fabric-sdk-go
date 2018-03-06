/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/options"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

type dsConnection interface {
	api.Connection
	Send(*ab.SeekInfo) error
}

// Dispatcher is responsible for handling all events, including connection and registration events originating from the client,
// and events originating from the channel event service. All events are processed in a single Go routine
// in order to avoid any race conditions and to ensure that events are processed in the order that they are received.
// This also avoids the need for synchronization.
type Dispatcher struct {
	clientdisp.Dispatcher
	seekRequest *SeekEvent
}

// New returns a new deliver dispatcher
func New(context fabcontext.Client, channelID string, connectionProvider api.ConnectionProvider, discoveryService fab.DiscoveryService, opts ...options.Opt) *Dispatcher {
	return &Dispatcher{
		Dispatcher: *clientdisp.New(context, channelID, connectionProvider, discoveryService, opts...),
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
		logger.Warnf("Unable to register channel since no connection was established.")
		return
	}

	ed.seekRequest = evt

	if err := ed.connection().Send(evt.SeekInfo); err != nil {
		evt.ErrCh <- errors.Wrapf(err, "error sending seek info for channel [%s]", ed.ChannelID())
		ed.seekRequest = nil
	}
}

func (ed *Dispatcher) handleDeliverResponseStatus(e esdispatcher.Event) {
	evt := e.(*pb.DeliverResponse_Status)

	if ed.seekRequest == nil {
		return
	}

	if ed.seekRequest.ErrCh != nil {
		if evt.Status != cb.Status_SUCCESS {
			ed.seekRequest.ErrCh <- errors.Errorf("received error status from seek info request: %s", evt.Status)
		} else {
			ed.seekRequest.ErrCh <- nil
		}
	}

	ed.seekRequest = nil
}

func (ed *Dispatcher) handleDeliverResponseBlock(e esdispatcher.Event) {
	ed.HandleBlock(e.(*pb.DeliverResponse_Block).Block)
}

func (ed *Dispatcher) handleDeliverResponseFilteredBlock(e esdispatcher.Event) {
	ed.HandleFilteredBlock(e.(*pb.DeliverResponse_FilteredBlock).FilteredBlock)
}

func (ed *Dispatcher) handleDisconnectedEvent(e esdispatcher.Event) {
	logger.Debug("Handling disconnected event...")

	if ed.seekRequest != nil && ed.seekRequest.ErrCh != nil {
		// We're in the middle of a seek request. Send an error response to the caller.
		ed.seekRequest.ErrCh <- errors.New("connection terminated")
	}
	ed.seekRequest = nil

	ed.Dispatcher.HandleDisconnectedEvent(e)
}

func (ed *Dispatcher) registerHandlers() {
	// Override Handlers
	ed.RegisterHandler(&clientdisp.DisconnectedEvent{}, ed.handleDisconnectedEvent)

	// Register handlers
	ed.RegisterHandler(&SeekEvent{}, ed.handleSeekEvent)
	ed.RegisterHandler(&pb.DeliverResponse_Status{}, ed.handleDeliverResponseStatus)
	ed.RegisterHandler(&pb.DeliverResponse_Block{}, ed.handleDeliverResponseBlock)
	ed.RegisterHandler(&pb.DeliverResponse_FilteredBlock{}, ed.handleDeliverResponseFilteredBlock)
}
