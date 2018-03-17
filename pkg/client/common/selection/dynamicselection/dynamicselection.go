/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicselection

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	copts "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection/pgresolver"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
)

const defaultCacheTimeout = 30 * time.Minute

// ChannelUser contains user(identity) info to be used for specific channel
type ChannelUser struct {
	ChannelID string
	Username  string
	OrgName   string
}

// SelectionProvider implements selection provider
// TODO: refactor users into client contexts
type SelectionProvider struct {
	config       core.Config
	users        []ChannelUser
	lbp          pgresolver.LoadBalancePolicy
	providers    api.Providers
	cacheTimeout time.Duration
}

// Opt applies a selection provider option
type Opt func(*SelectionProvider)

// WithLoadBalancePolicy sets the load-balance policy
func WithLoadBalancePolicy(lbp pgresolver.LoadBalancePolicy) Opt {
	return func(p *SelectionProvider) {
		p.lbp = lbp
	}
}

// WithCacheTimeout sets the expiration timeout of the cache
func WithCacheTimeout(timeout time.Duration) Opt {
	return func(p *SelectionProvider) {
		p.cacheTimeout = timeout
	}
}

// New returns dynamic selection provider
func New(config core.Config, users []ChannelUser, opts ...Opt) (*SelectionProvider, error) {
	p := &SelectionProvider{
		config:       config,
		users:        users,
		lbp:          pgresolver.NewRandomLBP(),
		cacheTimeout: defaultCacheTimeout,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p, nil
}

type selectionService struct {
	channelID        string
	pgResolvers      *lazycache.Cache
	pgLBP            pgresolver.LoadBalancePolicy
	ccPolicyProvider CCPolicyProvider
	discoveryService fab.DiscoveryService
	cacheTimeout     time.Duration
}

// Initialize allow for initializing providers
func (p *SelectionProvider) Initialize(providers contextAPI.Providers) error {
	p.providers = providers
	return nil
}

// CreateSelectionService creates a selection service
func (p *SelectionProvider) CreateSelectionService(channelID string) (fab.SelectionService, error) {
	if channelID == "" {
		return nil, errors.New("Must provide channel ID")
	}

	var channelUser *ChannelUser
	for _, p := range p.users {
		if p.ChannelID == channelID {
			channelUser = &p
			break
		}
	}

	if channelUser == nil {
		return nil, errors.New("Must provide user for channel")
	}

	ccPolicyProvider, err := newCCPolicyProvider(p.providers, channelID, channelUser.Username, channelUser.OrgName)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to create cc policy provider")
	}

	return newSelectionService(channelID, p.lbp, ccPolicyProvider, p.cacheTimeout)
}

func newSelectionService(channelID string, lbp pgresolver.LoadBalancePolicy, ccPolicyProvider CCPolicyProvider, cacheTimeout time.Duration) (*selectionService, error) {
	service := &selectionService{
		channelID:        channelID,
		pgLBP:            lbp,
		ccPolicyProvider: ccPolicyProvider,
	}

	service.pgResolvers = lazycache.New(
		"PG_Resolver_Cache",
		func(key lazycache.Key) (interface{}, error) {
			return lazyref.New(
				func() (interface{}, error) {
					return service.createPGResolver(key.(*resolverKey))
				},
				lazyref.WithAbsoluteExpiration(cacheTimeout),
			), nil
		},
	)

	return service, nil
}

func (s *selectionService) Initialize(context contextAPI.Channel) error {
	s.discoveryService = context.DiscoveryService()
	return nil
}

func (s *selectionService) GetEndorsersForChaincode(chaincodeIDs []string, opts ...copts.Opt) ([]fab.Peer, error) {
	if len(chaincodeIDs) == 0 {
		return nil, errors.New("no chaincode IDs provided")
	}

	params := options.NewParams(opts)

	resolver, err := s.getPeerGroupResolver(chaincodeIDs)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("Error getting peer group resolver for chaincodes [%v] on channel [%s]", chaincodeIDs, s.channelID))
	}
	return resolver.Resolve(params.PeerFilter).Peers(), nil
}

func (s *selectionService) getPeerGroupResolver(chaincodeIDs []string) (pgresolver.PeerGroupResolver, error) {
	value, err := s.pgResolvers.Get(newResolverKey(s.channelID, chaincodeIDs...))
	if err != nil {
		return nil, err
	}
	lazyRef := value.(*lazyref.Reference)
	resolver, err := lazyRef.Get()
	if err != nil {
		return nil, err
	}
	return resolver.(pgresolver.PeerGroupResolver), nil
}

func (s *selectionService) createPGResolver(key *resolverKey) (pgresolver.PeerGroupResolver, error) {
	// Retrieve the signature policies for all of the chaincodes
	var policyGroups []pgresolver.Group
	for _, ccID := range key.chaincodeIDs {
		policyGroup, err := s.getPolicyGroupForCC(key.channelID, ccID)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("error retrieving signature policy for chaincode [%s] on channel [%s]", ccID, key.channelID))
		}
		policyGroups = append(policyGroups, policyGroup)
	}

	// Perform an 'and' operation on all of the peer groups
	aggregatePolicyGroup, err := pgresolver.NewGroupOfGroups(policyGroups).Nof(int32(len(policyGroups)))
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("error computing signature policy for chaincode(s) [%v] on channel [%s]", key.chaincodeIDs, key.channelID))
	}

	// Create the resolver
	resolver, err := pgresolver.NewPeerGroupResolver(aggregatePolicyGroup, s.pgLBP)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("error creating peer group resolver for chaincodes [%v] on channel [%s]", key.chaincodeIDs, key.channelID))
	}
	return resolver, nil
}

func (s *selectionService) getPolicyGroupForCC(channelID string, ccID string) (pgresolver.Group, error) {
	sigPolicyEnv, err := s.ccPolicyProvider.GetChaincodePolicy(ccID)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("error querying chaincode [%s] on channel [%s]", ccID, channelID))
	}

	return pgresolver.NewSignaturePolicyCompiler(
		func(mspID string) []fab.Peer {
			return s.getAvailablePeers(mspID)
		},
	).Compile(sigPolicyEnv)
}

func (s *selectionService) getAvailablePeers(mspID string) []fab.Peer {
	channelPeers, err := s.discoveryService.GetPeers()
	if err != nil {
		logger.Errorf("Error retrieving peers from discovery service: %s", err)
		return nil
	}

	var peers []fab.Peer
	for _, peer := range channelPeers {
		if string(peer.MSPID()) == mspID {
			peers = append(peers, peer)
		}
	}

	if logging.IsEnabledFor(loggerModule, logging.DEBUG) {
		str := ""
		for i, peer := range peers {
			str += peer.URL()
			if i+1 < len(peers) {
				str += ","
			}
		}
		logger.Debugf("Available peers:\n%s\n", str)
	}

	return peers
}
