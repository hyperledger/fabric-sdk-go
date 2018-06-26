/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricselection

import (
	"context"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type selectionFilter struct {
	ctx    contextAPI.Client
	peers  []fab.Peer
	filter options.PeerFilter
}

func newFilter(ctx contextAPI.Client, filter options.PeerFilter, peers []fab.Peer) *selectionFilter {
	return &selectionFilter{
		ctx:    ctx,
		peers:  peers,
		filter: filter,
	}
}

func (s *selectionFilter) Exclude(endpoint discclient.Peer) bool {
	logger.Debugf("Calling peer filter on endpoint [%s]", endpoint.AliveMessage.GetAliveMsg().Membership.Endpoint)

	peer := asPeerValue(s.ctx, &endpoint)

	// The peer must be included in the set of peers returned from fab.DiscoveryService.
	// (Note that DiscoveryService may return a filtered set of peers, depending on how the
	// SDK was configured, so we need to exclude those peers from selection.)
	if !containsPeer(s.peers, peer) {
		logger.Debugf("Excluding peer [%s] since it isn't in the set of peers returned by the discovery service", peer.URL())
		return true
	}

	// Apply the PeerFilter (if any)
	if s.filter != nil && !s.filter(peer) {
		logger.Debugf("Excluding peer [%s] since it was excluded by the peer filter", peer.URL())
		return true
	}

	return false
}

type prioritySelector struct {
	ctx      contextAPI.Client
	selector options.PrioritySelector
}

func newSelector(ctx contextAPI.Client, selector options.PrioritySelector) discclient.PrioritySelector {
	if selector != nil {
		return &prioritySelector{ctx: ctx, selector: selector}
	}
	return discclient.PrioritiesByHeight
}

func (s *prioritySelector) Compare(endpoint1, endpoint2 discclient.Peer) discclient.Priority {
	logger.Debugf("Calling priority selector on endpoint1 [%s] and endpoint2 [%s]", endpoint1.AliveMessage.GetAliveMsg().Membership.Endpoint, endpoint2.AliveMessage.GetAliveMsg().Membership.Endpoint)
	return discclient.Priority(s.selector(asPeerValue(s.ctx, &endpoint1), asPeerValue(s.ctx, &endpoint2)))
}

// asPeerValue converts the discovery endpoint into a light-weight peer value (i.e. without the GRPC config)
// so that it may used by a peer filter
func asPeerValue(ctx contextAPI.Client, endpoint *discclient.Peer) fab.Peer {
	url := endpoint.AliveMessage.GetAliveMsg().GetMembership().Endpoint

	// Get the mapped URL of the peer
	peerConfig, found := ctx.EndpointConfig().PeerConfig(url)
	if found {
		url = peerConfig.URL
	} else {
		logger.Debugf("Peer config not found for url [%s]", url)
	}

	return &peerEndpointValue{
		mspID:       endpoint.MSPID,
		url:         url,
		blockHeight: endpoint.StateInfoMessage.GetStateInfo().GetProperties().LedgerHeight,
	}
}

func containsPeer(peers []fab.Peer, peer fab.Peer) bool {
	for _, p := range peers {
		if p.URL() == peer.URL() {
			return true
		}
	}
	return false
}

type peerEndpointValue struct {
	mspID       string
	url         string
	blockHeight uint64
}

func (p *peerEndpointValue) MSPID() string {
	return p.mspID
}

func (p *peerEndpointValue) URL() string {
	return p.url
}

func (p *peerEndpointValue) BlockHeight() uint64 {
	return p.blockHeight
}

func (p *peerEndpointValue) ProcessTransactionProposal(context.Context, fab.ProcessProposalRequest) (*fab.TransactionProposalResponse, error) {
	panic("not implemented")
}
