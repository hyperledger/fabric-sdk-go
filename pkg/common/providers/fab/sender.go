/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// TransactionRequest holds endorsed Transaction Proposals.
type TransactionRequest struct {
	Proposal          *TransactionProposal
	ProposalResponses []*TransactionProposalResponse
}

// Sender provides the ability for a transaction to be created and sent.
//
// TODO: CreateTransaction should be refactored as it is actually a factory method.
type Sender interface {
	CreateTransaction(request TransactionRequest) (*Transaction, error)
	SendTransaction(tx *Transaction) (*TransactionResponse, error)
}

// The Transaction object created from an endorsed proposal.
type Transaction struct {
	Proposal    *TransactionProposal
	Transaction *pb.Transaction
}

// TransactionResponse contains information returned by the orderer.
type TransactionResponse struct {
	Orderer string
}
