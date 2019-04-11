/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package keyvaluestore

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/pkg/errors"
)

var storePath = "/tmp/testkeyvaluestore"

func TestDefaultFKVS(t *testing.T) {
	testFKVS(t, nil)
}

func TestFKVSWithCustomKeySerializer(t *testing.T) {
	keySerializer := func(key interface{}) (string, error) {
		keyString, ok := key.(string)
		if !ok {
			return "", errors.New("converting key to string failed")
		}
		return filepath.Join(storePath, fmt.Sprintf("mypath/%s/valuefile", keyString)), nil
	}
	testFKVS(t, keySerializer)
}

func testFKVS(t *testing.T, KeySerializer KeySerializer) {
	var store core.KVStore
	var err error
	store, err = New(
		&FileKeyValueStoreOptions{
			Path:          storePath,
			KeySerializer: KeySerializer,
		})
	if err != nil {
		t.Fatalf("New failed [%s]", err)
	}
	if err1 := cleanup(storePath); err1 != nil {
		t.Fatalf("%s", err1)
	}
	defer cleanup(storePath)

	err = store.Store(nil, []byte("1234"))
	if err == nil || err.Error() != "key is nil" {
		t.Fatal("SetValue(nil, ...) should throw error")
	}
	err = store.Store("key", nil)
	if err == nil || err.Error() != "value is nil" {
		t.Fatal("Store(..., nil should throw error")
	}

	key1 := "key1"
	value1 := []byte("value1")
	key2 := "key2"
	value2 := []byte("value2")
	if err1 := store.Store(key1, value1); err1 != nil {
		t.Fatalf("SetValue %s failed [%s]", key1, err1)
	}
	if err1 := store.Store(key2, value2); err1 != nil {
		t.Fatalf("SetValue %s failed [%s]", key1, err1)
	}

	// Check key1, value1
	checkKeyValue(store, key1, value1, t)

	// Check ke2, value2
	checkKeyValue(store, key2, value2, t)

	// Check non-existing key
	checkNonExistingKey(store, t)

	// Check empty string value
	checkEmptyStringValue(store, t)
}

func checkKeyValue(store core.KVStore, key string, value []byte, t *testing.T) {
	if err := checkStoreValue(store, key, value); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", key, err)
	}
	if err := store.Delete(key); err != nil {
		t.Fatalf("Delete %s failed [%s]", key, err)
	}
	if err := checkStoreValue(store, key, nil); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", key, err)
	}
}

func checkNonExistingKey(store core.KVStore, t *testing.T) {
	_, err := store.Load("non-existing")
	if err == nil || err != core.ErrKeyValueNotFound {
		t.Fatal("fetching value for non-existing key should return ErrNotFound")
	}
}

func checkEmptyStringValue(store core.KVStore, t *testing.T) {
	keyEmptyString := "empty-string"
	valueEmptyString := []byte("")
	err := store.Store(keyEmptyString, valueEmptyString)
	if err != nil {
		t.Fatal("setting an empty string value shouldn't fail")
	}
	if err := checkStoreValue(store, keyEmptyString, valueEmptyString); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", keyEmptyString, err)
	}
}

func TestCreateNewFileKeyValueStore(t *testing.T) {

	_, err := New(
		&FileKeyValueStoreOptions{
			Path: "",
		})
	if err == nil || err.Error() != "FileKeyValueStore path is empty" {
		t.Fatal("File path validation on NewFileKeyValueStore is not working as expected")
	}

	_, err = New(nil)
	if err == nil || err.Error() != "FileKeyValueStoreOptions is nil" {
		t.Fatal("File path validation on NewFileKeyValueStore is not working as expected")
	}

	var store core.KVStore
	store, err = New(
		&FileKeyValueStoreOptions{
			Path: storePath,
		})
	if err != nil {
		t.Fatal("creating a store shouldn't fail")
	}
	if store == nil {
		t.Fatal("creating a store failed")
	}
}

func cleanup(storePath string) error {
	err := os.RemoveAll(storePath)
	if err != nil {
		return errors.Wrapf(err, "Cleaning up directory '%s' failed", storePath)
	}
	return nil
}

func checkStoreValue(store core.KVStore, key interface{}, expected []byte) error {
	v, err := store.Load(key)
	if err != nil {
		if err == core.ErrKeyValueNotFound && expected == nil {
			return nil
		}
		return err
	}
	if err = compare(v, expected); err != nil {
		return err
	}
	file, err := store.(*FileKeyValueStore).keySerializer(key)
	if err != nil {
		return err
	}
	if expected == nil {
		_, err1 := os.Stat(file)
		if err1 == nil {
			return fmt.Errorf("path shouldn't exist [%s]", file)
		}
		if !os.IsNotExist(err1) {
			return errors.Wrapf(err, "stat file failed [%s]", file)
		}
		// Doesn't exist, OK
		return nil
	}
	v, err = ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return compare(v, expected)
}

func compare(v interface{}, expected []byte) error {
	var vbytes []byte
	var ok bool
	if v == nil {
		vbytes = nil
	} else {
		vbytes, ok = v.([]byte)
		if !ok {
			return errors.New("value is not []byte")
		}
	}
	if !bytes.Equal(vbytes, expected) {
		return errors.New("value from store comparison failed")
	}
	return nil
}
