/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package txn enables creating, endorsing and sending transactions to Fabric peers and orderers.
package txn

import (
	reqContext "context"
	"math/rand"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	ctxprovider "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
)

var logger = logging.NewLogger("fabsdk/fab")

// CCProposalType reflects transitions in the chaincode lifecycle
type CCProposalType int

// Define chaincode proposal types
const (
	Instantiate CCProposalType = iota
	Upgrade
)

// New create a transaction with proposal response, following the endorsement policy.
func New(request fab.TransactionRequest) (*fab.Transaction, error) {
	if len(request.ProposalResponses) == 0 {
		return nil, errors.New("at least one proposal response is necessary")
	}

	proposal := request.Proposal

	// the original header
	hdr, err := protoutil.UnmarshalHeader(proposal.Header)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal proposal header failed")
	}

	// the original payload
	pPayl, err := protoutil.UnmarshalChaincodeProposalPayload(proposal.Payload)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal proposal payload failed")
	}

	responsePayload := request.ProposalResponses[0].ProposalResponse.Payload
	if vprErr := validateProposalResponses(request.ProposalResponses); vprErr != nil {
		return nil, vprErr
	}

	// fill endorsements
	endorsements := make([]*pb.Endorsement, len(request.ProposalResponses))
	for n, r := range request.ProposalResponses {
		endorsements[n] = r.ProposalResponse.Endorsement
	}

	// create ChaincodeEndorsedAction
	cea := &pb.ChaincodeEndorsedAction{ProposalResponsePayload: responsePayload, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := protoutil.GetBytesProposalPayloadForTx(pPayl)
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &pb.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := protoutil.GetBytesChaincodeActionPayload(cap)
	if err != nil {
		return nil, err
	}

	// create a transaction
	taa := &pb.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*pb.TransactionAction, 1)
	taas[0] = taa

	return &fab.Transaction{
		Transaction: &pb.Transaction{Actions: taas},
		Proposal:    proposal,
	}, nil
}

func validateProposalResponses(responses []*fab.TransactionProposalResponse) error {
	for _, r := range responses {
		if r.ProposalResponse.Response.Status < int32(common.Status_SUCCESS) || r.ProposalResponse.Response.Status >= int32(common.Status_BAD_REQUEST) {
			return errors.Errorf("proposal response was not successful, error code %d, msg %s", r.ProposalResponse.Response.Status, r.ProposalResponse.Response.Message)
		}
	}
	return nil
}

// Send send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
func Send(reqCtx reqContext.Context, tx *fab.Transaction, orderers []fab.Orderer) (*fab.TransactionResponse, error) {
	if len(orderers) == 0 {
		return nil, errors.New("orderers is nil")
	}
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	if tx.Proposal == nil || tx.Proposal.Proposal == nil {
		return nil, errors.New("proposal is nil")
	}

	// the original header
	hdr, err := protoutil.UnmarshalHeader(tx.Proposal.Proposal.Header)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal proposal header failed")
	}
	// serialize the tx
	txBytes, err := protoutil.GetBytesTransaction(tx.Transaction)
	if err != nil {
		return nil, err
	}

	// create the payload
	payload := common.Payload{Header: hdr, Data: txBytes}

	transactionResponse, err := BroadcastPayload(reqCtx, &payload, orderers)
	if err != nil {
		return nil, err
	}

	return transactionResponse, nil
}

// BroadcastPayload will send the given payload to some orderer, picking random endpoints
// until all are exhausted
func BroadcastPayload(reqCtx reqContext.Context, payload *common.Payload, orderers []fab.Orderer) (*fab.TransactionResponse, error) {
	// Check if orderers are defined
	if len(orderers) == 0 {
		return nil, errors.New("orderers not set")
	}

	ctx, ok := context.RequestClientContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for signPayload")
	}
	envelope, err := signPayload(ctx, payload)
	if err != nil {
		return nil, err
	}

	return broadcastEnvelope(reqCtx, envelope, orderers)
}

// broadcastEnvelope will send the given envelope to some orderer, picking random endpoints
// until all are exhausted
func broadcastEnvelope(reqCtx reqContext.Context, envelope *fab.SignedEnvelope, orderers []fab.Orderer) (*fab.TransactionResponse, error) {
	// Check if orderers are defined
	if len(orderers) == 0 {
		return nil, errors.New("orderers not set")
	}

	// Copy aside the ordering service endpoints
	randOrderers := []fab.Orderer{}
	randOrderers = append(randOrderers, orderers...)

	// get a context client instance to create child contexts with timeout read from the config in sendBroadcast()
	ctxClient, ok := context.RequestClientContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for SendTransaction")
	}

	// Iterate them in a random order and try broadcasting 1 by 1
	var errResp error
	for _, i := range rand.Perm(len(randOrderers)) {
		resp, err := sendBroadcast(reqCtx, envelope, randOrderers[i], ctxClient)
		if err != nil {
			errResp = err
		} else {
			return resp, nil
		}
	}
	return nil, errResp
}

func sendBroadcast(reqCtx reqContext.Context, envelope *fab.SignedEnvelope, orderer fab.Orderer, client ctxprovider.Client) (*fab.TransactionResponse, error) {
	logger.Debugf("Broadcasting envelope to orderer: %s\n", orderer.URL())
	// create a childContext for this SendBroadcast orderer using the config's timeout value
	// the parent context (reqCtx) should not have a timeout value
	childCtx, cancel := context.NewRequest(client, context.WithTimeoutType(fab.OrdererResponse), context.WithParent(reqCtx))
	defer cancel()

	// Send request
	if _, err := orderer.SendBroadcast(childCtx, envelope); err != nil {
		logger.Debugf("Receive Error Response from orderer: %s\n", err)
		return nil, errors.Wrapf(err, "calling orderer '%s' failed", orderer.URL())
	}

	logger.Debugf("Receive Success Response from orderer\n")
	return &fab.TransactionResponse{Orderer: orderer.URL()}, nil
}

// SendPayload sends the given payload to each orderer and returns a block response
func SendPayload(reqCtx reqContext.Context, payload *common.Payload, orderers []fab.Orderer) (*common.Block, error) {
	if len(orderers) == 0 {
		return nil, errors.New("orderers not set")
	}

	ctx, ok := context.RequestClientContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for signPayload")
	}
	envelope, err := signPayload(ctx, payload)
	if err != nil {
		return nil, err
	}

	// Copy aside the ordering service endpoints
	randOrderers := []fab.Orderer{}
	randOrderers = append(randOrderers, orderers...)

	// Iterate them in a random order and try broadcasting 1 by 1
	var errResp error
	for _, i := range rand.Perm(len(randOrderers)) {
		resp, err := sendEnvelope(reqCtx, envelope, randOrderers[i])
		if err != nil {
			errResp = err
		} else {
			return resp, nil
		}
	}
	return nil, errResp
}

// sendEnvelope sends the given envelope to each orderer and returns a block response
func sendEnvelope(reqCtx reqContext.Context, envelope *fab.SignedEnvelope, orderer fab.Orderer) (*common.Block, error) {

	logger.Debugf("Broadcasting envelope to orderer :%s\n", orderer.URL())
	blocks, errs := orderer.SendDeliver(reqCtx, envelope)

	// This function currently returns the last received block and error.
	var block *common.Block
	var err multi.Errors

read:
	for {
		select {
		case b, ok := <-blocks:
			// We need to block until SendDeliver releases the connection. Currently
			// this is triggered by the go chan closing.
			// TODO: we may want to refactor (e.g., adding a synchronous SendDeliver)
			if !ok {
				break read
			}
			block = b
		case e := <-errs:
			err = append(err, e)
		}
	}

	// drain remaining errors.
	for i := 0; i < len(errs); i++ {
		e := <-errs
		err = append(err, e)
	}

	return block, err.ToError()
}
