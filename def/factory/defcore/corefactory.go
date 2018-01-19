/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defcore

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicore"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"

	"github.com/hyperledger/fabric-sdk-go/def/provider/fabpvdr"
	cryptosuiteimpl "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	signingMgr "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/signingmgr"

	"github.com/hyperledger/fabric-sdk-go/pkg/logging/modlog"
)

// ProviderFactory represents the default SDK provider factory.
type ProviderFactory struct {
	// stateStoreOpts is deprecated
	stateStoreOpts StateStoreOptsDeprecated
}

// NewProviderFactory returns the default SDK provider factory.
func NewProviderFactory() *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// NewStateStoreProvider creates a KeyValueStore using the SDK's default implementation
func (f *ProviderFactory) NewStateStoreProvider(config apiconfig.Config) (fab.KeyValueStore, error) {

	var stateStorePath = f.stateStoreOpts.Path
	if stateStorePath == "" {
		clientCofig, err := config.Client()
		if err != nil {
			return nil, errors.WithMessage(err, "Unable to retrieve client config")
		}
		stateStorePath = clientCofig.CredentialStore.Path
	}

	stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateNewFileKeyValueStore failed")
	}
	return stateStore, nil
}

// NewCryptoSuiteProvider returns a new default implementation of BCCSP
func (f *ProviderFactory) NewCryptoSuiteProvider(config apiconfig.Config) (apicryptosuite.CryptoSuite, error) {
	cryptoSuiteProvider, err := cryptosuiteimpl.GetSuiteByConfig(config)
	return cryptoSuiteProvider, err
}

// NewSigningManager returns a new default implementation of signing manager
func (f *ProviderFactory) NewSigningManager(cryptoProvider apicryptosuite.CryptoSuite, config apiconfig.Config) (fab.SigningManager, error) {
	return signingMgr.NewSigningManager(cryptoProvider, config)
}

// NewFabricProvider returns a new default implementation of fabric primitives
func (f *ProviderFactory) NewFabricProvider(config apiconfig.Config, stateStore fab.KeyValueStore, cryptoSuite apicryptosuite.CryptoSuite, signer fab.SigningManager) (apicore.FabricProvider, error) {
	return fabpvdr.NewFabricProvider(config, stateStore, cryptoSuite, signer), nil
}

// NewLoggerProvider returns a new default implementation of a logger backend
// This function is separated from the factory to allow logger creation first.
func NewLoggerProvider() apilogging.LoggerProvider {
	return modlog.LoggerProvider()
}
