/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"testing"

	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/mocks"
	"github.com/stretchr/testify/assert"
)

func TestWithTargetURLsInvalid(t *testing.T) {
	ctx := setupMockTestContext("test", "Org1MSP")
	opt := WithTargetURLs("invalid")

	mockConfig := &fcmocks.MockConfig{}

	oConfig := &core.PeerConfig{
		URL: "127.0.0.1:7050",
	}

	mockConfig.SetCustomPeerCfg(oConfig)
	ctx.SetConfig(mockConfig)

	opts := requestOptions{}
	err := opt(ctx, &opts)
	assert.NotNil(t, err, "Should have failed for invalid target peer")
}

func TestWithTargetURLsValid(t *testing.T) {
	ctx := setupMockTestContext("test", "Org1MSP")
	opt := WithTargetURLs("127.0.0.1:7050")

	mockConfig := &fcmocks.MockConfig{}

	pConfig1 := core.PeerConfig{
		URL: "127.0.0.1:7050",
	}

	npConfig1 := core.NetworkPeer{
		PeerConfig: pConfig1,
		MSPID:      "MYMSP",
	}

	pConfig2 := core.PeerConfig{
		URL: "127.0.0.1:7051",
	}

	npConfig2 := core.NetworkPeer{
		PeerConfig: pConfig2,
		MSPID:      "OTHERMSP",
	}

	mockConfig.SetCustomPeerCfg(&pConfig1)
	mockConfig.SetCustomNetworkPeerCfg([]core.NetworkPeer{npConfig2, npConfig1})
	ctx.SetConfig(mockConfig)

	opts := requestOptions{}
	err := opt(ctx, &opts)
	assert.Nil(t, err, "Should have failed for invalid target peer")

	assert.Equal(t, 1, len(opts.Targets), "should have one peer")
	assert.Equal(t, pConfig1.URL, opts.Targets[0].URL(), "", "Wrong URL")
	assert.Equal(t, npConfig1.MSPID, opts.Targets[0].MSPID(), "", "Wrong MSP")
}

func setupMockTestContext(username string, mspID string) *fcmocks.MockContext {
	user := mspmocks.NewMockSigningIdentity(username, mspID)
	ctx := fcmocks.NewMockContext(user)
	return ctx
}

func TestTimeoutOptions(t *testing.T) {

	opts := requestOptions{}

	options := []RequestOption{WithTimeout(core.PeerResponse, 20*time.Second),
		WithTimeout(core.ResMgmt, 25*time.Second), WithTimeout(core.OrdererResponse, 30*time.Second),
		WithTimeout(core.EventHubConnection, 35*time.Second), WithTimeout(core.Execute, 40*time.Second),
		WithTimeout(core.Query, 45*time.Second)}

	for _, option := range options {
		option(nil, &opts)
	}

	assert.True(t, opts.Timeouts[core.PeerResponse] == 20*time.Second, "timeout value by type didn't match with one supplied")
	assert.True(t, opts.Timeouts[core.ResMgmt] == 25*time.Second, "timeout value by type didn't match with one supplied")
	assert.True(t, opts.Timeouts[core.OrdererResponse] == 30*time.Second, "timeout value by type didn't match with one supplied")
	assert.True(t, opts.Timeouts[core.EventHubConnection] == 35*time.Second, "timeout value by type didn't match with one supplied")
	assert.True(t, opts.Timeouts[core.Execute] == 40*time.Second, "timeout value by type didn't match with one supplied")
	assert.True(t, opts.Timeouts[core.Query] == 45*time.Second, "timeout value by type didn't match with one supplied")

}
