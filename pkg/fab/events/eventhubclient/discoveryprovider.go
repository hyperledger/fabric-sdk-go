/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package eventhubclient

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/endpoint"
	"github.com/pkg/errors"
)

// discoveryProvider is a wrapper around the discovery provider that
// converts each peer into an EventEndpoint (which provides the event URL).
type discoveryProvider struct {
	fab.DiscoveryProvider
	ctx context.Client
}

func newDiscoveryProvider(ctx context.Client) *discoveryProvider {
	return &discoveryProvider{
		DiscoveryProvider: ctx.DiscoveryProvider(),
		ctx:               ctx,
	}
}

// CreateDiscoveryService creates a new DiscoveryService for the given channel
func (p *discoveryProvider) CreateDiscoveryService(channelID string) (fab.DiscoveryService, error) {
	target, err := p.DiscoveryProvider.CreateDiscoveryService(channelID)
	if err != nil {
		return nil, err
	}
	return &discoveryService{
		DiscoveryService: target,
		ctx:              p.ctx,
	}, nil
}

type discoveryService struct {
	fab.DiscoveryService
	ctx context.Client
}

func (s *discoveryService) GetPeers() ([]fab.Peer, error) {
	var eventEndpoints []fab.Peer

	peers, err := s.DiscoveryService.GetPeers()
	if err != nil {
		return nil, err
	}

	// Choose only the peers from the MSP in context
	// since Event Hub connections are only allowed
	// using the local MSP.
	mspID := s.ctx.MspID()

	for _, peer := range peers {
		if peer.MSPID() != mspID {
			continue
		}

		peerConfig, err := s.ctx.Config().PeerConfigByURL(peer.URL())
		if err != nil {
			return nil, errors.Wrapf(err, "unable to determine event hub URL from [%s]", peer.URL())
		}
		if peerConfig == nil {
			return nil, errors.Errorf("unable to determine event hub URL from [%s]", peer.URL())
		}

		eventEndpoints = append(eventEndpoints,
			&endpoint.EventEndpoint{
				Peer:   peer,
				EvtURL: peerConfig.EventURL,
			},
		)
	}

	return eventEndpoints, nil
}
