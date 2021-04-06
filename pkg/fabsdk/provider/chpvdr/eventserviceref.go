/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"sync/atomic"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

const (
	defaultTimeout = 60 * time.Second
)

type eventClientProvider func() (fab.EventClient, error)

// EventClientRef holds a reference to the event client and manages its lifecycle.
// When the idle timeout has been reached then the event client is closed. The next time
// the event client ref is accessed, a new event client is created.
// The EventClientRef implements all of the functions of fab.EventService, so the
// EventClientRef may be used wherever an EventService is required.
type EventClientRef struct {
	ref         *lazyref.Reference
	provider    eventClientProvider
	eventClient fab.EventClient
	closed      int32
}

// NewEventClientRef returns a new EventClientRef
func NewEventClientRef(idleTimeout time.Duration, evtClientProvider eventClientProvider) *EventClientRef {
	clientRef := &EventClientRef{
		provider: evtClientProvider,
	}

	if idleTimeout == 0 {
		idleTimeout = defaultTimeout
	}

	clientRef.ref = lazyref.New(
		clientRef.initializer(),
		lazyref.WithFinalizer(clientRef.finalizer()),
		lazyref.WithIdleExpiration(idleTimeout),
	)

	return clientRef
}

// Close immediately closes the connection.
func (ref *EventClientRef) Close() {
	if !atomic.CompareAndSwapInt32(&ref.closed, 0, 1) {
		// Already closed
		return
	}

	logger.Debug("Closing the event client")
	ref.ref.Close()
}

// Closed returns true if the event client is closed
func (ref *EventClientRef) Closed() bool {
	return atomic.LoadInt32(&ref.closed) == 1
}

// RegisterBlockEvent registers for block events.
func (ref *EventClientRef) RegisterBlockEvent(filter ...fab.BlockFilter) (fab.Registration, <-chan *fab.BlockEvent, error) {
	service, err := ref.get()
	if err != nil {
		return nil, nil, err
	}
	return service.RegisterBlockEvent(filter...)
}

// RegisterFilteredBlockEvent registers for filtered block events.
func (ref *EventClientRef) RegisterFilteredBlockEvent() (fab.Registration, <-chan *fab.FilteredBlockEvent, error) {
	service, err := ref.get()
	if err != nil {
		return nil, nil, err
	}
	return service.RegisterFilteredBlockEvent()
}

// RegisterChaincodeEvent registers for chaincode events.
func (ref *EventClientRef) RegisterChaincodeEvent(ccID, eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	service, err := ref.get()
	if err != nil {
		return nil, nil, err
	}
	return service.RegisterChaincodeEvent(ccID, eventFilter)
}

// RegisterTxStatusEvent registers for transaction status events.
func (ref *EventClientRef) RegisterTxStatusEvent(txID string) (fab.Registration, <-chan *fab.TxStatusEvent, error) {
	service, err := ref.get()
	if err != nil {
		return nil, nil, err
	}
	return service.RegisterTxStatusEvent(txID)
}

// Unregister removes the given registration and closes the event channel.
func (ref *EventClientRef) Unregister(reg fab.Registration) {
	if service, err := ref.get(); err != nil {
		logger.Warnf("Error unregistering event registration: %s", err)
	} else {
		service.Unregister(reg)
	}
}

func (ref *EventClientRef) get() (fab.EventService, error) {
	if ref.Closed() {
		return nil, errors.New("event client is closed")
	}

	service, err := ref.ref.Get()
	if err != nil {
		return nil, err
	}
	return service.(fab.EventService), nil
}

func (ref *EventClientRef) initializer() lazyref.Initializer {
	return func() (interface{}, error) {
		if ref.eventClient != nil {
			// Already connected
			return ref.eventClient, nil
		}

		logger.Debug("Creating event client...")
		eventClient, err := ref.provider()
		if err != nil {
			return nil, err
		}
		logger.Debug("...connecting event client...")
		if err := eventClient.Connect(); err != nil {
			eventClient.Close()
			return nil, err
		}
		ref.eventClient = eventClient
		logger.Debug("...event client successfully connected.")
		return eventClient, nil
	}
}

func (ref *EventClientRef) finalizer() lazyref.Finalizer {
	return func(interface{}) {
		logger.Debug("Finalizer called")
		if ref.eventClient != nil {
			if ref.Closed() {
				logger.Debug("Forcing close the event client")
				ref.eventClient.Close()
			} else {
				logger.Debug("Closing the event client if no outstanding connections...")

				// Only close the client if there are not outstanding registrations
				if ref.eventClient.CloseIfIdle() {
					logger.Debug("... closed event client.")
					ref.eventClient = nil
				} else {
					logger.Debug("... event client was not closed since there are outstanding registrations.")
				}
			}
		}
	}
}
