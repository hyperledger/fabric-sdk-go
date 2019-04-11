/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/spf13/cast"
)

const (
	defEnabled       = true
	defHashAlgorithm = "SHA2"
	defLevel         = 256
	defProvider      = "SW"
	defSoftVerify    = true
)

//ConfigFromBackend returns CryptoSuite config implementation for given backend
func ConfigFromBackend(coreBackend ...core.ConfigBackend) core.CryptoSuiteConfig {
	return &Config{backend: lookup.New(coreBackend...)}
}

// Config represents the crypto suite configuration for the client
type Config struct {
	backend *lookup.ConfigLookup
}

// IsSecurityEnabled config used enable and disable security in cryptosuite
func (c *Config) IsSecurityEnabled() bool {
	val, ok := c.backend.Lookup("client.BCCSP.security.enabled")
	if !ok {
		return defEnabled
	}
	return cast.ToBool(val)
}

// SecurityAlgorithm returns cryptoSuite config hash algorithm
func (c *Config) SecurityAlgorithm() string {
	val, ok := c.backend.Lookup("client.BCCSP.security.hashAlgorithm")
	if !ok {
		return defHashAlgorithm
	}
	return cast.ToString(val)
}

// SecurityLevel returns cryptSuite config security level
func (c *Config) SecurityLevel() int {
	val, ok := c.backend.Lookup("client.BCCSP.security.level")
	if !ok {
		return defLevel
	}
	return cast.ToInt(val)
}

//SecurityProvider provider SW or PKCS11
func (c *Config) SecurityProvider() string {
	val, ok := c.backend.Lookup("client.BCCSP.security.default.provider")
	if !ok {
		return strings.ToLower(defProvider)
	}
	return strings.ToLower(cast.ToString(val))
}

//SoftVerify flag
func (c *Config) SoftVerify() bool {
	val, ok := c.backend.Lookup("client.BCCSP.security.softVerify")
	if !ok {
		return defSoftVerify
	}
	return cast.ToBool(val)
}

//SecurityProviderLibPath will be set only if provider is PKCS11
func (c *Config) SecurityProviderLibPath() string {
	configuredLibs := c.backend.GetString("client.BCCSP.security.library")
	libPaths := strings.Split(configuredLibs, ",")
	logger.Debugf("Configured BCCSP Lib Paths %s", libPaths)
	var lib string
	for _, path := range libPaths {
		if _, err := os.Stat(strings.TrimSpace(path)); !os.IsNotExist(err) {
			lib = strings.TrimSpace(path)
			break
		}
	}
	if lib != "" {
		logger.Debugf("Found softhsm library: %s", lib)
	} else {
		logger.Debug("Softhsm library was not found")
	}
	return lib
}

//SecurityProviderPin will be set only if provider is PKCS11
func (c *Config) SecurityProviderPin() string {
	return c.backend.GetString("client.BCCSP.security.pin")
}

//SecurityProviderLabel will be set only if provider is PKCS11
func (c *Config) SecurityProviderLabel() string {
	return c.backend.GetString("client.BCCSP.security.label")
}

// KeyStorePath returns the keystore path used by BCCSP
func (c *Config) KeyStorePath() string {
	keystorePath := pathvar.Subst(c.backend.GetString("client.credentialStore.cryptoStore.path"))
	return filepath.Join(keystorePath, "keystore")
}
