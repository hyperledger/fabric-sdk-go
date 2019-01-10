/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package configless

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration/e2e"
)

// this test mimics the original e2e test with the difference of injecting interface functions implementations
// to programmatically supply configs instead of using a yaml file. With this change, application developers can fetch
// configs from any source as long as they provide their own implementations.

func TestE2E(t *testing.T) {

	//Using same Run call as e2e package but with programmatically overriding interfaces
	// since in this configless test, we are overriding all the config's interfaces, there's no need to add a configProvider
	//
	// But if someone wants to partially override the configs interfaces (by setting only some functions of either
	// EndpointConfig, CryptoSuiteConfig and/or IdentityConfig) then they need to provide a configProvider
	// with a config file that contains at least the sections that are not overridden by the provided functions
	e2e.RunWithoutSetup(t, nil,
		fabsdk.WithEndpointConfig(endpointConfigImpls...),
		fabsdk.WithCryptoSuiteConfig(cryptoConfigImpls...),
		fabsdk.WithIdentityConfig(identityConfigImpls...),
		fabsdk.WithMetricsConfig(operationsConfigImpls...),
	)
}
