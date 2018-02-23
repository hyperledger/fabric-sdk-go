/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package txn enables creating, endorsing and sending transactions to Fabric peers and orderers.
package txn

import (
	"bytes"
	"math/rand"
	"sync"
	"time"

	"github.com/pkg/errors"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
)

var logger = logging.NewLogger("fabric_sdk_go")

// CCProposalType reflects transitions in the chaincode lifecycle
type CCProposalType int

// Define chaincode proposal types
const (
	Instantiate CCProposalType = iota
	Upgrade
)

type context interface {
	contextApi.ProviderContext
	contextApi.IdentityContext
}

// New create a transaction with proposal response, following the endorsement policy.
func New(request fab.TransactionRequest) (*fab.Transaction, error) {
	if len(request.ProposalResponses) == 0 {
		return nil, errors.New("at least one proposal response is necessary")
	}

	proposal := request.Proposal

	// the original header
	hdr, err := protos_utils.GetHeader(proposal.Header)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal proposal header failed")
	}

	// the original payload
	pPayl, err := protos_utils.GetChaincodeProposalPayload(proposal.Payload)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal proposal payload failed")
	}

	// get header extensions so we have the visibility field
	hdrExt, err := protos_utils.GetChaincodeHeaderExtension(hdr)
	if err != nil {
		return nil, err
	}

	responsePayload := request.ProposalResponses[0].ProposalResponse.Payload
	for _, r := range request.ProposalResponses {
		if r.ProposalResponse.Response.Status != 200 {
			return nil, errors.Errorf("proposal response was not successful, error code %d, msg %s", r.ProposalResponse.Response.Status, r.ProposalResponse.Response.Message)
		}
		if !bytes.Equal(responsePayload, r.ProposalResponse.Payload) {
			return nil, errors.Errorf("proposal response payloads are not the same (%v, %v)", responsePayload, r.ProposalResponse.Payload)
		}
	}

	// fill endorsements
	endorsements := make([]*pb.Endorsement, len(request.ProposalResponses))
	for n, r := range request.ProposalResponses {
		endorsements[n] = r.ProposalResponse.Endorsement
	}

	// create ChaincodeEndorsedAction
	cea := &pb.ChaincodeEndorsedAction{ProposalResponsePayload: responsePayload, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := protos_utils.GetBytesProposalPayloadForTx(pPayl, hdrExt.PayloadVisibility)
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &pb.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := protos_utils.GetBytesChaincodeActionPayload(cap)
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

// Send send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
func Send(ctx context, tx *fab.Transaction, orderers []fab.Orderer) (*fab.TransactionResponse, error) {
	if orderers == nil || len(orderers) == 0 {
		return nil, errors.New("orderers is nil")
	}
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	if tx.Proposal == nil || tx.Proposal.Proposal == nil {
		return nil, errors.New("proposal is nil")
	}

	// the original header
	hdr, err := protos_utils.GetHeader(tx.Proposal.Proposal.Header)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal proposal header failed")
	}
	// serialize the tx
	txBytes, err := protos_utils.GetBytesTransaction(tx.Transaction)
	if err != nil {
		return nil, err
	}

	// create the payload
	payload := common.Payload{Header: hdr, Data: txBytes}

	transactionResponse, err := BroadcastPayload(ctx, &payload, orderers)
	if err != nil {
		return nil, err
	}

	return transactionResponse, nil
}

// BroadcastPayload will send the given payload to some orderer, picking random endpoints
// until all are exhausted
func BroadcastPayload(ctx context, payload *common.Payload, orderers []fab.Orderer) (*fab.TransactionResponse, error) {
	// Check if orderers are defined
	if len(orderers) == 0 {
		return nil, errors.New("orderers not set")
	}

	envelope, err := signPayload(ctx, payload)
	if err != nil {
		return nil, err
	}

	return broadcastEnvelope(ctx, envelope, orderers)
}

// broadcastEnvelope will send the given envelope to some orderer, picking random endpoints
// until all are exhausted
func broadcastEnvelope(ctx context, envelope *fab.SignedEnvelope, orderers []fab.Orderer) (*fab.TransactionResponse, error) {
	// Check if orderers are defined
	if len(orderers) == 0 {
		return nil, errors.New("orderers not set")
	}

	// Copy aside the ordering service endpoints
	randOrderers := []fab.Orderer{}
	for _, o := range orderers {
		randOrderers = append(randOrderers, o)
	}

	// Iterate them in a random order and try broadcasting 1 by 1
	var errResp *fab.TransactionResponse
	for _, i := range rand.Perm(len(randOrderers)) {
		resp := sendBroadcast(envelope, randOrderers[i])
		if resp.Err != nil {
			errResp = resp
		} else {
			return resp, nil
		}
	}
	return errResp, nil
}

func sendBroadcast(envelope *fab.SignedEnvelope, orderer fab.Orderer) *fab.TransactionResponse {
	logger.Debugf("Broadcasting envelope to orderer :%s\n", orderer.URL())
	if _, err := orderer.SendBroadcast(envelope); err != nil {
		logger.Debugf("Receive Error Response from orderer :%v\n", err)
		return &fab.TransactionResponse{Orderer: orderer.URL(),
			Err: errors.Wrapf(err, "calling orderer '%s' failed", orderer.URL())}
	}

	logger.Debugf("Receive Success Response from orderer\n")
	return &fab.TransactionResponse{Orderer: orderer.URL(), Err: nil}
}

// SendPayload sends the given payload to each orderer and returns a block response
func SendPayload(ctx context, payload *common.Payload, orderers []fab.Orderer) (*common.Block, error) {
	if orderers == nil || len(orderers) == 0 {
		return nil, errors.New("orderers not set")
	}

	envelope, err := signPayload(ctx, payload)
	if err != nil {
		return nil, err
	}

	return sendEnvelope(ctx, envelope, orderers)
}

// sendEnvelope sends the given envelope to each orderer and returns a block response
func sendEnvelope(ctx context, envelope *fab.SignedEnvelope, orderers []fab.Orderer) (*common.Block, error) {

	var blockResponse *common.Block
	var errorResponse error
	var mutex sync.Mutex
	outstandingRequests := len(orderers)
	done := make(chan bool)

	// Send the request to all orderers and return as soon as one responds with a block.
	for _, o := range orderers {

		go func(orderer fab.Orderer) {
			logger.Debugf("Broadcasting envelope to orderer :%s\n", orderer.URL())

			blocks, errs, cancel := orderer.SendDeliver(envelope)
			defer cancel()

			select {
			case block := <-blocks:
				mutex.Lock()
				if blockResponse == nil {
					blockResponse = block
					done <- true
				}
				mutex.Unlock()

			case err := <-errs:
				mutex.Lock()
				if errorResponse == nil {
					errorResponse = err
				}
				outstandingRequests--
				if outstandingRequests == 0 {
					done <- true
				}
				mutex.Unlock()

			case <-time.After(ctx.Config().TimeoutOrDefault(core.OrdererResponse)):
				mutex.Lock()
				if errorResponse == nil {
					errorResponse = errors.New("timeout waiting for response from orderer")
				}
				outstandingRequests--
				if outstandingRequests == 0 {
					done <- true
				}
				mutex.Unlock()
			}
		}(o)
	}

	<-done

	if blockResponse != nil {
		return blockResponse, nil
	}

	// There must be an error
	if errorResponse != nil {
		return nil, errors.Wrap(errorResponse, "error returned from orderer service")
	}

	return nil, errors.New("unexpected: didn't receive a block from any of the orderer servces and didn't receive any error")
}

// Status is the transaction status returned from eventhub tx events
type Status struct {
	Code  pb.TxValidationCode
	Error error
}

// RegisterStatus registers on the given eventhub for the given transaction id
// returns a TxValidationCode channel which receives the validation code when the
// transaction completes. If the code is TxValidationCode_VALID then
// the transaction committed successfully, otherwise the code indicates the error
// that occurred.
func RegisterStatus(txID fab.TransactionID, eventHub fab.EventHub) chan Status {
	statusNotifier := make(chan Status)

	eventHub.RegisterTxEvent(txID, func(txId string, code pb.TxValidationCode, err error) {
		logger.Debugf("Received code(%s) for txid(%s) and err(%s)\n", code, txId, err)
		statusNotifier <- Status{Code: code, Error: err}
	})

	return statusNotifier
}
