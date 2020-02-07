/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	reqContext "context"
	"strings"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
)

// Transactor enables sending transactions and transaction proposals on the channel.
type Transactor struct {
	reqCtx    reqContext.Context
	ChannelID string
	orderers  []fab.Orderer
}

// NewTransactor returns a Transactor for the current context and channel config.
func NewTransactor(reqCtx reqContext.Context, cfg fab.ChannelCfg) (*Transactor, error) {

	ctx, ok := contextImpl.RequestClientContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for create new transactor")
	}

	orderers, err := orderersFromChannelCfg(ctx, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "reading orderers from channel config failed")
	}
	// TODO: adjust integration tests to always have valid orderers (even when doing only queries)
	//if len(orderers) == 0 {
	//	return nil, errors.New("orderers are not configured")
	//}

	t := Transactor{
		reqCtx:    reqCtx,
		ChannelID: cfg.ID(),
		orderers:  orderers,
	}
	return &t, nil
}

func orderersFromChannelCfg(ctx context.Client, cfg fab.ChannelCfg) ([]fab.Orderer, error) {

	//below call to get orderers from endpoint config 'channels.<CHANNEL-ID>.orderers' is not recommended.
	//To override any orderer configuration items, entity matchers should be used.
	orderers, err := orderersFromChannel(ctx, cfg.ID())
	if err != nil {
		return nil, err
	}
	if len(orderers) > 0 {
		logger.Debugf("there [%v] orderer(s) in SDK config returning these and skipping channelCfg", len(orderers))
		return orderers, nil
	}

	ordererDict := orderersByTarget(ctx)

	logger.Debugf("there are no 'channel orderer(s)' in SDK configs. Got [%v] 'orderer(s)' configs from SDK and try to lookup additional ones in channelCfg", len(ordererDict))

	// Add orderer if specified in channel config
	for _, target := range cfg.Orderers() {

		// Figure out orderer configuration
		oCfg, ok := ordererDict[target]

		//try entity matcher
		if !ok {
			logger.Debugf("Failed to get channel Cfg orderer [%s] from ordererDict, now trying orderer Matchers in Entity Matchers", target)
			// Try to find a match from entityMatchers config
			matchingOrdererConfig, found, ignore := ctx.EndpointConfig().OrdererConfig(strings.ToLower(target))
			if ignore {
				logger.Debugf("orderer [%s] is ignored and will not be added", target)
				continue
			}

			if found {
				logger.Debugf("Found matching ordererConfig from entity Matchers for channel Cfg Orderer [%s]", target)
				oCfg = *matchingOrdererConfig
				ok = true
			}
		}

		//create orderer using channel config block orderer address
		if !ok {
			logger.Debugf("Unable to find matching ordererConfig from entity Matchers for channel Cfg Orderer [%s]", target)
			oCfg = fab.OrdererConfig{
				URL: target,
			}
			logger.Debugf("Created a new OrdererConfig with URL as [%s]", target)
		}

		o, err := ctx.InfraProvider().CreateOrdererFromConfig(&oCfg)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create orderer from config")
		}
		orderers = append(orderers, o)

	}
	return orderers, nil
}

//deprecated
//orderersFromChannel returns list of fab.Orderer by channel id
//will return empty list when orderers are not found in endpoint config
func orderersFromChannel(ctx context.Client, channelID string) ([]fab.Orderer, error) {

	chNetworkConfig := ctx.EndpointConfig().ChannelConfig(channelID)
	orderers := []fab.Orderer{}
	for _, chOrderer := range chNetworkConfig.Orderers {

		ordererConfig, found, ignoreOrderer := ctx.EndpointConfig().OrdererConfig(chOrderer)
		if !found || ignoreOrderer {
			//continue if given channel orderer not found in endpoint config
			continue
		}

		orderer, err := ctx.InfraProvider().CreateOrdererFromConfig(ordererConfig)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create orderer from config")
		}

		orderers = append(orderers, orderer)
	}
	return orderers, nil
}

func orderersByTarget(ctx context.Client) map[string]fab.OrdererConfig {
	ordererDict := map[string]fab.OrdererConfig{}
	orderersConfig := ctx.EndpointConfig().OrderersConfig()

	for _, oc := range orderersConfig {
		address := endpoint.ToAddress(oc.URL)
		ordererDict[address] = oc
		logger.Debugf("ordererConfig from SDK to be added: %s", oc.URL)
	}
	return ordererDict
}

// CreateTransactionHeader creates a Transaction Header based on the current context.
func (t *Transactor) CreateTransactionHeader(opts ...fab.TxnHeaderOpt) (fab.TransactionHeader, error) {

	ctx, ok := contextImpl.RequestClientContext(t.reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for txn Header")
	}

	txh, err := txn.NewHeader(ctx, t.ChannelID, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "new transaction ID failed")
	}

	return txh, nil
}

// SendTransactionProposal sends a TransactionProposal to the target peers.
func (t *Transactor) SendTransactionProposal(proposal *fab.TransactionProposal, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {
	ctx, ok := contextImpl.RequestClientContext(t.reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for SendTransactionProposal")
	}

	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeoutType(fab.PeerResponse), contextImpl.WithParent(t.reqCtx))
	defer cancel()

	return txn.SendProposal(reqCtx, proposal, targets)
}

// CreateTransaction create a transaction with proposal response.
// TODO: should this be removed as it is purely a wrapper?
func (t *Transactor) CreateTransaction(request fab.TransactionRequest) (*fab.Transaction, error) {
	return txn.New(request)
}

// SendTransaction send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
func (t *Transactor) SendTransaction(tx *fab.Transaction) (*fab.TransactionResponse, error) {
	return txn.Send(t.reqCtx, tx, t.orderers)
}
