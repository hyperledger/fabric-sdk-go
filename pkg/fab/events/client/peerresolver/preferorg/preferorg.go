/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package preferorg

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver/minblockheight"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service"
)

var logger = logging.NewLogger("fabsdk/fab")

// PeerResolver is a peer resolver that determines which peers are suitable based on block height, although
// will prefer the peers in the current org (as long as their block height is above a configured threshold).
// If none of the peers from the current org are suitable then a peer from another org is chosen.
type PeerResolver struct {
	*params
	mspID               string
	blockHeightResolver *minblockheight.PeerResolver
}

// NewResolver returns a new "prefer org" resolver provider.
func NewResolver() peerresolver.Provider {
	return func(ed service.Dispatcher, context context.Client, channelID string, opts ...options.Opt) peerresolver.Resolver {
		return New(ed, context, channelID, opts...)
	}
}

// New returns a new "prefer org" resolver.
func New(dispatcher service.Dispatcher, context context.Client, channelID string, opts ...options.Opt) *PeerResolver {
	params := defaultParams(context, channelID)
	options.Apply(params, opts)

	mspID := context.Identifier().MSPID

	logger.Debugf("Creating new PreferOrg peer resolver with options: MSP ID [%s]", mspID)

	return &PeerResolver{
		params:              params,
		mspID:               mspID,
		blockHeightResolver: minblockheight.New(dispatcher, context, channelID, opts...),
	}
}

// Resolve uses the MinBlockHeight resolver to choose peers but will prefer peers in the given org.
func (r *PeerResolver) Resolve(peers []fab.Peer) (fab.Peer, error) {
	filteredPeers := r.blockHeightResolver.Filter(peers)

	var orgPeers []fab.Peer
	for _, p := range filteredPeers {
		if p.MSPID() == r.mspID {
			orgPeers = append(orgPeers, p)
		}
	}

	if len(orgPeers) > 0 {
		// Our org is in the list. Use the default balancer to balance between them.
		logger.Debugf("Choosing a peer from [%s]", r.mspID)
		return r.loadBalancePolicy.Choose(orgPeers)
	}

	logger.Debugf("Choosing a peer from another org since there are no peers from [%s] in the list of peers", r.mspID)
	return r.loadBalancePolicy.Choose(filteredPeers)
}

// ShouldDisconnect determines whether connected peer not in our org should be disconnected
// and reconnected to a peer in our org.
func (r *PeerResolver) ShouldDisconnect(peers []fab.Peer, connectedPeer fab.Peer) bool {
	if connectedPeer.MSPID() != r.mspID {
		// We're connected to a peer not in our org. Check if we can connect back to one of our peers.
		logger.Debugf("Currently connected to [%s]. Checking if there are any peers from [%s] that are suitable to connect to", connectedPeer.URL(), r.mspID)

		for _, p := range r.blockHeightResolver.Filter(peers) {
			if p.MSPID() == r.mspID {
				logger.Debugf("Peer [%s] in our preferred org [%s] suitable to connect to so the event client will be disconnected from the peer in the other org [%s]", p.URL(), r.mspID, connectedPeer.URL())
				return true
			}
		}
	}

	logger.Debugf("Using the min block height resolver to determine whether peer [%s] should be disconnected", connectedPeer.URL())
	return r.blockHeightResolver.ShouldDisconnect(peers, connectedPeer)
}
