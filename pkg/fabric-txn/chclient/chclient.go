/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package chclient enables channel client
package chclient

import (
	"fmt"
	"reflect"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/internal"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// ChannelClient enables access to a Fabric network.
type ChannelClient struct {
	client    fab.FabricClient
	channel   fab.Channel
	discovery fab.DiscoveryService
	eventHub  fab.EventHub
}

// NewChannelClient returns a ChannelClient instance.
func NewChannelClient(client fab.FabricClient, channel fab.Channel, discovery fab.DiscoveryService, eventHub fab.EventHub) (*ChannelClient, error) {

	channelClient := ChannelClient{client: client, channel: channel, discovery: discovery, eventHub: eventHub}

	return &channelClient, nil
}

// Query chaincode
func (cc *ChannelClient) Query(request apitxn.QueryRequest) ([]byte, error) {

	return cc.QueryWithOpts(request, apitxn.QueryOpts{})

}

// QueryWithOpts allows the user to provide options for query (sync vs async, etc.)
func (cc *ChannelClient) QueryWithOpts(request apitxn.QueryRequest, opts apitxn.QueryOpts) ([]byte, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, fmt.Errorf("Chaincode name and function name must be provided")
	}

	notifier := opts.Notifier
	if notifier == nil {
		notifier = make(chan apitxn.QueryResponse)
	}

	peers, err := cc.discovery.GetPeers(request.ChaincodeID)
	if err != nil {
		return nil, fmt.Errorf("Unable to get peers: %v", err)
	}

	txProcessors := peer.PeersToTxnProcessors(peers)

	go sendTransactionProposal(request, cc.channel, txProcessors, notifier)

	if opts.Notifier != nil {
		return nil, nil
	}

	timeout := cc.client.Config().TimeoutOrDefault(apiconfig.Query)
	if opts.Timeout != 0 {
		timeout = opts.Timeout
	}

	select {
	case response := <-notifier:
		return response.Response, response.Error
	case <-time.After(timeout):
		return nil, fmt.Errorf("Query request timed out")
	}

}

func sendTransactionProposal(request apitxn.QueryRequest, channel fab.Channel, proposalProcessors []apitxn.ProposalProcessor, notifier chan apitxn.QueryResponse) {

	// TODO: Temporary conversion until proposal sender is changed to handle [][]byte arguments
	ccArgs := toStringArray(request.Args)
	transactionProposalResponses, _, err := internal.CreateAndSendTransactionProposal(channel,
		request.ChaincodeID, request.Fcn, ccArgs, proposalProcessors, nil)

	if err != nil {
		notifier <- apitxn.QueryResponse{Response: nil, Error: err}
		return
	}

	response := transactionProposalResponses[0].ProposalResponse.GetResponse().Payload

	notifier <- apitxn.QueryResponse{Response: response, Error: nil}
}

// ExecuteTx prepares and executes transaction
func (cc *ChannelClient) ExecuteTx(request apitxn.ExecuteTxRequest) (apitxn.TransactionID, error) {

	return cc.ExecuteTxWithOpts(request, apitxn.ExecuteTxOpts{})
}

// ExecuteTxWithOpts allows the user to provide options for execute transaction:
// sync vs async, filter to inspect proposal response before commit etc)
func (cc *ChannelClient) ExecuteTxWithOpts(request apitxn.ExecuteTxRequest, opts apitxn.ExecuteTxOpts) (apitxn.TransactionID, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return apitxn.TransactionID{}, fmt.Errorf("Chaincode name and function name must be provided")
	}

	peers, err := cc.discovery.GetPeers(request.ChaincodeID)
	if err != nil {
		return apitxn.TransactionID{}, fmt.Errorf("Unable to get peers: %v", err)
	}

	// TODO: Temporary conversion until proposal sender is changed to handle [][]byte arguments
	ccArgs := toStringArray(request.Args)
	txProposalResponses, txID, err := internal.CreateAndSendTransactionProposal(cc.channel,
		request.ChaincodeID, request.Fcn, ccArgs, peer.PeersToTxnProcessors(peers), request.TransientMap)
	if err != nil {
		return apitxn.TransactionID{}, fmt.Errorf("CreateAndSendTransactionProposal returned error: %v", err)
	}

	if opts.TxFilter != nil {
		txProposalResponses, err = opts.TxFilter.ProcessTxProposalResponse(txProposalResponses)
		if err != nil {
			return txID, fmt.Errorf("TxFilter returned error: %v", err)
		}
	}

	notifier := opts.Notifier
	if notifier == nil {
		notifier = make(chan apitxn.ExecuteTxResponse)
	}

	timeout := cc.client.Config().TimeoutOrDefault(apiconfig.ExecuteTx)
	if opts.Timeout != 0 {
		timeout = opts.Timeout
	}

	go sendTransaction(cc.channel, txID, txProposalResponses, cc.eventHub, notifier, timeout)

	if opts.Notifier != nil {
		return txID, nil
	}

	select {
	case response := <-notifier:
		return response.Response, response.Error
	case <-time.After(timeout): // This should never happen since there's timeout in sendTransaction
		return txID, fmt.Errorf("ExecuteTx request timed out")
	}

}

func sendTransaction(channel fab.Channel, txID apitxn.TransactionID, txProposalResponses []*apitxn.TransactionProposalResponse, eventHub fab.EventHub, notifier chan apitxn.ExecuteTxResponse, timeout time.Duration) {

	if eventHub.IsConnected() == false {
		err := eventHub.Connect()
		if err != nil {
			notifier <- apitxn.ExecuteTxResponse{Response: apitxn.TransactionID{}, Error: err}
		}
	}

	chcode := internal.RegisterTxEvent(txID, eventHub)
	_, err := internal.CreateAndSendTransaction(channel, txProposalResponses)
	if err != nil {
		notifier <- apitxn.ExecuteTxResponse{Response: apitxn.TransactionID{}, Error: fmt.Errorf("CreateAndSendTransaction returned error: %v", err)}
		return
	}

	select {
	case code := <-chcode:
		if code == pb.TxValidationCode_VALID {
			notifier <- apitxn.ExecuteTxResponse{Response: txID, TxValidationCode: code}
		} else {
			notifier <- apitxn.ExecuteTxResponse{Response: txID, TxValidationCode: code, Error: fmt.Errorf("ExecuteTx received a failed transaction response from eventhub for txid(%s), code(%s)", txID, code)}
		}
	case <-time.After(timeout):
		notifier <- apitxn.ExecuteTxResponse{Response: txID, Error: fmt.Errorf("ExecuteTx didn't receive block event for txid(%s)", txID)}
	}
}

// Close releases channel client resources (disconnects event hub etc.)
func (cc *ChannelClient) Close() error {
	if cc.eventHub.IsConnected() == true {
		return cc.eventHub.Disconnect()
	}

	return nil
}

// RegisterChaincodeEvent registers chain code event
// @param {chan bool} channel which receives event details when the event is complete
// @returns {object} object handle that should be used to unregister
func (cc *ChannelClient) RegisterChaincodeEvent(notify chan<- *apitxn.CCEvent, chainCodeID string, eventID string) apitxn.Registration {

	// Register callback for CE
	rce := cc.eventHub.RegisterChaincodeEvent(chainCodeID, eventID, func(ce *fab.ChaincodeEvent) {
		notify <- &apitxn.CCEvent{ChaincodeID: ce.ChaincodeID, EventName: ce.EventName, TxID: ce.TxID, Payload: ce.Payload}
	})

	return rce
}

// UnregisterChaincodeEvent removes chain code event registration
func (cc *ChannelClient) UnregisterChaincodeEvent(registration apitxn.Registration) error {

	switch regType := registration.(type) {

	case *fab.ChainCodeCBE:
		cc.eventHub.UnregisterChaincodeEvent(regType)
	default:
		return fmt.Errorf("Unsupported registration type: %v", reflect.TypeOf(registration))
	}

	return nil

}

func toStringArray(byteArray [][]byte) []string {
	strArray := make([]string, len(byteArray))
	for i := 0; i < len(byteArray); i++ {
		strArray[i] = string(byteArray[i])
	}
	return strArray
}
