/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoint

import (
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

var logger = logging.NewLogger("fabsdk/fab")

// DiscoveryWrapper wraps a target discovery service and adds endpoint data to each
// of the discovered peers.
type DiscoveryWrapper struct {
	fab.DiscoveryService
	ctx     context.Client
	chPeers []fab.ChannelPeer
	filter  fab.TargetFilter
}

// Opt is a discoveryProvider option
type Opt func(p *DiscoveryWrapper)

// WithTargetFilter applies the target filter to the discovery provider
func WithTargetFilter(filter fab.TargetFilter) Opt {
	return func(p *DiscoveryWrapper) {
		p.filter = filter
	}
}

// NewEndpointDiscoveryWrapper returns a new event endpoint discovery service
// that wraps a given target discovery service and adds endpoint data to each
// of the discovered peers.
func NewEndpointDiscoveryWrapper(ctx context.Client, channelID string, target fab.DiscoveryService, opts ...Opt) (*DiscoveryWrapper, error) {
	chpeers := ctx.EndpointConfig().ChannelPeers(channelID)
	if len(chpeers) == 0 {
		return nil, errors.Errorf("no channel peers for channel [%s]", channelID)
	}

	s := &DiscoveryWrapper{
		DiscoveryService: target,
		chPeers:          chpeers,
		ctx:              ctx,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.filter != nil {
		s.DiscoveryService = discovery.NewDiscoveryFilterService(target, s.filter)
	}

	return s, nil
}

// GetPeers returns the discovered peers
func (s *DiscoveryWrapper) GetPeers() ([]fab.Peer, error) {
	var eventEndpoints []fab.Peer

	peers, err := s.DiscoveryService.GetPeers()
	if err != nil {
		return nil, err
	}

	for _, peer := range peers {

		var peerConfig *fab.PeerConfig
		var found bool

		chPeer := s.getChannelPeer(peer.URL())
		if chPeer != nil {
			peerConfig = &chPeer.PeerConfig
		} else {
			peerConfig, found = s.ctx.EndpointConfig().PeerConfig(peer.URL())
			if !found {
				continue
			}
			chPeer = s.getChannelPeer(peerConfig.URL)
		}

		logger.Debugf("Channel peer config for [%s]: %#v", peer.URL(), chPeer)

		if chPeer != nil && !chPeer.EventSource {
			logger.Debugf("Excluding peer [%s] since it is not configured as an event source", peer.URL())
			continue
		}

		eventEndpoint := FromPeerConfig(s.ctx.EndpointConfig(), peer, peerConfig)
		eventEndpoints = append(eventEndpoints, eventEndpoint)
	}

	return eventEndpoints, nil
}

func (s *DiscoveryWrapper) getChannelPeer(url string) *fab.ChannelPeer {
	for _, chpeer := range s.chPeers {
		if chpeer.URL == url {
			return &fab.ChannelPeer{
				PeerChannelConfig: chpeer.PeerChannelConfig,
				NetworkPeer:       chpeer.NetworkPeer,
			}
		}
	}
	return nil
}
