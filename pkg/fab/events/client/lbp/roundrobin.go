/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lbp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/rollingcounter"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// RoundRobin implements a round-robin load-balance policy
type RoundRobin struct {
	counter *rollingcounter.Counter
}

// NewRoundRobin returns a new RoundRobin load-balance policy
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{
		counter: rollingcounter.New(),
	}
}

// Choose chooses from the list of peers in round-robin fashion
func (lbp *RoundRobin) Choose(peers []fab.Peer) (fab.Peer, error) {
	if len(peers) == 0 {
		logger.Warn("No peers to choose from!")
		return nil, nil
	}
	return peers[lbp.counter.Next(len(peers))], nil
}
