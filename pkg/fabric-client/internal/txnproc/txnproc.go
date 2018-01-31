/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package txnproc provides functionality for processing fabric transactions.
package txnproc

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// SendTransactionProposalToProcessors sends a TransactionProposal to ProposalProcessor.
func SendTransactionProposalToProcessors(proposal *apifabclient.TransactionProposal, targets []apifabclient.ProposalProcessor) ([]*apifabclient.TransactionProposalResponse, error) {

	if proposal == nil || proposal.SignedProposal == nil {
		return nil, errors.New("signedProposal is required")
	}

	if len(targets) < 1 {
		return nil, errors.New("targets is required")
	}

	var responseMtx sync.Mutex
	var transactionProposalResponses []*apifabclient.TransactionProposalResponse
	var wg sync.WaitGroup

	for _, p := range targets {
		wg.Add(1)
		go func(processor apifabclient.ProposalProcessor) {
			defer wg.Done()

			r, err := processor.ProcessTransactionProposal(*proposal)
			if err != nil {
				logger.Debugf("Received error response from txn proposal processing: %v", err)
				// Error is handled downstream.
			}

			tpr := apifabclient.TransactionProposalResponse{
				TransactionProposalResult: r, Err: err}

			responseMtx.Lock()
			transactionProposalResponses = append(transactionProposalResponses, &tpr)
			responseMtx.Unlock()
		}(p)
	}
	wg.Wait()
	return transactionProposalResponses, nil
}
