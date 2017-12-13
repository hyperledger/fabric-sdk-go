/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package events

import (
	"reflect"
	"regexp"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	syncmap "golang.org/x/sync/syncmap"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	common "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/ledger/util"
	cnsmr "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/events/consumer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	consumer "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/consumer"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// EventHub allows a client to listen to event at a peer.
type EventHub struct {
	//Used for protecting parts of code from running concurrently
	mtx sync.RWMutex
	// Map of clients registered for chaincode events
	chaincodeRegistrants syncmap.Map
	// Array of clients registered for block events
	blockRegistrants []func(*common.Block)
	// Map of clients registered for transactional events
	txRegistrants syncmap.Map
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
		return nil, errors.New("Client is required")
	}
	eventHub := EventHub{
		blockRegistrants:    nil,
		interestedEvents:    nil,
		eventsClientFactory: &consumerClientFactory{},
		client:              client,
	}
	// register default transaction callback
	eventHub.RegisterBlockEvent(eventHub.txCallback)
	return &eventHub, nil
}

// NewEventHubFromConfig creates new event hub from client and peer config
func NewEventHubFromConfig(client fab.FabricClient, peerCfg *apiconfig.PeerConfig) (*EventHub, error) {

	eventHub, err := NewEventHub(client)
	if err != nil {
		return nil, err
	}

	serverHostOverride := ""
	if str, ok := peerCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}

	eventHub.peerAddr = peerCfg.EventURL
	eventHub.peerTLSCertificate = peerCfg.TLSCACerts.Path
	eventHub.peerTLSServerHostOverride = serverHostOverride

	return eventHub, nil
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
func (eventHub *EventHub) Disconnect() error {
	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	if !eventHub.connected {
		return nil
	}

	// Unregister interests with server and stop the stream
	err := eventHub.grpcClient.UnregisterAsync(eventHub.interestedEvents)
	if err != nil {
		return errors.WithMessage(err, "event client UnregisterAsync failed")
	}
	err = eventHub.grpcClient.Stop()
	if err != nil {
		return errors.WithMessage(err, "event client stop failed")
	}

	eventHub.connected = false
	return nil
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
		return errors.New("peerAddr is required")
	}

	if eventHub.interestedEvents == nil || len(eventHub.interestedEvents) == 0 {
		return errors.New("at least one event must be registered")
	}

	if eventHub.grpcClient == nil {
		eventsClient, _ := eventHub.eventsClientFactory.newEventsClient(eventHub.client,
			eventHub.peerAddr, eventHub.peerTLSCertificate, eventHub.peerTLSServerHostOverride,
			eventHub.client.Config().TimeoutOrDefault(apiconfig.EventReg), eventHub)
		eventHub.grpcClient = eventsClient
	}

	if err := eventHub.grpcClient.Start(); err != nil {
		eventHub.grpcClient.Stop()
		return errors.WithMessage(err, "event client start failed")
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
	// Deliver events asynchronously so that we can continue receiving events
	go func() {
		switch msg.Event.(type) {
		case *pb.Event_Block:
			blockEvent := msg.Event.(*pb.Event_Block)
			logger.Debugf("Recv blockEvent for block number [%d]", blockEvent.Block.Header.Number)
			for _, v := range eventHub.getBlockRegistrants() {
				v(blockEvent.Block)
			}
			txFilter := util.TxValidationFlags(blockEvent.Block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
			for i, tdata := range blockEvent.Block.Data.Data {
				if txFilter.IsValid(i) {
					if ccEvent, channelID, err := getChainCodeEvent(tdata); err != nil {
						logger.Warnf("getChainCodeEvent return error: %v\n", err)
					} else if ccEvent != nil {
						eventHub.notifyChaincodeRegistrants(channelID, ccEvent, true)
					}
				} else {
					logger.Debugf("received invalid transaction")
				}
			}
			return
		case *pb.Event_ChaincodeEvent:
			ccEvent := msg.Event.(*pb.Event_ChaincodeEvent)
			logger.Debugf("Recv ccEvent for txID [%s]", ccEvent.ChaincodeEvent.TxId)
			if ccEvent != nil {
				eventHub.notifyChaincodeRegistrants("", ccEvent.ChaincodeEvent, false)
			}
			return
		default:
			return
		}
	}()

	return true, nil
}

// Disconnected implements consumer.EventAdapter interface for receiving events
func (eventHub *EventHub) Disconnected(err error) {
	if err != nil {
		logger.Warnf("EventHub was disconnected unexpectedly: %s", err)
	}
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
	var cbeArray []*fab.ChainCodeCBE

	ccRegistrantArray, ok := eventHub.chaincodeRegistrants.Load(ccid)
	if !ok {
		cbeArray = make([]*fab.ChainCodeCBE, 0)
	} else {
		cbeArray = ccRegistrantArray.([]*fab.ChainCodeCBE)
	}
	cbeArray = append(cbeArray, &cbe)
	eventHub.chaincodeRegistrants.Store(ccid, cbeArray)

	return &cbe
}

// UnregisterChaincodeEvent unregisters chaincode event registration
// ChainCodeCBE: handle returned from call to registerChaincodeEvent.
func (eventHub *EventHub) UnregisterChaincodeEvent(cbe *fab.ChainCodeCBE) {
	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	eventHub.removeChaincodeInterest(cbe.CCID, cbe.EventNameFilter)

	ccRegistrantArray, ok := eventHub.chaincodeRegistrants.Load(cbe.CCID)
	if ok {
		cbeArray := ccRegistrantArray.([]*fab.ChainCodeCBE)
		if len(cbeArray) <= 0 {
			logger.Debugf("No event registration for ccid %s \n", cbe.CCID)
			return
		}

		for i, v := range cbeArray {
			if v == cbe {
				newCbeArray := append(cbeArray[:i], cbeArray[i+1:]...)
				if len(newCbeArray) <= 0 {
					eventHub.chaincodeRegistrants.Delete(cbe.CCID)
				} else {
					eventHub.chaincodeRegistrants.Store(cbe.CCID, newCbeArray)
				}
				break
			}
		}
	}
}

// RegisterTxEvent registers a callback function to receive transactional events.
// txid: transaction id
// callback: Function that takes a single parameter which
// is a json object representation of type "message Transaction"
func (eventHub *EventHub) RegisterTxEvent(txnID apitxn.TransactionID, callback func(string, pb.TxValidationCode, error)) {
	logger.Debugf("reg txid %s\n", txnID.ID)
	eventHub.txRegistrants.Store(txnID.ID, callback)
}

// UnregisterTxEvent unregister transactional event registration.
// txid: transaction id
func (eventHub *EventHub) UnregisterTxEvent(txnID apitxn.TransactionID) {
	logger.Debugf("un-reg txid %s\n", txnID.ID)
	eventHub.txRegistrants.Delete(txnID.ID)
}

/**
 * private internal callback for processing tx events
 * @param {object} block json object representing block of tx
 * from the fabric
 */
func (eventHub *EventHub) txCallback(block *common.Block) {
	txFilter := util.TxValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	for i, v := range block.Data.Data {

		if env, err := utils.GetEnvelopeFromBlock(v); err != nil {
			logger.Debugf("error extracting Envelope from block: %v\n", err)
			return
		} else if env != nil {
			// get the payload from the envelope
			payload, err := utils.GetPayload(env)
			if err != nil {
				logger.Debugf("error extracting Payload from envelope: %v\n", err)
				return
			}

			channelHeaderBytes := payload.Header.ChannelHeader
			channelHeader := &common.ChannelHeader{}
			err = proto.Unmarshal(channelHeaderBytes, channelHeader)
			if err != nil {
				logger.Debugf("error extracting ChannelHeader from payload: %v\n", err)
				return
			}
			callback := eventHub.getTXRegistrant(channelHeader.TxId)
			if callback != nil {
				if txFilter.IsInvalid(i) {
					callback(channelHeader.TxId, txFilter.Flag(i), errors.New("received invalid transaction"))
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
	copy(clone, eventHub.blockRegistrants)
	return clone
}

func (eventHub *EventHub) getChaincodeRegistrants(chaincodeID string) []*fab.ChainCodeCBE {
	eventHub.mtx.RLock()
	defer eventHub.mtx.RUnlock()

	registrants, ok := eventHub.chaincodeRegistrants.Load(chaincodeID)
	if !ok {
		return nil
	}
	cbeRegistrants := registrants.([]*fab.ChainCodeCBE)
	// Return a clone of the array to avoid race conditions
	clone := make([]*fab.ChainCodeCBE, len(cbeRegistrants))
	copy(clone, cbeRegistrants)
	return clone
}

func (eventHub *EventHub) getTXRegistrant(txID string) func(string, pb.TxValidationCode, error) {
	v, ok := eventHub.txRegistrants.Load(txID)
	if !ok {
		return nil
	}
	return v.(func(string, pb.TxValidationCode, error))
}

// getChainCodeEvents parses block events for chaincode events associated with individual transactions
func getChainCodeEvent(tdata []byte) (event *pb.ChaincodeEvent, channelID string, err error) {

	if tdata == nil {
		return nil, "", errors.New("Cannot extract payload from nil transaction")
	}

	if env, err := utils.GetEnvelopeFromBlock(tdata); err != nil {
		return nil, "", errors.Wrap(err, "tx from block failed")
	} else if env != nil {
		// get the payload from the envelope
		payload, err := utils.GetPayload(env)
		if err != nil {
			return nil, "", errors.Wrap(err, "extract payload from envelope failed")
		}

		channelHeaderBytes := payload.Header.ChannelHeader
		channelHeader := &common.ChannelHeader{}
		err = proto.Unmarshal(channelHeaderBytes, channelHeader)
		if err != nil {
			return nil, "", errors.Wrap(err, "unmarshal channel header failed")
		}

		channelID := channelHeader.ChannelId

		// Chaincode events apply to endorser transaction only
		if common.HeaderType(channelHeader.Type) == common.HeaderType_ENDORSER_TRANSACTION {
			tx, err := utils.GetTransaction(payload.Data)
			if err != nil {
				return nil, "", errors.Wrap(err, "unmarshal transaction payload")
			}
			chaincodeActionPayload, err := utils.GetChaincodeActionPayload(tx.Actions[0].Payload)
			if err != nil {
				return nil, "", errors.Wrap(err, "chaincode action payload retrieval failed")
			}
			propRespPayload, err := utils.GetProposalResponsePayload(chaincodeActionPayload.Action.ProposalResponsePayload)
			if err != nil {
				return nil, "", errors.Wrap(err, "proposal response payload retrieval failed")
			}
			caPayload, err := utils.GetChaincodeAction(propRespPayload.Extension)
			if err != nil {
				return nil, "", errors.Wrap(err, "chaincode action retrieval failed")
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
