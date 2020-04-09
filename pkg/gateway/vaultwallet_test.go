/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"testing"
)

const (
	vaultPath = "test"
	token     = "" // Vault token
)

func createVaultWallet() (*Wallet, error) {
	return NewVaultWallet(vaultPath, token, nil)
}

func TestVaultWalletSuite(t *testing.T) {
	testWalletSuite(t, createVaultWallet)

	// prune all
	wallet, err := NewVaultWallet(vaultPath, token, nil)
	if err != nil {
		t.Errorf("Pruning error: %s", err)
	}
	wallet.Remove("label1")
	wallet.Remove("label2")
	wallet.Remove("label3")
}
