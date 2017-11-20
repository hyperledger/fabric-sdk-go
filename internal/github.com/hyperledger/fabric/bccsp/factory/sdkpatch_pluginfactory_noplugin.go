// +build !linux,!nobccspplugin nobccspplugin

/*
Copyright IBM Corp., SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/
package factory

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"

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
func (f *PluginFactory) Get(config *FactoryOpts) (bccsp.BCCSP, error) {
	return nil, errors.New("not supported")
}
