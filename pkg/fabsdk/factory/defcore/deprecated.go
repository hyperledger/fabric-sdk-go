/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defcore

// StateStoreOptsDeprecated provides setup parameters for KeyValueStore
type StateStoreOptsDeprecated struct {
	Path string
}

// CreateProviderFactoryDeprecated returns the default SDK provider factory.
func CreateProviderFactoryDeprecated(stateStoreOpts StateStoreOptsDeprecated) *ProviderFactory {
	f := ProviderFactory{
		stateStoreOpts,
	}
	return &f
}
