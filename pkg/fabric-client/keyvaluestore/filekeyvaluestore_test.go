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
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/kvstore"
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
		return path.Join(storePath, fmt.Sprintf("mypath/%s/valuefile", keyString)), nil
	}
	testFKVS(t, keySerializer)
}

func testFKVS(t *testing.T, KeySerializer KeySerializer) {
	var store kvstore.KVStore
	var err error
	store, err = NewFileKeyValueStore(
		&FileKeyValueStoreOptions{
			Path:          storePath,
			KeySerializer: KeySerializer,
		})
	if err != nil {
		t.Fatalf("NewFileKeyValueStore failed [%s]", err)
	}
	if err := cleanup(storePath); err != nil {
		t.Fatalf("%s", err)
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
	if err := store.Store(key1, value1); err != nil {
		t.Fatalf("SetValue %s failed [%s]", key1, err)
	}
	if err := store.Store(key2, value2); err != nil {
		t.Fatalf("SetValue %s failed [%s]", key1, err)
	}

	// Check key1, value1
	if err := checkStoreValue(store, key1, value1); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", key1, err)
	}
	if err := store.Delete(key1); err != nil {
		t.Fatalf("Delete %s failed [%s]", key1, err)
	}
	if err := checkStoreValue(store, key1, nil); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", key1, err)
	}

	// Check ke2, value2
	if err := checkStoreValue(store, key2, value2); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", key2, err)
	}
	if err := store.Delete(key2); err != nil {
		t.Fatalf("Delete %s failed [%s]", key2, err)
	}
	if err := checkStoreValue(store, key2, nil); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", key2, err)
	}

	// Check non-existing key
	_, err = store.Load("non-existing")
	if err == nil || err != kvstore.ErrNotFound {
		t.Fatal("fetching value for non-existing key should return ErrNotFound")
	}

	// Check empty string value
	keyEmptyString := "empty-string"
	valueEmptyString := []byte("")
	err = store.Store(keyEmptyString, valueEmptyString)
	if err != nil {
		t.Fatal("setting an empty string value shouldn't fail")
	}
	if err := checkStoreValue(store, keyEmptyString, valueEmptyString); err != nil {
		t.Fatalf("checkStoreValue %s failed [%s]", keyEmptyString, err)
	}
}

func TestCreateNewFileKeyValueStore(t *testing.T) {

	_, err := NewFileKeyValueStore(
		&FileKeyValueStoreOptions{
			Path: "",
		})
	if err == nil || err.Error() != "FileKeyValueStore path is empty" {
		t.Fatal("File path validation on NewFileKeyValueStore is not working as expected")
	}

	_, err = NewFileKeyValueStore(nil)
	if err == nil || err.Error() != "FileKeyValueStoreOptions is nil" {
		t.Fatal("File path validation on NewFileKeyValueStore is not working as expected")
	}

	var store kvstore.KVStore
	store, err = NewFileKeyValueStore(
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

func checkStoreValue(store kvstore.KVStore, key interface{}, expected []byte) error {
	v, err := store.Load(key)
	if err != nil {
		if err == kvstore.ErrNotFound && expected == nil {
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
		_, err := os.Stat(file)
		if err == nil {
			return fmt.Errorf("path shouldn't exist [%s]", file)
		}
		if !os.IsNotExist(err) {
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
	if bytes.Compare(vbytes, expected) != 0 {
		return errors.New("value from store comparison failed")
	}
	return nil
}
