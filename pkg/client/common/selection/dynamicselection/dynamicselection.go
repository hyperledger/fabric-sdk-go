/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicselection

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"

	copts "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection/pgresolver"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
)

const defaultCacheTimeout = 30 * time.Minute

// Opt applies a selection provider option
type Opt func(*SelectionService)

// WithLoadBalancePolicy sets the load-balance policy
func WithLoadBalancePolicy(lbp pgresolver.LoadBalancePolicy) Opt {
	return func(s *SelectionService) {
		s.pgLBP = lbp
	}
}

// WithCacheTimeout sets the expiration timeout of the cache
func WithCacheTimeout(timeout time.Duration) Opt {
	return func(s *SelectionService) {
		s.cacheTimeout = timeout
	}
}

// SelectionService chooses endorsing peers for a given set of chaincodes using their chaincode policy
type SelectionService struct {
	channelID        string
	pgResolvers      *lazycache.Cache
	pgLBP            pgresolver.LoadBalancePolicy
	ccPolicyProvider CCPolicyProvider
	discoveryService fab.DiscoveryService
	cacheTimeout     time.Duration
}

type policyProviderFactory func() (CCPolicyProvider, error)

// NewService creates a new dynamic selection service
func NewService(context context.Client, channelID string, discovery fab.DiscoveryService, opts ...Opt) (*SelectionService, error) {
	return newService(context, channelID, discovery,
		func() (CCPolicyProvider, error) {
			return newCCPolicyProvider(context, discovery, channelID)
		}, opts...)
}

func newService(context context.Client, channelID string, discovery fab.DiscoveryService, factory policyProviderFactory, opts ...Opt) (*SelectionService, error) {
	ccPolicyProvider, err := factory()
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to create cc policy provider")
	}

	service := &SelectionService{
		channelID:        channelID,
		discoveryService: discovery,
		ccPolicyProvider: ccPolicyProvider,
		cacheTimeout:     defaultCacheTimeout,
		pgLBP:            pgresolver.NewRandomLBP(),
	}

	for _, opt := range opts {
		opt(service)
	}

	if service.cacheTimeout == 0 {
		service.cacheTimeout = context.EndpointConfig().Timeout(fab.SelectionServiceRefresh)
	}

	if service.pgLBP == nil {
		service.pgLBP = pgresolver.NewRandomLBP()
	}

	service.pgResolvers = lazycache.New(
		"PG_Resolver_Cache",
		func(key lazycache.Key) (interface{}, error) {
			return service.createPGResolver(key.(*resolverKey))
		},
		lazyref.WithAbsoluteExpiration(service.cacheTimeout),
	)

	return service, nil
}

// GetEndorsersForChaincode returns the endorsing peers for the given chaincodes
func (s *SelectionService) GetEndorsersForChaincode(chaincodes []*fab.ChaincodeCall, opts ...copts.Opt) ([]fab.Peer, error) {
	if len(chaincodes) == 0 {
		return nil, errors.New("no chaincode IDs provided")
	}

	params := options.NewParams(opts)

	var chaincodeIDs []string
	for _, cc := range chaincodes {
		chaincodeIDs = append(chaincodeIDs, cc.ID)
	}

	resolver, err := s.getPeerGroupResolver(chaincodeIDs)
	if err != nil {
		return nil, errors.WithMessagef(err, "Error getting peer group resolver for chaincodes [%v] on channel [%s]", chaincodeIDs, s.channelID)
	}

	peers, err := s.discoveryService.GetPeers()
	if err != nil {
		return nil, err
	}

	if params.PeerFilter != nil {
		var filteredPeers []fab.Peer
		for _, peer := range peers {
			if params.PeerFilter(peer) {
				filteredPeers = append(filteredPeers, peer)
			} else {
				logger.Debugf("Peer [%s] is not accepted by the filter and therefore peer group will be excluded.", peer.URL())
			}
		}
		peers = filteredPeers
	}

	if params.PeerSorter != nil {
		sortedPeers := make([]fab.Peer, len(peers))
		copy(sortedPeers, peers)
		peers = params.PeerSorter(sortedPeers)
	}

	peerGroup, err := resolver.Resolve(peers)
	if err != nil {
		return nil, err
	}
	return peerGroup.Peers(), nil
}

// Close closes all resources associated with the service
func (s *SelectionService) Close() {
	s.pgResolvers.Close()
}

func (s *SelectionService) getPeerGroupResolver(chaincodeIDs []string) (pgresolver.PeerGroupResolver, error) {
	resolver, err := s.pgResolvers.Get(newResolverKey(s.channelID, chaincodeIDs...))
	if err != nil {
		return nil, err
	}
	return resolver.(pgresolver.PeerGroupResolver), nil
}

func (s *SelectionService) createPGResolver(key *resolverKey) (pgresolver.PeerGroupResolver, error) {
	// Retrieve the signature policies for all of the chaincodes
	var policyGroups []pgresolver.GroupRetriever
	for _, ccID := range key.chaincodeIDs {
		policyGroup, err := s.getPolicyGroupForCC(key.channelID, ccID)
		if err != nil {
			return nil, errors.WithMessagef(err, "error retrieving signature policy for chaincode [%s] on channel [%s]", ccID, key.channelID)
		}
		policyGroups = append(policyGroups, policyGroup)
	}

	// Perform an 'and' operation on all of the peer groups
	aggregatePolicyGroupRetriever := func(peerRetriever pgresolver.MSPPeerRetriever) (pgresolver.GroupOfGroups, error) {
		var groups []pgresolver.Group
		for _, f := range policyGroups {
			grps, err := f(peerRetriever)
			if err != nil {
				return nil, err
			}
			groups = append(groups, grps)
		}
		return pgresolver.NewGroupOfGroups(groups).Nof(int32(len(policyGroups)))
	}

	// Create the resolver
	resolver, err := pgresolver.NewPeerGroupResolver(aggregatePolicyGroupRetriever, s.pgLBP)
	if err != nil {
		return nil, errors.WithMessagef(err, "error creating peer group resolver for chaincodes [%v] on channel [%s]", key.chaincodeIDs, key.channelID)
	}
	return resolver, nil
}

func (s *SelectionService) getPolicyGroupForCC(channelID string, ccID string) (pgresolver.GroupRetriever, error) {
	sigPolicyEnv, err := s.ccPolicyProvider.GetChaincodePolicy(ccID)
	if err != nil {
		return nil, errors.WithMessagef(err, "error querying chaincode [%s] on channel [%s]", ccID, channelID)
	}
	return pgresolver.CompileSignaturePolicy(sigPolicyEnv)
}
