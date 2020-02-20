/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"os"
	"path/filepath"
	"testing"
)

func createFileSystemWallet() (*Wallet, error) {
	dir := filepath.Join("testdata", "wallet", "unit")
	os.RemoveAll(dir)
	return NewFileSystemWallet(dir)
}

func TestFileSystemWalletSuite(t *testing.T) {
	testWalletSuite(t, createFileSystemWallet)
	os.RemoveAll(filepath.Join("testdata", "wallet", "unit"))
}

func TestFormatCompatibility(t *testing.T) {
	dir := filepath.Join("testdata", "wallet")
	wallet, err := NewFileSystemWallet(dir)
	if err != nil {
		t.Fatalf("Failed to create FileSystemWallet: %s", err)
	}

	id, err := wallet.Get("x509-v1")
	if err != nil {
		t.Fatalf("Failed to get identity from FileSystemWallet: %s", err)
	}

	x509 := id.(*X509Identity)

	if x509.mspID() != "mspId" {
		t.Fatalf("Incorrect MspID: %s", x509.MspID)
	}

	if x509.idType() != x509Type {
		t.Fatalf("Incorrect IDType: %s", x509.idType())
	}

	if x509.Version != 1 {
		t.Fatalf("Incorrect version: %d", x509.Version)
	}
}

func TestNonJSONFormat(t *testing.T) {
	dir := filepath.Join("testdata", "wallet")
	wallet, err := NewFileSystemWallet(dir)
	if err != nil {
		t.Fatalf("Failed to create FileSystemWallet: %s", err)
	}

	_, err = wallet.Get("invalid1")
	if err == nil {
		t.Fatal("Expected error to be thrown")
	}
}

func TestInvalidJSONFormat(t *testing.T) {
	dir := filepath.Join("testdata", "wallet")
	wallet, err := NewFileSystemWallet(dir)
	if err != nil {
		t.Fatalf("Failed to create FileSystemWallet: %s", err)
	}

	_, err = wallet.Get("invalid2")
	if err == nil {
		t.Fatal("Expected error to be thrown")
	}
}

func TestMissingTypeFormat(t *testing.T) {
	dir := filepath.Join("testdata", "wallet")
	wallet, err := NewFileSystemWallet(dir)
	if err != nil {
		t.Fatalf("Failed to create FileSystemWallet: %s", err)
	}

	_, err = wallet.Get("invalid3")
	if err == nil {
		t.Fatal("Expected error to be thrown")
	}
}

func TestInvalidTypeFormat(t *testing.T) {
	dir := filepath.Join("testdata", "wallet")
	wallet, err := NewFileSystemWallet(dir)
	if err != nil {
		t.Fatalf("Failed to create FileSystemWallet: %s", err)
	}

	_, err = wallet.Get("invalid4")
	if err == nil {
		t.Fatal("Expected error to be thrown")
	}
}
