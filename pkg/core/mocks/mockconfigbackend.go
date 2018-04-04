/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/mitchellh/mapstructure"
)

//MockConfigBackend mocks config backend for unit tests
type MockConfigBackend struct {
	KeyValueMap map[string]interface{}
}

//Lookup returns or unmarshals value for given key
func (b *MockConfigBackend) Lookup(key string, opts ...core.LookupOption) (interface{}, bool) {
	if len(opts) > 0 {
		lookupOpts := &core.LookupOpts{}
		for _, option := range opts {
			option(lookupOpts)
		}

		if lookupOpts.UnmarshalType != nil {
			v, ok := b.KeyValueMap[key]
			if !ok {
				return nil, false
			}

			decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
				Result:     lookupOpts.UnmarshalType,
			})
			if err != nil {
				return nil, false
			}
			err = decoder.Decode(v)
			if err != nil {
				return nil, false
			}
			return lookupOpts.UnmarshalType, true
		}
	}
	v, ok := b.KeyValueMap[key]
	return v, ok
}
