/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/api/kvstore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/modlog"
	"github.com/pkg/errors"
)

type mockCorePkg struct {
	stateStore     kvstore.KVStore
	cryptoSuite    apicryptosuite.CryptoSuite
	signingManager apifabclient.SigningManager
	fabricProvider apisdk.FabricProvider
}

func newMockCorePkg(config apiconfig.Config) (*mockCorePkg, error) {
	pkgSuite := defPkgSuite{}
	core, err := pkgSuite.Core()
	if err != nil {
		return nil, err
	}
	stateStore, err := core.NewStateStoreProvider(config)
	if err != nil {
		return nil, err
	}
	cs, err := core.NewCryptoSuiteProvider(config)
	if err != nil {
		return nil, err
	}
	sm, err := core.NewSigningManager(cs, config)
	if err != nil {
		return nil, err
	}

	ctx := mocks.NewMockProviderContextCustom(config, cs, sm)
	fp, err := core.NewFabricProvider(ctx)
	if err != nil {
		return nil, err
	}

	c := mockCorePkg{
		stateStore:     stateStore,
		cryptoSuite:    cs,
		signingManager: sm,
		fabricProvider: fp,
	}

	return &c, nil
}

func (mc *mockCorePkg) NewStateStoreProvider(config apiconfig.Config) (kvstore.KVStore, error) {
	return mc.stateStore, nil
}

func (mc *mockCorePkg) NewCryptoSuiteProvider(config apiconfig.Config) (apicryptosuite.CryptoSuite, error) {
	return mc.cryptoSuite, nil
}

func (mc *mockCorePkg) NewSigningManager(cryptoProvider apicryptosuite.CryptoSuite, config apiconfig.Config) (apifabclient.SigningManager, error) {
	return mc.signingManager, nil
}

func (mc *mockCorePkg) NewFabricProvider(ctx apifabclient.ProviderContext) (apisdk.FabricProvider, error) {
	return mc.fabricProvider, nil
}

type mockPkgSuite struct {
	errOnCore    bool
	errOnService bool
	errOnContext bool
	errOnSession bool
	errOnLogger  bool
}

func (ps *mockPkgSuite) Core() (apisdk.CoreProviderFactory, error) {
	if ps.errOnCore {
		return nil, errors.New("Error")
	}
	return defcore.NewProviderFactory(), nil
}

func (ps *mockPkgSuite) Service() (apisdk.ServiceProviderFactory, error) {
	if ps.errOnService {
		return nil, errors.New("Error")
	}
	return defsvc.NewProviderFactory(), nil
}

func (ps *mockPkgSuite) Context() (apisdk.OrgClientFactory, error) {
	if ps.errOnContext {
		return nil, errors.New("Error")
	}
	return defclient.NewOrgClientFactory(), nil
}

func (ps *mockPkgSuite) Session() (apisdk.SessionClientFactory, error) {
	if ps.errOnSession {
		return nil, errors.New("Error")
	}
	return defclient.NewSessionClientFactory(), nil
}

func (ps *mockPkgSuite) Logger() (apilogging.LoggerProvider, error) {
	if ps.errOnLogger {
		return nil, errors.New("Error")
	}
	return modlog.LoggerProvider(), nil
}
