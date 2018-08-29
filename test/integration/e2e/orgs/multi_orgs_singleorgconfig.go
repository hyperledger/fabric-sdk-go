// +build !prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/require"
)

//TestMultiOrgWithSingleOrgConfig uses new sdk instance with new config which only has entries for org1.
// this function tests,
// 			if discovered peer has MSP ID found by dynamic discovery service.
// 			if it is able to get endorsement from peers not mentioned in config
//			if tlscacerts are being used by channel block anchor peers if not found in config
func TestMultiOrgWithSingleOrgConfig(t *testing.T, examplecc string) {
	//Config containing references to org1 only
	configProvider := config.FromFile(pathvar.Subst(integration.ConfigPathSingleOrg))
	//if local test, add entity matchers to override URLs to localhost
	if integration.IsLocal() {
		configProvider = integration.AddLocalEntityMapping(configProvider)
	}

	org1sdk, err := fabsdk.New(configProvider)
	if err != nil {
		t.Fatal("failed to created SDK,", err)
	}
	defer org1sdk.Close()

	//prepare context
	org1ChannelClientContext := org1sdk.ChannelContext("orgchannel", fabsdk.WithUser("User1"), fabsdk.WithOrg("Org1"))

	// Org1 user connects to 'orgchannel'
	chClientOrg1User, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org1 user: %s", err)
	}

	req := channel.Request{
		ChaincodeID: examplecc,
		Fcn:         "invoke",
		Args:        integration.ExampleCCDefaultQueryArgs(),
	}
	resp, err := chClientOrg1User.Query(req, channel.WithRetry(retry.DefaultChannelOpts))

	require.NoError(t, err, "query funds failed")

	foundOrg2Endorser := false
	for _, v := range resp.Responses {
		//check if response endorser is org2 peer and MSP ID 'Org2MSP' is found
		if strings.Contains(string(v.Endorsement.Endorser), "Org2MSP") {
			foundOrg2Endorser = true
			break
		}
	}

	require.True(t, foundOrg2Endorser, "Org2 MSP ID was not in the endorsement")
}
