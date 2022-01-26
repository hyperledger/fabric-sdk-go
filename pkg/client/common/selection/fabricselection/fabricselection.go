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

	"github.com/hyperledger/fabric-protos-go/peer"
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/random"
	soptions "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	reqContext "github.com/hyperledger/fabric-sdk-go/pkg/context"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

const (
	moduleName = "fabsdk/client"

	// AccessDenied indicates that the user does not have permission to perform the operation
	AccessDenied = "access denied"
)

var logger = logging.NewLogger(moduleName)

// DiscoveryClient is the client to the discovery service
type DiscoveryClient interface {
	Send(ctx context.Context, req *fabdiscovery.Request, targets ...fab.PeerConfig) (<-chan fabdiscovery.Response, error)
}

// clientProvider is overridden by unit tests
var clientProvider = func(ctx contextAPI.Client) (DiscoveryClient, error) {
	return fabdiscovery.New(ctx)
}

// Service chooses endorsing peers for a given set of chaincodes using
// Fabric's Discovery Service
type Service struct {
	channelID       string
	responseTimeout time.Duration
	ctx             contextAPI.Client
	discovery       fab.DiscoveryService
	discClient      DiscoveryClient
	chResponseCache *lazycache.Cache
	retryOpts       retry.Opts
	errHandler      fab.ErrorHandler
	peerSorter      soptions.PeerSorter
}

// New creates a new dynamic selection service using Fabric's Discovery Service
func New(ctx contextAPI.Client, channelID string, discovery fab.DiscoveryService, opts ...coptions.Opt) (*Service, error) {
	options := params{retryOpts: getDefaultRetryOpts(ctx, channelID)}

	coptions.Apply(&options, opts)

	if options.refreshInterval == 0 {
		options.refreshInterval = ctx.EndpointConfig().Timeout(fab.SelectionServiceRefresh)
	}

	if options.responseTimeout == 0 {
		options.responseTimeout = ctx.EndpointConfig().Timeout(fab.DiscoveryResponse)
	}

	logger.Debugf("Cache refresh interval: %s", options.refreshInterval)
	logger.Debugf("Deliver service response timeout: %s", options.responseTimeout)
	logger.Debugf("Retry opts: %#v", options.retryOpts)

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
		errHandler:      options.errHandler,
		peerSorter:      resolvePeerSorter(channelID, ctx),
	}

	s.chResponseCache = lazycache.NewWithData(
		"Fabric_Selection_Cache",
		func(key lazycache.Key, data interface{}) (interface{}, error) {
			invocationChain := key.(*cacheKey).chaincodes
			if logging.IsEnabledFor(moduleName, logging.DEBUG) {
				key, err := json.Marshal(invocationChain)
				if err != nil {
					panic(fmt.Sprintf("marshal of chaincodes failed: %s", err))
				}
				logger.Debugf("Refreshing endorsers for chaincodes [%s] in channel [%s] from discovery service...", key, channelID)
			}

			ropts := s.retryOpts
			if data != nil {
				ropts = data.(retry.Opts)
				logger.Debugf("Overriding retry opts: %#v", ropts)
			}

			endorsers, err := s.queryEndorsers(invocationChain, ropts)
			if err != nil && s.errHandler != nil {
				logger.Debugf("[%s] Got error from discovery query: %s. Invoking error handler", s.channelID, err)
				s.errHandler(s.ctx, s.channelID, err)
			}
			return endorsers, err
		},
		lazyref.WithRefreshInterval(lazyref.InitImmediately, options.refreshInterval),
	)

	return s, nil
}

// GetEndorsersForChaincode returns the endorsing peers for the given chaincodes
func (s *Service) GetEndorsersForChaincode(chaincodes []*fab.ChaincodeCall, opts ...coptions.Opt) ([]fab.Peer, error) {
	if s.chResponseCache.IsClosed() {
		return nil, errors.Errorf("Selection service has been closed")
	}

	logger.Debugf("Getting endorsers for chaincodes [%#v]...", chaincodes)
	if len(chaincodes) == 0 {
		return nil, errors.New("no chaincode IDs provided")
	}

	params := soptions.Params{RetryOpts: s.retryOpts}
	coptions.Apply(&params, opts)

	if params.PeerSorter == nil {
		params.PeerSorter = s.peerSorter
	}

	chResponse, err := s.getChannelResponse(chaincodes, params.RetryOpts)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting channel response for channel [%s]", s.channelID)
	}

	// Execute getEndorsers with retries since the discovered peers may be out of sync with
	// the peers returned from the endorser query and it may take a while for them to sync.
	endpoints, err := retry.NewInvoker(retry.New(s.retryOpts)).Invoke(
		func() (interface{}, error) {
			return s.getEndorsers(chaincodes, chResponse, params.PeerFilter, params.PeerSorter)
		},
	)

	if err != nil || endpoints == nil {
		return nil, err
	}

	return asPeers(s.ctx, endpoints.(discclient.Endorsers)), nil
}

// Close closes all resources associated with the service
func (s *Service) Close() {
	logger.Debug("Closing channel response cache")
	s.chResponseCache.Close()
}

func (s *Service) getEndorsers(chaincodes []*fab.ChaincodeCall, chResponse discclient.ChannelResponse, peerFilter soptions.PeerFilter, sorter soptions.PeerSorter) (discclient.Endorsers, error) {
	peers, err := s.discovery.GetPeers()
	if err != nil {
		return nil, errors.Wrapf(err, "error getting peers from discovery service for channel [%s]", s.channelID)
	}

	endpoints, err := chResponse.Endorsers(asInvocationChain(chaincodes), newFilter(s.ctx, peers, peerFilter, sorter))
	if err != nil && newDiscoveryError(err).IsTransient() {
		return nil, status.New(status.DiscoveryServerStatus, int32(status.QueryEndorsers), fmt.Sprintf("error getting endorsers: %s", err), []interface{}{})
	}

	return endpoints, err
}

func (s *Service) getChannelResponse(chaincodes []*fab.ChaincodeCall, retryOpts retry.Opts) (discclient.ChannelResponse, error) {
	key := newCacheKey(chaincodes)
	chResp, err := s.chResponseCache.Get(key, retryOpts)
	if err != nil {
		return nil, err
	}
	return chResp.(discclient.ChannelResponse), nil
}

func (s *Service) queryEndorsers(chaincodes []*fab.ChaincodeCall, retryOpts retry.Opts) (discclient.ChannelResponse, error) {
	logger.Debugf("Querying discovery service for endorsers for chaincodes: %#v", chaincodes)

	targets, err := s.getTargets(s.ctx)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, errors.Errorf("no peers configured for channel [%s]", s.channelID)
	}

	req, err := fabdiscovery.NewRequest().OfChannel(s.channelID).AddEndorsersQuery(asChaincodeInterests(chaincodes))
	if err != nil {
		return nil, errors.Wrapf(err, "error creating endorser query request")
	}

	logger.Debugf("Querying Discovery Service with retry opts: %#v", retryOpts)
	chResponse, err := retry.NewInvoker(retry.New(retryOpts)).Invoke(
		func() (interface{}, error) {
			return s.query(req, chaincodes, targets)
		},
	)

	if err != nil {
		return nil, newDiscoveryError(err)
	}
	return chResponse.(discclient.ChannelResponse), nil
}

func (s *Service) query(req *fabdiscovery.Request, chaincodes []*fab.ChaincodeCall, targets []fab.PeerConfig) (discclient.ChannelResponse, error) {
	logger.Debugf("Querying Discovery Service for endorsers for chaincodes: %#v on channel [%s]", chaincodes, s.channelID)
	reqCtx, cancel := reqContext.NewRequest(s.ctx, reqContext.WithTimeout(s.responseTimeout))
	defer cancel()

	responses, err := s.discClient.Send(reqCtx, req, targets...)

	if err != nil {
		return nil, errors.Wrapf(err, "error calling discover service send for selection")
	}

	// TODO: In a future patch:
	// - validate the signatures in the responses
	// For now just pick the first successful response

	invocChain := asInvocationChain(chaincodes)

	var respErrors []error
	var discErrs []DiscoveryError

	for response := range responses {
		logger.Debugf("Checking response from [%s]...", response.Target())

		if err := response.Error(); err != nil {
			respErrors = append(respErrors, err)
			logger.Debugf("... got error response from [%s]: %s", response.Target(), err)
			continue
		}

		chResp := response.ForChannel(s.channelID)
		// Make sure the target didn't return an error

		if _, err := chResp.Endorsers(invocChain, discclient.NoFilter); err != nil {
			logger.Debugf("... got error response from [%s]: %s", response.Target(), err)

			discErrs = append(discErrs, newDiscoveryError(err))
			continue
		}

		logger.Debugf("... got success response from [%s]", response.Target())

		//got successful response, cancel all outstanding requests
		cancel()

		return chResp, nil
	}

	for _, e := range discErrs {
		if e.IsTransient() {
			logger.Debugf("Got transient error: %s", e)
			errMsg := fmt.Sprintf("error received from Discovery Server: %s", e)
			return nil, status.New(status.DiscoveryServerStatus, int32(status.QueryEndorsers), errMsg, []interface{}{})
		}

		logger.Debugf("Got non-transient error: %s", e)
		respErrors = append(respErrors, e)
	}

	return nil, errors.Wrap(multi.New(respErrors...), "no successful response received from any peer")
}

func (s *Service) getTargets(ctx contextAPI.Client) ([]fab.PeerConfig, error) {

	chpeers := ctx.EndpointConfig().ChannelPeers(s.channelID)
	if len(chpeers) == 0 {
		return nil, errors.Errorf("no channel peers configured for channel [%s]", s.channelID)
	}

	chConfig := ctx.EndpointConfig().ChannelConfig(s.channelID)

	//pick number of peers based on channel policy
	return random.PickRandomNPeerConfigs(chpeers, chConfig.Policies.Discovery.MaxTargets), nil
}

func asChaincodeInterests(chaincodes []*fab.ChaincodeCall) *peer.ChaincodeInterest {
	return &peer.ChaincodeInterest{
		Chaincodes: asInvocationChain(chaincodes),
	}
}

func asInvocationChain(chaincodes []*fab.ChaincodeCall) discclient.InvocationChain {
	var invocChain discclient.InvocationChain
	for _, cc := range chaincodes {
		invocChain = append(invocChain, &peer.ChaincodeCall{
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
			logger.Debugf(err.Error())
			continue
		}
		peers = append(peers, peer)
	}
	return peers
}

func asPeer(ctx contextAPI.Client, endpoint *discclient.Peer) (fab.Peer, error) {
	url := endpoint.AliveMessage.GetAliveMsg().Membership.Endpoint

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
		return nil, errors.Wrapf(err, "unable to create peer config for [%s]", url)
	}

	return &peerEndpoint{
		Peer: peer,
	}, nil
}

func getDefaultRetryOpts(ctx contextAPI.Client, channelID string) retry.Opts {
	return ctx.EndpointConfig().ChannelConfig(channelID).Policies.Discovery.RetryOpts
}

type peerEndpoint struct {
	fab.Peer
}

func (p *peerEndpoint) BlockHeight() uint64 {
	ledgerHeight, ok := p.Properties()[fab.PropertyLedgerHeight]
	if !ok {
		return 0
	}

	return ledgerHeight.(uint64)
}

// DiscoveryError is an error originating at the Discovery service
type DiscoveryError struct {
	error
}

func newDiscoveryError(err error) DiscoveryError {
	return DiscoveryError{err}
}

//IsTransient checks if error is about peer being unable to ding chaincode or endorsement combination wasn't satisfied
func (e DiscoveryError) IsTransient() bool {
	return strings.Contains(e.Error(), "failed constructing descriptor for chaincodes") ||
		strings.Contains(e.Error(), "no endorsement combination can be satisfied")
}

//IsAccessDenied checks if response contains access denied msg
func (e DiscoveryError) IsAccessDenied() bool {
	return strings.Contains(e.Error(), AccessDenied)
}
