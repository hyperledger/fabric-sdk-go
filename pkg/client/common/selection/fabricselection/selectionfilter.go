/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricselection

import (
	"context"
	"sort"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/balancer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/sorter/balancedsorter"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/sorter/blockheightsorter"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
)

type selectionFilter struct {
	ctx    contextAPI.Client
	peers  []fab.Peer
	filter options.PeerFilter
	sorter options.PeerSorter
}

var noFilter = func(fab.Peer) bool {
	return true
}

func newFilter(ctx contextAPI.Client, peers []fab.Peer, filter options.PeerFilter, sorter options.PeerSorter) *selectionFilter {
	if filter == nil {
		filter = noFilter
	}

	return &selectionFilter{
		ctx:    ctx,
		peers:  peers,
		filter: filter,
		sorter: sorter,
	}
}

func resolvePeerSorter(channelID string, ctx contextAPI.Client) options.PeerSorter {
	channelConfig := ctx.EndpointConfig().ChannelConfig(channelID)
	return resolveSortingStrategy(channelID, channelConfig, resolveBalancer(channelID, channelConfig))
}

func resolveSortingStrategy(channelID string, channelConfig *fab.ChannelEndpointConfig, balancer balancer.Balancer) options.PeerSorter {
	switch channelConfig.Policies.Selection.SortingStrategy {
	case fab.Balanced:
		logger.Debugf("Using balanced selection sorter for channel [%s]", channelID)
		return balancedsorter.New(balancedsorter.WithBalancer(balancer))
	default:
		logger.Debugf("Using block height priority selection sorter for channel [%s]", channelID)
		return blockheightsorter.New(blockheightsorter.WithBalancer(balancer))
	}
}

func resolveBalancer(channelID string, channelConfig *fab.ChannelEndpointConfig) balancer.Balancer {
	switch channelConfig.Policies.Selection.Balancer {
	case fab.Random:
		logger.Debugf("Using random selection balancer for channel [%s]", channelID)
		return balancer.Random()
	default:
		logger.Debugf("Using round-robin selection balancer for channel [%s]", channelID)
		return balancer.RoundRobin()
	}
}

func (s *selectionFilter) Filter(endorsers discclient.Endorsers) discclient.Endorsers {

	// Convert the endorsers to peers
	peers := s.asPeerValues(endorsers)

	// Sort the peers in alphabetical order so that they are always presented to the balancer in the same order
	peers = s.sortByURL(peers)

	// Filter out the peers that weren't included in the list of discovered peers
	discoveredPeers := s.filterDiscovered(peers)

	// Apply the peer filter (if any)
	filteredPeers := s.filterPeers(discoveredPeers)

	// Apply the peer sorter (if any)
	sortedPeers := s.sortPeers(filteredPeers)

	// Convert the filtered peers to endorsers
	return s.asEndorsers(endorsers, sortedPeers)
}

func (s *selectionFilter) sortPeers(peers []fab.Peer) []fab.Peer {
	if s.sorter == nil {
		return peers
	}

	logger.Debugf("Sorting peers")
	return s.sorter(peers)
}

func (s *selectionFilter) filterPeers(peers []fab.Peer) []fab.Peer {
	if s.filter == nil {
		return peers
	}

	var filteredPeers []fab.Peer
	for _, peer := range peers {
		// Apply the PeerFilter (if any)
		if s.filter(peer) {
			logger.Debugf("Including peer [%s] since it was included by the peer filter", peer.URL())
			filteredPeers = append(filteredPeers, peer)
		} else {
			logger.Debugf("Excluding peer [%s] since it was excluded by the peer filter", peer.URL())
		}
	}
	return filteredPeers
}

func (s *selectionFilter) filterDiscovered(peers []fab.Peer) []fab.Peer {
	// The peer must be included in the set of peers returned from fab.DiscoveryService.
	// (Note that DiscoveryService may return a filtered set of peers, depending on how the
	// SDK was configured, so we need to exclude those peers from selection.)
	var discoveryPeers []fab.Peer
	for _, peer := range peers {
		if containsPeer(s.peers, peer) {
			logger.Debugf("Including peer [%s] since it is in the set of peers returned by the discovery service", peer.URL())
			discoveryPeers = append(discoveryPeers, peer)
		} else {
			logger.Debugf("Excluding peer [%s] since it isn't in the set of peers returned by the discovery service", peer.URL())
		}
	}
	return discoveryPeers
}

// asPeerValue converts the discovery endpoint into a light-weight peer value (i.e. without the GRPC config)
// so that it may used by a peer filter
func (s *selectionFilter) asPeerValue(endpoint *discclient.Peer) fab.Peer {
	url := endpoint.AliveMessage.GetAliveMsg().GetMembership().Endpoint

	// Get the mapped URL of the peer if such defined in EndpointConfig
	peerConfig, found := s.ctx.EndpointConfig().PeerConfig(url)
	if found {
		url = peerConfig.URL
	}

	return &peerEndpointValue{
		mspID:      endpoint.MSPID,
		url:        url,
		properties: fabdiscovery.GetProperties(endpoint),
	}
}

func (s *selectionFilter) asPeerValues(endorsers discclient.Endorsers) []fab.Peer {
	var peers []fab.Peer
	for _, endorser := range endorsers {
		peer := s.asPeerValue(endorser)
		logger.Debugf("Adding peer [%s]", peer.URL())
		peers = append(peers, peer)
	}
	return peers
}

func (s *selectionFilter) sortByURL(peers []fab.Peer) []fab.Peer {
	sort.Sort(&peerSorter{
		peers: peers,
	})
	return peers
}

func (s *selectionFilter) asEndorsers(allEndorsers discclient.Endorsers, filteredPeers []fab.Peer) discclient.Endorsers {
	var filteredEndorsers discclient.Endorsers
	for _, peer := range filteredPeers {
		endorser, found := s.asEndorser(allEndorsers, peer)
		if !found {
			// This should never happen since the peer was composed from the initial list of endorsers
			logger.Warnf("Endorser [%s] not found. Endorser will be excluded.", peer.URL())
			continue
		}
		logger.Debugf("Adding endorser [%s]", peer.URL())
		filteredEndorsers = append(filteredEndorsers, endorser)
	}
	return filteredEndorsers
}

func (s *selectionFilter) asEndorser(endorsers discclient.Endorsers, peer fab.Peer) (*discclient.Peer, bool) {
	for _, endorser := range endorsers {
		url := s.asPeerValue(endorser).URL()
		if peer.URL() == url {
			return endorser, true
		}
	}
	return nil, false
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
	mspID      string
	url        string
	properties fab.Properties
}

func (p *peerEndpointValue) MSPID() string {
	return p.mspID
}

func (p *peerEndpointValue) URL() string {
	return p.url
}

func (p *peerEndpointValue) Properties() fab.Properties {
	return p.properties
}

func (p *peerEndpointValue) BlockHeight() uint64 {
	ledgerHeight, ok := p.properties[fab.PropertyLedgerHeight]
	if !ok {
		return 0
	}

	return ledgerHeight.(uint64)
}

func (p *peerEndpointValue) ProcessTransactionProposal(context.Context, fab.ProcessProposalRequest) (*fab.TransactionProposalResponse, error) {
	panic("not implemented")
}

type peers []fab.Peer

type peerSorter struct {
	peers
}

func (es *peerSorter) Len() int {
	return len(es.peers)
}

func (es *peerSorter) Less(i, j int) bool {
	return es.peers[i].URL() < es.peers[j].URL()
}

func (es *peerSorter) Swap(i, j int) {
	es.peers[i], es.peers[j] = es.peers[j], es.peers[i]
}
