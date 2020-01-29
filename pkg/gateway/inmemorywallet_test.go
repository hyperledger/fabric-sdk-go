/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"reflect"
	"sort"
	"testing"
)

func TestNewInMemoryWallet(t *testing.T) {
	wallet := NewInMemoryWallet()
	if wallet == nil {
		t.Fatal("Failed to create in memory wallet")
	}
}

func TestInsertionAndExistance(t *testing.T) {
	wallet := NewInMemoryWallet()
	wallet.Put("label1", NewX509Identity("testCert", "testPrivKey"))
	exists := wallet.Exists("label1")
	if exists != true {
		t.Fatal("Expected label1 to be in wallet")
	}
}

func TestNonExistance(t *testing.T) {
	wallet := NewInMemoryWallet()
	exists := wallet.Exists("label1")
	if exists != false {
		t.Fatal("Expected label1 to not be in wallet")
	}
}

func TestLookupNonExist(t *testing.T) {
	wallet := NewInMemoryWallet()
	_, err := wallet.Get("label1")
	if err == nil {
		t.Fatal("Expected error for label1 not in wallet")
	}
}

func TestInsertionAndLookup(t *testing.T) {
	wallet := NewInMemoryWallet()
	wallet.Put("label1", NewX509Identity("testCert", "testPrivKey"))
	entry, err := wallet.Get("label1")
	if err != nil {
		t.Fatalf("Failed to lookup identity: %s", err)
	}
	if entry.GetType() != "X509" {
		t.Fatalf("Unexpected identity type: %s", entry.GetType())
	}
}

func TestContentsOfWallet(t *testing.T) {
	wallet := NewInMemoryWallet()
	contents := wallet.List()
	if len(contents) != 0 {
		t.Fatal("Wallet should be empty")
	}
	wallet.Put("label1", NewX509Identity("testCert", "testPrivKey"))
	wallet.Put("label2", NewX509Identity("testCert", "testPrivKey"))
	contents = wallet.List()
	sort.Strings(contents)
	expected := []string{"label1", "label2"}
	if !reflect.DeepEqual(contents, expected) {
		t.Fatalf("Unexpected wallet contents: %s", contents)
	}
}

func TestRemovalFromWallet(t *testing.T) {
	wallet := NewInMemoryWallet()
	contents := wallet.List()
	wallet.Put("label1", NewX509Identity("testCert1", "testPrivKey"))
	wallet.Put("label2", NewX509Identity("testCert2", "testPrivKey"))
	wallet.Put("label3", NewX509Identity("testCert3", "testPrivKey"))
	wallet.Remove("label2")
	contents = wallet.List()
	sort.Strings(contents)
	expected := []string{"label1", "label3"}
	if !reflect.DeepEqual(contents, expected) {
		t.Fatalf("Unexpected wallet contents: %s", contents)
	}
}

func TestRemoveNonExist(t *testing.T) {
	wallet := NewInMemoryWallet()
	err := wallet.Remove("label1")
	if err != nil {
		t.Fatal("Remove should not throw error for non-existant label")
	}
}
