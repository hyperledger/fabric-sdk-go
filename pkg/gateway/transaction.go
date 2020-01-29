/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
)

type transaction struct {
	name           string
	contract       *contract
	request        *channel.Request
	endorsingPeers []string
}

func newTransaction(name string, contract *contract, options ...TransactionOption) (Transaction, error) {
	txn := &transaction{
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
	return func(txn *transaction) error {
		txn.request.TransientMap = data
		return nil
	}
}

// WithEndorsingPeers is an optional argument to the CreateTransaction method which
// sets the peers that should be used for endorsement of transaction submitted to the ledger using Submit()
func WithEndorsingPeers(peers ...string) TransactionOption {
	return func(txn *transaction) error {
		txn.endorsingPeers = peers
		return nil
	}
}

func (txn *transaction) Evaluate(args ...string) ([]byte, error) {
	bytes := make([][]byte, len(args))
	for i, v := range args {
		bytes[i] = []byte(v)
	}
	txn.request.Args = bytes

	response, err := txn.contract.client.Query(*txn.request, channel.WithTargets(txn.contract.network.peers[0]))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to evaluate")
	}

	return response.Payload, nil
}

func (txn *transaction) Submit(args ...string) ([]byte, error) {
	bytes := make([][]byte, len(args))
	for i, v := range args {
		bytes[i] = []byte(v)
	}
	txn.request.Args = bytes

	response, err := txn.contract.client.Execute(*txn.request)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to submit")
	}

	return response.Payload, nil
}
