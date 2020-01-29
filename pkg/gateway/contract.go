/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import "github.com/hyperledger/fabric-sdk-go/pkg/client/channel"

type contract struct {
	chaincodeID string
	name        string
	network     *network
	client      *channel.Client
}

func newContract(network *network, chaincodeID string, name string) *contract {
	return &contract{network: network, client: network.client, chaincodeID: chaincodeID, name: name}
}

func (c *contract) GetName() string {
	return c.chaincodeID
}

func (c *contract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	txn, err := c.CreateTransaction(name)

	if err != nil {
		return nil, err
	}

	return txn.Evaluate(args...)
}

func (c *contract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	txn, err := c.CreateTransaction(name)

	if err != nil {
		return nil, err
	}

	return txn.Submit(args...)
}

func (c *contract) CreateTransaction(name string, args ...TransactionOption) (Transaction, error) {
	return newTransaction(name, c, args...)
}
