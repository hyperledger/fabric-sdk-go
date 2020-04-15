/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"crypto/tls"
	"crypto/x509"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	commtls "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm/tls"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/cryptoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	grpcCodes "google.golang.org/grpc/codes"
)

var logger = logging.NewLogger("fabsdk/fab")
var defaultOrdererListenPort = 7050
var defaultPeerListenPort = 7051

const (
	defaultPeerConnectionTimeout          = time.Second * 10
	defaultPeerResponseTimeout            = time.Minute * 3
	defaultDiscoveryGreylistExpiryTimeout = time.Second * 10
	defaultEventRegTimeout                = time.Second * 15
	defaultOrdererConnectionTimeout       = time.Second * 15
	defaultOrdererResponseTimeout         = time.Minute * 2
	defaultQueryTimeout                   = time.Minute * 3
	defaultExecuteTimeout                 = time.Minute * 3
	defaultResMgmtTimeout                 = time.Minute * 3
	defaultDiscoveryConnectionTimeout     = time.Second * 15
	defaultDiscoveryResponseTimeout       = time.Second * 15
	defaultConnIdleInterval               = time.Second * 30
	defaultEventServiceIdleInterval       = time.Minute * 2
	defaultChannelConfigRefreshInterval   = time.Second * 90
	defaultChannelMemshpRefreshInterval   = time.Second * 60
	defaultDiscoveryRefreshInterval       = time.Second * 5
	defaultSelectionRefreshInterval       = time.Second * 5
	defaultCacheSweepInterval             = time.Second * 15

	defaultResolverStrategy                 = fab.PreferOrgStrategy
	defaultMinBlockHeightResolverMode       = fab.ResolveByThreshold
	defaultBalancer                         = fab.Random
	defaultBlockHeightLagThreshold          = 5
	defaultReconnectBlockHeightLagThreshold = 10
	defaultPeerMonitor                      = "" // The peer monitor will be enabled if necessary
	defaultPeerMonitorPeriod                = 5 * time.Second

	//default grpc opts
	defaultKeepAliveTime    = 0
	defaultKeepAliveTimeout = time.Second * 20
	defaultKeepAlivePermit  = false
	defaultFailFast         = false
	defaultAllowInsecure    = false

	defaultMaxTargets   = 2
	defaultMinResponses = 1

	defaultEntity = "_default"
)

var (
	defaultDiscoveryRetryableCodes = map[status.Group][]status.Code{
		status.GRPCTransportStatus: {
			status.Code(grpcCodes.Unavailable),
		},
		status.DiscoveryServerStatus: {
			status.QueryEndorsers,
		},
	}

	defaultDiscoveryRetryOpts = retry.Opts{
		Attempts:       6,
		InitialBackoff: 500 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		BackoffFactor:  1.75,
		RetryableCodes: defaultDiscoveryRetryableCodes,
	}

	defaultChannelPolicies = &ChannelPolicies{
		QueryChannelConfig: QueryChannelConfigPolicy{
			MaxTargets:   defaultMaxTargets,
			MinResponses: defaultMinResponses,
			RetryOpts:    retry.Opts{},
		},
		Discovery: DiscoveryPolicy{
			MaxTargets:   defaultMaxTargets,
			MinResponses: defaultMinResponses,
			RetryOpts:    defaultDiscoveryRetryOpts,
		},
		Selection: SelectionPolicy{
			SortingStrategy:         BlockHeightPriority,
			Balancer:                Random,
			BlockHeightLagThreshold: defaultBlockHeightLagThreshold,
		},
		EventService: EventServicePolicy{
			ResolverStrategy:                 string(fab.PreferOrgStrategy),
			MinBlockHeightResolverMode:       string(defaultMinBlockHeightResolverMode),
			Balancer:                         Random,
			PeerMonitor:                      defaultPeerMonitor,
			PeerMonitorPeriod:                defaultPeerMonitorPeriod,
			BlockHeightLagThreshold:          defaultBlockHeightLagThreshold,
			ReconnectBlockHeightLagThreshold: defaultReconnectBlockHeightLagThreshold,
		},
	}
)

//ConfigFromBackend returns endpoint config implementation for given backend
func ConfigFromBackend(coreBackend ...core.ConfigBackend) (fab.EndpointConfig, error) {

	config := &EndpointConfig{
		backend: lookup.New(coreBackend...),
	}

	if err := config.loadEndpointConfiguration(); err != nil {
		return nil, errors.WithMessage(err, "network configuration load failed")
	}

	//print deprecated warning
	detectDeprecatedNetworkConfig(config)

	return config, nil
}

// EndpointConfig represents the endpoint configuration for the client
type EndpointConfig struct {
	backend                  *lookup.ConfigLookup
	networkConfig            *fab.NetworkConfig
	tlsCertPool              commtls.CertPool
	entityMatchers           *entityMatchers
	peerConfigsByOrg         map[string][]fab.PeerConfig
	networkPeers             []fab.NetworkPeer
	ordererConfigs           []fab.OrdererConfig
	channelPeersByChannel    map[string][]fab.ChannelPeer
	channelOrderersByChannel map[string][]fab.OrdererConfig
	tlsClientCerts           []tls.Certificate
	peerMatchers             []matcherEntry
	ordererMatchers          []matcherEntry
	channelMatchers          []matcherEntry
	defaultPeerConfig        fab.PeerConfig
	defaultOrdererConfig     fab.OrdererConfig
	defaultChannelPolicies   fab.ChannelPolicies
	defaultChannel           *fab.ChannelEndpointConfig
}

//endpointConfigEntity contains endpoint config elements needed by endpointconfig
type endpointConfigEntity struct {
	Client        ClientConfig
	Channels      map[string]ChannelEndpointConfig
	Organizations map[string]OrganizationConfig
	Orderers      map[string]OrdererConfig
	Peers         map[string]PeerConfig
}

//entityMatchers for endpoint configuration
type entityMatchers struct {
	matchers map[string][]MatchConfig
}

//matcher entry mapping regex to match config
type matcherEntry struct {
	regex       *regexp.Regexp
	matchConfig MatchConfig
}

// Timeout reads timeouts for the given timeout type, if type is not found in the config
// then default is set as per the const value above for the corresponding type
func (c *EndpointConfig) Timeout(tType fab.TimeoutType) time.Duration {
	return c.getTimeout(tType)
}

// OrderersConfig returns a list of defined orderers
func (c *EndpointConfig) OrderersConfig() []fab.OrdererConfig {
	return c.ordererConfigs
}

// OrdererConfig returns the requested orderer
func (c *EndpointConfig) OrdererConfig(nameOrURL string) (*fab.OrdererConfig, bool, bool) {
	return c.tryMatchingOrdererConfig(nameOrURL, true)
}

// PeersConfig Retrieves the fabric peers for the specified org from the
// config file provided
func (c *EndpointConfig) PeersConfig(org string) ([]fab.PeerConfig, bool) {
	peerConfigs, ok := c.peerConfigsByOrg[strings.ToLower(org)]
	return peerConfigs, ok
}

// PeerConfig Retrieves a specific peer from the configuration by name or url
func (c *EndpointConfig) PeerConfig(nameOrURL string) (*fab.PeerConfig, bool) {
	return c.tryMatchingPeerConfig(nameOrURL, true)
}

// NetworkConfig returns the network configuration defined in the config file
func (c *EndpointConfig) NetworkConfig() *fab.NetworkConfig {
	return c.networkConfig
}

// NetworkPeers returns the network peers configuration, all the peers from all the orgs in config.
func (c *EndpointConfig) NetworkPeers() []fab.NetworkPeer {
	return c.networkPeers
}

// MappedChannelName will return channelName if it is an original channel name in the config
// if it is not, then it will try to find a channelMatcher and return its MappedName.
// If more than one matcher is found, then the first matcher in the list will be used.
func (c *EndpointConfig) mappedChannelName(networkConfig *fab.NetworkConfig, channelName string) string {

	// if channelName is the original key found in the Channels map config, then return it as is
	_, ok := networkConfig.Channels[strings.ToLower(channelName)]
	if ok {
		return channelName
	}

	// if !ok, then find a channelMatcher for channelName

	//Return if no channelMatchers are configured
	if len(c.channelMatchers) == 0 {
		return defaultEntity
	}

	//loop over channelMatchers to find the matching channel name
	for _, matcher := range c.channelMatchers {
		if matcher.regex.MatchString(channelName) {
			// get the matching matchConfig from the index number
			return matcher.matchConfig.MappedName
		}
	}

	return defaultEntity
}

// ChannelConfig returns the channel configuration
func (c *EndpointConfig) ChannelConfig(name string) *fab.ChannelEndpointConfig {

	// get the mapped channel Name
	mappedChannelName := c.mappedChannelName(c.networkConfig, name)
	if mappedChannelName == defaultEntity {
		return c.defaultChannel
	}

	//look up in network config by channelName
	ch, ok := c.networkConfig.Channels[strings.ToLower(mappedChannelName)]
	if !ok {
		return c.defaultChannel
	}
	return &ch
}

// ChannelPeers returns the channel peers configuration
func (c *EndpointConfig) ChannelPeers(name string) []fab.ChannelPeer {

	//get mapped channel name
	mappedChannelName := c.mappedChannelName(c.networkConfig, name)

	//look up in dictionary
	return c.channelPeersByChannel[strings.ToLower(mappedChannelName)]
}

// ChannelOrderers returns a list of channel orderers
func (c *EndpointConfig) ChannelOrderers(name string) []fab.OrdererConfig {
	//get mapped channel name
	mappedChannelName := c.mappedChannelName(c.networkConfig, name)

	//look up in dictionary
	return c.channelOrderersByChannel[strings.ToLower(mappedChannelName)]
}

// TLSCACertPool returns the configured cert pool. If a certConfig
// is provided, the certificate is added to the pool
func (c *EndpointConfig) TLSCACertPool() commtls.CertPool {
	return c.tlsCertPool
}

// TLSClientCerts loads the client's certs for mutual TLS
func (c *EndpointConfig) TLSClientCerts() []tls.Certificate {
	return c.tlsClientCerts
}

func (c *EndpointConfig) loadPrivateKeyFromConfig(clientConfig *ClientConfig, clientCerts tls.Certificate, cb []byte) ([]tls.Certificate, error) {

	kb := clientConfig.TLSCerts.Client.Key.Bytes()

	// load the key/cert pair from []byte
	clientCerts, err := tls.X509KeyPair(cb, kb)
	if err != nil {
		return nil, errors.Errorf("Error loading cert/key pair as TLS client credentials: %s", err)
	}

	logger.Debug("pk read from config successfully")

	return []tls.Certificate{clientCerts}, nil
}

// CryptoConfigPath ...
func (c *EndpointConfig) CryptoConfigPath() string {
	return pathvar.Subst(c.backend.GetString("client.cryptoconfig.path"))
}

func (c *EndpointConfig) getTimeout(tType fab.TimeoutType) time.Duration { //nolint
	var timeout time.Duration
	switch tType {
	case fab.PeerConnection:
		timeout = c.backend.GetDuration("client.peer.timeout.connection")
		if timeout == 0 {
			timeout = defaultPeerConnectionTimeout
		}
	case fab.PeerResponse:
		timeout = c.backend.GetDuration("client.peer.timeout.response")
		if timeout == 0 {
			timeout = defaultPeerResponseTimeout
		}
	case fab.DiscoveryGreylistExpiry:
		timeout = c.backend.GetDuration("client.peer.timeout.discovery.greylistExpiry")
		if timeout == 0 {
			timeout = defaultDiscoveryGreylistExpiryTimeout
		}
	case fab.EventReg:
		timeout = c.backend.GetDuration("client.eventService.timeout.registrationResponse")
		if timeout == 0 {
			timeout = defaultEventRegTimeout
		}
	case fab.OrdererConnection:
		timeout = c.backend.GetDuration("client.orderer.timeout.connection")
		if timeout == 0 {
			timeout = defaultOrdererConnectionTimeout
		}
	case fab.OrdererResponse:
		timeout = c.backend.GetDuration("client.orderer.timeout.response")
		if timeout == 0 {
			timeout = defaultOrdererResponseTimeout
		}
	case fab.DiscoveryConnection:
		timeout = c.backend.GetDuration("client.discovery.timeout.connection")
		if timeout == 0 {
			timeout = defaultDiscoveryConnectionTimeout
		}
	case fab.DiscoveryResponse:
		timeout = c.backend.GetDuration("client.discovery.timeout.response")
		if timeout == 0 {
			timeout = defaultDiscoveryResponseTimeout
		}
	case fab.Query:
		timeout = c.backend.GetDuration("client.global.timeout.query")
		if timeout == 0 {
			timeout = defaultQueryTimeout
		}
	case fab.Execute:
		timeout = c.backend.GetDuration("client.global.timeout.execute")
		if timeout == 0 {
			timeout = defaultExecuteTimeout
		}
	case fab.ResMgmt:
		timeout = c.backend.GetDuration("client.global.timeout.resmgmt")
		if timeout == 0 {
			timeout = defaultResMgmtTimeout
		}
	case fab.ConnectionIdle:
		timeout = c.backend.GetDuration("client.global.cache.connectionIdle")
		if timeout == 0 {
			timeout = defaultConnIdleInterval
		}
	case fab.EventServiceIdle:
		timeout = c.backend.GetDuration("client.global.cache.eventServiceIdle")
		if timeout == 0 {
			timeout = defaultEventServiceIdleInterval
		}
	case fab.ChannelConfigRefresh:
		timeout = c.backend.GetDuration("client.global.cache.channelConfig")
		if timeout == 0 {
			timeout = defaultChannelConfigRefreshInterval
		}
	case fab.ChannelMembershipRefresh:
		timeout = c.backend.GetDuration("client.global.cache.channelMembership")
		if timeout == 0 {
			timeout = defaultChannelMemshpRefreshInterval
		}
	case fab.DiscoveryServiceRefresh:
		timeout = c.backend.GetDuration("client.global.cache.discovery")
		if timeout == 0 {
			timeout = defaultDiscoveryRefreshInterval
		}
	case fab.SelectionServiceRefresh:
		timeout = c.backend.GetDuration("client.global.cache.selection")
		if timeout == 0 {
			timeout = defaultSelectionRefreshInterval
		}

	case fab.CacheSweepInterval: // EXPERIMENTAL - do we need this to be configurable?
		timeout = c.backend.GetDuration("client.cache.interval.sweep")
		if timeout == 0 {
			timeout = defaultCacheSweepInterval
		}
	}

	return timeout
}

func (c *EndpointConfig) loadEndpointConfiguration() error {

	endpointConfigurationEntity := endpointConfigEntity{}

	err := c.backend.UnmarshalKey("client", &endpointConfigurationEntity.Client)
	logger.Debugf("Client is: %+v", endpointConfigurationEntity.Client)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'client' config item to endpointConfigurationEntity.Client type")
	}

	err = c.backend.UnmarshalKey(
		"channels", &endpointConfigurationEntity.Channels,
		lookup.WithUnmarshalHookFunction(peerChannelConfigHookFunc()),
	)
	logger.Debugf("channels are: %+v", endpointConfigurationEntity.Channels)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'channels' config item to endpointConfigurationEntity.Channels type")
	}

	err = c.backend.UnmarshalKey("organizations", &endpointConfigurationEntity.Organizations)
	logger.Debugf("organizations are: %+v", endpointConfigurationEntity.Organizations)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'organizations' config item to endpointConfigurationEntity.Organizations type")
	}

	err = c.backend.UnmarshalKey("orderers", &endpointConfigurationEntity.Orderers)
	logger.Debugf("orderers are: %+v", endpointConfigurationEntity.Orderers)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'orderers' config item to endpointConfigurationEntity.Orderers type")
	}

	err = c.backend.UnmarshalKey("peers", &endpointConfigurationEntity.Peers)
	logger.Debugf("peers are: %+v", endpointConfigurationEntity.Peers)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'peers' config item to endpointConfigurationEntity.Peers type")
	}

	//load all endpointconfig entities
	err = c.loadEndpointConfigEntities(&endpointConfigurationEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to load channel configs")
	}

	return nil
}

func (c *EndpointConfig) loadEndpointConfigEntities(configEntity *endpointConfigEntity) error {

	//Compile the entityMatchers
	matchError := c.compileMatchers()
	if matchError != nil {
		return matchError
	}

	//load all TLS configs, before building any network config
	err := c.loadAllTLSConfig(configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to load network TLSConfig")
	}

	//load default configs
	err = c.loadDefaultConfigItems(configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to network config")
	}

	//load network config
	err = c.loadNetworkConfig(configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to network config")
	}

	//load peer configs by org dictionary
	c.loadPeerConfigsByOrg()

	//load network peers
	c.loadNetworkPeers()

	//load orderer configs
	err = c.loadOrdererConfigs()
	if err != nil {
		return errors.WithMessage(err, "failed to load orderer configs")
	}

	//load channel peers
	err = c.loadChannelPeers()
	if err != nil {
		return errors.WithMessage(err, "failed to load channel peers")
	}

	//load channel orderers
	err = c.loadChannelOrderers()
	if err != nil {
		return errors.WithMessage(err, "failed to load channel orderers")
	}

	//load tls cert pool
	err = c.loadTLSCertPool()
	if err != nil {
		return errors.WithMessage(err, "failed to load TLS cert pool")
	}

	c.loadDefaultChannel()

	return nil
}

func (c *EndpointConfig) loadDefaultChannel() {
	defChCfg, ok := c.networkConfig.Channels[defaultEntity]
	if ok {
		c.defaultChannel = &fab.ChannelEndpointConfig{Peers: defChCfg.Peers, Orderers: defChCfg.Orderers, Policies: defChCfg.Policies}
		delete(c.networkConfig.Channels, defaultEntity)
	} else {
		logger.Debugf("No default config. Returning hard-coded defaults.")
		c.defaultChannel = &fab.ChannelEndpointConfig{Policies: c.getChannelPolicies(defaultChannelPolicies)}
	}
}

func (c *EndpointConfig) loadDefaultConfigItems(configEntity *endpointConfigEntity) error {
	//default orderer config
	err := c.loadDefaultOrderer(configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to load default orderer")
	}

	//default peer config
	err = c.loadDefaultPeer(configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to load default peer")
	}

	//default channel policies
	c.loadDefaultChannelPolicies(configEntity)
	return nil
}

func (c *EndpointConfig) loadNetworkConfig(configEntity *endpointConfigEntity) error {

	networkConfig := fab.NetworkConfig{}

	//Channels
	networkConfig.Channels = make(map[string]fab.ChannelEndpointConfig)

	// Load default channel config first since it will be used for defaulting  other channels peers and orderers
	defChNwCfg, ok := configEntity.Channels[defaultEntity]
	if ok {
		networkConfig.Channels[defaultEntity] = c.loadChannelEndpointConfig(defChNwCfg, ChannelEndpointConfig{})
	} else {
		networkConfig.Channels[defaultEntity] = fab.ChannelEndpointConfig{Policies: c.getChannelPolicies(defaultChannelPolicies)}
	}

	for chID, chNwCfg := range configEntity.Channels {
		if chID == defaultEntity {
			// default entity has been loaded already
			continue
		}

		networkConfig.Channels[chID] = c.loadChannelEndpointConfig(chNwCfg, defChNwCfg)
	}

	//Organizations
	networkConfig.Organizations = make(map[string]fab.OrganizationConfig)
	for orgName, orgConfig := range configEntity.Organizations {

		tlsKeyCertPairs := make(map[string]fab.CertKeyPair)
		for user, tlsKeyPair := range orgConfig.Users {
			tlsKeyCertPairs[user] = fab.CertKeyPair{
				Cert: tlsKeyPair.Cert.Bytes(),
				Key:  tlsKeyPair.Key.Bytes(),
			}
		}

		networkConfig.Organizations[orgName] = fab.OrganizationConfig{
			MSPID:                  orgConfig.MSPID,
			CryptoPath:             orgConfig.CryptoPath,
			Peers:                  orgConfig.Peers,
			CertificateAuthorities: orgConfig.CertificateAuthorities,
			Users:                  tlsKeyCertPairs,
		}

	}

	//Orderers
	err := c.loadAllOrdererConfigs(&networkConfig, configEntity.Orderers)
	if err != nil {
		return err
	}

	//Peers
	err = c.loadAllPeerConfigs(&networkConfig, configEntity.Peers)
	if err != nil {
		return err
	}

	c.networkConfig = &networkConfig
	return nil
}

func (c *EndpointConfig) loadChannelEndpointConfig(chNwCfg ChannelEndpointConfig, defChNwCfg ChannelEndpointConfig) fab.ChannelEndpointConfig {

	chPeers := make(map[string]fab.PeerChannelConfig)

	chNwCfgPeers := chNwCfg.Peers
	if len(chNwCfgPeers) == 0 {
		//fill peers in with default channel peers
		chNwCfgPeers = defChNwCfg.Peers
	}

	for chPeer, chPeerCfg := range chNwCfgPeers {
		if c.isPeerToBeIgnored(chPeer) {
			//filter peer to be ignored
			continue
		}
		chPeers[chPeer] = fab.PeerChannelConfig{
			EndorsingPeer:  chPeerCfg.EndorsingPeer,
			ChaincodeQuery: chPeerCfg.ChaincodeQuery,
			LedgerQuery:    chPeerCfg.LedgerQuery,
			EventSource:    chPeerCfg.EventSource,
		}
	}

	chOrderers := []string{}

	chNwCfgOrderers := chNwCfg.Orderers
	if len(chNwCfgOrderers) == 0 {
		//fill orderers in with default channel orderers
		chNwCfgOrderers = defChNwCfg.Orderers
	}

	for _, name := range chNwCfgOrderers {
		if !c.isOrdererToBeIgnored(name) {
			//filter orderer to be ignored
			chOrderers = append(chOrderers, name)
		}
	}

	// Policies use default channel policies if info is missing
	return fab.ChannelEndpointConfig{
		Peers:    chPeers,
		Orderers: chOrderers,
		Policies: c.addMissingChannelPoliciesItems(chNwCfg),
	}
}

func (c *EndpointConfig) getChannelPolicies(policies *ChannelPolicies) fab.ChannelPolicies {

	discoveryPolicy := fab.DiscoveryPolicy{

		MaxTargets:   policies.Discovery.MaxTargets,
		MinResponses: policies.Discovery.MinResponses,
		RetryOpts:    policies.Discovery.RetryOpts,
	}

	selectionPolicy := fab.SelectionPolicy{

		SortingStrategy:         fab.SelectionSortingStrategy(policies.Selection.SortingStrategy),
		Balancer:                fab.BalancerType(policies.Selection.Balancer),
		BlockHeightLagThreshold: policies.Selection.BlockHeightLagThreshold,
	}

	channelCfgPolicy := fab.QueryChannelConfigPolicy{
		MaxTargets:   policies.QueryChannelConfig.MaxTargets,
		MinResponses: policies.QueryChannelConfig.MinResponses,
		RetryOpts:    policies.QueryChannelConfig.RetryOpts,
	}

	eventServicePolicy := fab.EventServicePolicy{
		ResolverStrategy:                 fab.ResolverStrategy(policies.EventService.ResolverStrategy),
		MinBlockHeightResolverMode:       fab.MinBlockHeightResolverMode(policies.EventService.MinBlockHeightResolverMode),
		Balancer:                         fab.BalancerType(policies.EventService.Balancer),
		BlockHeightLagThreshold:          policies.EventService.BlockHeightLagThreshold,
		PeerMonitor:                      fab.EnabledDisabled(policies.EventService.PeerMonitor),
		ReconnectBlockHeightLagThreshold: policies.EventService.ReconnectBlockHeightLagThreshold,
		PeerMonitorPeriod:                policies.EventService.PeerMonitorPeriod,
	}

	return fab.ChannelPolicies{
		Discovery:          discoveryPolicy,
		Selection:          selectionPolicy,
		QueryChannelConfig: channelCfgPolicy,
		EventService:       eventServicePolicy,
	}
}

func (c *EndpointConfig) addMissingChannelPoliciesItems(chNwCfg ChannelEndpointConfig) fab.ChannelPolicies {

	policies := c.getChannelPolicies(&chNwCfg.Policies)

	policies.Discovery = c.addMissingDiscoveryPolicyInfo(policies.Discovery)
	policies.Selection = c.addMissingSelectionPolicyInfo(policies.Selection)
	policies.QueryChannelConfig = c.addMissingQueryChannelConfigPolicyInfo(policies.QueryChannelConfig)
	policies.EventService = c.addMissingEventServicePolicyInfo(policies.EventService)

	return policies
}

func (c *EndpointConfig) addMissingDiscoveryPolicyInfo(policy fab.DiscoveryPolicy) fab.DiscoveryPolicy {

	if policy.MaxTargets == 0 {
		policy.MaxTargets = c.defaultChannelPolicies.Discovery.MaxTargets
	}

	if policy.MinResponses == 0 {
		policy.MinResponses = c.defaultChannelPolicies.Discovery.MinResponses
	}

	if isEmpty(policy.RetryOpts) {
		policy.RetryOpts = c.defaultChannelPolicies.Discovery.RetryOpts
	} else {
		policy.RetryOpts = addMissingRetryOpts(policy.RetryOpts, c.defaultChannelPolicies.Discovery.RetryOpts)
	}

	return policy
}

func (c *EndpointConfig) addMissingSelectionPolicyInfo(policy fab.SelectionPolicy) fab.SelectionPolicy {

	if policy.SortingStrategy == "" {
		policy.SortingStrategy = c.defaultChannelPolicies.Selection.SortingStrategy
	}

	if policy.Balancer == "" {
		policy.Balancer = c.defaultChannelPolicies.Selection.Balancer
	}

	if policy.BlockHeightLagThreshold == 0 {
		policy.BlockHeightLagThreshold = defaultBlockHeightLagThreshold
	}

	return policy
}

func (c *EndpointConfig) addMissingQueryChannelConfigPolicyInfo(policy fab.QueryChannelConfigPolicy) fab.QueryChannelConfigPolicy {

	if policy.MaxTargets == 0 {
		policy.MaxTargets = c.defaultChannelPolicies.QueryChannelConfig.MaxTargets
	}

	if policy.MinResponses == 0 {
		policy.MinResponses = c.defaultChannelPolicies.QueryChannelConfig.MinResponses
	}

	if isEmpty(policy.RetryOpts) {
		policy.RetryOpts = c.defaultChannelPolicies.QueryChannelConfig.RetryOpts
	} else {
		policy.RetryOpts = addMissingRetryOpts(policy.RetryOpts, c.defaultChannelPolicies.QueryChannelConfig.RetryOpts)
	}

	return policy
}

func (c *EndpointConfig) addMissingEventServicePolicyInfo(policy fab.EventServicePolicy) fab.EventServicePolicy {
	if policy.Balancer == "" {
		policy.Balancer = c.defaultChannelPolicies.EventService.Balancer
	}
	if policy.BlockHeightLagThreshold == 0 {
		policy.BlockHeightLagThreshold = c.defaultChannelPolicies.EventService.BlockHeightLagThreshold
	}
	if policy.ResolverStrategy == "" {
		policy.ResolverStrategy = c.defaultChannelPolicies.EventService.ResolverStrategy
	}
	if policy.MinBlockHeightResolverMode == "" {
		policy.MinBlockHeightResolverMode = c.defaultChannelPolicies.EventService.MinBlockHeightResolverMode
	}
	if policy.PeerMonitor == "" {
		policy.PeerMonitor = c.defaultChannelPolicies.EventService.PeerMonitor
	}
	if policy.ReconnectBlockHeightLagThreshold == 0 {
		policy.ReconnectBlockHeightLagThreshold = c.defaultChannelPolicies.EventService.ReconnectBlockHeightLagThreshold
	}
	if policy.PeerMonitorPeriod == 0 {
		policy.PeerMonitorPeriod = c.defaultChannelPolicies.EventService.PeerMonitorPeriod
	}

	return policy
}

func addMissingRetryOpts(opts retry.Opts, defaultOpts retry.Opts) retry.Opts {
	// If retry opts are defined then Attempts must be defined, otherwise
	// we cannot distinguish between default 0 and intentional 0 to disable retries for that channel

	empty := retry.Opts{}

	if opts.InitialBackoff == empty.InitialBackoff {
		opts.InitialBackoff = defaultOpts.InitialBackoff
	}

	if opts.BackoffFactor == empty.BackoffFactor {
		opts.BackoffFactor = defaultOpts.BackoffFactor
	}

	if opts.MaxBackoff == empty.MaxBackoff {
		opts.MaxBackoff = defaultOpts.MaxBackoff
	}

	if len(opts.RetryableCodes) == len(empty.RetryableCodes) {
		opts.RetryableCodes = defaultOpts.RetryableCodes
	}

	return opts
}

func isEmpty(opts retry.Opts) bool {

	empty := retry.Opts{}
	if opts.Attempts == empty.Attempts &&
		opts.InitialBackoff == empty.InitialBackoff &&
		opts.BackoffFactor == empty.BackoffFactor &&
		opts.MaxBackoff == empty.MaxBackoff &&
		len(opts.RetryableCodes) == len(empty.RetryableCodes) {
		return true
	}

	return false
}

func (c *EndpointConfig) loadAllPeerConfigs(networkConfig *fab.NetworkConfig, entityPeers map[string]PeerConfig) error {
	networkConfig.Peers = make(map[string]fab.PeerConfig)
	for name, peerConfig := range entityPeers {
		if name == defaultEntity || c.isPeerToBeIgnored(name) {
			//filter default and ignored peers
			continue
		}
		tlsCert, _, err := peerConfig.TLSCACerts.TLSCert()
		if err != nil {
			return errors.WithMessage(err, "failed to load peer network config")
		}
		networkConfig.Peers[name] = c.addMissingPeerConfigItems(name, fab.PeerConfig{
			URL:         peerConfig.URL,
			GRPCOptions: peerConfig.GRPCOptions,
			TLSCACert:   tlsCert,
		})
	}
	return nil
}

func (c *EndpointConfig) loadAllOrdererConfigs(networkConfig *fab.NetworkConfig, entityOrderers map[string]OrdererConfig) error {
	networkConfig.Orderers = make(map[string]fab.OrdererConfig)
	for name, ordererConfig := range entityOrderers {
		if name == defaultEntity || c.isOrdererToBeIgnored(name) {
			//filter default and ignored orderers
			continue
		}
		tlsCert, _, err := ordererConfig.TLSCACerts.TLSCert()
		if err != nil {
			return errors.WithMessage(err, "failed to load orderer network config")
		}
		networkConfig.Orderers[name] = c.addMissingOrdererConfigItems(name, fab.OrdererConfig{
			URL:         ordererConfig.URL,
			GRPCOptions: ordererConfig.GRPCOptions,
			TLSCACert:   tlsCert,
		})
	}
	return nil
}

func (c *EndpointConfig) addMissingPeerConfigItems(name string, config fab.PeerConfig) fab.PeerConfig {

	// peer URL
	if config.URL == "" {
		if c.defaultPeerConfig.URL == "" {
			config.URL = name + ":" + strconv.Itoa(defaultPeerListenPort)
		} else {
			config.URL = c.defaultPeerConfig.URL
		}
	}

	//tls ca certs
	if config.TLSCACert == nil {
		config.TLSCACert = c.defaultPeerConfig.TLSCACert
	}

	//if no grpc opts found
	if len(config.GRPCOptions) == 0 {
		config.GRPCOptions = c.defaultPeerConfig.GRPCOptions
		return config
	}

	//missing grpc opts
	for name, val := range c.defaultPeerConfig.GRPCOptions {
		_, ok := config.GRPCOptions[name]
		if !ok {
			config.GRPCOptions[name] = val
		}
	}

	return config
}

func (c *EndpointConfig) addMissingOrdererConfigItems(name string, config fab.OrdererConfig) fab.OrdererConfig {
	// orderer URL
	if config.URL == "" {
		if c.defaultOrdererConfig.URL == "" {
			config.URL = name + ":" + strconv.Itoa(defaultOrdererListenPort)
		} else {
			config.URL = c.defaultOrdererConfig.URL
		}
	}

	//tls ca certs
	if config.TLSCACert == nil {
		config.TLSCACert = c.defaultOrdererConfig.TLSCACert
	}

	//if no grpc opts found
	if len(config.GRPCOptions) == 0 {
		config.GRPCOptions = c.defaultOrdererConfig.GRPCOptions
		return config
	}

	//missing grpc opts
	for name, val := range c.defaultOrdererConfig.GRPCOptions {
		_, ok := config.GRPCOptions[name]
		if !ok {
			config.GRPCOptions[name] = val
		}
	}

	return config
}

func (c *EndpointConfig) loadDefaultOrderer(configEntity *endpointConfigEntity) error {

	defaultEntityOrderer, ok := configEntity.Orderers[defaultEntity]
	if !ok {
		defaultEntityOrderer = OrdererConfig{
			GRPCOptions: make(map[string]interface{}),
		}
	}

	c.defaultOrdererConfig = fab.OrdererConfig{
		GRPCOptions: defaultEntityOrderer.GRPCOptions,
	}

	//set defaults for missing grpc opts

	//keep-alive-time
	_, ok = c.defaultOrdererConfig.GRPCOptions["keep-alive-time"]
	if !ok {
		c.defaultOrdererConfig.GRPCOptions["keep-alive-time"] = defaultKeepAliveTime
	}

	//keep-alive-timeout
	_, ok = c.defaultOrdererConfig.GRPCOptions["keep-alive-timeout"]
	if !ok {
		c.defaultOrdererConfig.GRPCOptions["keep-alive-timeout"] = defaultKeepAliveTimeout
	}

	//keep-alive-permit
	_, ok = c.defaultOrdererConfig.GRPCOptions["keep-alive-permit"]
	if !ok {
		c.defaultOrdererConfig.GRPCOptions["keep-alive-permit"] = defaultKeepAlivePermit
	}

	//fail-fast
	_, ok = c.defaultOrdererConfig.GRPCOptions["fail-fast"]
	if !ok {
		c.defaultOrdererConfig.GRPCOptions["fail-fast"] = defaultFailFast
	}

	//allow-insecure
	_, ok = c.defaultOrdererConfig.GRPCOptions["allow-insecure"]
	if !ok {
		c.defaultOrdererConfig.GRPCOptions["allow-insecure"] = defaultAllowInsecure
	}

	var err error
	c.defaultOrdererConfig.TLSCACert, _, err = defaultEntityOrderer.TLSCACerts.TLSCert()
	if err != nil {
		return errors.WithMessage(err, "failed to load default orderer network config")
	}

	return nil
}

func (c *EndpointConfig) loadDefaultChannelPolicies(configEntity *endpointConfigEntity) {

	var defaultChPolicies fab.ChannelPolicies
	defaultChannel, ok := configEntity.Channels[defaultEntity]
	if !ok {
		defaultChPolicies = c.getChannelPolicies(defaultChannelPolicies)
	} else {
		defaultChPolicies = c.getChannelPolicies(&defaultChannel.Policies)
	}

	c.loadDefaultDiscoveryPolicy(&defaultChPolicies.Discovery)
	c.loadDefaultSelectionPolicy(&defaultChPolicies.Selection)
	c.loadDefaultQueryChannelPolicy(&defaultChPolicies.QueryChannelConfig)
	c.loadDefaultEventServicePolicy(&defaultChPolicies.EventService)

	c.defaultChannelPolicies = defaultChPolicies

}

func (c *EndpointConfig) loadDefaultDiscoveryPolicy(policy *fab.DiscoveryPolicy) {
	if policy.MaxTargets == 0 {
		policy.MaxTargets = defaultMaxTargets
	}

	if policy.MinResponses == 0 {
		policy.MinResponses = defaultMinResponses
	}

	if len(policy.RetryOpts.RetryableCodes) == 0 {
		policy.RetryOpts.RetryableCodes = defaultDiscoveryRetryableCodes
	}
}

func (c *EndpointConfig) loadDefaultSelectionPolicy(policy *fab.SelectionPolicy) {
	if policy.SortingStrategy == "" {
		policy.SortingStrategy = fab.BlockHeightPriority
	}

	if policy.Balancer == "" {
		policy.Balancer = fab.RoundRobin
	}

	if policy.BlockHeightLagThreshold == 0 {
		policy.BlockHeightLagThreshold = defaultBlockHeightLagThreshold
	}
}

func (c *EndpointConfig) loadDefaultQueryChannelPolicy(policy *fab.QueryChannelConfigPolicy) {
	if policy.MaxTargets == 0 {
		policy.MaxTargets = defaultMaxTargets
	}

	if policy.MinResponses == 0 {
		policy.MinResponses = defaultMinResponses
	}
}

func (c *EndpointConfig) loadDefaultEventServicePolicy(policy *fab.EventServicePolicy) {
	if policy.ResolverStrategy == "" {
		policy.ResolverStrategy = defaultResolverStrategy
	}

	if policy.MinBlockHeightResolverMode == "" {
		policy.MinBlockHeightResolverMode = defaultMinBlockHeightResolverMode
	}

	if policy.Balancer == "" {
		policy.Balancer = defaultBalancer
	}

	if policy.BlockHeightLagThreshold == 0 {
		policy.BlockHeightLagThreshold = defaultBlockHeightLagThreshold
	}

	if policy.ReconnectBlockHeightLagThreshold == 0 {
		policy.ReconnectBlockHeightLagThreshold = defaultReconnectBlockHeightLagThreshold
	}

	if policy.PeerMonitorPeriod == 0 {
		policy.PeerMonitorPeriod = defaultPeerMonitorPeriod
	}
}

func (c *EndpointConfig) loadDefaultPeer(configEntity *endpointConfigEntity) error {

	defaultEntityPeer, ok := configEntity.Peers[defaultEntity]
	if !ok {
		defaultEntityPeer = PeerConfig{
			GRPCOptions: make(map[string]interface{}),
		}
	}

	c.defaultPeerConfig = fab.PeerConfig{
		GRPCOptions: defaultEntityPeer.GRPCOptions,
	}

	//set defaults for missing grpc opts

	//keep-alive-time
	_, ok = c.defaultPeerConfig.GRPCOptions["keep-alive-time"]
	if !ok {
		c.defaultPeerConfig.GRPCOptions["keep-alive-time"] = defaultKeepAliveTime
	}

	//keep-alive-timeout
	_, ok = c.defaultPeerConfig.GRPCOptions["keep-alive-timeout"]
	if !ok {
		c.defaultPeerConfig.GRPCOptions["keep-alive-timeout"] = defaultKeepAliveTimeout
	}

	//keep-alive-permit
	_, ok = c.defaultPeerConfig.GRPCOptions["keep-alive-permit"]
	if !ok {
		c.defaultPeerConfig.GRPCOptions["keep-alive-permit"] = defaultKeepAlivePermit
	}

	//fail-fast
	_, ok = c.defaultPeerConfig.GRPCOptions["fail-fast"]
	if !ok {
		c.defaultPeerConfig.GRPCOptions["fail-fast"] = defaultFailFast
	}

	//allow-insecure
	_, ok = c.defaultPeerConfig.GRPCOptions["allow-insecure"]
	if !ok {
		c.defaultPeerConfig.GRPCOptions["allow-insecure"] = defaultAllowInsecure
	}

	var err error
	c.defaultPeerConfig.TLSCACert, _, err = defaultEntityPeer.TLSCACerts.TLSCert()
	if err != nil {
		return errors.WithMessage(err, "failed to load default peer network config")
	}

	return nil
}

//loadAllTLSConfig pre-loads all network TLS Configs
func (c *EndpointConfig) loadAllTLSConfig(configEntity *endpointConfigEntity) error {
	//resolve path and load bytes
	err := c.loadClientTLSConfig(configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to load client TLSConfig ")
	}

	//resolve path and load bytes
	err = c.loadOrgTLSConfig(configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to load org TLSConfig ")
	}

	//resolve path and load bytes
	err = c.loadOrdererPeerTLSConfig(configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to load orderer/peer TLSConfig ")
	}

	//preload TLS client certs
	err = c.loadTLSClientCerts(configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to load TLS client certs ")
	}

	return nil
}

//loadClientTLSConfig pre-loads all TLSConfig bytes in client config
func (c *EndpointConfig) loadClientTLSConfig(configEntity *endpointConfigEntity) error {
	//Clients Config
	//resolve paths and org name
	configEntity.Client.Organization = strings.ToLower(configEntity.Client.Organization)
	configEntity.Client.TLSCerts.Client.Key.Path = pathvar.Subst(configEntity.Client.TLSCerts.Client.Key.Path)
	configEntity.Client.TLSCerts.Client.Cert.Path = pathvar.Subst(configEntity.Client.TLSCerts.Client.Cert.Path)

	//pre load client key and cert bytes
	err := configEntity.Client.TLSCerts.Client.Key.LoadBytes()
	if err != nil {
		return errors.WithMessage(err, "failed to load client key")
	}

	err = configEntity.Client.TLSCerts.Client.Cert.LoadBytes()
	if err != nil {
		return errors.WithMessage(err, "failed to load client cert")
	}

	return nil
}

//loadOrgTLSConfig pre-loads all TLSConfig bytes in organizations
func (c *EndpointConfig) loadOrgTLSConfig(configEntity *endpointConfigEntity) error {

	//Organizations Config
	for org, orgConfig := range configEntity.Organizations {
		for user, userConfig := range orgConfig.Users {
			//resolve paths
			userConfig.Key.Path = pathvar.Subst(userConfig.Key.Path)
			userConfig.Cert.Path = pathvar.Subst(userConfig.Cert.Path)
			//pre load key and cert bytes
			err := userConfig.Key.LoadBytes()
			if err != nil {
				return errors.WithMessage(err, "failed to load org key")
			}

			err = userConfig.Cert.LoadBytes()
			if err != nil {
				return errors.WithMessage(err, "failed to load org cert")
			}
			orgConfig.Users[user] = userConfig
		}
		configEntity.Organizations[org] = orgConfig
	}

	return nil
}

//loadTLSConfig pre-loads all TLSConfig bytes in Orderer and Peer configs
func (c *EndpointConfig) loadOrdererPeerTLSConfig(configEntity *endpointConfigEntity) error {

	//Orderers Config
	for orderer, ordererConfig := range configEntity.Orderers {
		//resolve paths
		ordererConfig.TLSCACerts.Path = pathvar.Subst(ordererConfig.TLSCACerts.Path)
		//pre load key and cert bytes
		err := ordererConfig.TLSCACerts.LoadBytes()
		if err != nil {
			return errors.WithMessage(err, "failed to load orderer cert")
		}
		configEntity.Orderers[orderer] = ordererConfig
	}

	//Peer Config
	for peer, peerConfig := range configEntity.Peers {
		//resolve paths
		peerConfig.TLSCACerts.Path = pathvar.Subst(peerConfig.TLSCACerts.Path)
		//pre load key and cert bytes
		err := peerConfig.TLSCACerts.LoadBytes()
		if err != nil {
			return errors.WithMessage(err, "failed to load peer cert")
		}
		configEntity.Peers[peer] = peerConfig
	}

	return nil
}

func (c *EndpointConfig) loadPeerConfigsByOrg() {

	c.peerConfigsByOrg = make(map[string][]fab.PeerConfig)

	for orgName, orgConfig := range c.networkConfig.Organizations {
		orgPeers := orgConfig.Peers
		peers := []fab.PeerConfig{}

		for _, peerName := range orgPeers {
			p, ok := c.tryMatchingPeerConfig(peerName, false)
			if !ok {
				continue
			}

			if err := c.verifyPeerConfig(p, peerName, endpoint.IsTLSEnabled(p.URL)); err != nil {
				continue
			}

			peers = append(peers, *p)
		}
		c.peerConfigsByOrg[strings.ToLower(orgName)] = peers
	}

}

func (c *EndpointConfig) loadNetworkPeers() {

	var netPeers []fab.NetworkPeer
	for org, peerConfigs := range c.peerConfigsByOrg {

		orgConfig, ok := c.networkConfig.Organizations[org]
		if !ok {
			continue
		}

		for _, peerConfig := range peerConfigs {
			netPeers = append(netPeers, fab.NetworkPeer{PeerConfig: peerConfig, MSPID: orgConfig.MSPID})
		}
	}

	c.networkPeers = netPeers
}

func (c *EndpointConfig) loadOrdererConfigs() error {

	ordererConfigs := []fab.OrdererConfig{}
	for name := range c.networkConfig.Orderers {

		matchedOrderer, ok, ignoreOrderer := c.tryMatchingOrdererConfig(name, false)
		if !ok || ignoreOrderer {
			continue
		}

		if matchedOrderer.TLSCACert == nil && !c.backend.GetBool("client.tlsCerts.systemCertPool") {
			//check for TLS config only if secured connection is enabled
			allowInSecure := matchedOrderer.GRPCOptions["allow-insecure"] == true
			if endpoint.AttemptSecured(matchedOrderer.URL, allowInSecure) {
				return errors.Errorf("Orderer has no certs configured. Make sure TLSCACerts.Pem or TLSCACerts.Path is set for %s", matchedOrderer.URL)
			}
		}

		ordererConfigs = append(ordererConfigs, *matchedOrderer)
	}
	c.ordererConfigs = ordererConfigs
	return nil
}

func (c *EndpointConfig) loadChannelPeers() error {

	channelPeersByChannel := make(map[string][]fab.ChannelPeer)

	for channelID, channelConfig := range c.networkConfig.Channels {
		peers := []fab.ChannelPeer{}
		for peerName, chPeerConfig := range channelConfig.Peers {
			p, ok := c.tryMatchingPeerConfig(strings.ToLower(peerName), false)
			if !ok {
				continue
			}

			if err := c.verifyPeerConfig(p, peerName, endpoint.IsTLSEnabled(p.URL)); err != nil {
				logger.Debugf("Verify PeerConfig failed for peer [%s], cause : [%s]", peerName, err)
				return err
			}

			mspID, ok := c.peerMSPID(peerName)
			if !ok {
				return errors.Errorf("unable to find MSP ID for peer : %s", peerName)
			}

			networkPeer := fab.NetworkPeer{PeerConfig: *p, MSPID: mspID}

			peer := fab.ChannelPeer{PeerChannelConfig: chPeerConfig, NetworkPeer: networkPeer}

			peers = append(peers, peer)
		}
		channelPeersByChannel[strings.ToLower(channelID)] = peers
	}

	c.channelPeersByChannel = channelPeersByChannel

	return nil
}

func (c *EndpointConfig) loadChannelOrderers() error {

	channelOrderersByChannel := make(map[string][]fab.OrdererConfig)

	for channelID, channelConfig := range c.networkConfig.Channels {
		orderers := []fab.OrdererConfig{}
		for _, ordererName := range channelConfig.Orderers {

			orderer, ok, ignoreOrderer := c.tryMatchingOrdererConfig(strings.ToLower(ordererName), false)
			if ignoreOrderer {
				continue
			}

			if !ok {
				return errors.Errorf("Could not find Orderer Config for channel orderer [%s]", ordererName)
			}
			orderers = append(orderers, *orderer)
		}
		channelOrderersByChannel[strings.ToLower(channelID)] = orderers
	}

	c.channelOrderersByChannel = channelOrderersByChannel

	return nil
}

func (c *EndpointConfig) loadTLSCertPool() error {

	var err error
	c.tlsCertPool, err = commtls.NewCertPool(c.backend.GetBool("client.tlsCerts.systemCertPool"))
	if err != nil {
		return errors.WithMessage(err, "failed to create cert pool")
	}

	// preemptively add all TLS certs to cert pool as adding them at request time
	// is expensive
	certs, err := c.loadTLSCerts()
	if err != nil {
		logger.Infof("could not cache TLS certs: %s", err)
	}

	//add certs to cert pool
	c.tlsCertPool.Add(certs...)
	//update cetr pool
	if _, err := c.tlsCertPool.Get(); err != nil {
		return errors.WithMessage(err, "cert pool load failed")
	}
	return nil
}

// loadTLSClientCerts loads the client's certs for mutual TLS
// It checks the config for embedded pem files before looking for cert files
func (c *EndpointConfig) loadTLSClientCerts(configEntity *endpointConfigEntity) error {

	var clientCerts tls.Certificate
	cb := configEntity.Client.TLSCerts.Client.Cert.Bytes()
	if len(cb) == 0 {
		// if no cert found in the config, empty cert chain should be used
		c.tlsClientCerts = []tls.Certificate{clientCerts}
		return nil
	}

	// Load private key from cert using default crypto suite
	cs := cryptosuite.GetDefault()
	pk, err := cryptoutil.GetPrivateKeyFromCert(cb, cs)

	// If CryptoSuite fails to load private key from cert then load private key from config
	if err != nil || pk == nil {
		logger.Debugf("Reading pk from config, unable to retrieve from cert: %s", err)
		tlsClientCerts, error := c.loadPrivateKeyFromConfig(&configEntity.Client, clientCerts, cb)
		if error != nil {
			return errors.WithMessage(error, "failed to load TLS client certs")
		}
		c.tlsClientCerts = tlsClientCerts
		return nil
	}

	// private key was retrieved from cert
	clientCerts, err = cryptoutil.X509KeyPair(cb, pk, cs)
	if err != nil {
		return errors.WithMessage(err, "failed to load TLS client certs, failed to get X509KeyPair")
	}

	c.tlsClientCerts = []tls.Certificate{clientCerts}
	return nil
}

func (c *EndpointConfig) isPeerToBeIgnored(peerName string) bool {
	for _, matcher := range c.peerMatchers {
		if matcher.regex.MatchString(peerName) {
			return matcher.matchConfig.IgnoreEndpoint
		}
	}
	return false
}

func (c *EndpointConfig) isOrdererToBeIgnored(ordererName string) bool {
	for _, matcher := range c.ordererMatchers {
		if matcher.regex.MatchString(ordererName) {
			return matcher.matchConfig.IgnoreEndpoint
		}
	}
	return false
}

func (c *EndpointConfig) tryMatchingPeerConfig(peerSearchKey string, searchByURL bool) (*fab.PeerConfig, bool) {

	//loop over peer entity matchers to find the matching peer
	for _, matcher := range c.peerMatchers {
		if matcher.regex.MatchString(peerSearchKey) {
			return c.matchPeer(peerSearchKey, matcher)
		}
		logger.Debugf("Peer [%s] did not match using matcher [%s]", peerSearchKey, matcher.regex.String())
	}

	//direct lookup if peer matchers are not configured or no matchers matched
	peerConfig, ok := c.networkConfig.Peers[strings.ToLower(peerSearchKey)]
	if ok {
		return &peerConfig, true
	}

	if searchByURL {
		//lookup by URL
		for _, staticPeerConfig := range c.networkConfig.Peers {
			if strings.EqualFold(staticPeerConfig.URL, peerSearchKey) {
				return &fab.PeerConfig{
					URL:         staticPeerConfig.URL,
					GRPCOptions: staticPeerConfig.GRPCOptions,
					TLSCACert:   staticPeerConfig.TLSCACert,
				}, true
			}
		}
	}

	if searchByURL && strings.Contains(peerSearchKey, ":") {
		return &fab.PeerConfig{
			URL:         peerSearchKey,
			GRPCOptions: c.defaultPeerConfig.GRPCOptions,
			TLSCACert:   c.defaultPeerConfig.TLSCACert,
		}, true
	}

	return nil, false
}

func (c *EndpointConfig) matchPeer(peerSearchKey string, matcher matcherEntry) (*fab.PeerConfig, bool) {

	if matcher.matchConfig.IgnoreEndpoint {
		logger.Debugf("Ignoring peer `%s` since entity matcher IgnoreEndpoint flag is on", peerSearchKey)
		return nil, false
	}

	mappedHost := c.regexMatchAndReplace(matcher.regex, peerSearchKey, matcher.matchConfig.MappedHost)

	matchedPeer := c.getMappedPeer(mappedHost)
	if matchedPeer == nil {
		logger.Debugf("Could not find mapped host [%s] for peer [%s]", matcher.matchConfig.MappedHost, peerSearchKey)
		return nil, false
	}

	//URLSubstitutionExp if found use from entity matcher otherwise use from mapped host
	if matcher.matchConfig.URLSubstitutionExp != "" {
		matchedPeer.URL = c.regexMatchAndReplace(matcher.regex, peerSearchKey, matcher.matchConfig.URLSubstitutionExp)
	}

	//SSLTargetOverrideURLSubstitutionExp if found use from entity matcher otherwise use from mapped host
	if matcher.matchConfig.SSLTargetOverrideURLSubstitutionExp != "" {
		matchedPeer.GRPCOptions["ssl-target-name-override"] = c.regexMatchAndReplace(matcher.regex, peerSearchKey, matcher.matchConfig.SSLTargetOverrideURLSubstitutionExp)
	}

	//if no URL to add from entity matcher or from mapped host or from default peer
	if matchedPeer.URL == "" {
		matchedPeer.URL = c.getDefaultMatchingURL(peerSearchKey)
	}

	return matchedPeer, true
}

//getDefaultMatchingURL if search key is a URL then returns search key as URL otherwise returns empty
func (c *EndpointConfig) getDefaultMatchingURL(searchKey string) string {
	if strings.Contains(searchKey, ":") {
		return searchKey
	}
	return ""
}

func (c *EndpointConfig) getMappedPeer(host string) *fab.PeerConfig {
	//Get the peerConfig from mapped host
	peerConfig, ok := c.networkConfig.Peers[strings.ToLower(host)]
	if !ok {
		peerConfig = c.defaultPeerConfig
	}

	mappedConfig := fab.PeerConfig{
		URL:         peerConfig.URL,
		TLSCACert:   peerConfig.TLSCACert,
		GRPCOptions: make(map[string]interface{}),
	}

	for key, val := range peerConfig.GRPCOptions {
		mappedConfig.GRPCOptions[key] = val
	}

	return &mappedConfig
}

func (c *EndpointConfig) tryMatchingOrdererConfig(ordererSearchKey string, searchByURL bool) (*fab.OrdererConfig, bool, bool) {

	//loop over orderer entity matchers to find the matching orderer
	for _, matcher := range c.ordererMatchers {
		if matcher.regex.MatchString(ordererSearchKey) {
			return c.matchOrderer(ordererSearchKey, matcher)
		}
		logger.Debugf("Orderer [%s] did not match using matcher [%s]", ordererSearchKey, matcher.regex.String())
	}

	//direct lookup if orderer matchers are not configured or no matchers matched
	orderer, ok := c.networkConfig.Orderers[strings.ToLower(ordererSearchKey)]
	if ok {
		return &orderer, true, false
	}

	if searchByURL {
		//lookup by URL
		for _, ordererCfg := range c.OrderersConfig() {
			if strings.EqualFold(ordererCfg.URL, ordererSearchKey) {
				return &fab.OrdererConfig{
					URL:         ordererCfg.URL,
					GRPCOptions: ordererCfg.GRPCOptions,
					TLSCACert:   ordererCfg.TLSCACert,
				}, true, false
			}
		}
	}

	//In case of URL search, return default orderer config where URL=SearchKey
	if searchByURL && strings.Contains(ordererSearchKey, ":") {
		return &fab.OrdererConfig{
			URL:         ordererSearchKey,
			GRPCOptions: c.defaultOrdererConfig.GRPCOptions,
			TLSCACert:   c.defaultOrdererConfig.TLSCACert,
		}, true, false
	}

	return nil, false, false
}

func (c *EndpointConfig) matchOrderer(ordererSearchKey string, matcher matcherEntry) (*fab.OrdererConfig, bool, bool) {

	if matcher.matchConfig.IgnoreEndpoint {
		logger.Debugf(" Ignoring orderer `%s` since entity matcher IgnoreEndpoint flag is on", ordererSearchKey)
		// IgnoreEndpoint must force ignoring this matching orderer (weather found or not) and must be explicitly
		// mentioned. The third argument is explicitly used for this, all other cases will return false.
		return nil, false, true
	}

	mappedHost := c.regexMatchAndReplace(matcher.regex, ordererSearchKey, matcher.matchConfig.MappedHost)

	//Get the ordererConfig from mapped host
	matchedOrderer := c.getMappedOrderer(mappedHost)
	if matchedOrderer == nil {
		logger.Debugf("Could not find mapped host [%s] for orderer [%s]", matcher.matchConfig.MappedHost, ordererSearchKey)
		return nil, false, false
	}

	//URLSubstitutionExp if found use from entity matcher otherwise use from mapped host
	if matcher.matchConfig.URLSubstitutionExp != "" {
		matchedOrderer.URL = c.regexMatchAndReplace(matcher.regex, ordererSearchKey, matcher.matchConfig.URLSubstitutionExp)
	}

	//SSLTargetOverrideURLSubstitutionExp if found use from entity matcher otherwise use from mapped host
	if matcher.matchConfig.SSLTargetOverrideURLSubstitutionExp != "" {
		matchedOrderer.GRPCOptions["ssl-target-name-override"] = c.regexMatchAndReplace(matcher.regex, ordererSearchKey, matcher.matchConfig.SSLTargetOverrideURLSubstitutionExp)
	}

	//if no URL to add from entity matcher or from mapped host or from default peer
	if matchedOrderer.URL == "" {
		matchedOrderer.URL = c.getDefaultMatchingURL(ordererSearchKey)
	}

	return matchedOrderer, true, false
}

func (c *EndpointConfig) getMappedOrderer(host string) *fab.OrdererConfig {
	//Get the peerConfig from mapped host
	ordererConfig, ok := c.networkConfig.Orderers[strings.ToLower(host)]
	if !ok {
		ordererConfig = c.defaultOrdererConfig
	}

	mappedConfig := fab.OrdererConfig{
		URL:         ordererConfig.URL,
		TLSCACert:   ordererConfig.TLSCACert,
		GRPCOptions: make(map[string]interface{}),
	}

	for key, val := range ordererConfig.GRPCOptions {
		mappedConfig.GRPCOptions[key] = val
	}

	return &mappedConfig
}

func (c *EndpointConfig) compileMatchers() error {

	entMatchers := entityMatchers{}

	err := c.backend.UnmarshalKey("entityMatchers", &entMatchers.matchers)
	logger.Debugf("Matchers are: %+v", entMatchers)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'entMatchers' config item")
	}

	//return no error if entityMatchers is not configured
	if len(entMatchers.matchers) == 0 {
		logger.Debug("Entity matchers are not configured")
		return nil
	}

	err = c.compileAllMatchers(&entMatchers)
	if err != nil {
		return err
	}

	c.entityMatchers = &entMatchers
	return nil
}

func (c *EndpointConfig) compileAllMatchers(matcherConfig *entityMatchers) error {

	var err error
	if len(matcherConfig.matchers["channel"]) > 0 {
		c.channelMatchers, err = c.groupAllMatchers(matcherConfig.matchers["channel"])
		if err != nil {
			return err
		}
	}

	if len(matcherConfig.matchers["orderer"]) > 0 {
		c.ordererMatchers, err = c.groupAllMatchers(matcherConfig.matchers["orderer"])
		if err != nil {
			return err
		}
	}

	if len(matcherConfig.matchers["peer"]) > 0 {
		c.peerMatchers, err = c.groupAllMatchers(matcherConfig.matchers["peer"])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *EndpointConfig) groupAllMatchers(matchers []MatchConfig) ([]matcherEntry, error) {
	matcherEntries := make([]matcherEntry, len(matchers))
	for i, v := range matchers {
		regex, err := regexp.Compile(v.Pattern)
		if err != nil {
			return nil, err
		}
		matcherEntries[i] = matcherEntry{regex: regex, matchConfig: v}
	}
	return matcherEntries, nil
}

func (c *EndpointConfig) verifyPeerConfig(p *fab.PeerConfig, peerName string, tlsEnabled bool) error {
	if p == nil || p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && p.TLSCACert == nil && !c.backend.GetBool("client.tlsCerts.systemCertPool") {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

func (c *EndpointConfig) loadTLSCerts() ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	errs := multi.Errors{}

	for _, peer := range c.networkPeers {
		if peer.TLSCACert != nil {
			certs = append(certs, peer.TLSCACert)
		}
	}
	for _, orderer := range c.ordererConfigs {
		if orderer.TLSCACert != nil {
			certs = append(certs, orderer.TLSCACert)
		}
	}
	return certs, errs.ToError()
}

//ResetNetworkConfig clears network config cache
func (c *EndpointConfig) ResetNetworkConfig() error {
	c.networkConfig = nil
	return c.loadEndpointConfiguration()
}

// PeerMSPID returns msp that peer belongs to
func (c *EndpointConfig) peerMSPID(name string) (string, bool) {
	var mspID string
	// Find organisation/msp that peer belongs to
	for _, org := range c.networkConfig.Organizations {
		for i := 0; i < len(org.Peers); i++ {
			if strings.EqualFold(org.Peers[i], name) {
				// peer belongs to this org add org msp
				mspID = org.MSPID
				break
			} else {
				peer, ok := c.findMatchingPeer(org.Peers[i])
				if ok && strings.EqualFold(peer, name) {
					mspID = org.MSPID
					break
				}
			}
		}
	}

	return mspID, mspID != ""
}

func (c *EndpointConfig) findMatchingPeer(peerName string) (string, bool) {

	//Return if no peerMatchers are configured
	if len(c.peerMatchers) == 0 {
		return "", false
	}

	//loop over peerentityMatchers to find the matching peer
	for _, matcher := range c.peerMatchers {
		if matcher.regex.MatchString(peerName) {
			return matcher.matchConfig.MappedHost, true
		}
	}

	return "", false
}

//regexMatchAndReplace if 'repl' has $ then perform regex.ReplaceAllString otherwise return 'repl'
func (c *EndpointConfig) regexMatchAndReplace(regex *regexp.Regexp, src, repl string) string {
	if strings.Contains(repl, "$") {
		return regex.ReplaceAllString(src, repl)
	}
	return repl
}

//peerChannelConfigHookFunc returns hook function for unmarshalling 'fab.PeerChannelConfig'
// Rule : default set to 'true' if not provided in config
func peerChannelConfigHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {

		//Run through each PeerChannelConfig, create empty config map if value is nil
		if t == reflect.TypeOf(map[string]PeerChannelConfig{}) {
			dataMap, ok := data.(map[string]interface{})
			if ok {
				for k, v := range dataMap {
					if v == nil {
						// Make an empty map. It will be filled in with defaults
						// in other hook below
						dataMap[k] = make(map[string]interface{})
					}
				}
				return dataMap, nil
			}
		}

		//If target is of type 'fab.PeerChannelConfig', fill in defaults if not already specified
		if t == reflect.TypeOf(PeerChannelConfig{}) {
			dataMap, ok := data.(map[string]interface{})
			if ok {
				setDefault(dataMap, "endorsingpeer", true)
				setDefault(dataMap, "chaincodequery", true)
				setDefault(dataMap, "ledgerquery", true)
				setDefault(dataMap, "eventsource", true)

				return dataMap, nil
			}
		}

		return data, nil
	}
}

//setDefault sets default value provided to map if given key not found
func setDefault(dataMap map[string]interface{}, key string, defaultVal interface{}) {
	_, ok := dataMap[key]
	if !ok {
		dataMap[key] = defaultVal
	}
}

//detectDeprecatedConfigOptions detects deprecated config options and prints warnings
// currently detects: if channels.orderers are defined
func detectDeprecatedNetworkConfig(endpointConfig *EndpointConfig) {

	if endpointConfig.networkConfig == nil {
		return
	}

	//detect if channels orderers are mentioned
	for _, v := range endpointConfig.networkConfig.Channels {
		if len(v.Orderers) > 0 {
			logger.Warn("Getting orderers from endpoint config channels.orderer is deprecated, use entity matchers to override orderer configuration")
			logger.Warn("visit https://github.com/hyperledger/fabric-sdk-go/blob/master/test/fixtures/config/overrides/local_entity_matchers.yaml for samples")
			break
		}
	}
}
