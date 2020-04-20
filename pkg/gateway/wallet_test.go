/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"reflect"
	"sort"
	"testing"

	"github.com/pkg/errors"
)

type walletGenerator = func() (*Wallet, error)

func testWalletSuite(t *testing.T, gen walletGenerator) {
	tests := []struct {
		title string
		run   func(t *testing.T, wallet *Wallet)
	}{
		{"testInsertionAndExistance", testInsertionAndExistance},
		{"testNonExistance", testNonExistance},
		{"testLookupNonExist", testLookupNonExist},
		{"testInsertionAndLookup", testInsertionAndLookup},
		{"testContentsOfWallet", testContentsOfWallet},
		{"testRemovalFromWallet", testRemovalFromWallet},
		{"testRemoveNonExist", testRemoveNonExist},
		{"testPutInvalidID", testPutInvalidID},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			wallet, err := gen()
			if err != nil {
				t.Fatalf("Failed to create the wallet instance: %s", err)
			}
			test.run(t, wallet)
		})
	}
}

func testInsertionAndExistance(t *testing.T, wallet *Wallet) {
	wallet.Put("label1", NewX509Identity("msp", "testCert", "testPrivKey"))
	exists := wallet.Exists("label1")
	if exists != true {
		t.Fatal("Expected label1 to be in wallet")
	}
}

func testNonExistance(t *testing.T, wallet *Wallet) {
	exists := wallet.Exists("label1")
	if exists != false {
		t.Fatal("Expected label1 to not be in wallet")
	}
}

func testLookupNonExist(t *testing.T, wallet *Wallet) {
	_, err := wallet.Get("label1")
	if err == nil {
		t.Fatal("Expected error for label1 not in wallet")
	}
}

func testInsertionAndLookup(t *testing.T, wallet *Wallet) {
	wallet.Put("label1", NewX509Identity("msp", "testCert", "testPrivKey"))
	entry, err := wallet.Get("label1")
	if err != nil {
		t.Fatalf("Failed to lookup identity: %s", err)
	}
	if entry.idType() != x509Type {
		t.Fatalf("Unexpected identity type: %s", entry.idType())
	}
}

func testContentsOfWallet(t *testing.T, wallet *Wallet) {
	contents, _ := wallet.List()
	if len(contents) != 0 {
		t.Fatal("Wallet should be empty")
	}
	wallet.Put("label1", NewX509Identity("msp", "testCert", "testPrivKey"))
	wallet.Put("label2", NewX509Identity("msp", "testCert", "testPrivKey"))
	contents, _ = wallet.List()
	sort.Strings(contents)
	expected := []string{"label1", "label2"}
	if !reflect.DeepEqual(contents, expected) {
		t.Fatalf("Unexpected wallet contents: %s", contents)
	}
}

func testRemovalFromWallet(t *testing.T, wallet *Wallet) {
	contents, _ := wallet.List()
	wallet.Put("label1", NewX509Identity("msp", "testCert1", "testPrivKey"))
	wallet.Put("label2", NewX509Identity("msp", "testCert2", "testPrivKey"))
	wallet.Put("label3", NewX509Identity("msp", "testCert3", "testPrivKey"))
	wallet.Remove("label2")
	contents, _ = wallet.List()
	sort.Strings(contents)
	expected := []string{"label1", "label3"}
	if !reflect.DeepEqual(contents, expected) {
		t.Fatalf("Unexpected wallet contents: %s", contents)
	}
}

func testRemoveNonExist(t *testing.T, wallet *Wallet) {
	err := wallet.Remove("label1")
	if err != nil {
		t.Fatal("Remove should not throw error for non-existant label")
	}
}

func testPutInvalidID(t *testing.T, wallet *Wallet) {
	err := wallet.Put("label4", &badIdentity{})
	if err == nil {
		t.Fatal("Put should throw error for bad identity")
	}
}

func TestGetFromCorruptWallet(t *testing.T) {
	wallet := &Wallet{&corruptWallet{}}
	_, err := wallet.Get("user")
	if err == nil {
		t.Fatalf("Get should throw error for corrupt entry")
	}
}

type badIdentity struct{}

func (id *badIdentity) idType() string {
	return "bad"
}

func (id *badIdentity) mspID() string {
	return "mspid"
}

func (id *badIdentity) toJSON() ([]byte, error) {
	return nil, errors.New("toJSON error")
}

func (id *badIdentity) fromJSON(data []byte) (Identity, error) {
	return nil, errors.New("fromJSON error")
}

type corruptWallet struct{}

func (cw *corruptWallet) Put(label string, stream []byte) error {
	return nil
}

func (cw *corruptWallet) Get(label string) ([]byte, error) {
	return []byte("{\"type\":\"X.509\",\"credentials\":\"corrupt\"}"), nil
}

func (cw *corruptWallet) List() ([]string, error) {
	return nil, nil
}

func (cw *corruptWallet) Exists(label string) bool {
	return false
}

func (cw *corruptWallet) Remove(label string) error {
	return nil
}
