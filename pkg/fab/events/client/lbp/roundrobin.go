/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lbp

import (
	"math/rand"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// RoundRobin implements a round-robin load-balance policy
type RoundRobin struct {
	sync.Mutex
	index int
}

// NewRoundRobin returns a new RoundRobin load-balance policy
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{
		index: -1,
	}
}

// Choose chooses from the list of peers in round-robin fashion
func (lbp *RoundRobin) Choose(peers []fab.Peer) (fab.Peer, error) {
	if len(peers) == 0 {
		logger.Warn("No peers to choose from!")
		return nil, nil
	}

	lbp.Lock()
	defer lbp.Unlock()

	if lbp.index < 0 {
		// First time - start at a random index
		lbp.index = rand.Intn(len(peers))
	} else {
		lbp.index++
	}

	if lbp.index >= len(peers) {
		lbp.index = 0
	}

	logger.Debugf("Choosing peer at index %d", lbp.index)

	return peers[lbp.index], nil
}
