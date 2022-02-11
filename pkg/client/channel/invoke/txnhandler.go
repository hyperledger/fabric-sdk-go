/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package invoke

import (
	"bytes"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	selectopts "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
)

// TxnHeaderOptsProvider provides transaction header options which allow
// the provider to specify a custom creator and/or nonce.
type TxnHeaderOptsProvider func() []fab.TxnHeaderOpt

//EndorsementHandler for handling endorse transactions
type EndorsementHandler struct {
	next               Handler
	headerOptsProvider TxnHeaderOptsProvider
}

//Handle for endorsing transactions
func (e *EndorsementHandler) Handle(requestContext *RequestContext, clientContext *ClientContext) {

	if len(requestContext.Opts.Targets) == 0 {
		requestContext.Error = status.New(status.ClientStatus, status.NoPeersFound.ToInt32(), "targets were not provided", nil)
		return
	}

	// Endorse Tx
	var TxnHeaderOpts []fab.TxnHeaderOpt
	if e.headerOptsProvider != nil {
		TxnHeaderOpts = e.headerOptsProvider()
	}

	transactionProposalResponses, proposal, err := createAndSendTransactionProposal(
		clientContext.Transactor,
		&requestContext.Request,
		peer.PeersToTxnProcessors(requestContext.Opts.Targets),
		TxnHeaderOpts...,
	)

	requestContext.Response.Proposal = proposal
	requestContext.Response.TransactionID = proposal.TxnID // TODO: still needed?

	if err != nil {
		requestContext.Error = checkEndorserServerError(err)
		return
	}

	requestContext.Response.Responses = transactionProposalResponses
	if len(transactionProposalResponses) > 0 {
		responsePayload, err := getResultFromProposalResponse(transactionProposalResponses[0].ProposalResponse)
		if err != nil {
			requestContext.Error = err
			return
		}

		requestContext.Response.Payload = responsePayload
		requestContext.Response.ChaincodeStatus = transactionProposalResponses[0].ChaincodeStatus
	}

	//Delegate to next step if any
	if e.next != nil {
		e.next.Handle(requestContext, clientContext)
	}
}

func getResultFromProposalResponse(proposalResponse *pb.ProposalResponse) ([]byte, error) {
	responsePayload := &pb.ProposalResponsePayload{}
	if err := proto.Unmarshal(proposalResponse.GetPayload(), responsePayload); err != nil {
		return nil, errors.Wrap(err, "failed to deserialize proposal response payload")
	}

	return getResultFromProposalResponsePayload(responsePayload)
}

func getResultFromProposalResponsePayload(responsePayload *pb.ProposalResponsePayload) ([]byte, error) {
	chaincodeAction := &pb.ChaincodeAction{}
	if err := proto.Unmarshal(responsePayload.GetExtension(), chaincodeAction); err != nil {
		return nil, errors.Wrap(err, "failed to deserialize chaincode action")
	}

	return chaincodeAction.GetResponse().GetPayload(), nil
}

//ProposalProcessorHandler for selecting proposal processors
type ProposalProcessorHandler struct {
	next Handler
}

//Handle selects proposal processors
func (h *ProposalProcessorHandler) Handle(requestContext *RequestContext, clientContext *ClientContext) {
	//Get proposal processor, if not supplied then use selection service to get available peers as endorser
	if len(requestContext.Opts.Targets) == 0 {
		var selectionOpts []options.Opt
		if requestContext.SelectionFilter != nil {
			selectionOpts = append(selectionOpts, selectopts.WithPeerFilter(requestContext.SelectionFilter))
		}
		if requestContext.PeerSorter != nil {
			selectionOpts = append(selectionOpts, selectopts.WithPeerSorter(requestContext.PeerSorter))
		}

		endorsers, err := clientContext.Selection.GetEndorsersForChaincode(newInvocationChain(requestContext), selectionOpts...)
		if err != nil {
			requestContext.Error = errors.WithMessage(err, "Failed to get endorsing peers")
			return
		}
		requestContext.Opts.Targets = endorsers
	}

	//Delegate to next step if any
	if h.next != nil {
		h.next.Handle(requestContext, clientContext)
	}
}

func newInvocationChain(requestContext *RequestContext) []*fab.ChaincodeCall {
	invocChain := []*fab.ChaincodeCall{{ID: requestContext.Request.ChaincodeID}}
	for _, ccCall := range requestContext.Request.InvocationChain {
		if ccCall.ID == invocChain[0].ID {
			invocChain[0].Collections = ccCall.Collections
		} else {
			invocChain = append(invocChain, ccCall)
		}
	}
	return invocChain
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
	var a1 *pb.ProposalResponse
	for n, r := range txProposalResponse {
		response := r.ProposalResponse.GetResponse()
		if response.Status < int32(common.Status_SUCCESS) || response.Status >= int32(common.Status_BAD_REQUEST) {
			return status.NewFromProposalResponse(r.ProposalResponse, r.Endorser)
		}
		if n == 0 {
			a1 = r.ProposalResponse
			continue
		}

		if !bytes.Equal(a1.Payload, r.ProposalResponse.Payload) ||
			!bytes.Equal(a1.GetResponse().Payload, response.Payload) {
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
	txnID := requestContext.Response.TransactionID

	//Register Tx event
	reg, statusNotifier, err := clientContext.EventService.RegisterTxStatusEvent(string(txnID)) // TODO: Change func to use TransactionID instead of string
	if err != nil {
		requestContext.Error = errors.Wrap(err, "error registering for TxStatus event")
		return
	}
	defer clientContext.EventService.Unregister(reg)

	_, err = createAndSendTransaction(clientContext.Transactor, requestContext.Response.Proposal, requestContext.Response.Responses)
	if err != nil {
		requestContext.Error = errors.Wrap(err, "CreateAndSendTransaction failed")
		return
	}

	select {
	case txStatus := <-statusNotifier:
		requestContext.Response.TxValidationCode = txStatus.TxValidationCode

		if txStatus.TxValidationCode != pb.TxValidationCode_VALID {
			requestContext.Error = status.New(status.EventServerStatus, int32(txStatus.TxValidationCode),
				"received invalid transaction", nil)
			return
		}
	case <-requestContext.Ctx.Done():
		requestContext.Error = status.New(status.ClientStatus, status.Timeout.ToInt32(),
			"Execute didn't receive block event", nil)
		return
	}

	//Delegate to next step if any
	if c.next != nil {
		c.next.Handle(requestContext, clientContext)
	}
}

//NewQueryHandler returns query handler with chain of ProposalProcessorHandler, EndorsementHandler, EndorsementValidationHandler and SignatureValidationHandler
func NewQueryHandler(next ...Handler) Handler {
	return NewProposalProcessorHandler(
		NewEndorsementHandler(
			NewEndorsementValidationHandler(
				NewSignatureValidationHandler(next...),
			),
		),
	)
}

//NewExecuteHandler returns execute handler with chain of SelectAndEndorseHandler, EndorsementValidationHandler, SignatureValidationHandler and CommitHandler
func NewExecuteHandler(next ...Handler) Handler {
	return NewSelectAndEndorseHandler(
		NewEndorsementValidationHandler(
			NewSignatureValidationHandler(NewCommitHandler(next...)),
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

//NewEndorsementHandlerWithOpts returns a handler that endorses a transaction proposal
func NewEndorsementHandlerWithOpts(next Handler, provider TxnHeaderOptsProvider) *EndorsementHandler {
	return &EndorsementHandler{next: next, headerOptsProvider: provider}
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

	return transactionResponse, nil
}

func createAndSendTransactionProposal(transactor fab.ProposalSender, chrequest *Request, targets []fab.ProposalProcessor, opts ...fab.TxnHeaderOpt) ([]*fab.TransactionProposalResponse, *fab.TransactionProposal, error) {
	request := fab.ChaincodeInvokeRequest{
		ChaincodeID:  chrequest.ChaincodeID,
		Fcn:          chrequest.Fcn,
		Args:         chrequest.Args,
		TransientMap: chrequest.TransientMap,
		IsInit:       chrequest.IsInit,
	}

	txh, err := transactor.CreateTransactionHeader(opts...)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "creating transaction header failed")
	}

	proposal, err := txn.CreateChaincodeInvokeProposal(txh, request)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "creating transaction proposal failed")
	}

	transactionProposalResponses, err := transactor.SendTransactionProposal(proposal, targets)

	return transactionProposalResponses, proposal, err
}

func checkEndorserServerError(err error) error {
	if strings.Contains(err.Error(), "failed to distribute private collection") {
		return status.New(status.EndorserServerStatus, status.PvtDataDisseminationFailed.ToInt32(), err.Error(), nil)
	}

	return err
}
