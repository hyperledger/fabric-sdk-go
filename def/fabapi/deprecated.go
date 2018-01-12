/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabapi is deprecated and will be removed - see pkg/fabsdk
package fabapi

import (
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defclient"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Options is a wrapper configuration for the SDK
// Notice: this wrapper is deprecated and will be removed.
type Options struct {
	// Quick access options
	ConfigFile string
	ConfigByte []byte
	ConfigType string

	// Options for default providers
	StateStoreOpts StateStoreOpts

	// Factories to create clients and providers
	CoreFactory    apisdk.CoreProviderFactory
	ServiceFactory apisdk.ServiceProviderFactory
	ContextFactory apisdk.OrgClientFactory
	SessionFactory apisdk.SessionClientFactory
	LoggerFactory  apilogging.LoggerProvider
}

// StateStoreOpts provides setup parameters for KeyValueStore
type StateStoreOpts struct {
	Path string
}

// NewSDK wraps the NewSDK func moved to the pkg folder.
// Notice: this wrapper is deprecated and will be removed.
func NewSDK(options Options) (*fabsdk.FabricSDK, error) {
	sdk, err := fabsdk.New(
		fabsdk.ConfigBytes(options.ConfigByte, options.ConfigType),
		fabsdk.ConfigFile(options.ConfigFile),
		fabsdk.StateStorePath(options.StateStoreOpts.Path),
		pkgSuiteFromOptions(options))

	if err != nil {
		return nil, err
	}

	logger.Debug("fabapi.NewSDK is deprecated - please use fabsdk.New")

	return sdk, nil
}

func pkgSuiteFromOptions(options Options) fabsdk.SDKOption {
	impl := apisdk.PkgSuite{}
	if options.CoreFactory != nil {
		impl.Core = options.CoreFactory
	} else {
		impl.Core = defcore.NewProviderFactory()
	}

	if options.ServiceFactory != nil {
		impl.Service = options.ServiceFactory
	} else {
		impl.Service = defsvc.NewProviderFactory()
	}

	if options.ContextFactory != nil {
		impl.Context = options.ContextFactory
	} else {
		impl.Context = defclient.NewOrgClientFactory()
	}

	if options.SessionFactory != nil {
		impl.Session = options.SessionFactory
	} else {
		impl.Session = defclient.NewSessionClientFactory()
	}

	if options.LoggerFactory != nil {
		impl.Logger = options.LoggerFactory
	} else {
		impl.Logger = deflogger.LoggerProvider()
	}

	return fabsdk.PkgSuiteAsOpt(impl)
}
