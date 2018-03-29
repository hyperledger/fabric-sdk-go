/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"os"
	"path"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
)

// CryptoSuiteConfig represents the crypto suite configuration for the client
type CryptoSuiteConfig struct {
	backend *Backend
}

// IsSecurityEnabled config used enable and diable security in cryptosuite
func (c *CryptoSuiteConfig) IsSecurityEnabled() bool {
	return c.backend.getBool("client.BCCSP.security.enabled")
}

// SecurityAlgorithm returns cryptoSuite config hash algorithm
func (c *CryptoSuiteConfig) SecurityAlgorithm() string {
	return c.backend.getString("client.BCCSP.security.hashAlgorithm")
}

// SecurityLevel returns cryptSuite config security level
func (c *CryptoSuiteConfig) SecurityLevel() int {
	return c.backend.getInt("client.BCCSP.security.level")
}

//SecurityProvider provider SW or PKCS11
func (c *CryptoSuiteConfig) SecurityProvider() string {
	return c.backend.getString("client.BCCSP.security.default.provider")
}

//Ephemeral flag
func (c *CryptoSuiteConfig) Ephemeral() bool {
	return c.backend.getBool("client.BCCSP.security.ephemeral")
}

//SoftVerify flag
func (c *CryptoSuiteConfig) SoftVerify() bool {
	return c.backend.getBool("client.BCCSP.security.softVerify")
}

//SecurityProviderLibPath will be set only if provider is PKCS11
func (c *CryptoSuiteConfig) SecurityProviderLibPath() string {
	configuredLibs := c.backend.getString("client.BCCSP.security.library")
	libPaths := strings.Split(configuredLibs, ",")
	logger.Debug("Configured BCCSP Lib Paths %v", libPaths)
	var lib string
	for _, path := range libPaths {
		if _, err := os.Stat(strings.TrimSpace(path)); !os.IsNotExist(err) {
			lib = strings.TrimSpace(path)
			break
		}
	}
	if lib != "" {
		logger.Debug("Found softhsm library: %s", lib)
	} else {
		logger.Debug("Softhsm library was not found")
	}
	return lib
}

//SecurityProviderPin will be set only if provider is PKCS11
func (c *CryptoSuiteConfig) SecurityProviderPin() string {
	return c.backend.getString("client.BCCSP.security.pin")
}

//SecurityProviderLabel will be set only if provider is PKCS11
func (c *CryptoSuiteConfig) SecurityProviderLabel() string {
	return c.backend.getString("client.BCCSP.security.label")
}

// KeyStorePath returns the keystore path used by BCCSP
func (c *CryptoSuiteConfig) KeyStorePath() string {
	keystorePath := pathvar.Subst(c.backend.getString("client.credentialStore.cryptoStore.path"))
	return path.Join(keystorePath, "keystore")
}
