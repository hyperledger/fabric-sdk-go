/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"sync"
	"time"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

// clientProvider is overridden by unit tests
var clientProvider = func(ctx contextAPI.Client) (fabdiscovery.Client, error) {
	return fabdiscovery.New(ctx)
}

// service implements a dynamic Discovery Service that queries
// Fabric's Discovery service for information about the peers that
// are currently joined to the given channel.
type service struct {
	responseTimeout time.Duration
	lock            sync.RWMutex
	ctx             contextAPI.Client
	discClient      fabdiscovery.Client
	peersRef        *lazyref.Reference
	ErrHandler      fab.ErrorHandler
}

type queryPeers func() ([]fab.Peer, error)

func newService(config fab.EndpointConfig, query queryPeers, opts ...coptions.Opt) *service {
	opt := options{}
	coptions.Apply(&opt, opts)

	if opt.refreshInterval == 0 {
		opt.refreshInterval = config.Timeout(fab.DiscoveryServiceRefresh)
	}

	if opt.responseTimeout == 0 {
		opt.responseTimeout = config.Timeout(fab.DiscoveryResponse)
	}

	logger.Debugf("Cache refresh interval: %s", opt.refreshInterval)
	logger.Debugf("Deliver service response timeout: %s", opt.responseTimeout)

	return &service{
		responseTimeout: opt.responseTimeout,
		ErrHandler:      opt.errHandler,
		peersRef: lazyref.New(
			func() (interface{}, error) {
				return query()
			},
			lazyref.WithRefreshInterval(lazyref.InitOnFirstAccess, opt.refreshInterval),
		),
	}
}

// initialize initializes the service with client context
func (s *service) initialize(ctx contextAPI.Client) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.ctx != nil {
		// Already initialized
		logger.Debugf("Already initialized with context: %#v", s.ctx)
		return nil
	}

	discoveryClient, err := clientProvider(ctx)
	if err != nil {
		return errors.Wrap(err, "error creating discover client")
	}

	logger.Debugf("Initializing with context: %#v", ctx)
	s.ctx = ctx
	s.discClient = discoveryClient
	return nil
}

// Close stops the lazyref background refresh
func (s *service) Close() {
	logger.Debug("Closing peers ref...")
	s.peersRef.Close()
}

// GetPeers returns the available peers
func (s *service) GetPeers() ([]fab.Peer, error) {
	if s.peersRef.IsClosed() {
		return nil, errors.Errorf("Discovery client has been closed")
	}

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

func (s *service) discoveryClient() fabdiscovery.Client {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.discClient
}

func asPeers(ctx contextAPI.Client, endpoints []*discclient.Peer) []fab.Peer {
	var peers []fab.Peer
	for _, endpoint := range endpoints {
		peer, ok := asPeer(ctx, endpoint)
		if !ok {
			continue
		}
		peers = append(peers, peer)
	}
	return peers
}

func asPeer(ctx contextAPI.Client, endpoint *discclient.Peer) (fab.Peer, bool) {
	url := endpoint.AliveMessage.GetAliveMsg().Membership.Endpoint

	logger.Debugf("Adding endpoint [%s]", url)

	var (
		peer fab.Peer
		err  error
	)

	peerConfig, found := ctx.EndpointConfig().PeerConfig(url)
	if found {
		peer, err = ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{
			PeerConfig: *peerConfig,
			MSPID:      endpoint.MSPID,
			Properties: fabdiscovery.GetProperties(endpoint),
		})
	} else {
		peer, err = peerImpl.New(ctx.EndpointConfig(), peerImpl.FromPeerConfig(&fab.NetworkPeer{
			PeerConfig: fab.PeerConfig{
				URL: url,
			},
			MSPID:      endpoint.MSPID,
			Properties: fabdiscovery.GetProperties(endpoint),
		}))
	}

	if err != nil {
		logger.Warnf("Unable to create peer config for [%s]: %s", url, err)
		return nil, false
	}

	return peer, true
}
