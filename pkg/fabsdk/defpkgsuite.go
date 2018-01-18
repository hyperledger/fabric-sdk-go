/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defclient"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defsvc"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/modlog"
)

type defPkgSuite struct{}

func (ps *defPkgSuite) Core() (apisdk.CoreProviderFactory, error) {
	return defcore.NewProviderFactory(), nil
}

func (ps *defPkgSuite) Service() (apisdk.ServiceProviderFactory, error) {
	return defsvc.NewProviderFactory(), nil
}

func (ps *defPkgSuite) Context() (apisdk.OrgClientFactory, error) {
	return defclient.NewOrgClientFactory(), nil
}

func (ps *defPkgSuite) Session() (apisdk.SessionClientFactory, error) {
	return defclient.NewSessionClientFactory(), nil
}

func (ps *defPkgSuite) Logger() (apilogging.LoggerProvider, error) {
	return modlog.LoggerProvider(), nil
}
