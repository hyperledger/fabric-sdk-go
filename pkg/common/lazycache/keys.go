/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lazycache

// StringKey is a simple string cache key
type StringKey struct {
	key string
}

// NewStringKey returns a new StringKey
func NewStringKey(key string) *StringKey {
	return &StringKey{key: key}
}

// String returns the key as a string
func (k *StringKey) String() string {
	return k.key
}
