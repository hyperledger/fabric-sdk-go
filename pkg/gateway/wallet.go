/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type wallet interface {
	Put(label string, id Identity) error
	Get(label string) (Identity, error)
	Remove(label string) error
	Exists(label string) bool
	List() ([]string, error)
}

// A Wallet stores identity information used to connect to a Hyperledger Fabric network.
// Instances are created using factory methods on the implementing objects.
type Wallet struct {
	store WalletStore
}

// Put an identity into the wallet
func (w *Wallet) Put(label string, id Identity) error {
	content, err := id.toJSON()
	if err != nil {
		return err
	}

	return w.store.Put(label, content)
}

// Get an identity from the wallet. The implementation class of the identity object will vary depending on its type.
func (w *Wallet) Get(label string) (Identity, error) {
	content, err := w.store.Get(label)

	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, errors.Wrap(err, "Invalid identity format")
	}

	idType, ok := data["type"].(string)

	if !ok {
		return nil, errors.New("Invalid identity format: missing type property")
	}

	var id Identity

	switch idType {
	case x509Type:
		id = &X509Identity{}
	default:
		return nil, errors.New("Invalid identity format: unsupported identity type: " + idType)
	}

	return id.fromJSON(content)
}

// List returns the labels of all identities in the wallet.
func (w *Wallet) List() ([]string, error) {
	return w.store.List()
}

// Exists tests whether the wallet contains an identity for the given label.
func (w *Wallet) Exists(label string) bool {
	return w.store.Exists(label)
}

// Remove an identity from the wallet. If the identity does not exist, this method does nothing.
func (w *Wallet) Remove(label string) error {
	return w.store.Remove(label)
}
