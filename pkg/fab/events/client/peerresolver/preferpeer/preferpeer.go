/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package preferpeer

import (
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver/minblockheight"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/peerresolver/preferorg"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service"
)

var logger = logging.NewLogger("fabsdk/fab")

// PeerResolver is a peer resolver that determines which peers are suitable based on block height, although
// will prefer the peers in the provided list (as long as their block height is above a configured threshold).
// If none of the peers in the provided list are suitable then an attempt is made to select a peer from the
// current org will be selected. If none of the peers from the current org are suitable then a peer from another
// org is chosen.
type PeerResolver struct {
	*params
	preferredPeers         []string
	preferOrgResolver      *preferorg.PeerResolver
	minBlockHeightResolver *minblockheight.PeerResolver
}

// NewResolver returns a new "prefer peer" resolver provider.
func NewResolver(preferredPeers ...string) peerresolver.Provider {
	return func(ed service.Dispatcher, context context.Client, channelID string, opts ...options.Opt) peerresolver.Resolver {
		return New(ed, context, channelID, preferredPeers, opts...)
	}
}

// New returns a new "prefer peer" resolver.
func New(dispatcher service.Dispatcher, context context.Client, channelID string, preferredPeers []string, opts ...options.Opt) *PeerResolver {
	params := defaultParams(context, channelID)
	options.Apply(params, opts)

	logger.Debugf("Creating new PreferPeer peer resolver with options: Preferred Peers [%s]", preferredPeers)

	return &PeerResolver{
		params:                 params,
		preferredPeers:         preferredPeers,
		preferOrgResolver:      preferorg.New(dispatcher, context, channelID, opts...),
		minBlockHeightResolver: minblockheight.New(dispatcher, context, channelID, opts...),
	}
}

// Resolve uses the MinBlockHeight resolver to choose peers but will prefer the ones in the list of preferred peers.
func (r *PeerResolver) Resolve(peers []fab.Peer) (fab.Peer, error) {
	preferredPeers := r.getPreferredPeers(r.minBlockHeightResolver.Filter(peers))
	if len(preferredPeers) > 0 {
		// At least one of our preferred peers is suitable. Use the default balancer to balance between them.
		logger.Debugf("Choosing a peer from the list of preferred peers")
		return r.loadBalancePolicy.Choose(preferredPeers)
	}

	logger.Debugf("There are no suitable peers from the list of preferred peers [%s] so choosing another peer using the 'prefer org' resolver", r.preferredPeers)
	return r.preferOrgResolver.Resolve(peers)
}

// ShouldDisconnect determines whether the connected peer should be disconnected and reconnected to the preferred peer.
func (r *PeerResolver) ShouldDisconnect(peers []fab.Peer, connectedPeer fab.Peer) bool {
	if !r.isPreferred(connectedPeer) {
		// We're not connected to a preferred peer. Check if we can connect back to one.
		logger.Debugf("Currently connected to [%s]. Checking if any of the preferred peers [%s] is suitable to connect back to", connectedPeer.URL(), r.preferredPeers)

		if len(r.getPreferredPeers(r.minBlockHeightResolver.Filter(peers))) > 0 {
			logger.Debugf("At least one of our preferred peers is suitable to connect back to so the event client will be disconnected from peer [%s]", connectedPeer.URL())
			return true
		}

		logger.Debugf("None of our preferred peers is suitable to connect back to so the event client will NOT be disconnected from peer [%s]", connectedPeer.URL())
	}

	logger.Debugf("Using the 'prefer org' resolver to determine whether peer [%s] should be disconnected", connectedPeer.URL())
	return r.preferOrgResolver.ShouldDisconnect(peers, connectedPeer)
}

func (r *PeerResolver) getPreferredPeers(peers []fab.Peer) []fab.Peer {
	var preferredPeers []fab.Peer
	for _, p := range peers {
		if r.isPreferred(p) {
			preferredPeers = append(preferredPeers, p)
		}
	}
	return preferredPeers
}

func (r *PeerResolver) isPreferred(peer fab.Peer) bool {
	for _, url := range r.preferredPeers {
		if strings.Contains(peer.URL(), url) {
			return true
		}
	}
	return false
}
