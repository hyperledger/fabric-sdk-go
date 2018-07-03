/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkcs11

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	cryptosuite "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/pkcs11"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/test/integration/e2e"
)

const (
	// ConfigTestFile contains the path and filename of the config for integration tests
	ConfigTestFilename = "config_e2e_pkcs11.yaml"
)

func TestE2E(t *testing.T) {
	// Create SDK setup for the integration tests
	e2e.Run(t,
		config.FromFile(integration.GetConfigPath(ConfigTestFilename)),
		fabsdk.WithCorePkg(&CustomCryptoSuiteProviderFactory{}))
}

// CustomCryptoSuiteProviderFactory is will provide custom cryptosuite (bccsp.BCCSP)
type CustomCryptoSuiteProviderFactory struct {
	defcore.ProviderFactory
}

// CreateCryptoSuiteProvider returns a new default implementation of BCCSP
func (f *CustomCryptoSuiteProviderFactory) CreateCryptoSuiteProvider(config core.CryptoSuiteConfig) (core.CryptoSuite, error) {
	return cryptosuite.GetSuiteByConfig(config)
}
