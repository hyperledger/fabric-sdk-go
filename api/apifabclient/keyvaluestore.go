/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

// KeyValueStore ...
/**
 * Abstract class for a Key-Value store. The Chain class uses this store
 * to save sensitive information such as authenticated user's private keys,
 * certificates, etc.
 *
 */
type KeyValueStore interface {
	/**
	 * Get the value associated with name.
	 *
	 * @param {string} name of the key
	 * @returns {[]byte}
	 */
	Value(key string) ([]byte, error)

	/**
	 * Set the value associated with name.
	 * @param {string} name of the key to save
	 * @param {[]byte} value to save
	 */
	SetValue(key string, value []byte) error
}
