/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"testing"
)

func createInMemoryWallet() (*Wallet, error) {
	return NewInMemoryWallet(), nil
}

func TestNewInMemoryWallet(t *testing.T) {
	wallet := NewInMemoryWallet()
	if wallet == nil {
		t.Fatal("Failed to create in memory wallet")
	}
}

func TestInMemoryWalletSuite(t *testing.T) {
	testWalletSuite(t, createInMemoryWallet)
}
