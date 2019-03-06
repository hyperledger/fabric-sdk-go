// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/modlog"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	sdkApi "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defmsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/pkg/errors"
)

type mockCorePkg struct {
	cryptoSuite    core.CryptoSuite
	signingManager core.SigningManager
	infraProvider  fab.InfraProvider
}

func newMockCorePkg(configBackendProvider core.ConfigProvider) (*mockCorePkg, error) {

	configBackend, err := configBackendProvider()
	if err != nil {
		return nil, err
	}

	endpointConfig, err := fabImpl.ConfigFromBackend(configBackend...)
	if err != nil {
		return nil, err
	}

	cryptoSuiteConfig := cryptosuite.ConfigFromBackend(configBackend...)

	pkgSuite := defPkgSuite{}
	sdkcore, err := pkgSuite.Core()
	if err != nil {
		return nil, err
	}
	cs, err := sdkcore.CreateCryptoSuiteProvider(cryptoSuiteConfig)
	if err != nil {
		return nil, err
	}
	sm, err := sdkcore.CreateSigningManager(cs)
	if err != nil {
		return nil, err
	}
	fp, err := sdkcore.CreateInfraProvider(endpointConfig)
	if err != nil {
		return nil, err
	}

	c := mockCorePkg{
		cryptoSuite:    cs,
		signingManager: sm,
		infraProvider:  fp,
	}

	return &c, nil
}

func (mc *mockCorePkg) CreateCryptoSuiteProvider(config core.CryptoSuiteConfig) (core.CryptoSuite, error) {
	return mc.cryptoSuite, nil
}

func (mc *mockCorePkg) CreateSigningManager(cryptoProvider core.CryptoSuite) (core.SigningManager, error) {
	return mc.signingManager, nil
}

func (mc *mockCorePkg) CreateInfraProvider(config fab.EndpointConfig) (fab.InfraProvider, error) {
	return mc.infraProvider, nil
}

type mockPkgSuite struct {
	errOnCore    bool
	errOnMsp     bool
	errOnService bool
	errOnLogger  bool
}

func (ps *mockPkgSuite) Core() (sdkApi.CoreProviderFactory, error) {
	if ps.errOnCore {
		return nil, errors.New("Error")
	}
	return defcore.NewProviderFactory(), nil
}

func (ps *mockPkgSuite) MSP() (sdkApi.MSPProviderFactory, error) {
	if ps.errOnMsp {
		return nil, errors.New("Error")
	}
	return defmsp.NewProviderFactory(), nil
}

func (ps *mockPkgSuite) Service() (sdkApi.ServiceProviderFactory, error) {
	if ps.errOnService {
		return nil, errors.New("Error")
	}
	return defsvc.NewProviderFactory(), nil
}

func (ps *mockPkgSuite) Logger() (api.LoggerProvider, error) {
	if ps.errOnLogger {
		return nil, errors.New("Error")
	}
	return modlog.LoggerProvider(), nil
}
