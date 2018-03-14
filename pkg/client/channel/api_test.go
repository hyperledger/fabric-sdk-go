/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
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
	user := fcmocks.NewMockUserWithMSPID(username, mspID)
	ctx := fcmocks.NewMockContext(user)
	return ctx
}
