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

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
)

var logModules = [...]string{"fabsdk", "fabsdk/client", "fabsdk/core", "fabsdk/fab", "fabsdk/common",
	"fabsdk/msp", "fabsdk/util", "fabsdk/context"}

type options struct {
	envPrefix    string
	templatePath string
}

const (
	cmdRoot = "FABRIC_SDK"
)

// Option configures the package.
type Option func(opts *options) error

// FromReader loads configuration from in.
// configType can be "json" or "yaml".
func FromReader(in io.Reader, configType string, opts ...Option) core.ConfigProvider {
	return func() ([]core.ConfigBackend, error) {
		return initFromReader(in, configType, opts...)
	}
}

// FromFile reads from named config file
func FromFile(name string, opts ...Option) core.ConfigProvider {
	return func() ([]core.ConfigBackend, error) {
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
		if err != nil {
			return nil, errors.Wrapf(err, "loading config file failed: %s", name)
		}

		setLogLevel(backend)

		return []core.ConfigBackend{backend}, nil
	}
}

// FromRaw will initialize the configs from a byte array
func FromRaw(configBytes []byte, configType string, opts ...Option) core.ConfigProvider {
	return func() ([]core.ConfigBackend, error) {
		buf := bytes.NewBuffer(configBytes)
		return initFromReader(buf, configType, opts...)
	}
}

func initFromReader(in io.Reader, configType string, opts ...Option) ([]core.ConfigBackend, error) {
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
	err = backend.configViper.MergeConfig(in)
	if err != nil {
		return nil, err
	}
	setLogLevel(backend)

	return []core.ConfigBackend{backend}, nil
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

// setLogLevel will set the log level of the client
func setLogLevel(backend core.ConfigBackend) {
	loggingLevelString, _ := backend.Lookup("client.logging.level")
	logLevel := logging.INFO
	if loggingLevelString != nil {
		var err error
		logLevel, err = logging.LogLevel(loggingLevelString.(string))
		if err != nil {
			panic(err)
		}
	}

	// TODO: allow separate settings for each
	for _, logModule := range logModules {
		logging.SetLevel(logModule, logLevel)
	}
}
