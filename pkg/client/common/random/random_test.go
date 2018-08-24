/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package random

import (
	"fmt"
	"testing"

	pfab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/stretchr/testify/assert"
)

func TestPickRandomNPeerConfigs(t *testing.T) {
	counter := 20
	allChPeers := createNChannelPeers(counter)

	result := PickRandomNPeerConfigs(allChPeers, 4)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 4, len(result))
	verifyDuplicates(t, result)

	result = PickRandomNPeerConfigs(allChPeers, 1)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 1, len(result))
	verifyDuplicates(t, result)

	result = PickRandomNPeerConfigs(allChPeers, 19)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 19, len(result))
	verifyDuplicates(t, result)

	result = PickRandomNPeerConfigs(allChPeers, 20)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 20, len(result))
	verifyDuplicates(t, result)

	result = PickRandomNPeerConfigs(allChPeers, 21)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 20, len(result))
	verifyDuplicates(t, result)

	result = PickRandomNPeerConfigs(allChPeers, 24)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 20, len(result))
	verifyDuplicates(t, result)

	counter = 7
	allChPeers = createNChannelPeers(counter)

	result = PickRandomNPeerConfigs(allChPeers, 6)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 6, len(result))
	verifyDuplicates(t, result)

	result = PickRandomNPeerConfigs(allChPeers, 7)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 7, len(result))
	verifyDuplicates(t, result)

	result = PickRandomNPeerConfigs(allChPeers, 8)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 7, len(result))
	verifyDuplicates(t, result)

	counter = 2
	allChPeers = createNChannelPeers(counter)

	result = PickRandomNPeerConfigs(allChPeers, 2)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 2, len(result))
	verifyDuplicates(t, result)

	result = PickRandomNPeerConfigs(allChPeers, 24)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 2, len(result))
	verifyDuplicates(t, result)

	counter = 1
	allChPeers = createNChannelPeers(counter)

	result = PickRandomNPeerConfigs(allChPeers, 1)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 1, len(result))
	verifyDuplicates(t, result)

	result = PickRandomNPeerConfigs(allChPeers, 2)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 1, len(result))
	verifyDuplicates(t, result)

	result = PickRandomNPeerConfigs(allChPeers, 24)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 1, len(result))
	verifyDuplicates(t, result)

}

func createNChannelPeers(n int) []pfab.ChannelPeer {
	allChPeers := make([]pfab.ChannelPeer, n)
	for i := 0; i < n; i++ {
		allChPeers[i] = pfab.ChannelPeer{
			NetworkPeer: pfab.NetworkPeer{
				PeerConfig: pfab.PeerConfig{URL: fmt.Sprintf("URL-%d", i)},
			},
		}
	}
	return allChPeers
}

func verifyDuplicates(t *testing.T, chPeers []pfab.PeerConfig) {
	seen := make(map[string]bool)
	for _, v := range chPeers {
		if seen[v.URL] {
			t.Fatalf("found duplicate channel peer: %s", v.URL)
		}
		seen[v.URL] = true
	}
}

func TestPickMorePeersThanChannelPeers(t *testing.T) {

	// create 2 peers
	allChPeers := createNChannelPeers(2)

	// Ask for 4 peers
	result := PickRandomNPeerConfigs(allChPeers, 4)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
	assert.Equal(t, 2, len(result))
	verifyDuplicates(t, result)
}
