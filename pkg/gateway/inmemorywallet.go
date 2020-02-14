/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import "errors"

// InMemoryWallet stores identity information used to connect to a Hyperledger Fabric network.
// Instances are created using NewInMemoryWallet()
type InMemoryWallet struct {
	idhandler IDHandler
	storage   map[string]map[string]string
}

// NewInMemoryWallet creates an instance of a wallet, backed by files on the filesystem
func NewInMemoryWallet() *InMemoryWallet {
	return &InMemoryWallet{newX509IdentityHandler(), make(map[string]map[string]string, 10)}
}

// Put an identity into the wallet.
func (f *InMemoryWallet) Put(label string, id IdentityType) error {
	elements := f.idhandler.GetElements(id)
	f.storage[label] = elements
	return nil
}

// Get an identity from the wallet.
func (f *InMemoryWallet) Get(label string) (IdentityType, error) {
	if elements, ok := f.storage[label]; ok {
		return f.idhandler.FromElements(elements), nil
	}
	return nil, errors.New("label doesn't exist: " + label)
}

// Remove an identity from the wallet. If the identity does not exist, this method does nothing.
func (f *InMemoryWallet) Remove(label string) error {
	if _, ok := f.storage[label]; ok {
		delete(f.storage, label)
		return nil
	}
	return nil // what should we do here ?
}

// Exists returns true if the identity is in the wallet.
func (f *InMemoryWallet) Exists(label string) bool {
	_, ok := f.storage[label]
	return ok
}

// List all of the labels in the wallet.
func (f *InMemoryWallet) List() []string {
	labels := make([]string, 0, len(f.storage))
	for label := range f.storage {
		labels = append(labels, label)
	}
	return labels
}
