/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/api"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	sdkApi "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/modlog"
	"github.com/pkg/errors"
)

type mockCorePkg struct {
	stateStore      contextApi.KVStore
	cryptoSuite     core.CryptoSuite
	signingManager  contextApi.SigningManager
	identityManager map[string]contextApi.IdentityManager
	fabricProvider  sdkApi.FabricProvider
}

func newMockCorePkg(config core.Config) (*mockCorePkg, error) {
	pkgSuite := defPkgSuite{}
	core, err := pkgSuite.Core()
	if err != nil {
		return nil, err
	}
	stateStore, err := core.CreateStateStoreProvider(config)
	if err != nil {
		return nil, err
	}
	cs, err := core.CreateCryptoSuiteProvider(config)
	if err != nil {
		return nil, err
	}
	sm, err := core.CreateSigningManager(cs, config)
	if err != nil {
		return nil, err
	}
	netConfig, err := config.NetworkConfig()
	if err != nil {
		return nil, err
	}
	im := make(map[string]contextApi.IdentityManager)
	for orgName := range netConfig.Organizations {
		mgr, err := core.CreateIdentityManager(orgName, stateStore, cs, config)
		if err != nil {
			return nil, err
		}
		im[orgName] = mgr
	}

	ctx := mocks.NewMockProviderContextCustom(config, cs, sm)
	fp, err := core.CreateFabricProvider(ctx)
	if err != nil {
		return nil, err
	}

	c := mockCorePkg{
		stateStore:      stateStore,
		cryptoSuite:     cs,
		signingManager:  sm,
		identityManager: im,
		fabricProvider:  fp,
	}

	return &c, nil
}

func (mc *mockCorePkg) CreateStateStoreProvider(config core.Config) (contextApi.KVStore, error) {
	return mc.stateStore, nil
}

func (mc *mockCorePkg) CreateCryptoSuiteProvider(config core.Config) (core.CryptoSuite, error) {
	return mc.cryptoSuite, nil
}

func (mc *mockCorePkg) CreateSigningManager(cryptoProvider core.CryptoSuite, config core.Config) (contextApi.SigningManager, error) {
	return mc.signingManager, nil
}

func (mc *mockCorePkg) CreateIdentityManager(orgName string, stateStore contextApi.KVStore, cryptoProvider core.CryptoSuite, config core.Config) (contextApi.IdentityManager, error) {
	mgr, ok := mc.identityManager[orgName]
	if !ok {
		return nil, fmt.Errorf("identity manager not found for organization: %s", orgName)
	}
	return mgr, nil
}

func (mc *mockCorePkg) CreateFabricProvider(ctx context.ProviderContext) (sdkApi.FabricProvider, error) {
	return mc.fabricProvider, nil
}

type mockPkgSuite struct {
	errOnCore    bool
	errOnService bool
	errOnSession bool
	errOnLogger  bool
}

func (ps *mockPkgSuite) Core() (sdkApi.CoreProviderFactory, error) {
	if ps.errOnCore {
		return nil, errors.New("Error")
	}
	return defcore.NewProviderFactory(), nil
}

func (ps *mockPkgSuite) Service() (sdkApi.ServiceProviderFactory, error) {
	if ps.errOnService {
		return nil, errors.New("Error")
	}
	return defsvc.NewProviderFactory(), nil
}

func (ps *mockPkgSuite) Session() (sdkApi.SessionClientFactory, error) {
	if ps.errOnSession {
		return nil, errors.New("Error")
	}
	return defclient.NewSessionClientFactory(), nil
}

func (ps *mockPkgSuite) Logger() (api.LoggerProvider, error) {
	if ps.errOnLogger {
		return nil, errors.New("Error")
	}
	return modlog.LoggerProvider(), nil
}
