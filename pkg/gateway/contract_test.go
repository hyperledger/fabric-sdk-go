/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"testing"
)

func TestCreateTransaction(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &Gateway{
		options: &gatewayOptions{},
	}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")

	txn, err := contr.CreateTransaction("txn1")

	if err != nil {
		t.Fatalf("Failed to create transaction: %s", err)
	}

	name := txn.request.Fcn
	if name != "txn1" {
		t.Fatalf("Incorrect transaction name: %s", name)
	}
}

func TestCreateTransactionNamespaced(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &Gateway{
		options: &gatewayOptions{},
	}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContractWithName("contract1", "class1")

	txn, err := contr.CreateTransaction("txn1")

	if err != nil {
		t.Fatalf("Failed to create transaction: %s", err)
	}

	name := txn.request.Fcn
	if name != "class1:txn1" {
		t.Fatalf("Incorrect transaction name: %s", name)
	}
}

func TestSubmitTransaction(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &Gateway{
		options: &gatewayOptions{
			Timeout: defaultTimeout,
		},
	}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")

	result, err := contr.SubmitTransaction("txn1", "arg1", "arg2")

	if err != nil {
		t.Fatalf("Failed to submit transaction: %s", err)
	}

	if string(result) != "abc" {
		t.Fatalf("Incorrect transaction result: %s", result)
	}
}

func TestEvaluateTransaction(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &Gateway{
		options: &gatewayOptions{
			Timeout: defaultTimeout,
		},
	}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")

	result, err := contr.EvaluateTransaction("txn1", "arg1", "arg2")

	if err != nil {
		t.Fatalf("Failed to evaluate transaction: %s", err)
	}

	if string(result) != "abc" {
		t.Fatalf("Incorrect transaction result: %s", result)
	}
}

func TestContractEvent(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &Gateway{
		options: &gatewayOptions{
			Timeout: defaultTimeout,
		},
	}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")

	eventID := "test([a-zA-Z]+)"

	reg, _, err := contr.RegisterEvent(eventID)
	if err != nil {
		t.Fatalf("Failed to register contract event: %s", err)
	}
	defer contr.Unregister(reg)

}
