/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"testing"
)

func TestTransactionOptions(t *testing.T) {
	transient := make(map[string][]byte)
	transient["price"] = []byte("8500")
	
	c := mockChannelProvider("mychannel")

	gw := &Gateway{}

	nw, err := newNetwork(gw, c)

	if err != nil {
		t.Fatalf("Failed to create network: %s", err)
	}

	contr := nw.GetContract("contract1")

	txn, err := contr.CreateTransaction(
		"txn1", 
		WithTransient(transient),
		WithEndorsingPeers("peer1"),
	)

	if err != nil {
		t.Fatalf("Failed to create transaction: %s", err)
	}

	data := txn.request.TransientMap["price"]
	if string(data) != "8500" {
		t.Fatalf("Incorrect transient data: %s", string(data))
	}

	endorsers := txn.endorsingPeers
	if endorsers[0] != "peer1" {
		t.Fatalf("Incorrect endorsing peer: %s", endorsers[0])
	}
}
