/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/random"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
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
	channelID  string
	membership fab.ChannelMembership
}

// NewChannelService creates a Discovery Service to query the list of member peers on a given channel.
func NewChannelService(ctx contextAPI.Client, membership fab.ChannelMembership, channelID string, opts ...coptions.Opt) (*ChannelService, error) {
	logger.Debug("Creating new dynamic discovery service")
	s := &ChannelService{
		channelID:  channelID,
		membership: membership,
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
	peers, err := s.doQueryPeers()

	if err != nil && s.ErrHandler != nil {
		logger.Infof("[%s] Got error from discovery query: %s. Invoking error handler", s.channelID, err)
		s.ErrHandler(s.ctx, s.channelID, err)
	}

	return peers, err
}

func (s *ChannelService) doQueryPeers() ([]fab.Peer, error) {
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

	req := fabdiscovery.NewRequest().OfChannel(s.channelID).AddPeersQuery()
	responsesCh, err := s.discoveryClient().Send(reqCtx, req, targets...)

	if err != nil {
		return nil, errors.Wrapf(err, "error calling discover service send")
	}

	var respErrors []error

	for resp := range responsesCh {
		peers, err := s.evaluate(ctx, resp)

		if err == nil {
			//got successful response, cancel all outstanding requests to other targets
			cancel()

			return peers, nil
		}

		respErrors = append(respErrors, err)
	}

	return nil, errors.Wrap(multi.New(respErrors...), "no successful response received from any peer")
}

func (s *ChannelService) getTargets(ctx contextAPI.Client) ([]fab.PeerConfig, error) {
	chPeers := ctx.EndpointConfig().ChannelPeers(s.channelID)
	if len(chPeers) == 0 {
		return nil, errors.Errorf("no channel peers configured for channel [%s]", s.channelID)
	}

	chConfig := ctx.EndpointConfig().ChannelConfig(s.channelID)

	//pick number of peers given in channel policy
	return random.PickRandomNPeerConfigs(chPeers, chConfig.Policies.Discovery.MaxTargets), nil
}

// evaluate validates the responses and returns the peers
func (s *ChannelService) evaluate(clientCtx contextAPI.Client, response fabdiscovery.Response) ([]fab.Peer, error) {
	if err := response.Error(); err != nil {
		logger.Warnf("error from discovery request [%s]: %s", response.Target(), err)
		return nil, newDiscoveryError(err, response.Target())
	}

	endpoints, err := response.ForChannel(s.channelID).Peers()

	if err != nil {
		logger.Warnf("error getting peers from discovery response. target: %s. %s", response.Target(), err)
		return nil, newDiscoveryError(err, response.Target())
	}

	// TODO: In a future patch:
	// - validate the signatures in the responses
	// For now just pick the first successful response

	return s.asPeers(clientCtx, endpoints), nil
}

func (s *ChannelService) asPeers(ctx contextAPI.Client, endpoints []*discclient.Peer) []fab.Peer {
	var peers []fab.Peer
	for _, endpoint := range endpoints {
		peer, ok := asPeer(ctx, endpoint)
		if !ok {
			continue
		}

		//check if cache is updated with tlscert if this is a new org joined and membership is not done yet updating cache
		if s.membership.ContainsMSP(peer.MSPID()) {
			peers = append(peers, &peerEndpoint{
				Peer:        peer,
				blockHeight: endpoint.StateInfoMessage.GetStateInfo().GetProperties().LedgerHeight,
			})
		}
	}
	return peers
}

type peerEndpoint struct {
	fab.Peer
	blockHeight uint64
}

func (p *peerEndpoint) BlockHeight() uint64 {
	return p.blockHeight
}
