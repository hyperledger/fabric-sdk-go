/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient/connection"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

type ehConnection interface {
	api.Connection
	Send(emsg *pb.Event) error
}

// Dispatcher is responsible for handling all events, including connection and registration events originating from the client,
// and events originating from the event hub server. All events are processed in a single Go routine
// in order to avoid any race conditions. This avoids the need for synchronization.
type Dispatcher struct {
	clientdisp.Dispatcher
	regInterestsRequest   *RegisterInterestsEvent
	unregInterestsRequest *UnregisterInterestsEvent
}

// New creates a new event hub dispatcher
func New(context fabcontext.Client, chConfig fab.ChannelCfg, discovery fab.DiscoveryService, connectionProvider api.ConnectionProvider, opts ...options.Opt) *Dispatcher {
	return &Dispatcher{
		Dispatcher: *clientdisp.New(context, chConfig, discovery, connectionProvider, opts...),
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

func (ed *Dispatcher) connection() ehConnection {
	return ed.Dispatcher.Connection().(ehConnection)
}

func (ed *Dispatcher) handleRegInterestsEvent(e esdispatcher.Event) {
	evt := e.(*RegisterInterestsEvent)

	if ed.Connection() == nil {
		logger.Warn("Unable to register interests since no connection was established.")
		return
	}

	ed.regInterestsRequest = evt

	emsg := &pb.Event{
		Event: &pb.Event_Register{
			Register: &pb.Register{
				Events: evt.Interests,
			},
		},
	}

	if err := ed.connection().Send(emsg); err != nil {
		evt.ErrCh <- errors.Wrap(err, "error sending register interests event")
		ed.regInterestsRequest = nil
	}
}

func (ed *Dispatcher) handleUnregInterestsEvent(e esdispatcher.Event) {
	evt := e.(*UnregisterInterestsEvent)

	if ed.Connection() == nil {
		logger.Warn("Unable to unregister interests since no connection was established.")
		return
	}

	ed.unregInterestsRequest = evt

	emsg := &pb.Event{
		Event: &pb.Event_Unregister{
			Unregister: &pb.Unregister{
				Events: evt.Interests,
			},
		},
	}

	if err := ed.connection().Send(emsg); err != nil {
		evt.ErrCh <- errors.Wrap(err, "error sending unregister interests event")
		ed.unregInterestsRequest = nil
	}
}

func (ed *Dispatcher) handleRegInterestsResponse(e *pb.Event_Register) {
	if ed.regInterestsRequest == nil {
		return
	}

	if err := validateInterests(e.Register.Events, ed.regInterestsRequest.Interests); err != nil {
		logger.Warnf("Error registering interests: %s", err)
		if ed.regInterestsRequest.ErrCh != nil {
			ed.regInterestsRequest.ErrCh <- errors.Wrap(err, "error registering interests")
		}
	} else {
		// Send back a nil error to indicate that the operation was successful
		ed.regInterestsRequest.ErrCh <- nil
	}
	ed.regInterestsRequest = nil
}

func (ed *Dispatcher) handleUnregInterestsResponse(e *pb.Event_Unregister) {
	if ed.unregInterestsRequest == nil {
		return
	}

	if err := validateInterests(e.Unregister.Events, ed.unregInterestsRequest.Interests); err != nil {
		logger.Warnf("Error unregistering interests: %s", err)
		if ed.unregInterestsRequest.ErrCh != nil {
			ed.unregInterestsRequest.ErrCh <- errors.Wrap(err, "error unregistering interests")
		}
	} else {
		// Send back a nil error to indicate that the operation was successful
		ed.unregInterestsRequest.ErrCh <- nil
	}

	ed.unregInterestsRequest = nil
}

func validateInterests(have []*pb.Interest, want []*pb.Interest) error {
	if len(have) != len(want) {
		return errors.New("all interests were not registered/unregistered")
	}
	for _, hi := range have {
		found := false
		for _, wi := range want {
			if hi.EventType == wi.EventType {
				found = true
				break
			}
		}
		if !found {
			return errors.New("all interests were not registered/unregistered")
		}
	}
	return nil
}

func (ed *Dispatcher) handleEvent(e esdispatcher.Event) {
	ehevent := e.(*connection.Event)
	event := ehevent.Event.(*pb.Event)
	switch evt := event.Event.(type) {
	case *pb.Event_Block:
		ed.HandleBlock(evt.Block, ehevent.SourceURL)
	case *pb.Event_FilteredBlock:
		ed.HandleFilteredBlock(evt.FilteredBlock, ehevent.SourceURL)
	case *pb.Event_Register:
		ed.handleRegInterestsResponse(evt)
	case *pb.Event_Unregister:
		ed.handleUnregInterestsResponse(evt)
	default:
		logger.Warnf("Unsupported event type: %T", event.Event)
	}
}

func (ed *Dispatcher) handleDisconnectedEvent(e esdispatcher.Event) {
	if ed.regInterestsRequest != nil && ed.regInterestsRequest.ErrCh != nil {
		// We're in the middle of an interest registration. Send an error response to the caller.
		ed.regInterestsRequest.ErrCh <- errors.New("connection terminated")
	}
	ed.regInterestsRequest = nil
	ed.Dispatcher.HandleDisconnectedEvent(e)
}

func (ed *Dispatcher) registerHandlers() {
	// Override Handlers
	ed.RegisterHandler(&clientdisp.DisconnectedEvent{}, ed.handleDisconnectedEvent)

	// Register Handlers
	ed.RegisterHandler(&RegisterInterestsEvent{}, ed.handleRegInterestsEvent)
	ed.RegisterHandler(&UnregisterInterestsEvent{}, ed.handleUnregInterestsEvent)
	ed.RegisterHandler(&connection.Event{}, ed.handleEvent)
}
