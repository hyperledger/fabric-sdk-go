/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package balancer

import (
	"math/rand"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/rollingcounter"
)

var logger = logging.NewLogger("fabsdk/client")

// Balancer is a load-balancing function for peers
type Balancer func(peers []fab.Peer) []fab.Peer

// Random balances peers randomly
func Random() Balancer {
	logger.Debugf("Creating Random balancer")
	return func(peers []fab.Peer) []fab.Peer {
		logger.Debugf("Load balancing %d peers using Random strategy...", len(peers))

		balancedPeers := make([]fab.Peer, len(peers))
		for i, index := range rand.Perm(len(peers)) {
			balancedPeers[i] = peers[index]
		}
		return balancedPeers
	}
}

// RoundRobin balances peers in a round-robin fashion
func RoundRobin() Balancer {
	logger.Debugf("Creating Round-robin balancer")
	counter := rollingcounter.New()
	return func(peers []fab.Peer) []fab.Peer {
		logger.Debugf("Load balancing %d peers using Round-Robin strategy...", len(peers))

		index := counter.Next(len(peers))
		balancedPeers := make([]fab.Peer, len(peers))
		j := 0
		for i := index; i < len(peers); i++ {
			peer := peers[i]
			logger.Debugf("Adding peer [%s] at index %d", peer.URL(), i)
			balancedPeers[j] = peer
			j++
		}
		for i := 0; i < index; i++ {
			peer := peers[i]
			logger.Debugf("Adding peer [%s] at index %d", peer.URL(), i)
			balancedPeers[j] = peers[i]
			j++
		}

		return balancedPeers
	}
}
