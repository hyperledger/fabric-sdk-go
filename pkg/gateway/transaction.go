/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

// A Transaction represents a specific invocation of a transaction function, and provides
// flexibility over how that transaction is invoked. Applications should
// obtain instances of this class from a Contract using the
// Contract.CreateTransaction method.
//
// Instances of this class are stateful. A new instance <strong>must</strong>
// be created for each transaction invocation.
type Transaction struct {
	name           string
	contract       *Contract
	request        *channel.Request
	endorsingPeers []string
}

// TransactionOption functional arguments can be supplied when creating a transaction object
type TransactionOption = func(*Transaction) error

func newTransaction(name string, contract *Contract, options ...TransactionOption) (*Transaction, error) {
	txn := &Transaction{
		name:     name,
		contract: contract,
		request:  &channel.Request{ChaincodeID: contract.chaincodeID, Fcn: name},
	}

	for _, option := range options {
		err := option(txn)
		if err != nil {
			return nil, err
		}
	}

	return txn, nil
}

// WithTransient is an optional argument to the CreateTransaction method which
// sets the transient data that will be passed to the transaction function
// but will not be stored on the ledger. This can be used to pass
// private data to a transaction function.
func WithTransient(data map[string][]byte) TransactionOption {
	return func(txn *Transaction) error {
		txn.request.TransientMap = data
		return nil
	}
}

// WithEndorsingPeers is an optional argument to the CreateTransaction method which
// sets the peers that should be used for endorsement of transaction submitted to the ledger using Submit()
func WithEndorsingPeers(peers ...string) TransactionOption {
	return func(txn *Transaction) error {
		txn.endorsingPeers = peers
		return nil
	}
}

// Evaluate a transaction function and return its results.
// The transaction function will be evaluated on the endorsing peers but
// the responses will not be sent to the ordering service and hence will
// not be committed to the ledger. This can be used for querying the world state.
func (txn *Transaction) Evaluate(args ...string) ([]byte, error) {
	bytes := make([][]byte, len(args))
	for i, v := range args {
		bytes[i] = []byte(v)
	}
	txn.request.Args = bytes

	var options []channel.RequestOption
	options = append(options, channel.WithTimeout(fab.Query, txn.contract.network.gateway.options.Timeout))

	response, err := txn.contract.client.Query(
		*txn.request,
		options...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to evaluate")
	}

	return response.Payload, nil
}

// Submit a transaction to the ledger. The transaction function represented by this object
// will be evaluated on the endorsing peers and then submitted to the ordering service
// for committing to the ledger.
func (txn *Transaction) Submit(args ...string) ([]byte, error) {
	bytes := make([][]byte, len(args))
	for i, v := range args {
		bytes[i] = []byte(v)
	}
	txn.request.Args = bytes

	var options []channel.RequestOption
	if txn.endorsingPeers != nil {
		options = append(options, channel.WithTargetEndpoints(txn.endorsingPeers...))
	}
	options = append(options, channel.WithTimeout(fab.Execute, txn.contract.network.gateway.options.Timeout))

	response, err := txn.contract.client.Execute(
		*txn.request,
		options...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to submit")
	}

	return response.Payload, nil
}
