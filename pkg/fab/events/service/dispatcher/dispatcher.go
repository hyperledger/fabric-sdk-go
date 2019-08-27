/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"math"
	"reflect"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	ledgerutil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

const (
	dispatcherStateInitial = iota
	dispatcherStateStarted
	dispatcherStateStopped
)

// Handler is the handler for a given event type.
type Handler func(Event)

// HandlerRegistry contains the handlers for each type of event
type HandlerRegistry map[reflect.Type]Handler

// Dispatcher is responsible for handling all events, including connection and registration events originating from the client,
// and events originating from the channel event service. All events are processed in a single Go routine
// in order to avoid any race conditions and to ensure that events are processed in the order in which they are received.
// This also avoids the need for synchronization.
// The lastBlockNum member MUST be first to ensure it stays 64-bit aligned on 32-bit machines.
type Dispatcher struct {
	lastBlockNum uint64 // Must be first, do not move
	params
	updateLastBlockInfoOnly    bool
	state                      int32
	eventch                    chan interface{}
	blockRegistrations         []*BlockReg
	filteredBlockRegistrations []*FilteredBlockReg
	handlers                   map[reflect.Type]Handler
	txRegistrations            map[string]*TxStatusReg
	ccRegistrations            map[string]*ChaincodeReg
}

// New creates a new Dispatcher.
func New(opts ...options.Opt) *Dispatcher {
	logger.Debug("Creating new dispatcher.")

	params := defaultParams()
	options.Apply(params, opts)

	return &Dispatcher{
		params:          *params,
		handlers:        make(map[reflect.Type]Handler),
		eventch:         make(chan interface{}, params.eventConsumerBufferSize),
		txRegistrations: make(map[string]*TxStatusReg),
		ccRegistrations: make(map[string]*ChaincodeReg),
		state:           dispatcherStateInitial,
		lastBlockNum:    math.MaxUint64,
	}
}

// RegisterHandlers registers all of the handlers by event type
func (ed *Dispatcher) RegisterHandlers() {
	ed.RegisterHandler(&RegisterChaincodeEvent{}, ed.handleRegisterCCEvent)
	ed.RegisterHandler(&RegisterTxStatusEvent{}, ed.handleRegisterTxStatusEvent)
	ed.RegisterHandler(&RegisterBlockEvent{}, ed.handleRegisterBlockEvent)
	ed.RegisterHandler(&RegisterFilteredBlockEvent{}, ed.handleRegisterFilteredBlockEvent)
	ed.RegisterHandler(&UnregisterEvent{}, ed.handleUnregisterEvent)
	ed.RegisterHandler(&StopEvent{}, ed.HandleStopEvent)
	ed.RegisterHandler(&TransferEvent{}, ed.HandleTransferEvent)
	ed.RegisterHandler(&StopAndTransferEvent{}, ed.HandleStopAndTransferEvent)
	ed.RegisterHandler(&RegistrationInfoEvent{}, ed.handleRegistrationInfoEvent)

	// The following events are used for testing only
	ed.RegisterHandler(&fab.BlockEvent{}, ed.handleBlockEvent)
	ed.RegisterHandler(&fab.FilteredBlockEvent{}, ed.handleFilteredBlockEvent)
}

// EventCh returns the channel to which events may be posted
func (ed *Dispatcher) EventCh() (chan<- interface{}, error) {
	state := ed.getState()
	if state == dispatcherStateStarted {
		return ed.eventch, nil
	}
	return nil, errors.Errorf("dispatcher not started - Current state [%d]", state)
}

// Start starts dispatching events as they arrive. All events are processed in
// a single Go routine in order to avoid any race conditions
func (ed *Dispatcher) Start() error {
	if !ed.setState(dispatcherStateInitial, dispatcherStateStarted) {
		return errors.New("cannot start dispatcher since it's not in its initial state")
	}

	ed.RegisterHandlers()

	if err := ed.initRegistrations(); err != nil {
		return errors.WithMessage(err, "error initializing registrations")
	}

	if ed.initialLastBlockNum > 0 {
		if err := ed.updateLastBlockNum(ed.initialLastBlockNum); err != nil {
			logger.Warnf("Unable to update last block num to %d: %s", ed.initialLastBlockNum, err)
		}
	}

	go func() {
		for {
			if ed.getState() == dispatcherStateStopped {
				break
			}

			logger.Debug("Listening for events...")
			e, ok := <-ed.eventch
			if !ok {
				break
			}

			logger.Debugf("Received event: %+v", reflect.TypeOf(e))

			if handler, ok := ed.handlers[reflect.TypeOf(e)]; ok {
				logger.Debugf("Dispatching event: %+v", reflect.TypeOf(e))
				handler(e)
			} else {
				logger.Errorf("Handler not found for: %s", reflect.TypeOf(e))
			}
		}
		logger.Debug("Exiting event dispatcher")
	}()
	return nil
}

// LastBlockNum returns the block number of the last block for which an event was received.
func (ed *Dispatcher) LastBlockNum() uint64 {
	return atomic.LoadUint64(&ed.lastBlockNum)
}

// updateLastBlockNum updates the value of lastBlockNum and
// returns the updated value.
func (ed *Dispatcher) updateLastBlockNum(blockNum uint64) error {
	// The Deliver Service shouldn't be sending blocks out of order.
	// Log an error if we detect this happening.
	lastBlockNum := atomic.LoadUint64(&ed.lastBlockNum)
	if lastBlockNum == math.MaxUint64 || blockNum > lastBlockNum {
		atomic.StoreUint64(&ed.lastBlockNum, blockNum)
		logger.Debugf("Updated last block received to %d", blockNum)
		return nil
	}
	return errors.Errorf("Expecting a block number greater than %d but received block number %d", lastBlockNum, blockNum)
}

func (ed *Dispatcher) initRegistrations() error {
	logger.Debugf("Initializing registrations...")
	for _, reg := range ed.initialBlockRegistrations {
		logger.Debugf("Adding block registration")
		ed.registerBlockEvent(reg)
	}
	for _, reg := range ed.initialFilteredBlockRegistrations {
		logger.Debugf("Adding filtered block registration")
		ed.registerFilteredBlockEvent(reg)
	}
	for _, reg := range ed.initialCCRegistrations {
		logger.Debugf("Adding CC registration: CC ID [%s], Event filter [%s]", reg.ChaincodeID, reg.EventFilter)
		if err := ed.registerCCEvent(reg); err != nil {
			logger.Warnf("Error adding CC registration: %s", err)
			return err
		}
	}
	for _, reg := range ed.initialTxStatusRegistrations {
		logger.Debugf("Adding TxStatus registration: TxID [%s]", reg.TxID)
		if err := ed.registerTxStatusEvent(reg); err != nil {
			logger.Warnf("Error adding TxStatus registration: %s", err)
			return err
		}
	}

	return nil
}

func (ed *Dispatcher) clearRegistrations(closeChannel bool) {
	ed.clearBlockRegistrations(closeChannel)
	ed.clearFilteredBlockRegistrations(closeChannel)
	ed.clearTxRegistrations(closeChannel)
	ed.clearChaincodeRegistrations(closeChannel)
}

// clearBlockRegistrations removes all block registrations and closes the corresponding event channels.
// The listener will receive a 'closed' event to indicate that the channel has been closed.
func (ed *Dispatcher) clearBlockRegistrations(closeChannel bool) {
	if closeChannel {
		for _, reg := range ed.blockRegistrations {
			close(reg.Eventch)
		}
	}
	ed.blockRegistrations = nil
}

// clearFilteredBlockRegistrations removes all filtered block registrations and closes the corresponding event channels.
// The listener will receive a 'closed' event to indicate that the channel has been closed.
func (ed *Dispatcher) clearFilteredBlockRegistrations(closeChannel bool) {
	if closeChannel {
		for _, reg := range ed.filteredBlockRegistrations {
			close(reg.Eventch)
		}
	}
	ed.filteredBlockRegistrations = nil
}

// clearTxRegistrations removes all transaction registrations and closes the corresponding event channels.
// The listener will receive a 'closed' event to indicate that the channel has been closed.
func (ed *Dispatcher) clearTxRegistrations(closeChannel bool) {
	if closeChannel {
		for _, reg := range ed.txRegistrations {
			logger.Debugf("Closing TX registration event channel for TxID [%s].", reg.TxID)
			close(reg.Eventch)
		}
	}
	ed.txRegistrations = make(map[string]*TxStatusReg)
}

// clearChaincodeRegistrations removes all chaincode registrations and closes the corresponding event channels.
// The listener will receive a 'closed' event to indicate that the channel has been closed.
func (ed *Dispatcher) clearChaincodeRegistrations(closeChannel bool) {
	if closeChannel {
		for _, reg := range ed.ccRegistrations {
			logger.Debugf("Closing chaincode registration event channel for CC ID [%s] and event filter [%s].", reg.ChaincodeID, reg.EventFilter)
			close(reg.Eventch)
		}
	}
	ed.ccRegistrations = make(map[string]*ChaincodeReg)
}

// HandleStopEvent stops the dispatcher and unregisters all event registration.
// The Dispatcher is no longer usable.
func (ed *Dispatcher) HandleStopEvent(e Event) {
	event := e.(*StopEvent)

	logger.Debugf("Stopping dispatcher...")
	if !ed.setState(dispatcherStateStarted, dispatcherStateStopped) {
		logger.Warn("Cannot stop event dispatcher since it's already stopped.")
		event.ErrCh <- errors.New("dispatcher already stopped")
		return
	}

	// Remove all registrations and close the associated event channels
	// so that the client is notified that the registration has been removed
	ed.clearRegistrations(true)

	event.ErrCh <- nil
}

// HandleTransferEvent transfers all event registrations into a EventSnapshot.
func (ed *Dispatcher) HandleTransferEvent(e Event) {
	event := e.(*TransferEvent)

	event.SnapshotCh <- ed.newSnapshot()

	// Remove all registrations but don't close the associated event channels
	ed.clearRegistrations(false)
}

// HandleStopAndTransferEvent stops the dispatcher and transfers all event registrations
// into a EventSnapshot.
// The Dispatcher is no longer usable.
func (ed *Dispatcher) HandleStopAndTransferEvent(e Event) {
	event := e.(*StopAndTransferEvent)

	logger.Debugf("Stopping dispatcher...")
	if !ed.setState(dispatcherStateStarted, dispatcherStateStopped) {
		logger.Warn("Cannot stop event dispatcher since it's already stopped.")
		event.ErrCh <- errors.New("dispatcher already stopped")
		return
	}

	event.SnapshotCh <- ed.newSnapshot()

	// Remove all registrations but don't close the associated event channels
	ed.clearRegistrations(false)
}

func (ed *Dispatcher) handleRegisterBlockEvent(e Event) {
	event := e.(*RegisterBlockEvent)

	ed.registerBlockEvent(event.Reg)
	event.RegCh <- event.Reg
}

func (ed *Dispatcher) registerBlockEvent(reg *BlockReg) {
	ed.blockRegistrations = append(ed.blockRegistrations, reg)
}

func (ed *Dispatcher) handleRegisterFilteredBlockEvent(e Event) {
	event := e.(*RegisterFilteredBlockEvent)
	ed.registerFilteredBlockEvent(event.Reg)
	event.RegCh <- event.Reg
}

func (ed *Dispatcher) registerFilteredBlockEvent(reg *FilteredBlockReg) {
	ed.filteredBlockRegistrations = append(ed.filteredBlockRegistrations, reg)
}

func (ed *Dispatcher) handleRegisterCCEvent(e Event) {
	event := e.(*RegisterChaincodeEvent)

	regExp, err := regexp.Compile(event.Reg.EventFilter)
	if err != nil {
		event.ErrCh <- errors.Wrapf(err, "error compiling regular expression for event filter [%s]", event.Reg.EventFilter)
	} else {
		event.Reg.EventRegExp = regExp
		if err := ed.registerCCEvent(event.Reg); err != nil {
			event.ErrCh <- err
		} else {
			event.RegCh <- event.Reg
		}
	}
}

func (ed *Dispatcher) registerCCEvent(reg *ChaincodeReg) error {
	key := getCCKey(reg.ChaincodeID, reg.EventFilter)
	if _, exists := ed.ccRegistrations[key]; exists {
		return errors.Errorf("registration already exists for chaincode [%s] and event [%s]", reg.ChaincodeID, reg.EventFilter)
	}
	ed.ccRegistrations[key] = reg
	return nil
}

func (ed *Dispatcher) handleRegisterTxStatusEvent(e Event) {
	event := e.(*RegisterTxStatusEvent)

	if err := ed.registerTxStatusEvent(event.Reg); err != nil {
		event.ErrCh <- err
	} else {
		event.RegCh <- event.Reg
	}
}

func (ed *Dispatcher) registerTxStatusEvent(reg *TxStatusReg) error {
	if _, exists := ed.txRegistrations[reg.TxID]; exists {
		return errors.Errorf("registration already exists for TX ID [%s]", reg.TxID)
	}
	ed.txRegistrations[reg.TxID] = reg
	return nil
}

func (ed *Dispatcher) handleUnregisterEvent(e Event) {
	event := e.(*UnregisterEvent)

	var err error
	switch registration := event.Reg.(type) {
	case *BlockReg:
		err = ed.unregisterBlockEvents(registration)
	case *FilteredBlockReg:
		err = ed.unregisterFilteredBlockEvents(registration)
	case *ChaincodeReg:
		err = ed.unregisterCCEvents(registration)
	case *TxStatusReg:
		err = ed.unregisterTXEvents(registration)
	default:
		err = errors.Errorf("Unsupported registration type: %+v", reflect.TypeOf(registration))
	}
	if err != nil {
		logger.Warnf("Error in unregister: %s", err)
	}
}

func (ed *Dispatcher) handleBlockEvent(e Event) {
	evt := e.(*fab.BlockEvent)
	ed.HandleBlock(evt.Block, evt.SourceURL)
}

func (ed *Dispatcher) handleFilteredBlockEvent(e Event) {
	evt := e.(*fab.FilteredBlockEvent)
	ed.HandleFilteredBlock(evt.FilteredBlock, evt.SourceURL)
}

func (ed *Dispatcher) handleRegistrationInfoEvent(e Event) {
	evt := e.(*RegistrationInfoEvent)

	regInfo := &RegistrationInfo{
		NumBlockRegistrations:         len(ed.blockRegistrations),
		NumFilteredBlockRegistrations: len(ed.filteredBlockRegistrations),
		NumCCRegistrations:            len(ed.ccRegistrations),
		NumTxStatusRegistrations:      len(ed.txRegistrations),
	}

	regInfo.TotalRegistrations =
		regInfo.NumBlockRegistrations + regInfo.NumFilteredBlockRegistrations + regInfo.NumCCRegistrations + regInfo.NumTxStatusRegistrations

	evt.RegInfoCh <- regInfo
}

func (ed *Dispatcher) newSnapshot() fab.EventSnapshot {
	var ccRegistrations []*ChaincodeReg
	for _, reg := range ed.ccRegistrations {
		logger.Debugf("Adding CC registration to snaphot - CC ID [%s], Event filter [%s]", reg.ChaincodeID, reg.EventFilter)
		ccRegistrations = append(ccRegistrations, reg)
	}

	var txRegistrations []*TxStatusReg
	for _, reg := range ed.txRegistrations {
		logger.Debugf("Adding TxStatus registration to snaphot - TxID [%s]", reg.TxID)
		txRegistrations = append(txRegistrations, reg)
	}

	return &snapshot{
		lastBlockReceived:          ed.LastBlockNum(),
		blockRegistrations:         ed.blockRegistrations,
		filteredBlockRegistrations: ed.filteredBlockRegistrations,
		ccRegistrations:            ccRegistrations,
		txStatusRegistrations:      txRegistrations,
	}
}

// HandleBlock handles a block event
func (ed *Dispatcher) HandleBlock(block *cb.Block, sourceURL string) {
	logger.Debugf("Handling block event - Block #%d", block.Header.Number)

	if err := ed.updateLastBlockNum(block.Header.Number); err != nil {
		logger.Error(err.Error())
		return
	}

	if ed.updateLastBlockInfoOnly {
		ed.updateLastBlockInfoOnly = false
		return
	}

	logger.Debug("Publishing block event...")
	ed.publishBlockEvents(block, sourceURL)
	ed.publishFilteredBlockEvents(toFilteredBlock(block), sourceURL)
}

// HandleFilteredBlock handles a filtered block event
func (ed *Dispatcher) HandleFilteredBlock(fblock *pb.FilteredBlock, sourceURL string) {
	logger.Debugf("Handling filtered block event - Block #%d", fblock.Number)

	if err := ed.updateLastBlockNum(fblock.Number); err != nil {
		logger.Error(err.Error())
		return
	}

	if ed.updateLastBlockInfoOnly {
		ed.updateLastBlockInfoOnly = false
		return
	}

	logger.Debug("Publishing filtered block event...")
	ed.publishFilteredBlockEvents(fblock, sourceURL)
}

func (ed *Dispatcher) unregisterBlockEvents(registration *BlockReg) error {
	for i, reg := range ed.blockRegistrations {
		if reg == registration {
			// Move the 0'th item to i and then delete the 0'th item
			ed.blockRegistrations[i] = ed.blockRegistrations[0]
			ed.blockRegistrations = ed.blockRegistrations[1:]
			close(reg.Eventch)
			return nil
		}
	}
	return errors.New("the provided registration is invalid")
}

func (ed *Dispatcher) unregisterFilteredBlockEvents(registration *FilteredBlockReg) error {
	for i, reg := range ed.filteredBlockRegistrations {
		if reg == registration {
			// Move the 0'th item to i and then delete the 0'th item
			ed.filteredBlockRegistrations[i] = ed.filteredBlockRegistrations[0]
			ed.filteredBlockRegistrations = ed.filteredBlockRegistrations[1:]
			close(reg.Eventch)
			return nil
		}
	}
	return errors.New("the provided registration is invalid")
}

func (ed *Dispatcher) unregisterCCEvents(registration *ChaincodeReg) error {
	key := getCCKey(registration.ChaincodeID, registration.EventFilter)
	reg, ok := ed.ccRegistrations[key]
	if !ok {
		return errors.New("the provided registration is invalid")
	}

	logger.Debugf("Unregistering CC event for CC ID [%s] and event filter [%s]...", registration.ChaincodeID, registration.EventFilter)
	close(reg.Eventch)
	delete(ed.ccRegistrations, key)
	return nil
}

func (ed *Dispatcher) unregisterTXEvents(registration *TxStatusReg) error {
	reg, ok := ed.txRegistrations[registration.TxID]
	if !ok {
		return errors.New("the provided registration is invalid")
	}

	logger.Debugf("Unregistering Tx Status event for TxID [%s]...", registration.TxID)
	close(reg.Eventch)
	delete(ed.txRegistrations, registration.TxID)
	return nil
}

func (ed *Dispatcher) publishBlockEvents(block *cb.Block, sourceURL string) {
	for _, reg := range ed.blockRegistrations {
		if !reg.Filter(block) {
			logger.Debugf("Not sending block event for block #%d since it was filtered out.", block.Header.Number)
			continue
		}

		if ed.eventConsumerTimeout < 0 {
			select {
			case reg.Eventch <- NewBlockEvent(block, sourceURL):
			default:
				logger.Warn("Unable to send to block event channel.")
			}
		} else if ed.eventConsumerTimeout == 0 {
			reg.Eventch <- NewBlockEvent(block, sourceURL)
		} else {
			select {
			case reg.Eventch <- NewBlockEvent(block, sourceURL):
			case <-time.After(ed.eventConsumerTimeout):
				logger.Warn("Timed out sending block event.")
			}
		}
	}
}

func (ed *Dispatcher) publishFilteredBlockEvents(fblock *pb.FilteredBlock, sourceURL string) {
	if fblock == nil {
		logger.Warn("Filtered block is nil. Event will not be published")
		return
	}

	logger.Debugf("Publishing filtered block event: %#v", fblock)

	checkFilteredBlockRegistrations(ed, fblock, sourceURL)

	for _, tx := range fblock.FilteredTransactions {
		ed.publishTxStatusEvents(tx, fblock.Number, sourceURL)

		// Only send a chaincode event if the transaction has committed
		if tx.TxValidationCode == pb.TxValidationCode_VALID {
			txActions := tx.GetTransactionActions()
			if txActions == nil {
				continue
			}
			if len(txActions.ChaincodeActions) == 0 {
				logger.Debugf("No chaincode action found for TxID[%s], block[%d], source URL[%s]", tx.Txid, fblock.Number, sourceURL)
			}
			for _, action := range txActions.ChaincodeActions {
				if action.ChaincodeEvent != nil {
					ed.publishCCEvents(action.ChaincodeEvent, fblock.Number, sourceURL)
				}
			}
		} else {
			logger.Debugf("Cannot publish CCEvents for block[%d] and source URL[%s] since Tx Validation Code[%d] is not valid", fblock.Number, sourceURL, tx.TxValidationCode)
		}
	}
}

func checkFilteredBlockRegistrations(ed *Dispatcher, fblock *pb.FilteredBlock, sourceURL string) {
	for _, reg := range ed.filteredBlockRegistrations {
		if ed.eventConsumerTimeout < 0 {
			select {
			case reg.Eventch <- NewFilteredBlockEvent(fblock, sourceURL):
			default:
				logger.Warn("Unable to send to filtered block event channel.")
			}
		} else if ed.eventConsumerTimeout == 0 {
			reg.Eventch <- NewFilteredBlockEvent(fblock, sourceURL)
		} else {
			select {
			case reg.Eventch <- NewFilteredBlockEvent(fblock, sourceURL):
			case <-time.After(ed.eventConsumerTimeout):
				logger.Warn("Timed out sending filtered block event.")
			}
		}
	}
}

func (ed *Dispatcher) publishTxStatusEvents(tx *pb.FilteredTransaction, blockNum uint64, sourceURL string) {
	logger.Debugf("Publishing Tx Status event for TxID [%s]...", tx.Txid)
	if reg, ok := ed.txRegistrations[tx.Txid]; ok {
		logger.Debugf("Sending Tx Status event for TxID [%s] to registrant...", tx.Txid)

		if ed.eventConsumerTimeout < 0 {
			select {
			case reg.Eventch <- NewTxStatusEvent(tx.Txid, tx.TxValidationCode, blockNum, sourceURL):
			default:
				logger.Warn("Unable to send to Tx Status event channel.")
			}
		} else if ed.eventConsumerTimeout == 0 {
			reg.Eventch <- NewTxStatusEvent(tx.Txid, tx.TxValidationCode, blockNum, sourceURL)
		} else {
			select {
			case reg.Eventch <- NewTxStatusEvent(tx.Txid, tx.TxValidationCode, blockNum, sourceURL):
			case <-time.After(ed.eventConsumerTimeout):
				logger.Warn("Timed out sending Tx Status event.")
			}
		}
	}
}

func (ed *Dispatcher) publishCCEvents(ccEvent *pb.ChaincodeEvent, blockNum uint64, sourceURL string) {
	for _, reg := range ed.ccRegistrations {
		logger.Debugf("Matching CCEvent[%s,%s] against Reg[%s,%s] ...", ccEvent.ChaincodeId, ccEvent.EventName, reg.ChaincodeID, reg.EventFilter)
		if reg.ChaincodeID == ccEvent.ChaincodeId && reg.EventRegExp.MatchString(ccEvent.EventName) {
			logger.Debugf("... matched CCEvent[%s,%s] against Reg[%s,%s]", ccEvent.ChaincodeId, ccEvent.EventName, reg.ChaincodeID, reg.EventFilter)

			if ed.eventConsumerTimeout < 0 {
				select {
				case reg.Eventch <- NewChaincodeEvent(ccEvent.ChaincodeId, ccEvent.EventName, ccEvent.TxId, ccEvent.Payload, blockNum, sourceURL):
				default:
					logger.Warn("Unable to send to CC event channel.")
				}
			} else if ed.eventConsumerTimeout == 0 {
				reg.Eventch <- NewChaincodeEvent(ccEvent.ChaincodeId, ccEvent.EventName, ccEvent.TxId, ccEvent.Payload, blockNum, sourceURL)
			} else {
				select {
				case reg.Eventch <- NewChaincodeEvent(ccEvent.ChaincodeId, ccEvent.EventName, ccEvent.TxId, ccEvent.Payload, blockNum, sourceURL):
				case <-time.After(ed.eventConsumerTimeout):
					logger.Warn("Timed out sending CC event.")
				}
			}
		}
	}
}

// RegisterHandler registers an event handler
func (ed *Dispatcher) RegisterHandler(t interface{}, h Handler) {
	htype := reflect.TypeOf(t)
	if _, ok := ed.handlers[htype]; !ok {
		logger.Debugf("Registering handler for %s on dispatcher %T", htype, ed)
		ed.handlers[htype] = h
	} else {
		logger.Debugf("Cannot register handler %s on dispatcher %T since it's already registered", htype, ed)
	}
}

//UpdateLastBlockInfoOnly sets is next event should only be used for updating last block info.
func (ed *Dispatcher) UpdateLastBlockInfoOnly() {
	ed.updateLastBlockInfoOnly = true
}

func getCCKey(ccID, eventFilter string) string {
	return ccID + "/" + eventFilter
}

func toFilteredBlock(block *cb.Block) *pb.FilteredBlock {
	var channelID string
	var filteredTxs []*pb.FilteredTransaction
	txFilter := ledgerutil.TxValidationFlags(block.Metadata.Metadata[cb.BlockMetadataIndex_TRANSACTIONS_FILTER])

	for i, data := range block.Data.Data {
		filteredTx, chID, err := getFilteredTx(data, txFilter.Flag(i))
		if err != nil {
			logger.Warnf("error extracting Envelope from block: %s", err)
			continue
		}
		channelID = chID
		filteredTxs = append(filteredTxs, filteredTx)
	}

	return &pb.FilteredBlock{
		ChannelId:            channelID,
		Number:               block.Header.Number,
		FilteredTransactions: filteredTxs,
	}
}

func getFilteredTx(data []byte, txValidationCode pb.TxValidationCode) (*pb.FilteredTransaction, string, error) {
	env, err := protoutil.GetEnvelopeFromBlock(data)
	if err != nil {
		return nil, "", errors.Wrap(err, "error extracting Envelope from block")
	}
	if env == nil {
		return nil, "", errors.New("nil envelope")
	}

	payload, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		return nil, "", errors.Wrap(err, "error extracting Payload from envelope")
	}

	channelHeaderBytes := payload.Header.ChannelHeader
	channelHeader := &cb.ChannelHeader{}
	if err := proto.Unmarshal(channelHeaderBytes, channelHeader); err != nil {
		return nil, "", errors.Wrap(err, "error extracting ChannelHeader from payload")
	}

	filteredTx := &pb.FilteredTransaction{
		Type:             cb.HeaderType(channelHeader.Type),
		Txid:             channelHeader.TxId,
		TxValidationCode: txValidationCode,
	}

	if cb.HeaderType(channelHeader.Type) == cb.HeaderType_ENDORSER_TRANSACTION {
		actions, err := getFilteredTransactionActions(payload.Data)
		if err != nil {
			return nil, "", errors.Wrap(err, "error getting filtered transaction actions")
		}
		filteredTx.Data = actions
	}
	return filteredTx, channelHeader.ChannelId, nil
}

func getFilteredTransactionActions(data []byte) (*pb.FilteredTransaction_TransactionActions, error) {
	actions := &pb.FilteredTransaction_TransactionActions{
		TransactionActions: &pb.FilteredTransactionActions{},
	}
	tx, err := protoutil.UnmarshalTransaction(data)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling transaction payload")
	}
	chaincodeActionPayload, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling chaincode action payload")
	}
	propRespPayload, err := protoutil.UnmarshalProposalResponsePayload(chaincodeActionPayload.Action.ProposalResponsePayload)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling response payload")
	}
	ccAction, err := protoutil.UnmarshalChaincodeAction(propRespPayload.Extension)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling chaincode action")
	}
	ccEvent, err := protoutil.UnmarshalChaincodeEvents(ccAction.Events)
	if err != nil {
		return nil, errors.Wrap(err, "error getting chaincode events")
	}
	if ccEvent != nil {
		actions.TransactionActions.ChaincodeActions = append(actions.TransactionActions.ChaincodeActions, &pb.FilteredChaincodeAction{ChaincodeEvent: ccEvent})
	}
	return actions, nil
}

func (ed *Dispatcher) getState() int32 {
	return atomic.LoadInt32(&ed.state)
}

func (ed *Dispatcher) setState(expectedState, newState int32) bool {
	return atomic.CompareAndSwapInt32(&ed.state, expectedState, newState)
}
