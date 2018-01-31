/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// Sender provides the ability for a transaction to be created and sent.
type Sender interface {
	CreateTransaction(resps []*TransactionProposalResponse) (*Transaction, error)
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
	Err     error
}
