/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package txnproc provides functionality for processing fabric transactions.
package txnproc

import (
	"fmt"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	logging "github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

// SendTransactionProposalToProcessors sends a TransactionProposal to ProposalProcessor.
func SendTransactionProposalToProcessors(proposal *apitxn.TransactionProposal, targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, error) {

	if proposal == nil || proposal.SignedProposal == nil {
		return nil, fmt.Errorf("signedProposal is nil")
	}

	if len(targets) < 1 {
		return nil, fmt.Errorf("Missing peer objects for sending transaction proposal")
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
