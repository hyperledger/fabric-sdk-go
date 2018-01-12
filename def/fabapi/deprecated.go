/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabapi

import (
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
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
	opts := newSDKOptionsFromWrapper(options)
	sdk, err := fabsdk.NewSDK(opts)
	if err != nil {
		return nil, err
	}

	logger.Info("fabapi.NewSDK is depecated - please use fabsdk.NewSDK")

	return sdk, nil
}

// newSDKOptionsFromWrapper populates the SDK options with the default implementation referenced by the fabapi package
func newSDKOptionsFromWrapper(opt Options) fabsdk.Options {
	stateStoreOpts := fabsdk.StateStoreOpts{
		Path: opt.StateStoreOpts.Path,
	}

	sdkOpt := fabsdk.Options{
		ConfigFile:     opt.ConfigFile,
		ConfigByte:     opt.ConfigByte,
		ConfigType:     opt.ConfigType,
		StateStoreOpts: stateStoreOpts,
		CoreFactory:    opt.CoreFactory,
		ServiceFactory: opt.ServiceFactory,
		ContextFactory: opt.ContextFactory,
		SessionFactory: opt.SessionFactory,
		LoggerFactory:  opt.LoggerFactory,
	}

	PopulateSDKOpts(&sdkOpt)

	return sdkOpt
}
