/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package configless

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration/e2e"
)

// this test mimics the original e2e test with the difference of injecting EndpointConfig interface functions implementations
// to programmatically supply configs instead of using a yaml file. With this change, application developers can fetch
// configs from any source as long as they provide their own implementations.

func TestE2E(t *testing.T) {
	configPath := "../../../fixtures/config/config_test_crypto_bccsp.yaml"
	//Using same Run call as e2e package but with programmatically overriding interfaces
	e2e.RunWithoutSetup(t, config.FromFile(configPath),
		fabsdk.WithConfigEndpoint(endpointConfigImpls...))

	// TODO test with below line once IdentityConfig and CryptoConfig are split into
	// TODO sub interfaces like EndpointConfig and pass them in like WithConfigEndpoint,
	// TODO this will allow to test overriding all config interfaces without the need of a config file
	// TODO maybe add config.BareBone() in the SDK to get a configProvider without a config file instead
	// TODO of passing in an empty file as in below comment
	// use an empty config file to fully depend on injected EndpointConfig interfaces
	//configPath = "../../../pkg/core/config/testdata/viper-test.yaml"
}
