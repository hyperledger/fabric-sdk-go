/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package chclient enables channel client
package chclient

import (
	"bytes"
	"reflect"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/internal"
	"github.com/hyperledger/fabric-sdk-go/pkg/status"
)

// ChannelClient enables access to a Fabric network.
type ChannelClient struct {
	client    fab.Resource
	channel   fab.Channel
	discovery fab.DiscoveryService
	selection fab.SelectionService
	eventHub  fab.EventHub
}

// txProposalResponseFilter process transaction proposal response
type txProposalResponseFilter struct {
}

// ProcessTxProposalResponse process transaction proposal response
func (txProposalResponseFilter *txProposalResponseFilter) ProcessTxProposalResponse(txProposalResponse []*apitxn.TransactionProposalResponse) ([]*apitxn.TransactionProposalResponse, error) {
	var a1 []byte
	for n, r := range txProposalResponse {
		if r.ProposalResponse.GetResponse().Status != int32(common.Status_SUCCESS) {
			return nil, status.NewFromProposalResponse(r.ProposalResponse, r.Endorser)
		}
		if n == 0 {
			a1 = r.ProposalResponse.GetResponse().Payload
			continue
		}

		if bytes.Compare(a1, r.ProposalResponse.GetResponse().Payload) != 0 {
			return nil, status.New(status.EndorserClientStatus, status.EndorsementMismatch.ToInt32(),
				"ProposalResponsePayloads do not match", nil)
		}
	}

	return txProposalResponse, nil
}

// NewChannelClient returns a ChannelClient instance.
func NewChannelClient(client fab.Resource, channel fab.Channel, discovery fab.DiscoveryService, selection fab.SelectionService, eventHub fab.EventHub) (*ChannelClient, error) {

	channelClient := ChannelClient{client: client, channel: channel, discovery: discovery, selection: selection, eventHub: eventHub}

	return &channelClient, nil
}

// Query chaincode
func (cc *ChannelClient) Query(request apitxn.QueryRequest) ([]byte, error) {

	return cc.QueryWithOpts(request, apitxn.QueryOpts{})

}

// QueryWithOpts allows the user to provide options for query (sync vs async, etc.)
func (cc *ChannelClient) QueryWithOpts(request apitxn.QueryRequest, opts apitxn.QueryOpts) ([]byte, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, errors.New("ChaincodeID and Fcn are required")
	}

	notifier := opts.Notifier
	if notifier == nil {
		notifier = make(chan apitxn.QueryResponse)
	}

	txProcessors := opts.ProposalProcessors
	if len(txProcessors) == 0 {
		// Use discovery service to figure out proposal processors
		peers, err := cc.discovery.GetPeers()
		if err != nil {
			return nil, errors.WithMessage(err, "GetPeers failed")
		}
		endorsers := peers
		if cc.selection != nil {
			endorsers, err = cc.selection.GetEndorsersForChaincode(peers, request.ChaincodeID)
			if err != nil {
				return nil, errors.WithMessage(err, "Failed to get endorsing peers")
			}
		}
		txProcessors = peer.PeersToTxnProcessors(endorsers)
	}

	go sendTransactionProposal(request, cc.channel, txProcessors, opts.TxFilter, notifier)

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
		return nil, errors.New("query request timed out")
	}

}

func sendTransactionProposal(request apitxn.QueryRequest, channel fab.Channel, proposalProcessors []apitxn.ProposalProcessor, txFilter apitxn.TxProposalResponseFilter, notifier chan apitxn.QueryResponse) {

	transactionProposalResponses, _, err := internal.CreateAndSendTransactionProposal(channel,
		request.ChaincodeID, request.Fcn, request.Args, proposalProcessors, nil)

	if err != nil {
		notifier <- apitxn.QueryResponse{Response: nil, Error: err}
		return
	}

	if txFilter == nil {
		txFilter = &txProposalResponseFilter{}
	}

	transactionProposalResponses, err = txFilter.ProcessTxProposalResponse(transactionProposalResponses)
	if err != nil {
		notifier <- apitxn.QueryResponse{Response: nil, Error: errors.WithMessage(err, "TxFilter failed")}
		return
	}

	response := transactionProposalResponses[0].ProposalResponse.GetResponse().Payload

	notifier <- apitxn.QueryResponse{Response: response, Error: nil}
}

// ExecuteTx prepares and executes transaction
func (cc *ChannelClient) ExecuteTx(request apitxn.ExecuteTxRequest) ([]byte, apitxn.TransactionID, error) {

	return cc.ExecuteTxWithOpts(request, apitxn.ExecuteTxOpts{})
}

// ExecuteTxWithOpts allows the user to provide options for execute transaction:
// sync vs async, filter to inspect proposal response before commit etc)
func (cc *ChannelClient) ExecuteTxWithOpts(request apitxn.ExecuteTxRequest, opts apitxn.ExecuteTxOpts) ([]byte, apitxn.TransactionID, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, apitxn.TransactionID{}, errors.New("chaincode name and function name are required")
	}

	txProcessors := opts.ProposalProcessors
	if len(txProcessors) == 0 {
		// Use discovery service to figure out proposal processors
		peers, err := cc.discovery.GetPeers()
		if err != nil {
			return nil, apitxn.TransactionID{}, errors.WithMessage(err, "GetPeers failed")
		}
		endorsers := peers
		if cc.selection != nil {
			endorsers, err = cc.selection.GetEndorsersForChaincode(peers, request.ChaincodeID)
			if err != nil {
				return nil, apitxn.TransactionID{}, errors.WithMessage(err, "Failed to get endorsing peers for ExecuteTx")
			}
		}
		txProcessors = peer.PeersToTxnProcessors(endorsers)
	}

	txProposalResponses, txID, err := internal.CreateAndSendTransactionProposal(cc.channel,
		request.ChaincodeID, request.Fcn, request.Args, txProcessors, request.TransientMap)
	if err != nil {
		return nil, apitxn.TransactionID{}, errors.WithMessage(err, "CreateAndSendTransactionProposal failed")
	}

	if opts.TxFilter == nil {
		opts.TxFilter = &txProposalResponseFilter{}
	}

	var payloadResponse []byte
	txProposalResponses, err = opts.TxFilter.ProcessTxProposalResponse(txProposalResponses)
	if len(txProposalResponses) > 0 {
		payloadResponse = txProposalResponses[0].ProposalResponse.GetResponse().Payload
	}

	if err != nil {
		return payloadResponse, txID, errors.WithMessage(err, "TxFilter failed")
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
		return payloadResponse, txID, nil
	}

	select {
	case response := <-notifier:
		return payloadResponse, response.Response, response.Error
	case <-time.After(timeout): // This should never happen since there's timeout in sendTransaction
		return payloadResponse, txID, errors.New("ExecuteTx request timed out")
	}

}

func sendTransaction(channel fab.Channel, txID apitxn.TransactionID, txProposalResponses []*apitxn.TransactionProposalResponse, eventHub fab.EventHub, notifier chan apitxn.ExecuteTxResponse, timeout time.Duration) {
	if eventHub.IsConnected() == false {
		err := eventHub.Connect()
		if err != nil {
			notifier <- apitxn.ExecuteTxResponse{Response: apitxn.TransactionID{}, Error: err}
		}
	}

	statusNotifier := internal.RegisterTxEvent(txID, eventHub)
	_, err := internal.CreateAndSendTransaction(channel, txProposalResponses)
	if err != nil {
		notifier <- apitxn.ExecuteTxResponse{Response: apitxn.TransactionID{}, Error: errors.Wrap(err, "CreateAndSendTransaction failed")}
		return
	}

	select {
	case result := <-statusNotifier:
		if result.Error == nil {
			notifier <- apitxn.ExecuteTxResponse{Response: txID, TxValidationCode: result.Code}
		} else {
			notifier <- apitxn.ExecuteTxResponse{Response: txID, TxValidationCode: result.Code, Error: result.Error}
		}
	case <-time.After(timeout):
		notifier <- apitxn.ExecuteTxResponse{Response: txID, Error: errors.New("ExecuteTx didn't receive block event")}
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
		return errors.Errorf("Unsupported registration type: %v", reflect.TypeOf(registration))
	}

	return nil

}
