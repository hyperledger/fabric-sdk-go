/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txnhandler

import (
	"time"

	"bytes"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/txnhandler"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/internal"
	"github.com/hyperledger/fabric-sdk-go/pkg/status"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

//EndorseTxHandler for handling endorse transactions
type EndorseTxHandler struct {
	next txnhandler.Handler
}

//Handle for endorsing transactions
func (e *EndorseTxHandler) Handle(requestContext *txnhandler.RequestContext, clientContext *txnhandler.ClientContext) {

	//Get proposal processor, if not supplied then use discovery service to get available peers as endorser
	//If selection service available then get endorser peers for this chaincode
	txProcessors := requestContext.Opts.ProposalProcessors
	if len(txProcessors) == 0 {
		// Use discovery service to figure out proposal processors
		peers, err := clientContext.Discovery.GetPeers()
		if err != nil {
			requestContext.Response = apitxn.Response{Payload: nil, Error: errors.WithMessage(err, "GetPeers failed")}
			return
		}
		endorsers := peers
		if clientContext.Selection != nil {
			endorsers, err = clientContext.Selection.GetEndorsersForChaincode(peers, requestContext.Request.ChaincodeID)
			if err != nil {
				requestContext.Response = apitxn.Response{Payload: nil, Error: errors.WithMessage(err, "Failed to get endorsing peers")}
				return
			}
		}
		txProcessors = peer.PeersToTxnProcessors(endorsers)
	}

	// Endorse Tx
	transactionProposalResponses, txnID, err := internal.CreateAndSendTransactionProposal(clientContext.Channel,
		requestContext.Request.ChaincodeID, requestContext.Request.Fcn, requestContext.Request.Args, txProcessors, requestContext.Request.TransientMap)

	if err != nil {
		requestContext.Response = apitxn.Response{Payload: nil, TransactionID: txnID, Error: err}
		return
	}

	requestContext.Response.Responses = transactionProposalResponses
	requestContext.Response.TransactionID = txnID

	//Delegate to next step if any
	if e.next != nil {
		e.next.Handle(requestContext, clientContext)
	} else {
		var response []byte
		if len(transactionProposalResponses) > 0 {
			response = transactionProposalResponses[0].ProposalResponse.GetResponse().Payload
		}
		requestContext.Response = apitxn.Response{Payload: response, TransactionID: txnID, Responses: transactionProposalResponses}
	}
}

//EndorsementValidationHandler for transaction proposal response filtering
type EndorsementValidationHandler struct {
	next txnhandler.Handler
}

//Handle for Filtering proposal response
func (f *EndorsementValidationHandler) Handle(requestContext *txnhandler.RequestContext, clientContext *txnhandler.ClientContext) {

	//Filter tx proposal responses
	err := f.validate(requestContext.Response.Responses)
	if err != nil {
		requestContext.Response = apitxn.Response{Payload: nil, TransactionID: requestContext.Response.TransactionID,
			Error: errors.WithMessage(err, "TxFilter failed")}
		return
	}

	var response []byte
	if len(requestContext.Response.Responses) > 0 {
		response = requestContext.Response.Responses[0].ProposalResponse.GetResponse().Payload
	}

	requestContext.Response.Payload = response

	//Delegate to next step if any
	if f.next != nil {
		f.next.Handle(requestContext, clientContext)
	} else {
		requestContext.Response = apitxn.Response{Payload: response, Error: nil}
	}
}

func (f *EndorsementValidationHandler) validate(txProposalResponse []*apitxn.TransactionProposalResponse) error {
	var a1 []byte
	for n, r := range txProposalResponse {
		if r.ProposalResponse.GetResponse().Status != int32(common.Status_SUCCESS) {
			return status.NewFromProposalResponse(r.ProposalResponse, r.Endorser)
		}
		if n == 0 {
			a1 = r.ProposalResponse.GetResponse().Payload
			continue
		}

		if bytes.Compare(a1, r.ProposalResponse.GetResponse().Payload) != 0 {
			return status.New(status.EndorserClientStatus, status.EndorsementMismatch.ToInt32(),
				"ProposalResponsePayloads do not match", nil)
		}
	}

	return nil
}

//CommitTxHandler for committing transactions
type CommitTxHandler struct {
	next txnhandler.Handler
}

//Handle handles commit tx
func (c *CommitTxHandler) Handle(requestContext *txnhandler.RequestContext, clientContext *txnhandler.ClientContext) {

	//Connect to Event hub if not yet connected
	if clientContext.EventHub.IsConnected() == false {
		err := clientContext.EventHub.Connect()
		if err != nil {
			requestContext.Response = apitxn.Response{TransactionID: apitxn.TransactionID{}, Error: err}
			return
		}
	}

	txnID := requestContext.Response.TransactionID

	//Register Tx event
	statusNotifier := internal.RegisterTxEvent(txnID, clientContext.EventHub)
	_, err := internal.CreateAndSendTransaction(clientContext.Channel, requestContext.Response.Responses)
	if err != nil {
		requestContext.Response = apitxn.Response{TransactionID: apitxn.TransactionID{}, Error: errors.Wrap(err, "CreateAndSendTransaction failed")}
		return
	}

	select {
	case result := <-statusNotifier:
		if result.Error == nil {
			requestContext.Response = apitxn.Response{Payload: requestContext.Response.Payload, TransactionID: txnID, TxValidationCode: result.Code}
		} else {
			requestContext.Response = apitxn.Response{Payload: requestContext.Response.Payload, TransactionID: txnID, TxValidationCode: result.Code, Error: result.Error}
			return
		}
	case <-time.After(requestContext.Opts.Timeout):
		requestContext.Response = apitxn.Response{TransactionID: txnID, Error: errors.New("Execute didn't receive block event")}
		return
	}

	//Delegate to next step if any
	if c.next != nil {
		c.next.Handle(requestContext, clientContext)
	}
}

//NewQueryHandler returns query handler with EndorseTxHandler & EndorsementValidationHandler Chained
func NewQueryHandler() txnhandler.Handler {
	return &EndorseTxHandler{&EndorsementValidationHandler{}}
}

//NewExecuteHandler returns query handler with EndorseTxHandler, EndorsementValidationHandler & CommitTxHandler Chained
func NewExecuteHandler() txnhandler.Handler {
	return &EndorseTxHandler{&EndorsementValidationHandler{&CommitTxHandler{}}}
}
