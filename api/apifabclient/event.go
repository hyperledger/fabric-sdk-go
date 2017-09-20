/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	common "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	ehpb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// EventHub ...
type EventHub interface {
	SetPeerAddr(peerURL string, certificate string, serverHostOverride string)
	IsConnected() bool
	Connect() error
	Disconnect() error
	RegisterChaincodeEvent(ccid string, eventname string, callback func(*ChaincodeEvent)) *ChainCodeCBE
	UnregisterChaincodeEvent(cbe *ChainCodeCBE)
	RegisterTxEvent(txnID apitxn.TransactionID, callback func(string, pb.TxValidationCode, error))
	UnregisterTxEvent(txnID apitxn.TransactionID)
	RegisterBlockEvent(callback func(*common.Block))
	UnregisterBlockEvent(callback func(*common.Block))
}

//EventsClient holds the stream and adapter for consumer to work with
type EventsClient interface {
	RegisterAsync(ies []*ehpb.Interest) error
	UnregisterAsync(ies []*ehpb.Interest) error
	Unregister(ies []*ehpb.Interest) error
	Recv() (*ehpb.Event, error)
	Start() error
	Stop() error
}

// The EventHubExt interface allows extensions of the SDK to add functionality to EventHub overloads.
type EventHubExt interface {
	SetInterests(block bool)
}

// ChainCodeCBE ...
/**
 * The ChainCodeCBE is used internal to the EventHub to hold chaincode
 * event registration callbacks.
 */
type ChainCodeCBE struct {
	// chaincode id
	CCID string
	// event name regex filter
	EventNameFilter string
	// callback function to invoke on successful filter match
	CallbackFunc func(*ChaincodeEvent)
}

// ChaincodeEvent contains the current event data for the event handler
type ChaincodeEvent struct {
	ChaincodeID string
	TxID        string
	EventName   string
	Payload     []byte
	ChannelID   string
}
