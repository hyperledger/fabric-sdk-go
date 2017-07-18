/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package opt

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
