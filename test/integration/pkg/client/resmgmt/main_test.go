/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package resmgmt

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/integration/util/runner"
	"github.com/stretchr/testify/require"
)

const (
	org1Name      = "Org1"
	org2Name      = "Org2"
	org1AdminUser = "Admin"
	org2AdminUser = "Admin"
	orgChannelID  = "orgchannel"
)

var mainSDK *fabsdk.FabricSDK
var mainTestSetup *integration.BaseSetupImpl
var mainChaincodeID string

func TestMain(m *testing.M) {
	r := runner.NewWithExampleCC()
	r.Initialize()
	mainSDK = r.SDK()
	mainTestSetup = r.TestSetup()
	mainChaincodeID = r.ExampleChaincodeID()

	r.Run(m)
}

func setupMultiOrgContext(t *testing.T, sdk *fabsdk.FabricSDK) []*integration.OrgContext {
	orgContext, err := integration.SetupMultiOrgContext(sdk, org1Name, org2Name, org1AdminUser, org2AdminUser)
	require.NoError(t, err)
	return orgContext
}
