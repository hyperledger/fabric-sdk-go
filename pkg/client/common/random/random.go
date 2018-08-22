/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package random

import (
	"math/rand"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

//PickRandomNPeerConfigs picks N random  unique peer configs from given channel peer list
func PickRandomNPeerConfigs(chPeers []fab.ChannelPeer, n int) []fab.PeerConfig {

	var result []fab.PeerConfig
	for _, index := range rand.Perm(len(chPeers)) {
		result = append(result, chPeers[index].PeerConfig)
		if len(result) == n {
			break
		}
	}
	return result
}
