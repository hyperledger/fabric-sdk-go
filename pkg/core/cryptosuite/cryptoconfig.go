/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"os"
	"path"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
)

//ConfigFromBackend returns CryptoSuite config implementation for given backend
func ConfigFromBackend(coreBackend ...core.ConfigBackend) core.CryptoSuiteConfig {
	return &Config{backend: lookup.New(coreBackend...)}
}

// Config represents the crypto suite configuration for the client
type Config struct {
	backend *lookup.ConfigLookup
}

// IsSecurityEnabled config used enable and diable security in cryptosuite
func (c *Config) IsSecurityEnabled() bool {
	return c.backend.GetBool("client.BCCSP.security.enabled")
}

// SecurityAlgorithm returns cryptoSuite config hash algorithm
func (c *Config) SecurityAlgorithm() string {
	return c.backend.GetString("client.BCCSP.security.hashAlgorithm")
}

// SecurityLevel returns cryptSuite config security level
func (c *Config) SecurityLevel() int {
	return c.backend.GetInt("client.BCCSP.security.level")
}

//SecurityProvider provider SW or PKCS11
func (c *Config) SecurityProvider() string {
	return c.backend.GetLowerString("client.BCCSP.security.default.provider")
}

//SoftVerify flag
func (c *Config) SoftVerify() bool {
	return c.backend.GetBool("client.BCCSP.security.softVerify")
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
	return path.Join(keystorePath, "keystore")
}
