/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defmsp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/msppvdr"
	mspimpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/pkg/errors"
)

// ProviderFactory represents the default MSP provider factory.
type ProviderFactory struct {
}

// NewProviderFactory returns the default MSP provider factory.
func NewProviderFactory() *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// CreateUserStore creates a UserStore using the SDK's default implementation
func (f *ProviderFactory) CreateUserStore(config msp.IdentityConfig) (msp.UserStore, error) {

	stateStorePath := config.Client().CredentialStore.Path

	var userStore msp.UserStore

	if stateStorePath == "" {
		userStore = mspimpl.NewMemoryUserStore()
	} else {
		stateStore, err := kvs.New(&kvs.FileKeyValueStoreOptions{Path: stateStorePath})
		if err != nil {
			return nil, errors.WithMessage(err, "CreateNewFileKeyValueStore failed")
		}
		userStore, err = mspimpl.NewCertFileUserStore1(stateStore)
		if err != nil {
			return nil, errors.Wrapf(err, "creating a user store failed")
		}
	}

	return userStore, nil
}

// CreateIdentityManagerProvider returns a new default implementation of MSP provider
func (f *ProviderFactory) CreateIdentityManagerProvider(endpointConfig fab.EndpointConfig, cryptoProvider core.CryptoSuite, userStore msp.UserStore) (msp.IdentityManagerProvider, error) {
	return msppvdr.New(endpointConfig, cryptoProvider, userStore)
}
