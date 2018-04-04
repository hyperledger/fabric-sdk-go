/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// defConfigBackend represents the default config backend
type defConfigBackend struct {
	configViper *viper.Viper
	opts        options
}

// Lookup gets the config item value by Key
func (c *defConfigBackend) Lookup(key string, opts ...core.LookupOption) (interface{}, bool) {
	if len(opts) > 0 {
		lookupOpts := &core.LookupOpts{}
		for _, option := range opts {
			option(lookupOpts)
		}

		if lookupOpts.UnmarshalType != nil {
			err := c.configViper.UnmarshalKey(key, lookupOpts.UnmarshalType)
			if err != nil {
				//TODO may need debug logger here
				return nil, false
			}
			return lookupOpts.UnmarshalType, true
		}
	}
	value := c.configViper.Get(key)
	if value == nil {
		return nil, false
	}
	return value, true
}

// load Default config
func (c *defConfigBackend) loadTemplateConfig() error {
	// get Environment Default Config Path
	templatePath := c.opts.templatePath
	if templatePath == "" {
		return nil
	}

	// if set, use it to load default config
	c.configViper.AddConfigPath(pathvar.Subst(templatePath))
	err := c.configViper.ReadInConfig() // Find and read the config file
	if err != nil {                     // Handle errors reading the config file
		return errors.Wrap(err, "loading config file failed")
	}
	return nil
}
