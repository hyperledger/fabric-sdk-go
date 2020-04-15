/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
)

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

// Set sets a backend value
func (b *MockConfigBackend) Set(key string, value interface{}) bool {
	if key != "" {
		node := b.KeyValueMap
		ss := strings.Split(key, ".")
		for i, s := range ss {
			if i == len(ss)-1 {
				node[s] = value
				return true
			}
			v, ok := node[s]
			if !ok {
				node[s] = map[string]interface{}{}
				node = node[s].(map[string]interface{})
			} else {
				node, ok = v.(map[string]interface{})
				if !ok {
					return false
				}
			}

		}
	}
	return false
}

// Get returns a value from the backend
func (b *MockConfigBackend) Get(key string) (interface{}, bool) {
	if key != "" {
		node := b.KeyValueMap
		ss := strings.Split(key, ".")
		for i, s := range ss {
			v, ok := node[s]
			if i == len(ss)-1 {
				return v, ok
			}
			if !ok {
				return nil, ok
			}
			node = v.(map[string]interface{})
		}
	}
	return nil, false
}

// BackendFromFile returns MockConfigBackend populated from file
func BackendFromFile(configPath string) (*MockConfigBackend, error) {
	b, err := config.FromFile(configPath)()
	if err != nil {
		return nil, err
	}
	configBackend := b[0]

	backendMap := make(map[string]interface{})
	backendMap["client"], _ = configBackend.Lookup("client")
	backendMap["certificateAuthorities"], _ = configBackend.Lookup("certificateAuthorities")
	backendMap["entityMatchers"], _ = configBackend.Lookup("entityMatchers")
	backendMap["peers"], _ = configBackend.Lookup("peers")
	backendMap["organizations"], _ = configBackend.Lookup("organizations")
	backendMap["orderers"], _ = configBackend.Lookup("orderers")
	backendMap["channels"], _ = configBackend.Lookup("channels")

	return &MockConfigBackend{KeyValueMap: backendMap}, nil
}
