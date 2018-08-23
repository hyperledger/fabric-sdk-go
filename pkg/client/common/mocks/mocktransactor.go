/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/pkg/errors"
)

// MockTransactor provides an implementation of Transactor that exposes all its context.
type MockTransactor struct {
	Ctx       context.Client
	ChannelID string
	Orderers  []fab.Orderer
}

// CreateTransactionHeader creates a Transaction Header based on the current context.
func (t *MockTransactor) CreateTransactionHeader(opts ...fab.TxnHeaderOpt) (fab.TransactionHeader, error) {
	txh, err := txn.NewHeader(t.Ctx, t.ChannelID)
	if err != nil {
		return nil, errors.WithMessage(err, "new transaction ID failed")
	}

	return txh, nil
}

// SendTransactionProposal sends a TransactionProposal to the target peers.
func (t *MockTransactor) SendTransactionProposal(proposal *fab.TransactionProposal, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {
	rqtx, cancel := contextImpl.NewRequest(t.Ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	return txn.SendProposal(rqtx, proposal, targets)
}

// CreateTransaction create a transaction with proposal response.
func (t *MockTransactor) CreateTransaction(request fab.TransactionRequest) (*fab.Transaction, error) {
	return txn.New(request)
}

// SendTransaction send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
func (t *MockTransactor) SendTransaction(tx *fab.Transaction) (*fab.TransactionResponse, error) {
	rqtx, cancel := contextImpl.NewRequest(t.Ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	return txn.Send(rqtx, tx, t.Orderers)
}
