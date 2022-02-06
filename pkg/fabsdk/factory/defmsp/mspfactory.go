/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defmsp

import (
	"io/fs"

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
	opts providerFactoryOptions
}

type ProviderFactoryOption func(*providerFactoryOptions) error

// WithFS creates ProviderFactory with fs.FS read only based storage.
func WithFS(filesystem fs.FS) ProviderFactoryOption {
	return func(pfo *providerFactoryOptions) error {
		pfo.filesystem = filesystem
		return nil
	}
}

type providerFactoryOptions struct {
	filesystem fs.FS
}

// NewProviderFactory returns the default MSP provider factory.
func NewProviderFactory(opts ...ProviderFactoryOption) *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// CreateUserStore creates a UserStore using the SDK's default implementation
func (f *ProviderFactory) CreateUserStore(config msp.IdentityConfig) (msp.UserStore, error) {
	stateStorePath := config.CredentialStorePath()

	var userStore msp.UserStore
	if stateStorePath == "" {
		return mspimpl.NewMemoryUserStore(), nil
	}

	stateStore, err := kvs.New(&kvs.FileKeyValueStoreOptions{
		Path:       stateStorePath,
		Filesystem: f.opts.filesystem,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "CreateNewFileKeyValueStore failed")
	}
	userStore, err = mspimpl.NewCertFileUserStore1(stateStore)
	if err != nil {
		return nil, errors.Wrapf(err, "creating a user store failed")
	}

	return userStore, nil
}

// CreateIdentityManagerProvider returns a new default implementation of MSP provider
func (f *ProviderFactory) CreateIdentityManagerProvider(endpointConfig fab.EndpointConfig, cryptoProvider core.CryptoSuite, userStore msp.UserStore) (msp.IdentityManagerProvider, error) {
	return msppvdr.New(endpointConfig, cryptoProvider, userStore)
}
