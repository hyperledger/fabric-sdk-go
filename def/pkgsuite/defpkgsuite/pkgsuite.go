/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package defpkgsuite provides a default implementation of the fabric API for fabsdk
package defpkgsuite

import (
	"github.com/hyperledger/fabric-sdk-go/def/factory/defclient"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	sdkapi "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

// SDKOpt provides the default implementation for the SDK
func SDKOpt() fabsdk.SDKOption {
	return fabsdk.PkgSuiteAsOpt(newPkgSuite())
}

func newPkgSuite() sdkapi.PkgSuite {
	pkgSuite := sdkapi.PkgSuite{
		Core:    defcore.NewProviderFactory(),
		Service: defsvc.NewProviderFactory(),
		Context: defclient.NewOrgClientFactory(),
		Session: defclient.NewSessionClientFactory(),
		Logger:  deflogger.LoggerProvider(),
	}
	return pkgSuite
}
