/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package fab

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/integration/util/runner"
)

var mainSDK *fabsdk.FabricSDK
var mainTestSetup *integration.BaseSetupImpl
var mainChaincodeID string

const (
	org1Name = "Org1"
	org1User = "User1"
)

func TestMain(m *testing.M) {
	r := runner.NewWithExampleCC()
	r.Initialize()
	mainSDK = r.SDK()
	mainTestSetup = r.TestSetup()
	mainChaincodeID = r.ExampleChaincodeID()

	r.Run(m)
}
