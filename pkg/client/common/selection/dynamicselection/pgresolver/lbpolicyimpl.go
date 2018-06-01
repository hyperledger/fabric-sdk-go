/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pgresolver

import (
	"math/rand"
)

type randomLBP struct {
}

// NewRandomLBP returns a random load-balance policy
func NewRandomLBP() LoadBalancePolicy {
	return &randomLBP{}
}

func (lbp *randomLBP) Choose(peerGroups []PeerGroup) PeerGroup {
	logger.Debug("Invoking random LBP\n")

	if len(peerGroups) == 0 {
		logger.Warn("No available peer groups\n")
		// Return an empty PeerGroup
		return NewPeerGroup()
	}

	index := rand.Intn(len(peerGroups))

	logger.Debugf("randomLBP - Choosing index %d\n", index)
	return peerGroups[index]
}

type roundRobinLBP struct {
	index int
}

// NewRoundRobinLBP returns a round-robin load-balance policy
func NewRoundRobinLBP() LoadBalancePolicy {
	return &roundRobinLBP{index: -1}
}

func (lbp *roundRobinLBP) Choose(peerGroups []PeerGroup) PeerGroup {
	if len(peerGroups) == 0 {
		logger.Warn("No available peer groups\n")
		// Return an empty PeerGroup
		return NewPeerGroup()
	}

	if lbp.index == -1 {
		lbp.index = rand.Intn(len(peerGroups))
	} else {
		lbp.index++
	}

	if lbp.index >= len(peerGroups) {
		lbp.index = 0
	}

	logger.Debugf("roundRobinLBP - Choosing index %d\n", lbp.index)

	return peerGroups[lbp.index]
}
