/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package keyvaluestore

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/pkg/errors"
)

const (
	newDirMode  = 0700
	newFileMode = 0600
)

// KeySerializer converts a key to a unique fila path
type KeySerializer func(key interface{}) (string, error)

// Marshaller marshals a value into a byte array
type Marshaller func(value interface{}) ([]byte, error)

// Unmarshaller unmarshals a value from a byte array
type Unmarshaller func(value []byte) (interface{}, error)

// FileKeyValueStore stores each value into a separate file.
// KeySerializer maps a key to a unique file path (raletive to the store path)
// ValueSerializer and ValueDeserializer serializes/de-serializes a value
// to and from a byte array that is stored in the path derived from the key.
type FileKeyValueStore struct {
	path          string
	keySerializer KeySerializer
	marshaller    Marshaller
	unmarshaller  Unmarshaller
}

// FileKeyValueStoreOptions allow overriding store defaults
type FileKeyValueStoreOptions struct {
	// Store path, mandatory
	Path string
	// Optional. If not provided, default key serializer is used.
	KeySerializer KeySerializer
	// Optional. If not provided, default Marshaller is used.
	Marshaller Marshaller
	// Optional. If not provided, default Unmarshaller is used.
	Unmarshaller Unmarshaller
}

// Default Marshaller
func defaultMarshaller(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	valueBytes, ok := value.([]byte)
	if !ok {
		return nil, errors.New("converting value to byte array failed")
	}
	return valueBytes, nil
}

// Default Unmarshaller
func defaultUnmarshaller(value []byte) (interface{}, error) {
	return value, nil
}

// GetPath returns the store path
func (fkvs *FileKeyValueStore) GetPath() string {
	return fkvs.path
}

// New creates a new instance of FileKeyValueStore using provided options
func New(opts *FileKeyValueStoreOptions) (*FileKeyValueStore, error) {
	if opts == nil {
		return nil, errors.New("FileKeyValueStoreOptions is nil")
	}
	if opts.Path == "" {
		return nil, errors.New("FileKeyValueStore path is empty")
	}
	if opts.KeySerializer == nil {
		// Default key serializer
		opts.KeySerializer = func(key interface{}) (string, error) {
			keyString, ok := key.(string)
			if !ok {
				return "", errors.New("converting key to string failed")
			}
			return filepath.Join(opts.Path, keyString), nil
		}
	}
	if opts.Marshaller == nil {
		opts.Marshaller = defaultMarshaller
	}
	if opts.Unmarshaller == nil {
		opts.Unmarshaller = defaultUnmarshaller
	}
	return &FileKeyValueStore{
		path:          opts.Path,
		keySerializer: opts.KeySerializer,
		marshaller:    opts.Marshaller,
		unmarshaller:  opts.Unmarshaller,
	}, nil
}

// Load returns the value stored in the store for a key.
// If a value for the key was not found, returns (nil, ErrNotFound)
func (fkvs *FileKeyValueStore) Load(key interface{}) (interface{}, error) {
	file, err := fkvs.keySerializer(key)
	if err != nil {
		return nil, err
	}
	if _, err1 := os.Stat(file); os.IsNotExist(err1) {
		return nil, core.ErrKeyValueNotFound
	}
	bytes, err := ioutil.ReadFile(file) // nolint: gas
	if err != nil {
		return nil, err
	}
	if bytes == nil {
		return nil, core.ErrKeyValueNotFound
	}
	return fkvs.unmarshaller(bytes)
}

// Store sets the value for the key.
func (fkvs *FileKeyValueStore) Store(key interface{}, value interface{}) error {
	if key == nil {
		return errors.New("key is nil")
	}
	if value == nil {
		return errors.New("value is nil")
	}
	file, err := fkvs.keySerializer(key)
	if err != nil {
		return err
	}
	valueBytes, err := fkvs.marshaller(value)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(file), newDirMode)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, valueBytes, newFileMode)
}

// Delete deletes the value for a key.
func (fkvs *FileKeyValueStore) Delete(key interface{}) error {
	if key == nil {
		return errors.New("key is nil")
	}
	file, err := fkvs.keySerializer(key)
	if err != nil {
		return err
	}
	_, err = os.Stat(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "stat dir failed")
		}
		// Doesn't exist, OK
		return nil
	}
	return os.Remove(file)
}
