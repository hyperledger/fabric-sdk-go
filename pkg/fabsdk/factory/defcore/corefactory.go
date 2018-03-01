/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defcore

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/api"

	cryptosuiteimpl "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/identitymgr"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	signingMgr "github.com/hyperledger/fabric-sdk-go/pkg/fab/signingmgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	"github.com/pkg/errors"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	sdkApi "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/modlog"
)

// ProviderFactory represents the default SDK provider factory.
type ProviderFactory struct {
}

// NewProviderFactory returns the default SDK provider factory.
func NewProviderFactory() *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// CreateStateStoreProvider creates a KeyValueStore using the SDK's default implementation
func (f *ProviderFactory) CreateStateStoreProvider(config core.Config) (contextApi.KVStore, error) {

	clientCofig, err := config.Client()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to retrieve client config")
	}
	stateStorePath := clientCofig.CredentialStore.Path

	stateStore, err := kvs.New(&kvs.FileKeyValueStoreOptions{Path: stateStorePath})
	if err != nil {
		return nil, errors.WithMessage(err, "CreateNewFileKeyValueStore failed")
	}
	return stateStore, nil
}

// CreateCryptoSuiteProvider returns a new default implementation of BCCSP
func (f *ProviderFactory) CreateCryptoSuiteProvider(config core.Config) (core.CryptoSuite, error) {
	cryptoSuiteProvider, err := cryptosuiteimpl.GetSuiteByConfig(config)
	return cryptoSuiteProvider, err
}

// CreateSigningManager returns a new default implementation of signing manager
func (f *ProviderFactory) CreateSigningManager(cryptoProvider core.CryptoSuite, config core.Config) (contextApi.SigningManager, error) {
	return signingMgr.New(cryptoProvider, config)
}

// CreateIdentityManager returns a new default implementation of identity manager
func (f *ProviderFactory) CreateIdentityManager(org string, stateStore contextApi.KVStore, cryptoProvider core.CryptoSuite, config core.Config) (contextApi.IdentityManager, error) {
	return identitymgr.New(org, stateStore, cryptoProvider, config)
}

// CreateFabricProvider returns a new default implementation of fabric primitives
func (f *ProviderFactory) CreateFabricProvider(context context.ProviderContext) (sdkApi.FabricProvider, error) {
	return fabpvdr.New(context), nil
}

// NewLoggerProvider returns a new default implementation of a logger backend
// This function is separated from the factory to allow logger creation first.
func NewLoggerProvider() api.LoggerProvider {
	return modlog.LoggerProvider()
}
