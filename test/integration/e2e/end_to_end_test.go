/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

func TestE2E(t *testing.T) {

	//End to End testing
	Run(t, config.FromFile("../../fixtures/config/config_test.yaml"))

	//Using setup done set above by end to end test, run below test with new config which has no orderer config inside
	runWithNoOrdererConfig(t, integration.ConfigNoOrdererBackend)
}
