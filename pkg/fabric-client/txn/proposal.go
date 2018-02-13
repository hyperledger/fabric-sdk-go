/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txn

import (
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/multi"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
)

// NewProposal creates a proposal for transaction. This involves assembling the proposal with the data
// (chaincodeName, function to call, arguments, transient data, etc.) and signing it using identity in the current context.
func NewProposal(ctx context, channelID string, request fab.ChaincodeInvokeRequest) (*fab.TransactionProposal, error) {
	if request.ChaincodeID == "" {
		return nil, errors.New("ChaincodeID is required")
	}

	if request.Fcn == "" {
		return nil, errors.New("Fcn is required")
	}

	txid, err := NewID(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to create a transaction ID")
	}

	// Add function name to arguments
	argsArray := make([][]byte, len(request.Args)+1)
	argsArray[0] = []byte(request.Fcn)
	for i, arg := range request.Args {
		argsArray[i+1] = arg
	}

	// create invocation spec to target a chaincode with arguments
	ccis := &pb.ChaincodeInvocationSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: request.ChaincodeID},
		Input: &pb.ChaincodeInput{Args: argsArray}}}

	// create a proposal from a ChaincodeInvocationSpec
	creator, err := ctx.Identity()
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to get user context identity")
	}

	proposal, _, err := protos_utils.CreateChaincodeProposalWithTxIDNonceAndTransient(txid.ID, common.HeaderType_ENDORSER_TRANSACTION, channelID, ccis, txid.Nonce, creator, request.TransientMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chaincode proposal")
	}

	// sign proposal bytes
	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "marshal proposal failed")
	}

	signingMgr := ctx.SigningManager()
	if signingMgr == nil {
		return nil, errors.New("signing manager is nil")
	}

	signature, err := signingMgr.Sign(proposalBytes, ctx.PrivateKey())
	if err != nil {
		return nil, err
	}

	// construct the transaction proposal
	signedProposal := pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}
	tp := fab.TransactionProposal{
		TxnID:          txid,
		SignedProposal: &signedProposal,
		Proposal:       proposal,
	}

	return &tp, nil
}

// SignProposal creates a SignedProposal based on the current context.
func SignProposal(ctx context, proposal *pb.Proposal) (*pb.SignedProposal, error) {
	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "mashal proposal failed")
	}

	signingMgr := ctx.SigningManager()
	if signingMgr == nil {
		return nil, errors.New("signing manager is nil")
	}

	signature, err := signingMgr.Sign(proposalBytes, ctx.PrivateKey())
	if err != nil {
		return nil, errors.WithMessage(err, "signing proposal failed")
	}

	return &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}, nil
}

// SendProposal sends a TransactionProposal to ProposalProcessor.
func SendProposal(proposal *fab.TransactionProposal, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {

	if proposal == nil || proposal.SignedProposal == nil {
		return nil, errors.New("signedProposal is required")
	}

	if len(targets) < 1 {
		return nil, errors.New("targets is required")
	}

	var responseMtx sync.Mutex
	var transactionProposalResponses []*fab.TransactionProposalResponse
	var wg sync.WaitGroup
	errs := multi.Errors{}

	for _, p := range targets {
		wg.Add(1)
		go func(processor fab.ProposalProcessor) {
			defer wg.Done()

			resp, err := processor.ProcessTransactionProposal(*proposal)
			if err != nil {
				logger.Debugf("Received error response from txn proposal processing: %v", err)
				responseMtx.Lock()
				errs = append(errs, err)
				responseMtx.Unlock()
				return
			}

			responseMtx.Lock()
			transactionProposalResponses = append(transactionProposalResponses, &resp)
			responseMtx.Unlock()
		}(p)
	}
	wg.Wait()

	return transactionProposalResponses, errs.ToError()
}
