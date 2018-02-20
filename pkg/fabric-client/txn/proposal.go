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

// CreateChaincodeInvokeProposal creates a proposal for transaction.
func CreateChaincodeInvokeProposal(txid fab.TransactionID, channelID string, request fab.ChaincodeInvokeRequest) (*fab.TransactionProposal, error) {
	if request.ChaincodeID == "" {
		return nil, errors.New("ChaincodeID is required")
	}

	if request.Fcn == "" {
		return nil, errors.New("Fcn is required")
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

	proposal, _, err := protos_utils.CreateChaincodeProposalWithTxIDNonceAndTransient(txid.ID, common.HeaderType_ENDORSER_TRANSACTION, channelID, ccis, txid.Nonce, txid.Creator, request.TransientMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chaincode proposal")
	}

	tp := fab.TransactionProposal{
		TxnID:    txid,
		Proposal: proposal,
	}

	return &tp, nil
}

// signProposal creates a SignedProposal based on the current context.
func signProposal(ctx context, proposal *pb.Proposal) (*pb.SignedProposal, error) {
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
		return nil, errors.WithMessage(err, "sign failed")
	}

	return &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}, nil
}

// SendProposal sends a TransactionProposal to ProposalProcessor.
func SendProposal(ctx context, proposal *fab.TransactionProposal, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {

	if proposal == nil {
		return nil, errors.New("proposal is required")
	}

	if len(targets) < 1 {
		return nil, errors.New("targets is required")
	}

	signedProposal, err := signProposal(ctx, proposal.Proposal)
	if err != nil {
		return nil, errors.WithMessage(err, "sign proposal failed")
	}

	request := fab.ProcessProposalRequest{SignedProposal: signedProposal}

	var responseMtx sync.Mutex
	var transactionProposalResponses []*fab.TransactionProposalResponse
	var wg sync.WaitGroup
	errs := multi.Errors{}

	for _, p := range targets {
		wg.Add(1)
		go func(processor fab.ProposalProcessor) {
			defer wg.Done()

			resp, err := processor.ProcessTransactionProposal(request)
			if err != nil {
				logger.Debugf("Received error response from txn proposal processing: %v", err)
				responseMtx.Lock()
				errs = append(errs, err)
				responseMtx.Unlock()
				return
			}

			responseMtx.Lock()
			transactionProposalResponses = append(transactionProposalResponses, resp)
			responseMtx.Unlock()
		}(p)
	}
	wg.Wait()

	return transactionProposalResponses, errs.ToError()
}
