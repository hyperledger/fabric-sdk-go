/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoint

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

// DiscoveryProvider is a wrapper around a discovery provider that
// converts each peer into an EventEndpoint. The EventEndpoint
// provides additional connection options.
type DiscoveryProvider struct {
	fab.DiscoveryProvider
	ctx    context.Client
	filter fab.TargetFilter
}

// Opt is a discoveryProvider option
type Opt func(p *DiscoveryProvider)

// WithTargetFilter applies the target filter to the discovery provider
func WithTargetFilter(filter fab.TargetFilter) Opt {
	return func(p *DiscoveryProvider) {
		p.filter = filter
	}
}

// NewDiscoveryProvider returns a new event endpoint discovery provider
func NewDiscoveryProvider(ctx context.Client, opts ...Opt) *DiscoveryProvider {
	p := &DiscoveryProvider{
		DiscoveryProvider: ctx.DiscoveryProvider(),
		ctx:               ctx,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// CreateDiscoveryService creates a new DiscoveryService for the given channel
func (p *DiscoveryProvider) CreateDiscoveryService(channelID string) (fab.DiscoveryService, error) {
	target, err := p.DiscoveryProvider.CreateDiscoveryService(channelID)
	if err != nil {
		return nil, err
	}

	if p.filter != nil {
		target = discovery.NewDiscoveryFilterService(target, p.filter)
	}

	chpeers, err := p.ctx.Config().ChannelPeers(channelID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get channel peers for channel [%s]", channelID)
	}

	return &discoveryService{
		DiscoveryService: target,
		ctx:              p.ctx,
		channelID:        channelID,
		chPeers:          chpeers,
	}, nil
}

type discoveryService struct {
	fab.DiscoveryService
	ctx       context.Client
	channelID string
	chPeers   []core.ChannelPeer
}

func (s *discoveryService) GetPeers() ([]fab.Peer, error) {
	var eventEndpoints []fab.Peer

	peers, err := s.DiscoveryService.GetPeers()
	if err != nil {
		return nil, err
	}

	for _, peer := range peers {
		peerConfig, err := s.ctx.Config().PeerConfigByURL(peer.URL())
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get peer config from [%s]", peer.URL())
		}
		if peerConfig == nil {
			return nil, errors.Errorf("unable to get peer config from [%s]", peer.URL())
		}

		chPeer := s.getChannelPeer(peerConfig)

		logger.Debugf("Channel peer config for [%s]: %#v", peer.URL(), chPeer)

		if chPeer != nil && !chPeer.EventSource {
			logger.Debugf("Excluding peer [%s] since it is not configured as an event source", peer.URL())
			continue
		}

		eventEndpoint, err := FromPeerConfig(s.ctx.Config(), peer, peerConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create event endpoint for [%s]", peer.URL())
		}
		eventEndpoints = append(eventEndpoints, eventEndpoint)
	}

	return eventEndpoints, nil
}

func (s *discoveryService) getChannelPeer(peerConfig *core.PeerConfig) *core.ChannelPeer {
	for _, chpeer := range s.chPeers {
		if chpeer.URL == peerConfig.URL {
			return &chpeer
		}
	}
	return nil
}
