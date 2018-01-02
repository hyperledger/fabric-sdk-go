/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/
package plugin

import (
	"errors"
	"fmt"
	"os"
	"plugin"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
)

const (
	// PluginFactoryName is the factory name for BCCSP plugins
	PluginFactoryName = "PLUGIN"
)

// PluginOpts contains the options for the PluginFactory
type PluginOpts struct {
	// Path to plugin library
	Library string
	// Config map for the plugin library
	Config map[string]interface{}
}

// PluginFactory is the factory for BCCSP plugins
type PluginFactory struct{}

// Name returns the name of this factory
func (f *PluginFactory) Name() string {
	return PluginFactoryName
}

// Get returns an instance of BCCSP using Opts.
func (f *PluginFactory) Get(pluginOpts *PluginOpts) (bccsp.BCCSP, error) {
	// check for valid config
	if pluginOpts == nil {
		return nil, errors.New("Invalid config. It must not be nil.")
	}

	// Library is required property
	if pluginOpts.Library == "" {
		return nil, errors.New("Invalid config: missing property 'Library'")
	}

	// make sure the library exists
	if _, err := os.Stat(pluginOpts.Library); err != nil {
		return nil, fmt.Errorf("Could not find library '%s' [%s]", pluginOpts.Library, err)
	}

	// attempt to load the library as a plugin
	plug, err := plugin.Open(pluginOpts.Library)
	if err != nil {
		return nil, fmt.Errorf("Failed to load plugin '%s' [%s]", pluginOpts.Library, err)
	}

	// lookup the required symbol 'New'
	sym, err := plug.Lookup("New")
	if err != nil {
		return nil, fmt.Errorf("Could not find required symbol 'CryptoServiceProvider' [%s]", err)
	}

	// check to make sure symbol New meets the required function signature
	new, ok := sym.(func(config map[string]interface{}) (bccsp.BCCSP, error))
	if !ok {
		return nil, fmt.Errorf("Plugin does not implement the required function signature for 'New'")
	}

	return new(pluginOpts.Config)
}
