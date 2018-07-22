/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package selection

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/integration/util/runner"
	"github.com/stretchr/testify/require"
)

const (
	org1Name     = "Org1"
	org2Name     = "Org2"
	adminUser    = "Admin"
	org1User     = "User1"
	orgChannelID = "orgchannel"
)

var mainSDK *fabsdk.FabricSDK
var mainTestSetup *integration.BaseSetupImpl

func TestMain(m *testing.M) {
	r := runner.New()
	r.Initialize()
	mainSDK = r.SDK()
	mainTestSetup = r.TestSetup()

	r.Run(m)
}

func setupMultiOrgContext(t *testing.T, sdk *fabsdk.FabricSDK) []*integration.OrgContext {
	orgContext, err := integration.SetupMultiOrgContext(sdk, org1Name, org2Name, adminUser, adminUser)
	require.NoError(t, err)

	return orgContext
}
