/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package events

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sync"

	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	consumer "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/consumer"
	cnsmr "github.com/hyperledger/fabric/events/consumer"

	"github.com/hyperledger/fabric/core/ledger/util"
	common "github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

// EventHub allows a client to listen to event at a peer.
type EventHub struct {
	// Protects chaincodeRegistrants, blockRegistrants and txRegistrants
	mtx sync.RWMutex
	// Map of clients registered for chaincode events
	chaincodeRegistrants map[string][]*fab.ChainCodeCBE
	// Map of clients registered for block events
	blockRegistrants []func(*common.Block)
	// Map of clients registered for transactional events
	txRegistrants map[string]func(string, pb.TxValidationCode, error)
	// peer addr to connect to
	peerAddr string
	// peer tls certificate
	peerTLSCertificate string
	// peer tls server host override
	peerTLSServerHostOverride string
	// grpc event client interface
	grpcClient fab.EventsClient
	// fabric connection state of this eventhub
	connected bool
	// List of events client is interested in
	interestedEvents []*pb.Interest
	// Factory that creates EventsClient
	eventsClientFactory eventClientFactory
	// FabricClient
	client fab.FabricClient
}

// eventClientFactory creates an EventsClient instance
type eventClientFactory interface {
	newEventsClient(client fab.FabricClient, peerAddress string, certificate string, serverHostOverride string, regTimeout time.Duration, adapter cnsmr.EventAdapter) (fab.EventsClient, error)
}

// consumerClientFactory is the default implementation oif the eventClientFactory
type consumerClientFactory struct{}

func (ccf *consumerClientFactory) newEventsClient(client fab.FabricClient, peerAddress string, certificate string, serverHostOverride string, regTimeout time.Duration, adapter cnsmr.EventAdapter) (fab.EventsClient, error) {
	return consumer.NewEventsClient(client, peerAddress, certificate, serverHostOverride, regTimeout, adapter)
}

// NewEventHub ...
func NewEventHub(client fab.FabricClient) (*EventHub, error) {

	if client == nil {
		return nil, fmt.Errorf("Client is nil")
	}
	chaincodeRegistrants := make(map[string][]*fab.ChainCodeCBE)
	txRegistrants := make(map[string]func(string, pb.TxValidationCode, error))

	eventHub := EventHub{
		chaincodeRegistrants: chaincodeRegistrants,
		blockRegistrants:     nil,
		txRegistrants:        txRegistrants,
		interestedEvents:     nil,
		eventsClientFactory:  &consumerClientFactory{},
		client:               client,
	}

	// register default transaction callback
	eventHub.RegisterBlockEvent(eventHub.txCallback)

	return &eventHub, nil
}

// SetInterests clears all interests and sets the interests for BLOCK type of events.
func (eventHub *EventHub) SetInterests(block bool) {
	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	eventHub.interestedEvents = make([]*pb.Interest, 0)
	eventHub.blockRegistrants = make([]func(*common.Block), 0)

	if block {
		eventHub.blockRegistrants = append(eventHub.blockRegistrants, eventHub.txCallback)
		eventHub.interestedEvents = append(eventHub.interestedEvents, &pb.Interest{EventType: pb.EventType_BLOCK})
	}
}

// Disconnect disconnects from peer event source
func (eventHub *EventHub) Disconnect() {
	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	if !eventHub.connected {
		return
	}

	// Unregister interests with server and stop the stream
	eventHub.grpcClient.Unregister(eventHub.interestedEvents)
	eventHub.grpcClient.Stop()

	eventHub.connected = false
}

// RegisterBlockEvent - register callback function for block events
func (eventHub *EventHub) RegisterBlockEvent(callback func(*common.Block)) {
	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	eventHub.blockRegistrants = append(eventHub.blockRegistrants, callback)

	// Register interest for blocks (only declare interest once, so do this for the first registrant)
	if len(eventHub.blockRegistrants) == 1 {
		eventHub.interestedEvents = append(eventHub.interestedEvents, &pb.Interest{EventType: pb.EventType_BLOCK})
	}
}

// UnregisterBlockEvent unregister callback for block event
func (eventHub *EventHub) UnregisterBlockEvent(callback func(*common.Block)) {
	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	f1 := reflect.ValueOf(callback)

	for i := range eventHub.blockRegistrants {
		f2 := reflect.ValueOf(eventHub.blockRegistrants[i])
		if f1.Pointer() == f2.Pointer() {
			eventHub.blockRegistrants = append(eventHub.blockRegistrants[:i], eventHub.blockRegistrants[i+1:]...)
			break
		}
	}

	// Unregister interest for blocks if there are no more registrants
	if len(eventHub.blockRegistrants) < 1 {
		blockEventInterest := pb.Interest{EventType: pb.EventType_BLOCK}
		eventHub.grpcClient.UnregisterAsync([]*pb.Interest{&blockEventInterest})
		for i, v := range eventHub.interestedEvents {
			if *v == blockEventInterest {
				eventHub.interestedEvents = append(eventHub.interestedEvents[:i], eventHub.interestedEvents[i+1:]...)
			}
		}
	}
}

// addChaincodeInterest adds interest for specific CHAINCODE events.
func (eventHub *EventHub) addChaincodeInterest(ChaincodeID string, EventName string) {
	ccInterest := &pb.Interest{
		EventType: pb.EventType_CHAINCODE,
		RegInfo: &pb.Interest_ChaincodeRegInfo{
			ChaincodeRegInfo: &pb.ChaincodeReg{
				ChaincodeId: ChaincodeID,
				EventName:   EventName,
			},
		},
	}

	eventHub.interestedEvents = append(eventHub.interestedEvents, ccInterest)

	if eventHub.IsConnected() {
		eventHub.grpcClient.RegisterAsync([]*pb.Interest{ccInterest})
	}

}

// removeChaincodeInterest remove interest for specific CHAINCODE event
func (eventHub *EventHub) removeChaincodeInterest(ChaincodeID string, EventName string) {
	ccInterest := &pb.Interest{
		EventType: pb.EventType_CHAINCODE,
		RegInfo: &pb.Interest_ChaincodeRegInfo{
			ChaincodeRegInfo: &pb.ChaincodeReg{
				ChaincodeId: ChaincodeID,
				EventName:   EventName,
			},
		},
	}

	for i, v := range eventHub.interestedEvents {
		if v.EventType == ccInterest.EventType && *(v.GetChaincodeRegInfo()) == *(ccInterest.GetChaincodeRegInfo()) {
			eventHub.interestedEvents = append(eventHub.interestedEvents[:i], eventHub.interestedEvents[i+1:]...)
		}
	}

	if eventHub.IsConnected() {
		eventHub.grpcClient.UnregisterAsync([]*pb.Interest{ccInterest})
	}

}

// SetPeerAddr set peer url for event source
// peeraddr peer url
// peerTLSCertificate peer tls certificate
// peerTLSServerHostOverride tls serverhostoverride
func (eventHub *EventHub) SetPeerAddr(peerURL string, peerTLSCertificate string, peerTLSServerHostOverride string) {
	eventHub.peerAddr = peerURL
	eventHub.peerTLSCertificate = peerTLSCertificate
	eventHub.peerTLSServerHostOverride = peerTLSServerHostOverride

}

// IsConnected gets connected state of eventhub
// Returns true if connected to event source, false otherwise
func (eventHub *EventHub) IsConnected() bool {
	return eventHub.connected
}

// Connect establishes connection with peer event source
func (eventHub *EventHub) Connect() error {

	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	if eventHub.connected {
		logger.Debugf("Nothing to do - EventHub already connected")
		return nil
	}

	if eventHub.peerAddr == "" {
		return fmt.Errorf("eventHub.peerAddr is empty")
	}

	if eventHub.interestedEvents == nil || len(eventHub.interestedEvents) == 0 {
		return fmt.Errorf("You must register for at least one event before connecting")
	}

	if eventHub.grpcClient == nil {
		eventsClient, _ := eventHub.eventsClientFactory.newEventsClient(eventHub.client,
			eventHub.peerAddr, eventHub.peerTLSCertificate, eventHub.peerTLSServerHostOverride,
			eventHub.client.Config().TimeoutOrDefault(apiconfig.EventReg), eventHub)
		eventHub.grpcClient = eventsClient
	}

	if err := eventHub.grpcClient.Start(); err != nil {
		eventHub.grpcClient.Stop()
		return fmt.Errorf("Error from eventsClient.Start (%s)", err.Error())
	}

	eventHub.connected = true

	return nil
}

//GetInterestedEvents implements consumer.EventAdapter interface for registering interested events
func (eventHub *EventHub) GetInterestedEvents() ([]*pb.Interest, error) {
	return eventHub.interestedEvents, nil
}

//Recv implements consumer.EventAdapter interface for receiving events
func (eventHub *EventHub) Recv(msg *pb.Event) (bool, error) {
	switch msg.Event.(type) {
	case *pb.Event_Block:
		blockEvent := msg.Event.(*pb.Event_Block)
		logger.Debugf("Recv blockEvent:%v\n", blockEvent)
		for _, v := range eventHub.getBlockRegistrants() {
			v(blockEvent.Block)
		}

		for _, tdata := range blockEvent.Block.Data.Data {
			if ccEvent, channelID, err := getChainCodeEvent(tdata); err != nil {
				logger.Warningf("getChainCodeEvent return error: %v\n", err)
			} else if ccEvent != nil {
				eventHub.notifyChaincodeRegistrants(channelID, ccEvent, true)
			}
		}
		return true, nil
	case *pb.Event_ChaincodeEvent:
		ccEvent := msg.Event.(*pb.Event_ChaincodeEvent)
		logger.Debugf("Recv ccEvent:%v\n", ccEvent)

		if ccEvent != nil {
			eventHub.notifyChaincodeRegistrants("", ccEvent.ChaincodeEvent, false)
		}
		return true, nil
	default:
		return true, nil
	}
}

// Disconnected implements consumer.EventAdapter interface for receiving events
func (eventHub *EventHub) Disconnected(err error) {
	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	if !eventHub.connected {
		return
	}

	eventHub.grpcClient.Stop()
	eventHub.connected = false
}

// RegisterChaincodeEvent registers a callback function to receive chaincode events.
// ccid: chaincode id
// eventname: The regex string used to filter events
// callback: Callback function for filter matches that takes a single parameter which is a json object representation
//           of type "message ChaincodeEvent"
// Returns ChainCodeCBE object that should be treated as an opaque
// handle used to unregister (see unregisterChaincodeEvent)
func (eventHub *EventHub) RegisterChaincodeEvent(ccid string, eventname string, callback func(*fab.ChaincodeEvent)) *fab.ChainCodeCBE {
	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	eventHub.addChaincodeInterest(ccid, eventname)

	cbe := fab.ChainCodeCBE{CCID: ccid, EventNameFilter: eventname, CallbackFunc: callback}
	cbeArray := eventHub.chaincodeRegistrants[ccid]
	if cbeArray == nil && len(cbeArray) <= 0 {
		cbeArray = make([]*fab.ChainCodeCBE, 0)
		cbeArray = append(cbeArray, &cbe)
		eventHub.chaincodeRegistrants[ccid] = cbeArray
	} else {
		cbeArray = append(cbeArray, &cbe)
		eventHub.chaincodeRegistrants[ccid] = cbeArray
	}
	return &cbe
}

// UnregisterChaincodeEvent unregisters chaincode event registration
// ChainCodeCBE: handle returned from call to registerChaincodeEvent.
func (eventHub *EventHub) UnregisterChaincodeEvent(cbe *fab.ChainCodeCBE) {
	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	eventHub.removeChaincodeInterest(cbe.CCID, cbe.EventNameFilter)

	cbeArray := eventHub.chaincodeRegistrants[cbe.CCID]
	if len(cbeArray) <= 0 {
		logger.Debugf("No event registration for ccid %s \n", cbe.CCID)
		return
	}

	for i, v := range cbeArray {
		if v == cbe {
			newCbeArray := append(cbeArray[:i], cbeArray[i+1:]...)
			if len(newCbeArray) <= 0 {
				delete(eventHub.chaincodeRegistrants, cbe.CCID)
			} else {
				eventHub.chaincodeRegistrants[cbe.CCID] = newCbeArray
			}
			break
		}
	}
}

// RegisterTxEvent registers a callback function to receive transactional events.
// txid: transaction id
// callback: Function that takes a single parameter which
// is a json object representation of type "message Transaction"
func (eventHub *EventHub) RegisterTxEvent(txnID apitxn.TransactionID, callback func(string, pb.TxValidationCode, error)) {
	logger.Debugf("reg txid %s\n", txnID.ID)

	eventHub.mtx.Lock()
	eventHub.txRegistrants[txnID.ID] = callback
	eventHub.mtx.Unlock()
}

// UnregisterTxEvent unregister transactional event registration.
// txid: transaction id
func (eventHub *EventHub) UnregisterTxEvent(txnID apitxn.TransactionID) {
	eventHub.mtx.Lock()
	delete(eventHub.txRegistrants, txnID.ID)
	eventHub.mtx.Unlock()
}

/**
 * private internal callback for processing tx events
 * @param {object} block json object representing block of tx
 * from the fabric
 */
func (eventHub *EventHub) txCallback(block *common.Block) {
	logger.Debugf("txCallback block=%v\n", block)

	txFilter := util.TxValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	for i, v := range block.Data.Data {

		if env, err := utils.GetEnvelopeFromBlock(v); err != nil {
			logger.Errorf("error extracting Envelope from block: %v\n", err)
			return
		} else if env != nil {
			// get the payload from the envelope
			payload, err := utils.GetPayload(env)
			if err != nil {
				logger.Errorf("error extracting Payload from envelope: %v\n", err)
				return
			}

			channelHeaderBytes := payload.Header.ChannelHeader
			channelHeader := &common.ChannelHeader{}
			err = proto.Unmarshal(channelHeaderBytes, channelHeader)
			if err != nil {
				logger.Errorf("error extracting ChannelHeader from payload: %v\n", err)
				return
			}

			callback := eventHub.getTXRegistrant(channelHeader.TxId)
			if callback != nil {
				if txFilter.IsInvalid(i) {
					callback(channelHeader.TxId, txFilter.Flag(i), fmt.Errorf("Received invalid transaction from channel %s", channelHeader.ChannelId))
				} else {
					callback(channelHeader.TxId, txFilter.Flag(i), nil)
				}
			} else {
				logger.Debugf("No callback registered for TxID: %s\n", channelHeader.TxId)
			}
		}
	}
}

func (eventHub *EventHub) getBlockRegistrants() []func(*common.Block) {
	eventHub.mtx.RLock()
	defer eventHub.mtx.RUnlock()

	// Return a clone of the array to avoid race conditions
	clone := make([]func(*common.Block), len(eventHub.blockRegistrants))
	for i, registrant := range eventHub.blockRegistrants {
		clone[i] = registrant
	}
	return clone
}

func (eventHub *EventHub) getChaincodeRegistrants(chaincodeID string) []*fab.ChainCodeCBE {
	eventHub.mtx.RLock()
	defer eventHub.mtx.RUnlock()

	registrants, ok := eventHub.chaincodeRegistrants[chaincodeID]
	if !ok {
		return nil
	}

	// Return a clone of the array to avoid race conditions
	clone := make([]*fab.ChainCodeCBE, len(registrants))
	for i, registrants := range registrants {
		clone[i] = registrants
	}
	return clone
}

func (eventHub *EventHub) getTXRegistrant(txID string) func(string, pb.TxValidationCode, error) {
	eventHub.mtx.RLock()
	defer eventHub.mtx.RUnlock()
	return eventHub.txRegistrants[txID]
}

// getChainCodeEvents parses block events for chaincode events associated with individual transactions
func getChainCodeEvent(tdata []byte) (event *pb.ChaincodeEvent, channelID string, err error) {

	if tdata == nil {
		return nil, "", errors.New("Cannot extract payload from nil transaction")
	}

	if env, err := utils.GetEnvelopeFromBlock(tdata); err != nil {
		return nil, "", fmt.Errorf("Error getting tx from block(%s)", err)
	} else if env != nil {
		// get the payload from the envelope
		payload, err := utils.GetPayload(env)
		if err != nil {
			return nil, "", fmt.Errorf("Could not extract payload from envelope, err %s", err)
		}

		channelHeaderBytes := payload.Header.ChannelHeader
		channelHeader := &common.ChannelHeader{}
		err = proto.Unmarshal(channelHeaderBytes, channelHeader)
		if err != nil {
			return nil, "", fmt.Errorf("Could not extract channel header from envelope, err %s", err)
		}

		channelID := channelHeader.ChannelId

		// Chaincode events apply to endorser transaction only
		if common.HeaderType(channelHeader.Type) == common.HeaderType_ENDORSER_TRANSACTION {
			tx, err := utils.GetTransaction(payload.Data)
			if err != nil {
				return nil, "", fmt.Errorf("Error unmarshalling transaction payload for block event: %s", err)
			}
			chaincodeActionPayload, err := utils.GetChaincodeActionPayload(tx.Actions[0].Payload)
			if err != nil {
				return nil, "", fmt.Errorf("Error unmarshalling transaction action payload for block event: %s", err)
			}
			propRespPayload, err := utils.GetProposalResponsePayload(chaincodeActionPayload.Action.ProposalResponsePayload)
			if err != nil {
				return nil, "", fmt.Errorf("Error unmarshalling proposal response payload for block event: %s", err)
			}
			caPayload, err := utils.GetChaincodeAction(propRespPayload.Extension)
			if err != nil {
				return nil, "", fmt.Errorf("Error unmarshalling chaincode action for block event: %s", err)
			}
			ccEvent, err := utils.GetChaincodeEvents(caPayload.Events)

			if ccEvent != nil {
				return ccEvent, channelID, nil
			}
		}
	}
	return nil, "", nil
}

// Utility function to fire callbacks for chaincode registrants
func (eventHub *EventHub) notifyChaincodeRegistrants(channelID string, ccEvent *pb.ChaincodeEvent, patternMatch bool) {

	cbeArray := eventHub.getChaincodeRegistrants(ccEvent.ChaincodeId)
	if len(cbeArray) <= 0 {
		logger.Debugf("No event registration for ccid %s \n", ccEvent.ChaincodeId)
	}

	for _, v := range cbeArray {
		match := v.EventNameFilter == ccEvent.EventName
		if !match && patternMatch {
			match, _ = regexp.MatchString(v.EventNameFilter, ccEvent.EventName)
		}
		if match {
			callback := v.CallbackFunc
			if callback != nil {
				callback(&fab.ChaincodeEvent{
					ChaincodeID: ccEvent.ChaincodeId,
					TxID:        ccEvent.TxId,
					EventName:   ccEvent.EventName,
					Payload:     ccEvent.Payload,
					ChannelID:   channelID,
				})
			}
		}
	}
}
