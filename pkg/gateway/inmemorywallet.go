/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import "errors"

// InMemoryWalletStore stores identity information used to connect to a Hyperledger Fabric network.
// Instances are created using NewInMemoryWallet()
type inMemoryWalletStore struct {
	storage map[string][]byte
}

// NewInMemoryWallet creates an instance of a wallet, held in memory.
//
//  Returns:
//  A Wallet object.
func NewInMemoryWallet() *Wallet {
	store := &inMemoryWalletStore{make(map[string][]byte, 10)}
	return &Wallet{store}
}

// Put an identity into the wallet.
func (f *inMemoryWalletStore) Put(label string, content []byte) error {
	f.storage[label] = content
	return nil
}

// Get an identity from the wallet.
func (f *inMemoryWalletStore) Get(label string) ([]byte, error) {
	if content, ok := f.storage[label]; ok {
		return content, nil
	}
	return nil, errors.New("label doesn't exist: " + label)
}

// Remove an identity from the wallet. If the identity does not exist, this method does nothing.
func (f *inMemoryWalletStore) Remove(label string) error {
	if _, ok := f.storage[label]; ok {
		delete(f.storage, label)
		return nil
	}
	return nil // what should we do here ?
}

// Exists returns true if the identity is in the wallet.
func (f *inMemoryWalletStore) Exists(label string) bool {
	_, ok := f.storage[label]
	return ok
}

// List all of the labels in the wallet.
func (f *inMemoryWalletStore) List() ([]string, error) {
	labels := make([]string, 0, len(f.storage))
	for label := range f.storage {
		labels = append(labels, label)
	}
	return labels, nil
}
