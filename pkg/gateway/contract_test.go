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

	gw := &gateway{}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")

	txn, err := contr.CreateTransaction("txn1")

	if err != nil {
		t.Fatalf("Failed to create transaction: %s", err)
	}

	name := txn.(*transaction).name
	if name != "txn1" {
		t.Fatalf("Incorrect transaction name: %s", name)
	}
}

func TestSubmitTransaction(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &gateway{}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")

	result, err := contr.SubmitTransaction("txn1")

	if err != nil {
		t.Fatalf("Failed to submit transaction: %s", err)
	}

	if string(result) != "abc" {
		t.Fatalf("Incorrect transaction result: %s", result)
	}
}

func TestEvaluateTransaction(t *testing.T) {
	c := mockChannelProvider("mychannel")

	gw := &gateway{}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")

	result, err := contr.EvaluateTransaction("txn1")

	if err != nil {
		t.Fatalf("Failed to evaluate transaction: %s", err)
	}

	if string(result) != "abc" {
		t.Fatalf("Incorrect transaction result: %s", result)
	}
}
