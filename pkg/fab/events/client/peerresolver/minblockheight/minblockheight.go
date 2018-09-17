/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package minblockheight

import (
	"math"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service"
)

var logger = logging.NewLogger("fabsdk/fab")

// PeerResolver is a peer resolver that chooses the best peer according to a block height lag threshold.
// The maximum block height of all peers is determined and the peers whose block heights are under
// the maximum height but above a provided "lag" threshold are load balanced. The other peers are
// not considered.
type PeerResolver struct {
	*params
	dispatcher service.Dispatcher
}

// NewResolver returns a new "min block height" peer resolver provider.
func NewResolver() peerresolver.Provider {
	return func(ed service.Dispatcher, context context.Client, channelID string, opts ...options.Opt) peerresolver.Resolver {
		return New(ed, context, channelID, opts...)
	}
}

// New returns a new "min block height" peer resolver.
func New(dispatcher service.Dispatcher, context context.Client, channelID string, opts ...options.Opt) *PeerResolver {
	params := defaultParams(context, channelID)
	options.Apply(params, opts)

	logger.Debugf("Creating new min block height peer resolver with options: blockHeightLagThreshold: %d, reconnectBlockHeightLagThreshold: %d", params.blockHeightLagThreshold, params.reconnectBlockHeightLagThreshold)

	return &PeerResolver{
		params:     params,
		dispatcher: dispatcher,
	}
}

// Resolve returns the best peer according to a block height lag threshold. The maximum block height of
// all peers is determined and the peers that are within a provided "lag" threshold are load balanced.
func (r *PeerResolver) Resolve(peers []fab.Peer) (fab.Peer, error) {
	return r.loadBalancePolicy.Choose(r.Filter(peers))
}

// Filter returns the peers that are within a provided "lag" threshold from the highest block height of all peers.
func (r *PeerResolver) Filter(peers []fab.Peer) []fab.Peer {
	var minBlockHeight uint64
	if r.minBlockHeight > 0 {
		lastBlockReceived := r.dispatcher.LastBlockNum()
		if lastBlockReceived == math.MaxUint64 {
			// No blocks received yet
			logger.Debugf("Min block height was specified: %d", r.minBlockHeight)
			minBlockHeight = r.minBlockHeight
		} else {
			// Make sure minBlockHeight is greater than the last block received
			if r.minBlockHeight > lastBlockReceived {
				minBlockHeight = r.minBlockHeight
			} else {
				minBlockHeight = lastBlockReceived + 1
				logger.Debugf("Min block height was specified as %d but the last block received was %d. Setting min height to %d", r.minBlockHeight, lastBlockReceived, minBlockHeight)
			}
		}
	}

	retPeers := r.doFilterByBlockHeight(minBlockHeight, peers)
	if len(retPeers) == 0 && minBlockHeight > 0 {
		// The last block that was received may have been the last block in the channel. Try again with lastBlock-1.
		logger.Debugf("No peers at the minimum height %d. Trying again with min height %d ...", minBlockHeight, minBlockHeight-1)
		minBlockHeight--
		retPeers = r.doFilterByBlockHeight(minBlockHeight, peers)
		if len(retPeers) == 0 {
			// No peers at the given height. Try again without min height
			logger.Debugf("No peers at the minimum height %d. Trying again without min height ...", minBlockHeight)
			retPeers = r.doFilterByBlockHeight(0, peers)
		}
	}

	return retPeers
}

// ShouldDisconnect checks the current peer's block height relative to the block heights of the
// other peers and disconnects the peer if the configured threshold is reached.
// Returns false if the block height is acceptable; true if the client should be disconnected from the peer
func (r *PeerResolver) ShouldDisconnect(peers []fab.Peer, connectedPeer fab.Peer) bool {
	// Check if the peer should be disconnected
	peerState, ok := connectedPeer.(fab.PeerState)
	if !ok {
		logger.Debugf("Peer does not contain state")
		return false
	}

	lastBlockReceived := r.dispatcher.LastBlockNum()
	connectedPeerBlockHeight := peerState.BlockHeight()

	maxHeight := getMaxBlockHeight(peers)

	logger.Debugf("Block height of connected peer [%s] from Discovery: %d, Last block received: %d, Max block height from Discovery: %d", connectedPeer.URL(), connectedPeerBlockHeight, lastBlockReceived, maxHeight)

	if maxHeight <= uint64(r.reconnectBlockHeightLagThreshold) {
		logger.Debugf("Max block height of peers is %d and reconnect lag threshold is %d so event client will not be disconnected from peer", maxHeight, r.reconnectBlockHeightLagThreshold)
		return false
	}

	// The last block received may be lagging the actual block height of the peer
	if lastBlockReceived+1 < connectedPeerBlockHeight {
		// We can still get more blocks from the connected peer. Don't disconnect
		logger.Debugf("Block height of connected peer [%s] from Discovery is %d which is greater than last block received+1: %d. Won't disconnect from this peer since more blocks can still be retrieved from the peer", connectedPeer.URL(), connectedPeerBlockHeight, lastBlockReceived+1)
		return false
	}

	cutoffHeight := maxHeight - uint64(r.reconnectBlockHeightLagThreshold)
	peerBlockHeight := lastBlockReceived + 1

	if peerBlockHeight >= cutoffHeight {
		logger.Debugf("Block height from connected peer [%s] is %d which is greater than or equal to the cutoff %d so event client will not be disconnected from peer", connectedPeer.URL(), peerBlockHeight, cutoffHeight)
		return false
	}

	logger.Debugf("Block height from connected peer is %d which is less than the cutoff %d. Peer should be disconnected.", peerBlockHeight, cutoffHeight)

	return true
}

func (r *PeerResolver) doFilterByBlockHeight(minBlockHeight uint64, peers []fab.Peer) []fab.Peer {
	var cutoffHeight uint64
	if minBlockHeight > 0 {
		logger.Debugf("Setting cutoff height to be min block height: %d ...", minBlockHeight)
		cutoffHeight = minBlockHeight
	} else {
		if r.blockHeightLagThreshold < 0 || len(peers) == 1 {
			logger.Debugf("Returning all peers")
			return peers
		}

		maxHeight := getMaxBlockHeight(peers)
		logger.Debugf("Max block height of peers: %d", maxHeight)

		if maxHeight <= uint64(r.blockHeightLagThreshold) {
			logger.Debugf("Max block height of peers is %d and lag threshold is %d so returning all peers", maxHeight, r.blockHeightLagThreshold)
			return peers
		}
		cutoffHeight = maxHeight - uint64(r.blockHeightLagThreshold)
	}

	logger.Debugf("Choosing peers whose block heights are at least the cutoff height %d ...", cutoffHeight)

	var retPeers []fab.Peer
	for _, p := range peers {
		peerState, ok := p.(fab.PeerState)
		if !ok {
			logger.Debugf("Accepting peer [%s] since it does not have state (may be a local peer)", p.URL())
			retPeers = append(retPeers, p)
		} else if peerState.BlockHeight() >= cutoffHeight {
			logger.Debugf("Accepting peer [%s] at block height %d which is greater than or equal to the cutoff %d", p.URL(), peerState.BlockHeight(), cutoffHeight)
			retPeers = append(retPeers, p)
		} else {
			logger.Debugf("Rejecting peer [%s] at block height %d which is less than the cutoff %d", p.URL(), peerState.BlockHeight(), cutoffHeight)
		}
	}
	return retPeers
}

func getMaxBlockHeight(peers []fab.Peer) uint64 {
	var maxHeight uint64
	for _, peer := range peers {
		peerState, ok := peer.(fab.PeerState)
		if ok {
			blockHeight := peerState.BlockHeight()
			if blockHeight > maxHeight {
				maxHeight = blockHeight
			}
		}
	}
	return maxHeight
}
