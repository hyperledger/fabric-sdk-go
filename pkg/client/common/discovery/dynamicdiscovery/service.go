/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"context"
	"sync"
	"time"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	reqContext "github.com/hyperledger/fabric-sdk-go/pkg/context"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

type discoverClient interface {
	Send(ctx context.Context, req *discclient.Request, targets ...fab.PeerConfig) ([]fabdiscovery.Response, error)
}

// clientProvider is overridden by unit tests
var clientProvider = func(ctx contextAPI.Client) (discoverClient, error) {
	return fabdiscovery.New(ctx)
}

// Service implements a dynamic Discovery Service that queries
// Fabric's Discovery service for information about the peers that
// are currently joined to the given channel.
type Service struct {
	responseTimeout time.Duration
	lock            sync.RWMutex
	chContext       contextAPI.Channel
	discClient      discoverClient
	peersRef        *lazyref.Reference
}

// NewService creates a Discovery Service to query the list of member peers on a given channel.
func newService(options options) *Service {
	logger.Debugf("Creating new dynamic discovery service with cache refresh interval %s", options.refreshInterval)

	s := &Service{
		responseTimeout: options.responseTimeout,
	}
	s.peersRef = lazyref.New(
		func() (interface{}, error) {
			return s.queryPeers()
		},
		lazyref.WithRefreshInterval(lazyref.InitOnFirstAccess, options.refreshInterval),
	)
	return s
}

// Initialize initializes the service with channel context
func (s *Service) Initialize(ctx contextAPI.Channel) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.chContext != nil {
		// Already initialized
		logger.Debugf("Already initialized with context: %#v", s.chContext)
		return nil
	}

	discoverClient, err := clientProvider(ctx)
	if err != nil {
		return errors.Wrapf(err, "error creating discover client")
	}

	logger.Debugf("Initializing with context: %#v", ctx)
	s.chContext = ctx
	s.discClient = discoverClient
	return nil
}

// Close stops the lazyref background refresh
func (s *Service) Close() {
	logger.Debugf("Closing peers ref...")
	s.peersRef.Close()
}

// GetPeers returns the available peers for the channel
func (s *Service) GetPeers() ([]fab.Peer, error) {
	refValue, err := s.peersRef.Get()
	if err != nil {
		return nil, err
	}
	peers, ok := refValue.([]fab.Peer)
	if !ok {
		return nil, errors.New("get peersRef didn't return Peer type")
	}
	return peers, nil
}

func (s *Service) channelContext() contextAPI.Channel {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.chContext
}

func (s *Service) discoverClient() discoverClient {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.discClient
}

func (s *Service) queryPeers() ([]fab.Peer, error) {
	logger.Debugf("Refreshing peers of channel [%s] from discovery service...", s.chContext.ChannelID())

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
	responses, err := s.discoverClient().Send(reqCtx, req, targets...)
	if err != nil {
		if len(responses) == 0 {
			return nil, errors.Wrapf(err, "error calling discover service send")
		}
		logger.Warnf("Received %d response(s) and one or more errors from discovery client: %s", len(responses), err)
	}
	return s.evaluate(channelContext, responses)
}

func (s *Service) getTargets(ctx contextAPI.Channel) ([]fab.PeerConfig, error) {
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
func (s *Service) evaluate(ctx contextAPI.Channel, responses []fabdiscovery.Response) ([]fab.Peer, error) {
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

func asPeers(ctx contextAPI.Client, endpoints []*discclient.Peer) []fab.Peer {
	var peers []fab.Peer
	for _, endpoint := range endpoints {
		url := endpoint.AliveMessage.GetAliveMsg().Membership.Endpoint

		logger.Debugf("Adding endpoint [%s]", url)

		peerConfig, err := ctx.EndpointConfig().PeerConfigByURL(url)
		if err != nil {
			logger.Warnf("Error getting peer config for url [%s]: %s", err)
			continue
		}
		if peerConfig == nil {
			logger.Warnf("Unable to resolve peer config for [%s]", url)
			continue
		}
		peer, err := ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: *peerConfig, MSPID: endpoint.MSPID})
		if err != nil {
			logger.Warnf("Unable to create peer config for [%s]: %s", url, err)
			continue
		}
		peers = append(peers, peer)
	}

	return peers
}
