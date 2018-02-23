/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chclient

import (
	"time"

	"bytes"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
)

//EndorsementHandler for handling endorse transactions
type EndorsementHandler struct {
	next Handler
}

//Handle for endorsing transactions
func (e *EndorsementHandler) Handle(requestContext *RequestContext, clientContext *ClientContext) {

	if len(requestContext.Opts.ProposalProcessors) == 0 {
		requestContext.Error = status.New(status.ClientStatus, status.NoPeersFound.ToInt32(), "targets were not provided", nil)
		return
	}

	// Endorse Tx
	transactionProposalResponses, proposal, err := createAndSendTransactionProposal(clientContext.Transactor, &requestContext.Request, requestContext.Opts.ProposalProcessors)

	requestContext.Response.Proposal = proposal
	requestContext.Response.TransactionID = proposal.TxnID // TODO: still needed?

	if err != nil {
		requestContext.Error = err
		return
	}

	requestContext.Response.Responses = transactionProposalResponses
	if len(transactionProposalResponses) > 0 {
		requestContext.Response.Payload = transactionProposalResponses[0].ProposalResponse.GetResponse().Payload
	}

	//Delegate to next step if any
	if e.next != nil {
		e.next.Handle(requestContext, clientContext)
	}
}

//ProposalProcessorHandler for selecting proposal processors
type ProposalProcessorHandler struct {
	next Handler
}

//Handle selects proposal processors
func (h *ProposalProcessorHandler) Handle(requestContext *RequestContext, clientContext *ClientContext) {
	//Get proposal processor, if not supplied then use discovery service to get available peers as endorser
	//If selection service available then get endorser peers for this chaincode
	if len(requestContext.Opts.ProposalProcessors) == 0 {
		// Use discovery service to figure out proposal processors
		peers, err := clientContext.Discovery.GetPeers()
		if err != nil {
			requestContext.Error = errors.WithMessage(err, "GetPeers failed")
			return
		}
		endorsers := peers
		if clientContext.Selection != nil {
			endorsers, err = clientContext.Selection.GetEndorsersForChaincode(peers, requestContext.Request.ChaincodeID)
			if err != nil {
				requestContext.Error = errors.WithMessage(err, "Failed to get endorsing peers")
				return
			}
		}
		requestContext.Opts.ProposalProcessors = peer.PeersToTxnProcessors(endorsers)
	}

	//Delegate to next step if any
	if h.next != nil {
		h.next.Handle(requestContext, clientContext)
	}
}

//EndorsementValidationHandler for transaction proposal response filtering
type EndorsementValidationHandler struct {
	next Handler
}

//Handle for Filtering proposal response
func (f *EndorsementValidationHandler) Handle(requestContext *RequestContext, clientContext *ClientContext) {

	//Filter tx proposal responses
	err := f.validate(requestContext.Response.Responses)
	if err != nil {
		requestContext.Error = errors.WithMessage(err, "endorsement validation failed")
		return
	}

	//Delegate to next step if any
	if f.next != nil {
		f.next.Handle(requestContext, clientContext)
	}
}

func (f *EndorsementValidationHandler) validate(txProposalResponse []*fab.TransactionProposalResponse) error {
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
	next Handler
}

//Handle handles commit tx
func (c *CommitTxHandler) Handle(requestContext *RequestContext, clientContext *ClientContext) {

	//Connect to Event hub if not yet connected
	if clientContext.EventHub.IsConnected() == false {
		err := clientContext.EventHub.Connect()
		if err != nil {
			requestContext.Error = err
			return
		}
	}

	txnID := requestContext.Response.TransactionID

	//Register Tx event
	statusNotifier := txn.RegisterStatus(txnID, clientContext.EventHub)
	_, err := createAndSendTransaction(clientContext.Transactor, requestContext.Response.Proposal, requestContext.Response.Responses)
	if err != nil {
		requestContext.Error = errors.Wrap(err, "CreateAndSendTransaction failed")
		return
	}

	select {
	case result := <-statusNotifier:
		requestContext.Response.TxValidationCode = result.Code

		if result.Error != nil {
			requestContext.Error = result.Error
			return
		}
	case <-time.After(requestContext.Opts.Timeout):
		requestContext.Error = errors.New("Execute didn't receive block event")
		return
	}

	//Delegate to next step if any
	if c.next != nil {
		c.next.Handle(requestContext, clientContext)
	}
}

//NewQueryHandler returns query handler with EndorseTxHandler & EndorsementValidationHandler Chained
func NewQueryHandler(next ...Handler) Handler {
	return NewProposalProcessorHandler(
		NewEndorsementHandler(
			NewEndorsementValidationHandler(
				NewSignatureValidationHandler(next...),
			),
		),
	)
}

//NewExecuteHandler returns query handler with EndorseTxHandler, EndorsementValidationHandler & CommitTxHandler Chained
func NewExecuteHandler(next ...Handler) Handler {
	return NewProposalProcessorHandler(
		NewEndorsementHandler(
			NewEndorsementValidationHandler(
				NewSignatureValidationHandler(NewCommitHandler(next...)),
			),
		),
	)
}

//NewProposalProcessorHandler returns a handler that selects proposal processors
func NewProposalProcessorHandler(next ...Handler) *ProposalProcessorHandler {
	return &ProposalProcessorHandler{next: getNext(next)}
}

//NewEndorsementHandler returns a handler that endorses a transaction proposal
func NewEndorsementHandler(next ...Handler) *EndorsementHandler {
	return &EndorsementHandler{next: getNext(next)}
}

//NewEndorsementValidationHandler returns a handler that validates an endorsement
func NewEndorsementValidationHandler(next ...Handler) *EndorsementValidationHandler {
	return &EndorsementValidationHandler{next: getNext(next)}
}

//NewCommitHandler returns a handler that commits transaction propsal responses
func NewCommitHandler(next ...Handler) *CommitTxHandler {
	return &CommitTxHandler{next: getNext(next)}
}

func getNext(next []Handler) Handler {
	if len(next) > 0 {
		return next[0]
	}
	return nil
}

func createAndSendTransaction(sender fab.Sender, proposal *fab.TransactionProposal, resps []*fab.TransactionProposalResponse) (*fab.TransactionResponse, error) {

	txnRequest := fab.TransactionRequest{
		Proposal:          proposal,
		ProposalResponses: resps,
	}

	tx, err := sender.CreateTransaction(txnRequest)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateTransaction failed")
	}

	transactionResponse, err := sender.SendTransaction(tx)
	if err != nil {
		return nil, errors.WithMessage(err, "SendTransaction failed")

	}
	if transactionResponse.Err != nil {
		logger.Debugf("orderer %s failed (%s)", transactionResponse.Orderer, transactionResponse.Err.Error())
		return nil, errors.Wrap(transactionResponse.Err, "orderer failed")
	}

	return transactionResponse, nil
}

func createAndSendTransactionProposal(sender fab.ProposalSender, chrequest *Request, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, *fab.TransactionProposal, error) {
	request := fab.ChaincodeInvokeRequest{
		ChaincodeID:  chrequest.ChaincodeID,
		Fcn:          chrequest.Fcn,
		Args:         chrequest.Args,
		TransientMap: chrequest.TransientMap,
	}

	proposal, err := sender.CreateChaincodeInvokeProposal(request)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "creating transaction proposal failed")
	}

	transactionProposalResponses, err := sender.SendTransactionProposal(proposal, targets)
	return transactionProposalResponses, proposal, err
}
