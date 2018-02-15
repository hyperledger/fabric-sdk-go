/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
	"github.com/pkg/errors"
)

// MockTransactor provides an implementation of Transactor that exposes all its context.
type MockTransactor struct {
	Ctx       fab.Context
	ChannelID string
	Orderers  []fab.Orderer
}

// CreateChaincodeInvokeProposal creates a Transaction Proposal based on the current context and channel config.
func (t *MockTransactor) CreateChaincodeInvokeProposal(request fab.ChaincodeInvokeRequest) (*fab.TransactionProposal, error) {
	tp, err := txn.CreateChaincodeInvokeProposal(t.Ctx, t.ChannelID, request)
	if err != nil {
		return nil, errors.WithMessage(err, "new transaction proposal failed")
	}

	return tp, nil
}

// SendTransactionProposal ...
func (t *MockTransactor) SendTransactionProposal(proposal *fab.TransactionProposal, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {
	return txn.SendProposal(t.Ctx, proposal, targets)
}

// CreateTransaction create a transaction with proposal response.
func (t *MockTransactor) CreateTransaction(request fab.TransactionRequest) (*fab.Transaction, error) {
	return txn.New(request)
}

// SendTransaction send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
func (t *MockTransactor) SendTransaction(tx *fab.Transaction) (*fab.TransactionResponse, error) {
	return txn.Send(t.Ctx, tx, t.Orderers)
}
