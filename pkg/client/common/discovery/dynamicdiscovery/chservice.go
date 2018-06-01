/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	reqContext "github.com/hyperledger/fabric-sdk-go/pkg/context"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	"github.com/pkg/errors"
)

// ChannelService implements a dynamic Discovery Service that queries
// Fabric's Discovery service for information about the peers that
// are currently joined to the given channel.
type ChannelService struct {
	*service
	channelID string
}

// NewChannelService creates a Discovery Service to query the list of member peers on a given channel.
func NewChannelService(ctx contextAPI.Client, channelID string, opts ...coptions.Opt) (*ChannelService, error) {
	logger.Debug("Creating new dynamic discovery service")
	s := &ChannelService{
		channelID: channelID,
	}
	s.service = newService(ctx.EndpointConfig(), s.queryPeers, opts...)
	err := s.service.initialize(ctx)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Close releases resources
func (s *ChannelService) Close() {
	logger.Debugf("Closing discovery service for channel [%s]", s.channelID)
	s.service.Close()
}

func (s *ChannelService) queryPeers() ([]fab.Peer, error) {
	logger.Debugf("Refreshing peers of channel [%s] from discovery service...", s.channelID)

	ctx := s.context()

	targets, err := s.getTargets(ctx)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, errors.Errorf("no peers configured for channel [%s]", s.channelID)
	}

	reqCtx, cancel := reqContext.NewRequest(ctx, reqContext.WithTimeout(s.responseTimeout))
	defer cancel()

	req := discclient.NewRequest().OfChannel(s.channelID).AddPeersQuery()
	responses, err := s.discoveryClient().Send(reqCtx, req, targets...)
	if err != nil {
		if len(responses) == 0 {
			return nil, errors.Wrapf(err, "error calling discover service send")
		}
		logger.Warnf("Received %d response(s) and one or more errors from discovery client: %s", len(responses), err)
	}
	return s.evaluate(ctx, responses)
}

func (s *ChannelService) getTargets(ctx contextAPI.Client) ([]fab.PeerConfig, error) {
	// TODO: The number of peers to query should be retrieved from the channel policy.
	// This will done in a future patch.
	chpeers, ok := ctx.EndpointConfig().ChannelPeers(s.channelID)
	if !ok {
		return nil, errors.Errorf("failed to get peer configs for channel [%s]", s.channelID)
	}
	targets := make([]fab.PeerConfig, len(chpeers))
	for i := 0; i < len(targets); i++ {
		targets[i] = chpeers[i].NetworkPeer.PeerConfig
	}
	return targets, nil
}

// evaluate validates the responses and returns the peers
func (s *ChannelService) evaluate(ctx contextAPI.Client, responses []fabdiscovery.Response) ([]fab.Peer, error) {
	if len(responses) == 0 {
		return nil, errors.New("no successful response received from any peer")
	}

	// TODO: In a future patch:
	// - validate the signatures in the responses
	// For now just pick the first successful response

	var lastErr error
	for _, response := range responses {
		endpoints, err := response.ForChannel(s.channelID).Peers()
		if err != nil {
			lastErr = errors.Wrap(err, "error getting peers from discovery response")
			logger.Warn(lastErr.Error())
			continue
		}
		return asPeers(ctx, endpoints), nil
	}
	return nil, lastErr
}
