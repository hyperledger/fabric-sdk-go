/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package keyvaluestore

import (
	"testing"
)

func TestFKVSMethods(t *testing.T) {
	stateStore, err := CreateNewFileKeyValueStore("/tmp/keyvaluestore")
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	stateStore.SetValue("testvalue", []byte("data"))
	value, err := stateStore.Value("testvalue")
	if err != nil {
		t.Fatalf("stateStore.GetValue return error[%s]", err)
	}
	if string(value) != "data" {
		t.Fatalf("stateStore.GetValue didn't return the right value")
	}
}

func TestFKVSMethodsForFailures(t *testing.T) {

	stateStore, err := CreateNewFileKeyValueStore("")

	if err == nil || err.Error() != "FileKeyValueStore path is empty" {
		t.Fatal("File path validation on CreateNewFileKeyValueStore is not working as expected")
	}

	stateStore, err = CreateNewFileKeyValueStore("/tmp/keyvaluestore")

	_, err = stateStore.Value("invalid")
	if err == nil {
		t.Fatal(" fetching value was supposed to fail")
	}

	err = stateStore.SetValue("testvalue.json//C;", []byte(""))
	if err == nil {
		t.Fatal(" setting value was supposed to fail")
	}

}
