/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"bytes"
	"io"
	"strings"

	"github.com/spf13/viper"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/pkg/errors"

	"regexp"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

var logger = logging.NewLogger("fabsdk/core")

var logModules = [...]string{"fabsdk", "fabsdk/client", "fabsdk/core", "fabsdk/fab", "fabsdk/common",
	"fabsdk/msp", "fabsdk/util", "fabsdk/context"}

type options struct {
	envPrefix    string
	templatePath string
}

// Option configures the package.
type Option func(opts *options) error

//Provider provides all config implementations
//TODO to be removed once FromBackend in replaced with 3 functions for each config Type
type Provider func() (core.CryptoSuiteConfig, fab.EndpointConfig, msp.IdentityConfig, error)

// FromReader loads configuration from in.
// configType can be "json" or "yaml".
func FromReader(in io.Reader, configType string, opts ...Option) core.ConfigProvider {
	return func() (core.ConfigBackend, error) {
		backend, err := newBackend(opts...)
		if err != nil {
			return nil, err
		}

		if configType == "" {
			return nil, errors.New("empty config type")
		}

		// read config from bytes array, but must set ConfigType
		// for viper to properly unmarshal the bytes array
		backend.configViper.SetConfigType(configType)
		backend.configViper.MergeConfig(in)

		return backend, nil
	}
}

// FromFile reads from named config file
func FromFile(name string, opts ...Option) core.ConfigProvider {
	return func() (core.ConfigBackend, error) {
		backend, err := newBackend(opts...)
		if err != nil {
			return nil, err
		}

		if name == "" {
			return nil, errors.New("filename is required")
		}

		// create new viper
		backend.configViper.SetConfigFile(name)

		// If a config file is found, read it in.
		err = backend.configViper.MergeInConfig()
		if err == nil {
			logger.Debugf("Using config file: %s", backend.configViper.ConfigFileUsed())
		} else {
			return nil, errors.Wrap(err, "loading config file failed")
		}

		return backend, nil
	}
}

// FromRaw will initialize the configs from a byte array
func FromRaw(configBytes []byte, configType string, opts ...Option) core.ConfigProvider {
	buf := bytes.NewBuffer(configBytes)
	logger.Debugf("config.FromRaw buf Len is %d, Cap is %d: %s", buf.Len(), buf.Cap(), buf)

	return FromReader(buf, configType, opts...)
}

// FromBackend Creates config provider from config backend
//TODO to be replaced with 3 functions to get 3 kinds of configs
func FromBackend(backend core.ConfigBackend) Provider {
	return func() (core.CryptoSuiteConfig, fab.EndpointConfig, msp.IdentityConfig, error) {
		return initConfig(backend)
	}
}

// WithEnvPrefix defines the prefix for environment variable overrides.
// See viper SetEnvPrefix for more information.
func WithEnvPrefix(prefix string) Option {
	return func(opts *options) error {
		opts.envPrefix = prefix
		return nil
	}
}

func newBackend(opts ...Option) (*defConfigBackend, error) {
	o := options{
		envPrefix: cmdRoot,
	}

	for _, option := range opts {
		err := option(&o)
		if err != nil {
			return nil, errors.WithMessage(err, "Error in options passed to create new config backend")
		}
	}

	v := newViper(o.envPrefix)

	//default backend for config
	backend := &defConfigBackend{
		configViper: v,
		opts:        o,
	}

	err := backend.loadTemplateConfig()
	if err != nil {
		return nil, err
	}

	return backend, nil
}

func newViper(cmdRootPrefix string) *viper.Viper {
	myViper := viper.New()
	myViper.SetEnvPrefix(cmdRootPrefix)
	myViper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	myViper.SetEnvKeyReplacer(replacer)
	return myViper
}

func initConfig(backend core.ConfigBackend) (core.CryptoSuiteConfig, fab.EndpointConfig, msp.IdentityConfig, error) {

	configBackend := &Backend{coreBackend: backend}
	setLogLevel(configBackend)
	for _, logModule := range logModules {
		logger.Debugf("config %s logging level is set to: %s", logModule, logging.ParseString(logging.GetLevel(logModule)))
	}

	cryptoConfig := &CryptoSuiteConfig{backend: configBackend}
	endpointConfig := &EndpointConfig{backend: configBackend}

	if err := endpointConfig.cacheNetworkConfiguration(); err != nil {
		return nil, nil, nil, errors.WithMessage(err, "network configuration load failed")
	}
	//Compile the entityMatchers
	endpointConfig.peerMatchers = make(map[int]*regexp.Regexp)
	endpointConfig.ordererMatchers = make(map[int]*regexp.Regexp)
	endpointConfig.caMatchers = make(map[int]*regexp.Regexp)

	matchError := endpointConfig.compileMatchers()
	if matchError != nil {
		return nil, nil, nil, matchError
	}

	identityConfig := &IdentityConfig{endpointConfig: endpointConfig}

	return cryptoConfig, endpointConfig, identityConfig, nil
}

// setLogLevel will set the log level of the client
func setLogLevel(backend *Backend) {
	loggingLevelString := backend.getString("client.logging.level")
	logLevel := logging.INFO
	if loggingLevelString != "" {
		const logModule = "fabsdk" // TODO: allow more flexability in setting levels for different modules
		logger.Debugf("%s logging level from the config: %v", logModule, loggingLevelString)
		var err error
		logLevel, err = logging.LogLevel(loggingLevelString)
		if err != nil {
			panic(err)
		}
	}

	// TODO: allow separate settings for each
	for _, logModule := range logModules {
		logging.SetLevel(logModule, logLevel)
	}
}
