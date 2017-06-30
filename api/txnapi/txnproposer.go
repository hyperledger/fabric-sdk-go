/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package txnapi allows SDK users to plugin their own implementations of transaction processing.
package txnapi

import (
	pb "github.com/hyperledger/fabric/protos/peer"
)

// TxnProposalProcessor simulates transaction proposal, so that a client can submit the result for ordering.
type TxnProposalProcessor interface {
	ProcessTransactionProposal(proposal TransactionProposal) (TransactionProposalResult, error)
}

// TransactionProposal requests simulation of a proposed transaction from transaction processors.
type TransactionProposal struct {
	TransactionID string

	SignedProposal *pb.SignedProposal
	Proposal       *pb.Proposal
}

// TransactionProposalResponse encapsulates both the result of transaction proposal processing and errors.
type TransactionProposalResponse struct {
	TransactionProposalResult
	Err error // TODO: consider refactoring
}

// TransactionProposalResult respresents the result of transaction proposal processing.
type TransactionProposalResult struct {
	Endorser string
	Status   int32

	Proposal         TransactionProposal
	ProposalResponse *pb.ProposalResponse
}

// TODO: TransactionProposalResponse and TransactionProposalResult may need better names.
