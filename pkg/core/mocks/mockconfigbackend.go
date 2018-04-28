/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

//MockConfigBackend mocks config backend for unit tests
type MockConfigBackend struct {
	//KeyValueMap map to override CustomBackend key-values.
	KeyValueMap map[string]interface{}
}

//Lookup returns or unmarshals value for given key
func (b *MockConfigBackend) Lookup(key string) (interface{}, bool) {
	v, ok := b.KeyValueMap[key]
	return v, ok
}
