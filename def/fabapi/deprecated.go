/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabapi is deprecated and will be removed - see pkg/fabsdk
package fabapi

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	"github.com/pkg/errors"
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

func configFromOptions(options *Options) apiconfig.ConfigProvider {
	if options.ConfigByte != nil {
		return config.FromRaw(options.ConfigByte, options.ConfigType)
	}

	if options.ConfigFile != "" {
		return config.FromFile(options.ConfigFile)
	}

	return func() (apiconfig.Config, error) {
		return nil, errors.New("No configuration provided")
	}
}

// NewSDK wraps the NewSDK func moved to the pkg folder.
// Notice: this wrapper is deprecated and will be removed.
func NewSDK(options Options) (*fabsdk.FabricSDK, error) {
	sdk, err := fabsdk.New(configFromOptions(&options),
		sdkOptionsFromDeprecatedOptions(options)...)

	if err != nil {
		return nil, err
	}

	logger.Debug("fabapi.NewSDK is deprecated - please use fabsdk.New")

	return sdk, nil
}

func sdkOptionsFromDeprecatedOptions(options Options) []fabsdk.Option {
	opts := []fabsdk.Option{}

	if options.CoreFactory != nil {
		opts = append(opts, fabsdk.WithCorePkg(options.CoreFactory))
	} else {
		stateStoreOpts := defcore.StateStoreOptsDeprecated{
			Path: options.StateStoreOpts.Path,
		}
		core := defcore.NewProviderFactoryDeprecated(stateStoreOpts)
		opts = append(opts, fabsdk.WithCorePkg(core))
	}

	if options.ServiceFactory != nil {
		opts = append(opts, fabsdk.WithServicePkg(options.ServiceFactory))
	}

	if options.ContextFactory != nil {
		opts = append(opts, fabsdk.WithContextPkg(options.ContextFactory))
	}

	if options.SessionFactory != nil {
		opts = append(opts, fabsdk.WithSessionPkg(options.SessionFactory))
	}

	if options.LoggerFactory != nil {
		opts = append(opts, fabsdk.WithLoggerPkg(options.LoggerFactory))
	}

	return opts
}
