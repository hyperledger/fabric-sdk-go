/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockheightsorter

import (
	"sort"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

var logger = logging.NewLogger("fabsdk/client")

// New returns a peer sorter that uses block height and the provided balancer to sort the peers.
// This sorter uses a block-height-lag-threshold property which is the number of blocks from
// the highest block of a group of peers that a peer can lag behind and still be considered to be
// up-to-date. These peers are sorted using the given Balancer.
// If a peer's block height falls behind this "lag" threshold then it will be demoted to a lower
// priority list of peers which will be sorted according to block height.
func New(opts ...coptions.Opt) options.PeerSorter {
	params := defaultParams()
	coptions.Apply(params, opts)

	sorter := &sorter{
		params: params,
	}

	return func(peers []fab.Peer) []fab.Peer {
		return sorter.Sort(peers)
	}
}

type sorter struct {
	*params
}

// Sort sorts the given peers according to block height and lag threshold.
func (f *sorter) Sort(peers []fab.Peer) []fab.Peer {
	if len(peers) <= 1 {
		return peers
	}

	if f.blockHeightLagThreshold < 0 {
		logger.Debugf("Returning all peers")
		return f.balancer(peers)
	}

	maxHeight := getMaxBlockHeight(peers)
	logger.Debugf("Max block height of peers: %d", maxHeight)

	if maxHeight <= uint64(f.blockHeightLagThreshold) {
		logger.Debugf("Max block height of peers is %d and lag threshold is %d so returning peers unsorted", maxHeight, f.blockHeightLagThreshold)
		return f.balancer(peers)
	}

	cutoffHeight := maxHeight - uint64(f.blockHeightLagThreshold)

	logger.Debugf("Choosing peers whose block heights are greater than the cutoff height %d ...", cutoffHeight)

	// preferredPeers are all of the peers that have the same priority
	var preferredPeers []fab.Peer

	// otherPeers are peers that did not make the cutoff
	var otherPeers []fab.Peer

	for _, p := range peers {
		peerState, ok := p.(fab.PeerState)
		if !ok {
			logger.Debugf("Accepting peer [%s] since it does not have state (may be a local peer)", p.URL())
			preferredPeers = append(preferredPeers, p)
		} else if peerState.BlockHeight() >= cutoffHeight {
			logger.Debugf("Accepting peer [%s] at block height %d which is greater than or equal to the cutoff %d", p.URL(), peerState.BlockHeight(), cutoffHeight)
			preferredPeers = append(preferredPeers, p)
		} else {
			logger.Debugf("Rejecting peer [%s] at block height %d which is less than the cutoff %d", p.URL(), peerState.BlockHeight(), cutoffHeight)
			otherPeers = append(otherPeers, p)
		}
	}

	// Apply the balancer on the prefferred peers
	preferredPeers = f.balancer(preferredPeers)

	// Sort the remaining peers in reverse order of block height
	sort.Sort(sort.Reverse(&peerSorter{
		peers: otherPeers,
	}))

	return append(preferredPeers, otherPeers...)
}

func getMaxBlockHeight(channelPeers []fab.Peer) uint64 {
	var maxHeight uint64
	for _, peer := range channelPeers {
		peerState, ok := peer.(fab.PeerState)
		if !ok {
			logger.Debugf("Peer [%s] does not have block state", peer.URL())
			continue
		}
		blockHeight := peerState.BlockHeight()
		if blockHeight > maxHeight {
			maxHeight = blockHeight
		}
	}
	return maxHeight
}

type peers []fab.Peer

type peerSorter struct {
	peers
}

func (es *peerSorter) Len() int {
	return len(es.peers)
}

func (es *peerSorter) Less(i, j int) bool {
	state1, ok := es.peers[i].(fab.PeerState)
	if !ok {
		return false
	}

	state2, ok := es.peers[j].(fab.PeerState)
	if !ok {
		return false
	}

	return state1.BlockHeight() < state2.BlockHeight()
}

func (es *peerSorter) Swap(i, j int) {
	es.peers[i], es.peers[j] = es.peers[j], es.peers[i]
}
