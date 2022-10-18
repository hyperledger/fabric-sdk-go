/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

const dataFileExtension string = ".id"
const extensionLength = 3

// FileSystemWalletStore stores identity information used to connect to a Hyperledger Fabric network.
// Instances are created using NewFileSystemWallet()
type fileSystemWalletStore struct {
	path string
}

// NewFileSystemWallet creates an instance of a wallet, held in memory.
// This implementation is not backed by a persistent store.
//  Parameters:
//  path specifies where on the filesystem to store the wallet.
//
//  Returns:
//  A Wallet object.
func NewFileSystemWallet(path string) (*Wallet, error) {
	cleanPath := filepath.Clean(path)
	err := os.MkdirAll(cleanPath, os.ModePerm)

	if err != nil {
		return nil, err
	}

	store := &fileSystemWalletStore{cleanPath}
	return NewWalletWithStore(store), nil

}

// Put an identity into the wallet.
func (fsw *fileSystemWalletStore) Put(label string, content []byte) error {
	pathname := filepath.Join(fsw.path, label) + dataFileExtension

	f, err := os.OpenFile(filepath.Clean(pathname), os.O_RDWR|os.O_CREATE, 0600)

	if err != nil {
		return err
	}

	if _, err := f.Write(content); err != nil {
		_ = f.Close() // ignore error; Write error takes precedence
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

// Get an identity from the wallet.
func (fsw *fileSystemWalletStore) Get(label string) ([]byte, error) {
	pathname := filepath.Join(fsw.path, label) + dataFileExtension

	return ioutil.ReadFile(filepath.Clean(pathname))
}

// Remove an identity from the wallet. If the identity does not exist, this method does nothing.
func (fsw *fileSystemWalletStore) Remove(label string) error {
	pathname := filepath.Join(fsw.path, label) + dataFileExtension
	_ = os.Remove(filepath.Clean(pathname))
	return nil
}

// Exists tests the existence of an identity in the wallet.
func (fsw *fileSystemWalletStore) Exists(label string) bool {
	pathname := filepath.Join(fsw.path, label) + dataFileExtension

	_, err := os.Stat(filepath.Clean(pathname))
	return err == nil
}

// List all of the labels in the wallet.
func (fsw *fileSystemWalletStore) List() ([]string, error) {
	files, err := ioutil.ReadDir(fsw.path)

	if err != nil {
		return nil, err
	}

	var labels []string
	for _, file := range files {
		name := file.Name()
		if filepath.Ext(name) == dataFileExtension {
			labels = append(labels, name[:len(name)-extensionLength])
		}
	}

	return labels, nil
}
