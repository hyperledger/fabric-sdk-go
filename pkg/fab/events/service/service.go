/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package service

import (
	"runtime/debug"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/blockfilter"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	"github.com/pkg/errors"
)

const (
	// stopTimeout is the time that we wait for the dispatcher to stop.
	// It's hard-coded here since (at this point) it doesn't make sense to
	// expose it as an option.
	stopTimeout = 5 * time.Second
)

var logger = logging.NewLogger("fabsdk/fab")

// EventProducer produces events which are dispatched to clients
type EventProducer interface {
	// Register registers the given event channel with the event producer
	// and events are sent to this channel.
	Register(eventch chan<- interface{})
}

// Dispatcher is responsible for processing registration requests and block/filtered block events.
type Dispatcher interface {
	// Start starts the dispatcher, i.e. the dispatcher starts listening for requests/events
	Start() error

	// EventCh is the event channel over which to communicate with the dispatcher
	EventCh() (chan<- interface{}, error)

	// LastBlockNum returns the block number of the last block for which an event was received.
	LastBlockNum() uint64
}

// Service allows clients to register for channel events, such as filtered block, chaincode, and transaction status events.
type Service struct {
	params
	dispatcher Dispatcher
}

// New returns a new event service initialized with the given Dispatcher
func New(dispatcher Dispatcher, opts ...options.Opt) *Service {
	params := defaultParams()
	options.Apply(params, opts)

	return &Service{
		params:     *params,
		dispatcher: dispatcher,
	}
}

// Start starts the event service
func (s *Service) Start() error {
	return s.dispatcher.Start()
}

// Stop stops the event service
func (s *Service) Stop() {
	eventch, err := s.dispatcher.EventCh()
	if err != nil {
		logger.Warnf("Error stopping event service: %s", err)
		return
	}

	errch := make(chan error, 1)
	eventch <- dispatcher.NewStopEvent(errch)

	select {
	case err := <-errch:
		if err != nil {
			logger.Warnf("Error while stopping dispatcher: %s", err)
		}
	case <-time.After(stopTimeout):
		logger.Infof("Timed out waiting for dispatcher to stop")
	}
}

// StopAndTransfer stops the event service and transfers all event registrations into a snapshot.
func (s *Service) StopAndTransfer() (fab.EventSnapshot, error) {
	eventch, err := s.dispatcher.EventCh()
	if err != nil {
		logger.Warnf("Error stopping event service: %s", err)
		return nil, err
	}

	snapshotch := make(chan fab.EventSnapshot, 1)
	errch := make(chan error, 1)
	eventch <- dispatcher.NewStopAndTransferEvent(snapshotch, errch)

	select {
	case snapshot := <-snapshotch:
		return snapshot, nil
	case err := <-errch:
		logger.Warnf("Error while stopping dispatcher: %s", err)
		return nil, err
	case <-time.After(stopTimeout):
		logger.Warnf("Timed out waiting for dispatcher to stop")
		return nil, errors.New("timed out waiting for dispatcher to stop")
	}
}

// Transfer transfers all event registrations into a snapshot.
func (s *Service) Transfer() (fab.EventSnapshot, error) {
	eventch, err := s.dispatcher.EventCh()
	if err != nil {
		logger.Warnf("Error transferring registrations: %s", err)
		return nil, err
	}

	snapshotch := make(chan fab.EventSnapshot, 1)
	errch := make(chan error, 1)
	eventch <- dispatcher.NewTransferEvent(snapshotch, errch)

	select {
	case snapshot := <-snapshotch:
		return snapshot, nil
	case err := <-errch:
		logger.Warnf("Error while transferring event registrations into snapshot: %s", err)
		return nil, err
	case <-time.After(stopTimeout):
		logger.Warnf("Timed out waiting to transfer event registrations")
		return nil, errors.New("timed out waiting to transfer event registrations")
	}
}

// Submit submits an event for processing
func (s *Service) Submit(event interface{}) error {
	defer func() {
		// During shutdown, events may still be produced and we may
		// get a 'send on closed channel' panic. Just log and ignore the error.
		if p := recover(); p != nil {
			logger.Warnf("panic while submitting event: %s", p)
			debug.PrintStack()
		}
	}()

	eventch, err := s.dispatcher.EventCh()
	if err != nil {
		return errors.WithMessage(err, "Error submitting to event dispatcher")
	}
	eventch <- event

	return nil
}

// Dispatcher returns the event dispatcher
func (s *Service) Dispatcher() Dispatcher {
	return s.dispatcher
}

// RegisterBlockEvent registers for block events. If the client is not authorized to receive
// block events then an error is returned.
func (s *Service) RegisterBlockEvent(filter ...fab.BlockFilter) (fab.Registration, <-chan *fab.BlockEvent, error) {
	eventch := make(chan *fab.BlockEvent, s.eventConsumerBufferSize)
	regch := make(chan fab.Registration)
	errch := make(chan error)

	blockFilter := blockfilter.AcceptAny
	if len(filter) > 1 {
		return nil, nil, errors.New("only one block filter may be specified")
	}

	if len(filter) == 1 {
		blockFilter = filter[0]
	}

	if err := s.Submit(dispatcher.NewRegisterBlockEvent(blockFilter, eventch, regch, errch)); err != nil {
		return nil, nil, errors.WithMessage(err, "error registering for block events")
	}

	select {
	case response := <-regch:
		return response, eventch, nil
	case err := <-errch:
		return nil, nil, err
	}
}

// RegisterFilteredBlockEvent registers for filtered block events. If the client is not authorized to receive
// filtered block events then an error is returned.
func (s *Service) RegisterFilteredBlockEvent() (fab.Registration, <-chan *fab.FilteredBlockEvent, error) {
	eventch := make(chan *fab.FilteredBlockEvent, s.eventConsumerBufferSize)
	regch := make(chan fab.Registration)
	errch := make(chan error)

	if err := s.Submit(dispatcher.NewRegisterFilteredBlockEvent(eventch, regch, errch)); err != nil {
		return nil, nil, errors.WithMessage(err, "error registering for filtered block events")
	}

	select {
	case response := <-regch:
		return response, eventch, nil
	case err := <-errch:
		return nil, nil, err
	}
}

// RegisterChaincodeEvent registers for chaincode events. If the client is not authorized to receive
// chaincode events then an error is returned.
// - ccID is the chaincode ID for which events are to be received
// - eventFilter is the chaincode event name for which events are to be received
func (s *Service) RegisterChaincodeEvent(ccID, eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	if ccID == "" {
		return nil, nil, errors.New("chaincode ID is required")
	}
	if eventFilter == "" {
		return nil, nil, errors.New("event filter is required")
	}

	eventch := make(chan *fab.CCEvent, s.eventConsumerBufferSize)
	regch := make(chan fab.Registration)
	errch := make(chan error)

	if err := s.Submit(dispatcher.NewRegisterChaincodeEvent(ccID, eventFilter, eventch, regch, errch)); err != nil {
		return nil, nil, errors.WithMessage(err, "error registering for chaincode events")
	}

	select {
	case response := <-regch:
		return response, eventch, nil
	case err := <-errch:
		return nil, nil, err
	}
}

// RegisterTxStatusEvent registers for transaction status events. If the client is not authorized to receive
// transaction status events then an error is returned.
// - txID is the transaction ID for which events are to be received
func (s *Service) RegisterTxStatusEvent(txID string) (fab.Registration, <-chan *fab.TxStatusEvent, error) {
	if txID == "" {
		return nil, nil, errors.New("txID must be provided")
	}

	eventch := make(chan *fab.TxStatusEvent, s.eventConsumerBufferSize)
	regch := make(chan fab.Registration)
	errch := make(chan error)

	if err := s.Submit(dispatcher.NewRegisterTxStatusEvent(txID, eventch, regch, errch)); err != nil {
		return nil, nil, errors.WithMessage(err, "error registering for Tx Status events")
	}

	select {
	case response := <-regch:
		return response, eventch, nil
	case err := <-errch:
		return nil, nil, err
	}
}

// Unregister unregisters the given registration.
// - reg is the registration handle that was returned from one of the RegisterXXX functions
func (s *Service) Unregister(reg fab.Registration) {
	if err := s.Submit(dispatcher.NewUnregisterEvent(reg)); err != nil {
		logger.Warnf("Error unregistering: %s", err)
	}
}
