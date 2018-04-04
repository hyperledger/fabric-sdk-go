/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lookup

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/spf13/cast"
)

//New providers lookup wrapper around given backend
func New(coreBackend core.ConfigBackend) *ConfigLookup {
	return &ConfigLookup{backend: coreBackend}
}

//ConfigLookup is wrapper for core.ConfigBackend which performs key lookup and unmarshalling
type ConfigLookup struct {
	backend core.ConfigBackend
}

//GetBool returns bool value for given key
func (c *ConfigLookup) GetBool(key string) bool {
	value, ok := c.backend.Lookup(key)
	if !ok {
		return false
	}
	return cast.ToBool(value)
}

//GetString returns string value for given key
func (c *ConfigLookup) GetString(key string) string {
	value, ok := c.backend.Lookup(key)
	if !ok {
		return ""
	}
	return cast.ToString(value)
}

//GetInt returns int value for given key
func (c *ConfigLookup) GetInt(key string) int {
	value, ok := c.backend.Lookup(key)
	if !ok {
		return 0
	}
	return cast.ToInt(value)
}

//GetDuration returns time.Duration value for given key
func (c *ConfigLookup) GetDuration(key string) time.Duration {
	value, ok := c.backend.Lookup(key)
	if !ok {
		return 0
	}
	return cast.ToDuration(value)
}

//UnmarshalKey unmarshals value for given key to rawval type
func (c *ConfigLookup) UnmarshalKey(key string, rawVal interface{}) bool {
	_, ok := c.backend.Lookup(key, core.WithUnmarshalType(rawVal))
	return ok
}
