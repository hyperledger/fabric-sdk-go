/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	sdkApi "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defmsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/modlog"
)

type defPkgSuite struct{}

func (ps *defPkgSuite) Core() (sdkApi.CoreProviderFactory, error) {
	return defcore.NewProviderFactory(), nil
}

func (ps *defPkgSuite) MSP() (sdkApi.MSPProviderFactory, error) {
	return defmsp.NewProviderFactory(), nil
}

func (ps *defPkgSuite) Service() (sdkApi.ServiceProviderFactory, error) {
	return defsvc.NewProviderFactory(), nil
}

func (ps *defPkgSuite) Logger() (api.LoggerProvider, error) {
	return modlog.LoggerProvider(), nil
}
