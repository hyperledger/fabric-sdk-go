/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	reqContext "github.com/hyperledger/fabric-sdk-go/pkg/context"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	"github.com/pkg/errors"
)

// channelService implements a dynamic Discovery Service that queries
// Fabric's Discovery service for information about the peers that
// are currently joined to the given channel.
type channelService struct {
	*service
}

// newChannelService creates a Discovery Service to query the list of member peers on a given channel.
func newChannelService(options options) *channelService {
	logger.Debugf("Creating new dynamic discovery service with cache refresh interval %s", options.refreshInterval)

	s := &channelService{}
	s.service = newService(s.queryPeers, options)
	return s
}

// Initialize initializes the service with channel context
func (s *channelService) Initialize(ctx contextAPI.Channel) error {
	return s.service.Initialize(ctx)
}

func (s *channelService) channelContext() contextAPI.Channel {
	return s.context().(contextAPI.Channel)
}

func (s *channelService) queryPeers() ([]fab.Peer, error) {
	logger.Debugf("Refreshing peers of channel [%s] from discovery service...", s.channelContext().ChannelID())

	channelContext := s.channelContext()
	if channelContext == nil {
		return nil, errors.Errorf("the service has not been initialized")
	}

	targets, err := s.getTargets(channelContext)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, errors.Errorf("no peers configured for channel [%s]", channelContext.ChannelID())
	}

	reqCtx, cancel := reqContext.NewRequest(channelContext, reqContext.WithTimeout(s.responseTimeout))
	defer cancel()

	req := discclient.NewRequest().OfChannel(channelContext.ChannelID()).AddPeersQuery()
	responses, err := s.discoveryClient().Send(reqCtx, req, targets...)
	if err != nil {
		if len(responses) == 0 {
			return nil, errors.Wrapf(err, "error calling discover service send")
		}
		logger.Warnf("Received %d response(s) and one or more errors from discovery client: %s", len(responses), err)
	}
	return s.evaluate(channelContext, responses)
}

func (s *channelService) getTargets(ctx contextAPI.Channel) ([]fab.PeerConfig, error) {
	// TODO: The number of peers to query should be retrieved from the channel policy.
	// This will done in a future patch.
	chpeers, err := ctx.EndpointConfig().ChannelPeers(ctx.ChannelID())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get peer configs for channel [%s]", ctx.ChannelID())
	}
	targets := make([]fab.PeerConfig, len(chpeers))
	for i := 0; i < len(targets); i++ {
		targets[i] = chpeers[i].NetworkPeer.PeerConfig
	}
	return targets, nil
}

// evaluate validates the responses and returns the peers
func (s *channelService) evaluate(ctx contextAPI.Channel, responses []fabdiscovery.Response) ([]fab.Peer, error) {
	if len(responses) == 0 {
		return nil, errors.New("no successful response received from any peer")
	}

	// TODO: In a future patch:
	// - validate the signatures in the responses
	// - ensure N responses match according to the policy
	// For now just pick the first response
	response := responses[0]
	endpoints, err := response.ForChannel(ctx.ChannelID()).Peers()
	if err != nil {
		return nil, errors.Wrapf(err, "error getting peers from discovery response")
	}

	return asPeers(ctx, endpoints), nil
}
