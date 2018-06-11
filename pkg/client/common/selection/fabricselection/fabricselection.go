/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricselection

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	reqContext "github.com/hyperledger/fabric-sdk-go/pkg/context"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

const moduleName = "fabsdk/client"

var logger = logging.NewLogger(moduleName)

// PeerState provides state information about the Peer
type PeerState interface {
	BlockHeight() uint64
}

type discoveryClient interface {
	Send(ctx context.Context, req *discclient.Request, targets ...fab.PeerConfig) ([]fabdiscovery.Response, error)
}

// clientProvider is overridden by unit tests
var clientProvider = func(ctx contextAPI.Client) (discoveryClient, error) {
	return fabdiscovery.New(ctx)
}

// Service chooses endorsing peers for a given set of chaincodes using
// Fabric's Discovery Service
type Service struct {
	channelID       string
	responseTimeout time.Duration
	ctx             contextAPI.Client
	discovery       fab.DiscoveryService
	discClient      discoveryClient
	chResponseCache *lazycache.Cache
	retryOpts       retry.Opts
}

// New creates a new dynamic selection service using Fabric's Discovery Service
func New(ctx contextAPI.Client, channelID string, discovery fab.DiscoveryService, opts ...coptions.Opt) (*Service, error) {
	options := params{retryOpts: retry.DefaultResMgmtOpts}
	coptions.Apply(&options, opts)

	if options.refreshInterval == 0 {
		// Use DiscoveryServiceRefresh since the selection algorithm depends on up-to-date
		// information from the Discovery Client.
		options.refreshInterval = ctx.EndpointConfig().Timeout(fab.DiscoveryServiceRefresh)
	}

	if options.responseTimeout == 0 {
		options.responseTimeout = ctx.EndpointConfig().Timeout(fab.DiscoveryResponse)
	}

	logger.Debugf("Cache refresh interval: %s", options.refreshInterval)
	logger.Debugf("Deliver service response timeout: %s", options.responseTimeout)

	discoveryClient, err := clientProvider(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error creating discover client")
	}

	s := &Service{
		channelID:       channelID,
		ctx:             ctx,
		responseTimeout: options.responseTimeout,
		discovery:       discovery,
		discClient:      discoveryClient,
		retryOpts:       options.retryOpts,
	}

	s.chResponseCache = lazycache.New(
		"Channel_Response_Cache",
		func(key lazycache.Key) (interface{}, error) {
			return s.newChannelResponseRef(key.(*cacheKey).chaincodes, options.refreshInterval), nil
		},
	)

	return s, nil
}

// GetEndorsersForChaincode returns the endorsing peers for the given chaincodes
func (s *Service) GetEndorsersForChaincode(chaincodes []*fab.ChaincodeCall, opts ...coptions.Opt) ([]fab.Peer, error) {
	logger.Debugf("Getting endorsers for chaincodes [%#v]...", chaincodes)
	if len(chaincodes) == 0 {
		return nil, errors.New("no chaincode IDs provided")
	}

	chResponse, err := s.getChannelResponse(chaincodes)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting channel response for channel [%s]", s.channelID)
	}

	peers, err := s.discovery.GetPeers()
	if err != nil {
		return nil, errors.Wrapf(err, "error getting peers from discovery service for channel [%s]", s.channelID)
	}

	params := options.NewParams(opts)

	endpoints, err := chResponse.Endorsers(asInvocationChain(chaincodes), newSelector(params.PrioritySelector), newFilter(params.PeerFilter, peers))
	if err != nil {
		return nil, errors.Wrap(err, "error getting endorsers from channel response")
	}

	return asPeers(s.ctx, endpoints), nil
}

// Close closes all resources associated with the service
func (s *Service) Close() {
	logger.Debug("Closing channel response cache")
	s.chResponseCache.Close()
}

func (s *Service) newChannelResponseRef(chaincodes []*fab.ChaincodeCall, refreshInterval time.Duration) *lazyref.Reference {
	return lazyref.New(
		func() (interface{}, error) {
			if logging.IsEnabledFor(moduleName, logging.DEBUG) {
				key, _ := json.Marshal(chaincodes)
				logger.Debugf("Refreshing endorsers for chaincodes [%s] in channel [%s] from discovery service...", key, s.channelID)
			}
			return s.queryEndorsers(chaincodes)
		},
		lazyref.WithRefreshInterval(lazyref.InitImmediately, refreshInterval),
	)
}

func (s *Service) getChannelResponse(chaincodes []*fab.ChaincodeCall) (discclient.ChannelResponse, error) {
	key := newCacheKey(chaincodes)
	ref, err := s.chResponseCache.Get(key)
	if err != nil {
		return nil, err
	}
	chResp, err := ref.(*lazyref.Reference).Get()
	if err != nil {
		return nil, err
	}
	return chResp.(discclient.ChannelResponse), nil
}

func (s *Service) queryEndorsers(chaincodes []*fab.ChaincodeCall) (discclient.ChannelResponse, error) {
	logger.Debugf("Querying discovery service for endorsers for chaincodes: %#v", chaincodes)

	targets, err := s.getTargets(s.ctx)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, errors.Errorf("no peers configured for channel [%s]", s.channelID)
	}

	req, err := discclient.NewRequest().OfChannel(s.channelID).AddEndorsersQuery(asChaincodeInterests(chaincodes))
	if err != nil {
		return nil, errors.Wrapf(err, "error creating endorser query request")
	}

	chResponse, err := retry.NewInvoker(retry.New(s.retryOpts)).Invoke(
		func() (interface{}, error) {
			return s.query(req, chaincodes, targets)
		},
	)

	if err != nil {
		return nil, err
	}
	return chResponse.(discclient.ChannelResponse), err
}

func (s *Service) query(req *discclient.Request, chaincodes []*fab.ChaincodeCall, targets []fab.PeerConfig) (discclient.ChannelResponse, error) {
	logger.Debugf("Querying Discovery Service for endorsers for chaincodes: %#v on channel [%s]", chaincodes, s.channelID)
	reqCtx, cancel := reqContext.NewRequest(s.ctx, reqContext.WithTimeout(s.responseTimeout))
	defer cancel()

	responses, err := s.discClient.Send(reqCtx, req, targets...)
	if err != nil {
		if len(responses) == 0 {
			return nil, errors.Wrapf(err, "error calling discover service send")
		}
		logger.Warnf("Received %d response(s) and one or more errors from discovery client: %s", len(responses), err)
	}

	if len(responses) == 0 {
		return nil, errors.New("no successful response received from any peer")
	}

	// TODO: In a future patch:
	// - validate the signatures in the responses
	// For now just pick the first successful response

	invocChain := asInvocationChain(chaincodes)

	var lastErr error
	for _, response := range responses {
		chResp := response.ForChannel(s.channelID)
		// Make sure the target didn't return an error
		_, err := chResp.Endorsers(invocChain, discclient.NoPriorities, discclient.NoExclusion)
		if err != nil {
			lastErr = errors.Wrapf(err, "error getting endorsers from target [%s]", response.Target())
			logger.Debugf(lastErr.Error())
			continue
		}
		return chResp, nil
	}

	logger.Warn(lastErr.Error())

	if strings.Contains(lastErr.Error(), "failed constructing descriptor for chaincodes") {
		errMsg := fmt.Sprintf("error received from Discovery Server: %s", lastErr)
		return nil, status.New(status.DiscoveryServerStatus, int32(status.QueryEndorsers), errMsg, []interface{}{})
	}

	return nil, lastErr
}

func (s *Service) getTargets(ctx contextAPI.Client) ([]fab.PeerConfig, error) {
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

func asChaincodeInterests(chaincodes []*fab.ChaincodeCall) *discovery.ChaincodeInterest {
	return &discovery.ChaincodeInterest{
		Chaincodes: asInvocationChain(chaincodes),
	}
}

func asInvocationChain(chaincodes []*fab.ChaincodeCall) discclient.InvocationChain {
	var invocChain discclient.InvocationChain
	for _, cc := range chaincodes {
		invocChain = append(invocChain, &discovery.ChaincodeCall{
			Name:            cc.ID,
			CollectionNames: cc.Collections,
		})
	}
	return invocChain
}

func asPeers(ctx contextAPI.Client, endpoints []*discclient.Peer) []fab.Peer {
	var peers []fab.Peer
	for _, endpoint := range endpoints {
		peer, err := asPeer(ctx, endpoint)
		if err != nil {
			logger.Warnf(err.Error())
			continue
		}
		peers = append(peers, peer)
	}
	return peers
}

func asPeer(ctx contextAPI.Client, endpoint *discclient.Peer) (fab.Peer, error) {
	url := endpoint.AliveMessage.GetAliveMsg().Membership.Endpoint

	peerConfig, found := ctx.EndpointConfig().PeerConfig(url)
	if !found {
		logger.Warnf("Peer config not found for url [%s]", url)
		return nil, errors.Errorf("peer config not found for [%s]", url)
	}

	peer, err := ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: *peerConfig, MSPID: endpoint.MSPID})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create peer config for [%s]", url)
	}

	return &peerEndpoint{
		Peer:        peer,
		blockHeight: endpoint.StateInfoMessage.GetStateInfo().GetProperties().LedgerHeight,
	}, nil
}

type peerEndpoint struct {
	fab.Peer
	blockHeight uint64
}

func (p *peerEndpoint) BlockHeight() uint64 {
	return p.blockHeight
}
