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
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

type discoveryClient interface {
	Send(ctx context.Context, req *discclient.Request, targets ...fab.PeerConfig) ([]fabdiscovery.Response, error)
}

// clientProvider is overridden by unit tests
var clientProvider = func(ctx contextAPI.Client) (discoveryClient, error) {
	return fabdiscovery.New(ctx)
}

// service implements a dynamic Discovery Service that queries
// Fabric's Discovery service for information about the peers that
// are currently joined to the given channel.
type service struct {
	responseTimeout time.Duration
	lock            sync.RWMutex
	ctx             contextAPI.Client
	discClient      discoveryClient
	peersRef        *lazyref.Reference
}

type queryPeers func() ([]fab.Peer, error)

func newService(query queryPeers, options options) *service {
	logger.Debugf("Creating new dynamic discovery service with cache refresh interval %s", options.refreshInterval)
	return &service{
		responseTimeout: options.responseTimeout,
		peersRef: lazyref.New(
			func() (interface{}, error) {
				return query()
			},
			lazyref.WithRefreshInterval(lazyref.InitOnFirstAccess, options.refreshInterval),
		),
	}
}

// Initialize initializes the service with local context
func (s *service) Initialize(ctx contextAPI.Client) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.ctx != nil {
		// Already initialized
		logger.Debugf("Already initialized with context: %#v", s.ctx)
		return nil
	}

	discoveryClient, err := clientProvider(ctx)
	if err != nil {
		return errors.Wrapf(err, "error creating discover client")
	}

	logger.Debugf("Initializing with context: %#v", ctx)
	s.ctx = ctx
	s.discClient = discoveryClient
	return nil
}

// Close stops the lazyref background refresh
func (s *service) Close() {
	logger.Debugf("Closing peers ref...")
	s.peersRef.Close()
}

// GetPeers returns the available peers
func (s *service) GetPeers() ([]fab.Peer, error) {
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

func (s *service) context() contextAPI.Client {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ctx
}

func (s *service) discoveryClient() discoveryClient {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.discClient
}

func asPeers(ctx contextAPI.Client, endpoints []*discclient.Peer) []fab.Peer {
	var peers []fab.Peer
	for _, endpoint := range endpoints {
		url := endpoint.AliveMessage.GetAliveMsg().Membership.Endpoint

		logger.Debugf("Adding endpoint [%s]", url)

		peerConfig, err := ctx.EndpointConfig().PeerConfig(url)
		if err != nil {
			logger.Warnf("Error getting peer config for url [%s]: %s", err)
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
