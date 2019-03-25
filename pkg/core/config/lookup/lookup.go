/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lookup

import (
	"time"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
)

//New providers lookup wrapper around given backend
func New(coreBackends ...core.ConfigBackend) *ConfigLookup {
	return &ConfigLookup{backends: coreBackends}
}

//unmarshalOpts opts for unmarshal key function
type unmarshalOpts struct {
	hooks []mapstructure.DecodeHookFunc
}

// UnmarshalOption describes a functional parameter unmarshaling
type UnmarshalOption func(o *unmarshalOpts)

// WithUnmarshalHookFunction provides an option to pass Custom Decode Hook Func
// for unmarshaling
func WithUnmarshalHookFunction(hookFunction mapstructure.DecodeHookFunc) UnmarshalOption {
	return func(o *unmarshalOpts) {
		o.hooks = append(o.hooks, hookFunction)
	}
}

//ConfigLookup is wrapper for core.ConfigBackend which performs key lookup and unmarshalling
type ConfigLookup struct {
	backends []core.ConfigBackend
}

//Lookup returns value for given key
func (c *ConfigLookup) Lookup(key string) (interface{}, bool) {
	//loop through each backend to find the value by key, fallback to next one if not found
	for _, backend := range c.backends {
		if backend == nil {
			continue
		}
		val, ok := backend.Lookup(key)
		if ok {
			return val, true
		}
	}
	return nil, false
}

//GetBool returns bool value for given key
func (c *ConfigLookup) GetBool(key string) bool {
	value, ok := c.Lookup(key)
	if !ok {
		return false
	}
	return cast.ToBool(value)
}

//GetString returns string value for given key
func (c *ConfigLookup) GetString(key string) string {
	value, ok := c.Lookup(key)
	if !ok {
		return ""
	}
	return cast.ToString(value)
}

//GetLowerString returns lower case string value for given key
func (c *ConfigLookup) GetLowerString(key string) string {
	value, ok := c.Lookup(key)
	if !ok {
		return ""
	}
	return strings.ToLower(cast.ToString(value))
}

//GetInt returns int value for given key
func (c *ConfigLookup) GetInt(key string) int {
	value, ok := c.Lookup(key)
	if !ok {
		return 0
	}
	return cast.ToInt(value)
}

//GetDuration returns time.Duration value for given key
func (c *ConfigLookup) GetDuration(key string) time.Duration {
	value, ok := c.Lookup(key)
	if !ok {
		return 0
	}
	return cast.ToDuration(value)
}

//UnmarshalKey unmarshals value for given key to rawval type
func (c *ConfigLookup) UnmarshalKey(key string, rawVal interface{}, opts ...UnmarshalOption) error {
	value, ok := c.Lookup(key)
	if !ok {
		return nil
	}

	//mandatory hook func
	var unmarshalHooks []mapstructure.DecodeHookFunc
	unmarshalHooks = append(unmarshalHooks, mapstructure.StringToTimeDurationHookFunc())

	//check for opts
	unmarshalOptions := unmarshalOpts{}
	for _, param := range opts {
		param(&unmarshalOptions)
	}

	//compose multiple hook funcs to one if found in opts
	hookFn := mapstructure.ComposeDecodeHookFunc(append(unmarshalHooks, unmarshalOptions.hooks...)...)

	//build decoder
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: hookFn,
		Result:     rawVal,
	})
	if err != nil {
		return err
	}

	//decode
	return decoder.Decode(value)
}
