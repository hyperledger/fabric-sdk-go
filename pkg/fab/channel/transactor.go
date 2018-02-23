/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
)

// Transactor enables sending transactions and transaction proposals on the channel.
type Transactor struct {
	ctx       context.Context
	ChannelID string
	orderers  []fab.Orderer
}

// NewTransactor returns a Transactor for the current context and channel config.
func NewTransactor(ctx context.Context, cfg fab.ChannelCfg) (*Transactor, error) {
	orderers, err := orderersFromChannelCfg(ctx, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "reading orderers from channel config failed")
	}
	// TODO: adjust integration tests to always have valid orderers (even when doing only queries)
	//if len(orderers) == 0 {
	//	return nil, errors.New("orderers are not configured")
	//}

	t := Transactor{
		ctx:       ctx,
		ChannelID: cfg.Name(),
		orderers:  orderers,
	}
	return &t, nil
}

func orderersFromChannelCfg(ctx context.Context, cfg fab.ChannelCfg) ([]fab.Orderer, error) {
	orderers := []fab.Orderer{}
	ordererDict, err := orderersByTarget(ctx)
	if err != nil {
		return nil, err
	}

	// Add orderer if specified in config
	for _, target := range cfg.Orderers() {

		// Figure out orderer configuration
		oCfg, ok := ordererDict[target]

		if !ok {
			// TODO: need default options
			o, err := orderer.New(ctx.Config(), orderer.WithURL(target))
			// TODO: should we fail hard if we cannot configure a default orderer?
			//if err != nil {
			//	return nil, errors.WithMessage(err, "failed to create orderer from defaults")
			//}
			if err == nil {
				orderers = append(orderers, o)
			}
		} else {
			o, err := orderer.New(ctx.Config(), orderer.FromOrdererConfig(&oCfg))
			if err != nil {
				return nil, errors.WithMessage(err, "failed to create orderer from config")
			}
			orderers = append(orderers, o)
		}
	}
	return orderers, nil
}

func orderersByTarget(ctx context.Context) (map[string]core.OrdererConfig, error) {
	ordererDict := map[string]core.OrdererConfig{}
	orderersConfig, err := ctx.Config().OrderersConfig()
	if err != nil {
		return nil, errors.WithMessage(err, "loading orderers config failed")
	}

	for _, oc := range orderersConfig {
		address := urlutil.ToAddress(oc.URL)
		ordererDict[address] = oc
	}
	return ordererDict, nil
}

// CreateTransactionID creates a Transaction ID based on the current context.
func (t *Transactor) CreateTransactionID() (fab.TransactionID, error) {
	txid, err := txn.NewID(t.ctx)
	if err != nil {
		return fab.TransactionID{}, errors.WithMessage(err, "new transaction ID failed")
	}

	return txid, nil
}

// CreateChaincodeInvokeProposal creates a Transaction Proposal based on the current context and channel ID.
func (t *Transactor) CreateChaincodeInvokeProposal(request fab.ChaincodeInvokeRequest) (*fab.TransactionProposal, error) {
	txid, err := t.CreateTransactionID()
	if err != nil {
		return nil, errors.WithMessage(err, "create transaction ID failed")
	}

	tp, err := txn.CreateChaincodeInvokeProposal(txid, t.ChannelID, request)
	if err != nil {
		return nil, errors.WithMessage(err, "new transaction proposal failed")
	}

	return tp, nil
}

// SendTransactionProposal sends a TransactionProposal to the target peers.
func (t *Transactor) SendTransactionProposal(proposal *fab.TransactionProposal, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {
	return txn.SendProposal(t.ctx, proposal, targets)
}

// CreateTransaction create a transaction with proposal response.
// TODO: should this be removed as it is purely a wrapper?
func (t *Transactor) CreateTransaction(request fab.TransactionRequest) (*fab.Transaction, error) {
	return txn.New(request)
}

// SendTransaction send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
func (t *Transactor) SendTransaction(tx *fab.Transaction) (*fab.TransactionResponse, error) {
	return txn.Send(t.ctx, tx, t.orderers)
}
