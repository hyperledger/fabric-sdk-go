/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package txnproc provides functionality for processing fabric transactions.
package txnproc

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// SendTransactionProposalToProcessors sends a TransactionProposal to ProposalProcessor.
func SendTransactionProposalToProcessors(proposal *apitxn.TransactionProposal, targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, error) {

	if proposal == nil || proposal.SignedProposal == nil {
		return nil, errors.New("signedProposal is required")
	}

	if len(targets) < 1 {
		return nil, errors.New("targets is required")
	}

	var responseMtx sync.Mutex
	var transactionProposalResponses []*apitxn.TransactionProposalResponse
	var wg sync.WaitGroup

	for _, p := range targets {
		wg.Add(1)
		go func(processor apitxn.ProposalProcessor) {
			defer wg.Done()

			r, err := processor.ProcessTransactionProposal(*proposal)
			if err != nil {
				logger.Debugf("Received error response from txn proposal processing: %v", err)
				// Error is handled downstream.
			}

			tpr := apitxn.TransactionProposalResponse{
				TransactionProposalResult: r, Err: err}

			responseMtx.Lock()
			transactionProposalResponses = append(transactionProposalResponses, &tpr)
			responseMtx.Unlock()
		}(p)
	}
	wg.Wait()
	return transactionProposalResponses, nil
}
