/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"

//MockConfigBackend mocks config backend for unit tests
type MockConfigBackend struct {
	//KeyValueMap map to override CustomBackend key-values.
	KeyValueMap map[string]interface{}
	//CustomBackend config backend
	CustomBackend core.ConfigBackend
}

//Lookup returns or unmarshals value for given key
func (b *MockConfigBackend) Lookup(key string) (interface{}, bool) {
	v, ok := b.KeyValueMap[key]
	//if not found in custom map then try with backend
	if !ok && b.CustomBackend != nil {
		return b.CustomBackend.Lookup(key)
	}
	return v, ok
}
