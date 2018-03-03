/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package core

import "github.com/pkg/errors"

var (
	// ErrKeyValueNotFound indicates that a value for the key does not exist
	ErrKeyValueNotFound = errors.New("value for key not found")
)

// KVStore is a generic key-value store interface.
type KVStore interface {

	/**
	 * Store sets the value for the key.
	 */
	Store(key interface{}, value interface{}) error

	/**
	 * Load returns the value stored in the store for a key.
	 * If a value for the key was not found, returns (nil, ErrNotFound)
	 */
	Load(key interface{}) (interface{}, error)

	/**
	 * Delete deletes the value for a key.
	 */
	Delete(key interface{}) error
}
