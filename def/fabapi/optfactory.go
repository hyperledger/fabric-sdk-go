/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabapi

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// SDKOpts provides bootstrap setup
type SDKOpts struct {
	ConfigFile string
}

// ConfigOpts provides setup parameters for Config
type ConfigOpts struct { // TODO (moved ConfigFile to SDKOpts to make setup easier for API consumer)
}

// StateStoreOpts provides setup parameters for KeyValueStore
type StateStoreOpts struct {
	Path string
}

// NewDefaultConfig creates a Config using the SDK's default implementation
func NewDefaultConfig(o ConfigOpts, a SDKOpts) (apiconfig.Config, error) {
	return NewConfigManager(a.ConfigFile)
}

// NewDefaultStateStore creates a KeyValueStore using the SDK's default implementation
func NewDefaultStateStore(o StateStoreOpts, config apiconfig.Config) (fab.KeyValueStore, error) {
	return NewKVStore(o.Path) // TODO: config should have this capability
}
