/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// A Contract object represents a smart contract instance in a network.
// Applications should get a Contract instance from a Network using the GetContract method
type Contract struct {
	chaincodeID string
	name        string
	network     *Network
	client      *channel.Client
}

func newContract(network *Network, chaincodeID string, name string) *Contract {
	return &Contract{network: network, client: network.client, chaincodeID: chaincodeID, name: name}
}

// Name returns the name of the smart contract
func (c *Contract) Name() string {
	qualifiedName := c.chaincodeID
	if len(c.name) != 0 {
		qualifiedName += ":" + c.name
	}
	return qualifiedName
}

// EvaluateTransaction will evaluate a transaction function and return its results.
// The transaction function 'name'
// will be evaluated on the endorsing peers but the responses will not be sent to
// the ordering service and hence will not be committed to the ledger.
// This can be used for querying the world state.
//  Parameters:
//  name is the name of the transaction function to be invoked in the smart contract.
//  args are the arguments to be sent to the transaction function.
//
//  Returns:
//  The return value of the transaction function in the smart contract.
func (c *Contract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	txn, err := c.CreateTransaction(name)

	if err != nil {
		return nil, err
	}

	return txn.Evaluate(args...)
}

// SubmitTransaction will submit a transaction to the ledger. The transaction function 'name'
// will be evaluated on the endorsing peers and then submitted to the ordering service
// for committing to the ledger.
//  Parameters:
//  name is the name of the transaction function to be invoked in the smart contract.
//  args are the arguments to be sent to the transaction function.
//
//  Returns:
//  The return value of the transaction function in the smart contract.
func (c *Contract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	txn, err := c.CreateTransaction(name)

	if err != nil {
		return nil, err
	}

	return txn.Submit(args...)
}

// CreateTransaction creates an object representing a specific invocation of a transaction
// function implemented by this contract, and provides more control over
// the transaction invocation using the optional arguments. A new transaction object must
// be created for each transaction invocation.
//  Parameters:
//  name is the name of the transaction function to be invoked in the smart contract.
//  opts are the options to be associated with the transaction.
//
//  Returns:
//  A Transaction object for subsequent evaluation or submission.
func (c *Contract) CreateTransaction(name string, opts ...TransactionOption) (*Transaction, error) {
	return newTransaction(name, c, opts...)
}

// RegisterEvent registers for chaincode events. Unregister must be called when the registration is no longer needed.
//  Parameters:
//  eventFilter is the chaincode event filter (regular expression) for which events are to be received
//
//  Returns:
//  the registration and a channel that is used to receive events. The channel is closed when Unregister is called.
func (c *Contract) RegisterEvent(eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	return c.network.event.RegisterChaincodeEvent(c.chaincodeID, eventFilter)
}

// Unregister removes the given registration and closes the event channel.
//  Parameters:
//  registration is the registration handle that was returned from RegisterContractEvent method
func (c *Contract) Unregister(registration fab.Registration) {
	c.network.event.Unregister(registration)
}
