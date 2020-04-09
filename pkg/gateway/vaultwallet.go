/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"encoding/json"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"path/filepath"
)

// VaultWalletStore stores identity information used to connect to a Hyperledger Fabric network.
// Instances are created using NewVaultWallet()
type VaultWalletStore struct {
	path   string
	client *api.Logical
}

// NewVaultWallet creates an instance of a wallet,  backed by key/values in Vault
func NewVaultWallet(path, token string, vaultConfig *api.Config) (*Wallet, error) {

	if path == "" {
		return nil, errors.New("VaultConfig is empty")
	}
	if token == "" {
		return nil, errors.New("token is empty")
	}
	if vaultConfig == nil {
		vaultConfig = &api.Config{Address: "http://localhost:8200"}
	}

	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, errors.Wrap(err, "can't create Vault client")
	}

	client.SetToken(token)
	logicalBackend := client.Logical()

	store := &VaultWalletStore{path, logicalBackend}

	return &Wallet{store}, nil

}

// Put an identity into the wallet.
func (fsw *VaultWalletStore) Put(label string, content []byte) error {
	pathname := filepath.Join("secret/data", fsw.path, label)
	_, err := fsw.client.WriteBytes(pathname, content)
	return err
}

// Get an identity from the wallet.
func (fsw *VaultWalletStore) Get(label string) ([]byte, error) {
	pathname := filepath.Join("secret/data", fsw.path, label)
	data, err := fsw.client.Read(pathname)
	if err != nil {
		return nil, errors.Wrap(err, "Can't to read value from Vault")
	}
	if data == nil {
		return nil, nil
	}

	serializedIdentity, err := json.Marshal(data.Data)
	if err != nil {
		return nil, errors.Wrap(err, "Can't to serialize identity")
	}

	return serializedIdentity, err
}

// Remove an identity from the wallet. If the identity does not exist, this method does nothing.
func (fsw *VaultWalletStore) Remove(label string) error {
	pathname := filepath.Join("secret/data", fsw.path, label)
	data, err := fsw.client.Read(pathname)
	if err != nil {
		return errors.Wrap(err, "can't to read value from Vault")
	}
	if data == nil {
		return nil
	}

	if _, err = fsw.client.Delete(pathname); err != nil {
		return errors.Wrap(err, "can't to delete value from Vault")
	}

	return nil
}

// Exists tests the existence of an identity in the wallet.
func (fsw *VaultWalletStore) Exists(label string) bool {
	pathname := filepath.Join("secret/data", fsw.path, label)
	data, err := fsw.client.Read(pathname)
	if err != nil {
		return false
	}
	if data == nil {
		return false
	}
	return true
}

// List all of the labels in the wallet.
func (fsw *VaultWalletStore) List() ([]string, error) {
	pathname := filepath.Join("secret/data", fsw.path)
	data, err := fsw.client.List(pathname)
	if err != nil {
		return nil, errors.Wrap(err, "can't to read value from Vault")
	}
	if data == nil {
		return nil, errors.New("wallet is empty")
	}
	var responseArray []string

	keys, ok := data.Data["keys"].([]interface{})
	if !ok {
		return nil, errors.New("can't to cast empty interfaces array from Vault to strings array")
	}

	for _, keyvalue := range keys {
		keyvalueStr, ok := keyvalue.(string)
		if !ok {
			return nil, errors.New("can't to cast value from Vault to string")
		}
		responseArray = append(responseArray, keyvalueStr)
	}
	return responseArray, nil
}