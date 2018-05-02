/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	commtls "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm/tls"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/cryptoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

const (
	defaultEndorserConnectionTimeout      = time.Second * 10
	defaultPeerResponseTimeout            = time.Minute * 3
	defaultDiscoveryGreylistExpiryTimeout = time.Second * 10
	defaultEventHubConnectionTimeout      = time.Second * 15
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
	defaultDiscoveryRefreshInterval       = time.Second * 10

	defaultCacheSweepInterval = time.Second * 15
)

//ConfigFromBackend returns endpoint config implementation for given backend
func ConfigFromBackend(coreBackend ...core.ConfigBackend) (fab.EndpointConfig, error) {

	config := &EndpointConfig{
		backend:         lookup.New(coreBackend...),
		peerMatchers:    make(map[int]*regexp.Regexp),
		ordererMatchers: make(map[int]*regexp.Regexp),
		caMatchers:      make(map[int]*regexp.Regexp),
		channelMatchers: make(map[int]*regexp.Regexp),
	}

	if err := config.cacheNetworkConfiguration(); err != nil {
		return nil, errors.WithMessage(err, "network configuration load failed")
	}

	//Compile the entityMatchers
	matchError := config.compileMatchers()
	if matchError != nil {
		return nil, matchError
	}

	config.tlsCertPool = commtls.NewCertPool(config.backend.GetBool("client.tlsCerts.systemCertPool"))

	// preemptively add all TLS certs to cert pool as adding them at request time
	// is expensive
	certs, err := config.loadTLSCerts()
	if err != nil {
		logger.Infof("could not cache TLS certs", err.Error())
	}
	if _, err := config.TLSCACertPool(certs...); err != nil {
		return nil, errors.WithMessage(err, "cert pool load failed")
	}

	return config, nil
}

// EndpointConfig represents the endpoint configuration for the client
type EndpointConfig struct {
	backend             *lookup.ConfigLookup
	networkConfig       *fab.NetworkConfig
	tlsCertPool         commtls.CertPool
	networkConfigCached bool
	peerMatchers        map[int]*regexp.Regexp
	ordererMatchers     map[int]*regexp.Regexp
	caMatchers          map[int]*regexp.Regexp
	channelMatchers     map[int]*regexp.Regexp
}

// Timeout reads timeouts for the given timeout type, if type is not found in the config
// then default is set as per the const value above for the corresponding type
func (c *EndpointConfig) Timeout(tType fab.TimeoutType) time.Duration {
	return c.getTimeout(tType)
}

// MSPID returns the MSP ID for the requested organization
func (c *EndpointConfig) MSPID(org string) (string, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}
	// viper lowercases all key maps, org is lower case
	mspID := config.Organizations[strings.ToLower(org)].MSPID
	if mspID == "" {
		return "", errors.Errorf("MSP ID is empty for org: %s", org)
	}

	return mspID, nil
}

// PeerMSPID returns msp that peer belongs to
func (c *EndpointConfig) PeerMSPID(name string) (string, error) {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}

	var mspID string

	// Find organisation/msp that peer belongs to
	for _, org := range netConfig.Organizations {
		for i := 0; i < len(org.Peers); i++ {
			if strings.EqualFold(org.Peers[i], name) {
				// peer belongs to this org add org msp
				mspID = org.MSPID
				break
			} else {
				peer, err := c.findMatchingPeer(org.Peers[i])
				if err == nil && strings.EqualFold(peer, name) {
					mspID = org.MSPID
					break
				}
			}
		}
	}

	return mspID, nil

}

// OrderersConfig returns a list of defined orderers
func (c *EndpointConfig) OrderersConfig() ([]fab.OrdererConfig, error) {
	orderers := []fab.OrdererConfig{}
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	for name, orderer := range config.Orderers {

		matchedOrderer := c.tryMatchingOrdererConfig(config, name)
		if matchedOrderer != nil {
			//if found in entity matcher then use the matched one
			orderer = *matchedOrderer
		}

		if orderer.TLSCACerts.Path != "" {
			orderer.TLSCACerts.Path = pathvar.Subst(orderer.TLSCACerts.Path)
		} else if len(orderer.TLSCACerts.Pem) == 0 && !c.backend.GetBool("client.tlsCerts.systemCertPool") {
			return nil, errors.Errorf("Orderer has no certs configured. Make sure TLSCACerts.Pem or TLSCACerts.Path is set for %s", orderer.URL)
		}

		orderers = append(orderers, orderer)
	}

	return orderers, nil
}

// OrdererConfig returns the requested orderer
func (c *EndpointConfig) OrdererConfig(nameOrURL string) (*fab.OrdererConfig, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	orderer, ok := networkConfig.Orderers[strings.ToLower(nameOrURL)]

	if !ok {
		ordererCfgs, err := c.OrderersConfig()
		if err != nil {
			return nil, err
		}
		for _, ordererCfg := range ordererCfgs {
			if strings.EqualFold(ordererCfg.URL, nameOrURL) {
				orderer = ordererCfg
				ok = true
				break
			}
		}
	}

	if !ok {
		logger.Debugf("Could not find Orderer for [%s], trying with Entity Matchers", nameOrURL)
		matchingOrdererConfig := c.tryMatchingOrdererConfig(networkConfig, strings.ToLower(nameOrURL))
		if matchingOrdererConfig == nil {
			return nil, errors.WithStack(status.New(status.ClientStatus, status.NoMatchingOrdererEntity.ToInt32(), "no matching orderer config found", nil))
		}
		logger.Debugf("Found matching Orderer Config for [%s]", nameOrURL)
		orderer = *matchingOrdererConfig
	}

	if orderer.TLSCACerts.Path != "" {
		orderer.TLSCACerts.Path = pathvar.Subst(orderer.TLSCACerts.Path)
	}

	return &orderer, nil
}

// PeersConfig Retrieves the fabric peers for the specified org from the
// config file provided
func (c *EndpointConfig) PeersConfig(org string) ([]fab.PeerConfig, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	peersConfig := networkConfig.Organizations[strings.ToLower(org)].Peers
	peers := []fab.PeerConfig{}

	for _, peerName := range peersConfig {
		p := networkConfig.Peers[strings.ToLower(peerName)]
		if err = c.verifyPeerConfig(p, peerName, endpoint.IsTLSEnabled(p.URL)); err != nil {
			logger.Debugf("Could not verify Peer for [%s], trying with Entity Matchers", peerName)
			matchingPeerConfig := c.tryMatchingPeerConfig(networkConfig, peerName)
			if matchingPeerConfig == nil {
				continue
			}
			logger.Debugf("Found a matchingPeerConfig for [%s]", peerName)
			p = *matchingPeerConfig
		}
		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = pathvar.Subst(p.TLSCACerts.Path)
		}

		peers = append(peers, p)
	}
	return peers, nil
}

// PeerConfig Retrieves a specific peer from the configuration by org and name
func (c *EndpointConfig) PeerConfig(nameOrURL string) (*fab.PeerConfig, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	//lookup by name in config
	peerConfig, ok := networkConfig.Peers[strings.ToLower(nameOrURL)]

	var matchPeerConfig *fab.PeerConfig
	if ok {
		matchPeerConfig = &peerConfig
	} else {
		for _, staticPeerConfig := range networkConfig.Peers {
			if strings.EqualFold(staticPeerConfig.URL, nameOrURL) {
				matchPeerConfig = c.tryMatchingPeerConfig(networkConfig, nameOrURL)
				if matchPeerConfig == nil {
					matchPeerConfig = &staticPeerConfig
				}
				break
			}
		}
	}

	//Not found through config lookup by name or URL, try matcher now
	if matchPeerConfig == nil {
		logger.Debugf("Could not find Peer for name/url [%s], trying with Entity Matchers", nameOrURL)
		//try to match nameOrURL with peer entity matchers
		matchPeerConfig = c.tryMatchingPeerConfig(networkConfig, nameOrURL)
	}

	if matchPeerConfig == nil {
		return nil, errors.WithStack(status.New(status.ClientStatus, status.NoMatchingPeerEntity.ToInt32(), "no matching peer config found", nil))
	}

	logger.Debugf("Found MatchingPeerConfig for name/url [%s]", nameOrURL)

	if matchPeerConfig.TLSCACerts.Path != "" {
		matchPeerConfig.TLSCACerts.Path = pathvar.Subst(peerConfig.TLSCACerts.Path)
	}

	return matchPeerConfig, nil
}

// NetworkConfig returns the network configuration defined in the config file
func (c *EndpointConfig) NetworkConfig() (*fab.NetworkConfig, error) {
	if c.networkConfigCached {
		return c.networkConfig, nil
	}

	if err := c.cacheNetworkConfiguration(); err != nil {
		return nil, errors.WithMessage(err, "network configuration load failed")
	}
	return c.networkConfig, nil
}

// NetworkPeers returns the network peers configuration, all the peers from all the orgs in config.
func (c *EndpointConfig) NetworkPeers() ([]fab.NetworkPeer, error) {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	var netPeers []fab.NetworkPeer
	for org, orgConfig := range netConfig.Organizations {
		orgPeers, err := c.PeersConfig(org)
		if err != nil {
			return nil, err
		}

		for _, orgPeer := range orgPeers {
			netPeers = append(netPeers, fab.NetworkPeer{PeerConfig: orgPeer, MSPID: orgConfig.MSPID})
		}
	}

	return netPeers, nil
}

// MappedChannelName will return channelName if it is an original channel name in the config
// if it is not, then it will try to find a channelMatcher and return its MappedName.
// If more than one matcher is found, then the first matcher in the list will be used.
// TODO expose this function if it's needed elsewhere in the sdk
func (c *EndpointConfig) mappedChannelName(channelName string) (string, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}
	// if channelName is the original key found in the Channels map config, then return it as is
	_, ok := networkConfig.Channels[strings.ToLower(channelName)]
	if ok {
		return channelName, nil
	}

	// if !ok, then find a channelMatcher for channelName

	//Return if no channelMatchers are configured
	if len(c.channelMatchers) == 0 {
		return "", errors.New("no Channel entityMatchers found")
	}

	//sort the keys
	var keys []int
	for k := range c.channelMatchers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//loop over channelMatchers to find the matching channel name
	for _, k := range keys {
		v := c.channelMatchers[k]
		if v.MatchString(channelName) {
			// get the matching matchConfig from the index number
			channelMatchConfig := networkConfig.EntityMatchers["channel"][k]
			return channelMatchConfig.MappedName, nil
		}
	}

	// not matchers found, return an error
	return "", errors.WithStack(status.New(status.ClientStatus, status.NoMatchingChannelEntity.ToInt32(), "no matching channel config found", nil))
}

// ChannelConfig returns the channel configuration
func (c *EndpointConfig) ChannelConfig(name string) (*fab.ChannelNetworkConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	// viper lowercases all key maps
	ch, ok := config.Channels[strings.ToLower(name)]
	if !ok {
		matchingChannel, _, matchErr := c.tryMatchingChannelConfig(name)
		if matchErr != nil {
			return nil, errors.WithMessage(matchErr, "channel config not found")
		}
		return matchingChannel, nil
	}

	return &ch, nil
}

// ChannelPeers returns the channel peers configuration
func (c *EndpointConfig) ChannelPeers(name string) ([]fab.ChannelPeer, error) {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	peers := []fab.ChannelPeer{}

	// viper lowercases all key maps
	chConfig, ok := netConfig.Channels[strings.ToLower(name)]
	if !ok {
		matchingChannel, _, matchErr := c.tryMatchingChannelConfig(name)
		if matchErr != nil {
			return peers, nil
		}

		// reset 'name' with the mappedChannel as it's referenced further below
		chConfig = *matchingChannel
	}

	for peerName, chPeerConfig := range chConfig.Peers {

		// Get generic peer configuration
		p, ok := netConfig.Peers[strings.ToLower(peerName)]
		if !ok {
			logger.Debugf("Could not find Peer for [%s], trying with Entity Matchers", peerName)
			matchingPeerConfig := c.tryMatchingPeerConfig(netConfig, strings.ToLower(peerName))
			if matchingPeerConfig == nil {
				continue
			}
			logger.Debugf("Found matchingPeerConfig for [%s]", peerName)
			p = *matchingPeerConfig
		}

		if err = c.verifyPeerConfig(p, peerName, endpoint.IsTLSEnabled(p.URL)); err != nil {
			return nil, err
		}

		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = pathvar.Subst(p.TLSCACerts.Path)
		}

		mspID, err := c.PeerMSPID(peerName)
		if err != nil {
			return nil, errors.Errorf("failed to retrieve msp id for peer %s", peerName)
		}

		networkPeer := fab.NetworkPeer{PeerConfig: p, MSPID: mspID}

		peer := fab.ChannelPeer{PeerChannelConfig: chPeerConfig, NetworkPeer: networkPeer}

		peers = append(peers, peer)
	}

	return peers, nil

}

// ChannelOrderers returns a list of channel orderers
func (c *EndpointConfig) ChannelOrderers(name string) ([]fab.OrdererConfig, error) {
	orderers := []fab.OrdererConfig{}
	channel, err := c.ChannelConfig(name)
	if err != nil || channel == nil {
		return nil, errors.Errorf("Unable to retrieve channel config: %s", err)
	}

	for _, chOrderer := range channel.Orderers {
		orderer, err := c.OrdererConfig(chOrderer)
		if err != nil || orderer == nil {
			return nil, errors.Errorf("unable to retrieve orderer config: %s", err)
		}

		orderers = append(orderers, *orderer)
	}

	return orderers, nil
}

// TLSCACertPool returns the configured cert pool. If a certConfig
// is provided, the certificate is added to the pool
func (c *EndpointConfig) TLSCACertPool(certs ...*x509.Certificate) (*x509.CertPool, error) {
	return c.tlsCertPool.Get(certs...)
}

// EventServiceType returns the type of event service client to use
func (c *EndpointConfig) EventServiceType() fab.EventServiceType {
	etype := c.backend.GetString("client.eventService.type")
	switch etype {
	case "eventhub":
		return fab.EventHubEventServiceType
	default:
		return fab.DeliverEventServiceType
	}
}

// TLSClientCerts loads the client's certs for mutual TLS
// It checks the config for embedded pem files before looking for cert files
func (c *EndpointConfig) TLSClientCerts() ([]tls.Certificate, error) {
	clientConfig, err := c.client()
	if err != nil {
		return nil, err
	}
	var clientCerts tls.Certificate
	var cb []byte
	cb, err = clientConfig.TLSCerts.Client.Cert.Bytes()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load tls client cert")
	}

	if len(cb) == 0 {
		// if no cert found in the config, return empty cert chain
		return []tls.Certificate{clientCerts}, nil
	}

	// Load private key from cert using default crypto suite
	cs := cryptosuite.GetDefault()
	pk, err := cryptoutil.GetPrivateKeyFromCert(cb, cs)

	// If CryptoSuite fails to load private key from cert then load private key from config
	if err != nil || pk == nil {
		logger.Debugf("Reading pk from config, unable to retrieve from cert: %s", err)
		return c.loadPrivateKeyFromConfig(clientConfig, clientCerts, cb)
	}

	// private key was retrieved from cert
	clientCerts, err = cryptoutil.X509KeyPair(cb, pk, cs)
	if err != nil {
		return nil, err
	}

	return []tls.Certificate{clientCerts}, nil
}

func (c *EndpointConfig) loadPrivateKeyFromConfig(clientConfig *msp.ClientConfig, clientCerts tls.Certificate, cb []byte) ([]tls.Certificate, error) {
	var kb []byte
	var err error
	if clientConfig.TLSCerts.Client.Key.Pem != "" {
		kb = []byte(clientConfig.TLSCerts.Client.Key.Pem)
	} else if clientConfig.TLSCerts.Client.Key.Path != "" {
		kb, err = loadByteKeyOrCertFromFile(clientConfig, true)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to load key from file path '%s'", clientConfig.TLSCerts.Client.Key.Path)
		}
	}

	// load the key/cert pair from []byte
	clientCerts, err = tls.X509KeyPair(cb, kb)
	if err != nil {
		return nil, errors.Errorf("Error loading cert/key pair as TLS client credentials: %v", err)
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
	case fab.EndorserConnection:
		timeout = c.backend.GetDuration("client.peer.timeout.connection")
		if timeout == 0 {
			timeout = defaultEndorserConnectionTimeout
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
	case fab.EventHubConnection:
		timeout = c.backend.GetDuration("client.eventService.timeout.connection")
		if timeout == 0 {
			timeout = defaultEventHubConnectionTimeout
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

	case fab.CacheSweepInterval: // EXPERIMENTAL - do we need this to be configurable?
		timeout = c.backend.GetDuration("client.cache.interval.sweep")
		if timeout == 0 {
			timeout = defaultCacheSweepInterval
		}
	}

	return timeout
}

func (c *EndpointConfig) cacheNetworkConfiguration() error {
	networkConfig := fab.NetworkConfig{}
	networkConfig.Name = c.backend.GetString("name")
	networkConfig.Description = c.backend.GetString("description")
	networkConfig.Version = c.backend.GetString("version")

	err := c.backend.UnmarshalKey("client", &networkConfig.Client)
	logger.Debugf("Client is: %+v", networkConfig.Client)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'client' config item to networkConfig.Client type")
	}

	err = c.backend.UnmarshalKey("channels", &networkConfig.Channels, lookup.WithUnmarshalHookFunction(peerChannelConfigHookFunc()))
	logger.Debugf("channels are: %+v", networkConfig.Channels)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'channels' config item to networkConfig.Channels type")
	}

	err = c.backend.UnmarshalKey("organizations", &networkConfig.Organizations)
	logger.Debugf("organizations are: %+v", networkConfig.Organizations)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'organizations' config item to networkConfig.Organizations type")
	}

	err = c.backend.UnmarshalKey("orderers", &networkConfig.Orderers)
	logger.Debugf("orderers are: %+v", networkConfig.Orderers)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'orderers' config item to networkConfig.Orderers type")
	}

	err = c.backend.UnmarshalKey("peers", &networkConfig.Peers)
	logger.Debugf("peers are: %+v", networkConfig.Peers)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'peers' config item to networkConfig.Peers type")
	}

	err = c.backend.UnmarshalKey("certificateAuthorities", &networkConfig.CertificateAuthorities)
	logger.Debugf("certificateAuthorities are: %+v", networkConfig.CertificateAuthorities)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'certificateAuthorities' config item to networkConfig.CertificateAuthorities type")
	}

	err = c.backend.UnmarshalKey("entityMatchers", &networkConfig.EntityMatchers)
	logger.Debugf("Matchers are: %+v", networkConfig.EntityMatchers)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'entityMatchers' config item to networkConfig.EntityMatchers type")
	}

	c.networkConfig = &networkConfig
	c.networkConfigCached = true
	return nil
}

func (c *EndpointConfig) getPortIfPresent(url string) (int, bool) {
	s := strings.Split(url, ":")
	if len(s) > 1 {
		if port, err := strconv.Atoi(s[len(s)-1]); err == nil {
			return port, true
		}
	}
	return 0, false
}

func (c *EndpointConfig) tryMatchingPeerConfig(networkConfig *fab.NetworkConfig, peerName string) *fab.PeerConfig {

	//Return if no peerMatchers are configured
	if len(c.peerMatchers) == 0 {
		return nil
	}

	//sort the keys
	var keys []int
	for k := range c.peerMatchers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//loop over peerentityMatchers to find the matching peer
	for _, k := range keys {
		v := c.peerMatchers[k]
		if v.MatchString(peerName) {
			return c.matchPeer(networkConfig, peerName, k, v)
		}
	}

	return nil
}

func (c *EndpointConfig) matchPeer(networkConfig *fab.NetworkConfig, peerName string, k int, v *regexp.Regexp) *fab.PeerConfig {
	// get the matching matchConfig from the index number
	peerMatchConfig := networkConfig.EntityMatchers["peer"][k]
	//Get the peerConfig from mapped host
	peerConfig, ok := networkConfig.Peers[strings.ToLower(peerMatchConfig.MappedHost)]
	if !ok {
		return nil
	}

	// Make a copy of GRPC options (as it is manipulated below)
	peerConfig.GRPCOptions = copyPropertiesMap(peerConfig.GRPCOptions)

	_, isPortPresentInPeerName := c.getPortIfPresent(peerName)
	//if substitution url is empty, use the same network peer url
	if peerMatchConfig.URLSubstitutionExp == "" {
		peerConfig.URL = getPeerConfigURL(c, peerName, peerConfig.URL, isPortPresentInPeerName)
	} else {
		//else, replace url with urlSubstitutionExp if it doesnt have any variable declarations like $
		if !strings.Contains(peerMatchConfig.URLSubstitutionExp, "$") {
			peerConfig.URL = peerMatchConfig.URLSubstitutionExp
		} else {
			//if the urlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with substituionexp pattern
			peerConfig.URL = v.ReplaceAllString(peerName, peerMatchConfig.URLSubstitutionExp)
		}

	}

	//if eventSubstitution url is empty, use the same network peer url
	if peerMatchConfig.EventURLSubstitutionExp == "" {
		peerConfig.EventURL = getPeerConfigURL(c, peerName, peerConfig.EventURL, isPortPresentInPeerName)
	} else {
		//else, replace url with eventUrlSubstitutionExp if it doesnt have any variable declarations like $
		if !strings.Contains(peerMatchConfig.EventURLSubstitutionExp, "$") {
			peerConfig.EventURL = peerMatchConfig.EventURLSubstitutionExp
		} else {
			//if the eventUrlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with eventsubstituionexp pattern
			peerConfig.EventURL = v.ReplaceAllString(peerName, peerMatchConfig.EventURLSubstitutionExp)
		}

	}

	//if sslTargetOverrideUrlSubstitutionExp is empty, use the same network peer host
	if peerMatchConfig.SSLTargetOverrideURLSubstitutionExp == "" {
		if !strings.Contains(peerName, ":") {
			peerConfig.GRPCOptions["ssl-target-name-override"] = peerName
		} else {
			//Remove port and protocol of the peerName
			s := strings.Split(peerName, ":")
			if isPortPresentInPeerName {
				peerConfig.GRPCOptions["ssl-target-name-override"] = s[len(s)-2]
			} else {
				peerConfig.GRPCOptions["ssl-target-name-override"] = s[len(s)-1]
			}
		}

	} else {
		//else, replace url with sslTargetOverrideUrlSubstitutionExp if it doesnt have any variable declarations like $
		if !strings.Contains(peerMatchConfig.SSLTargetOverrideURLSubstitutionExp, "$") {
			peerConfig.GRPCOptions["ssl-target-name-override"] = peerMatchConfig.SSLTargetOverrideURLSubstitutionExp
		} else {
			//if the sslTargetOverrideUrlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with eventsubstituionexp pattern
			peerConfig.GRPCOptions["ssl-target-name-override"] = v.ReplaceAllString(peerName, peerMatchConfig.SSLTargetOverrideURLSubstitutionExp)
		}

	}
	return &peerConfig
}

func getPeerConfigURL(c *EndpointConfig, peerName, peerConfigURL string, isPortPresentInPeerName bool) string {
	port, isPortPresent := c.getPortIfPresent(peerConfigURL)
	url := peerName
	//append port of matched config
	if isPortPresent && !isPortPresentInPeerName {
		url += ":" + strconv.Itoa(port)
	}
	return url
}

func (c *EndpointConfig) tryMatchingOrdererConfig(networkConfig *fab.NetworkConfig, ordererName string) *fab.OrdererConfig {

	//Return if no ordererMatchers are configured
	if len(c.ordererMatchers) == 0 {
		return nil
	}

	//sort the keys
	var keys []int
	for k := range c.ordererMatchers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//loop over ordererentityMatchers to find the matching orderer
	for _, k := range keys {
		v := c.ordererMatchers[k]
		if v.MatchString(ordererName) {
			return c.matchOrderer(networkConfig, ordererName, k, v)
		}
	}

	return nil
}

func (c *EndpointConfig) matchOrderer(networkConfig *fab.NetworkConfig, ordererName string, k int, v *regexp.Regexp) *fab.OrdererConfig {
	// get the matching matchConfig from the index number
	ordererMatchConfig := networkConfig.EntityMatchers["orderer"][k]
	//Get the ordererConfig from mapped host
	ordererConfig, ok := networkConfig.Orderers[strings.ToLower(ordererMatchConfig.MappedHost)]
	if !ok {
		return nil
	}

	// Make a copy of GRPC options (as it is manipulated below)
	ordererConfig.GRPCOptions = copyPropertiesMap(ordererConfig.GRPCOptions)

	_, isPortPresentInOrdererName := c.getPortIfPresent(ordererName)
	//if substitution url is empty, use the same network orderer url
	if ordererMatchConfig.URLSubstitutionExp == "" {
		port, isPortPresent := c.getPortIfPresent(ordererConfig.URL)
		ordererConfig.URL = ordererName

		//append port of matched config
		if isPortPresent && !isPortPresentInOrdererName {
			ordererConfig.URL += ":" + strconv.Itoa(port)
		}
	} else {
		//else, replace url with urlSubstitutionExp if it doesnt have any variable declarations like $
		if !strings.Contains(ordererMatchConfig.URLSubstitutionExp, "$") {
			ordererConfig.URL = ordererMatchConfig.URLSubstitutionExp
		} else {
			//if the urlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with substituionexp pattern
			ordererConfig.URL = v.ReplaceAllString(ordererName, ordererMatchConfig.URLSubstitutionExp)
		}
	}

	//if sslTargetOverrideUrlSubstitutionExp is empty, use the same network peer host
	if ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp == "" {
		if !strings.Contains(ordererName, ":") {
			ordererConfig.GRPCOptions["ssl-target-name-override"] = ordererName
		} else {
			//Remove port and protocol of the ordererName
			s := strings.Split(ordererName, ":")
			if isPortPresentInOrdererName {
				ordererConfig.GRPCOptions["ssl-target-name-override"] = s[len(s)-2]
			} else {
				ordererConfig.GRPCOptions["ssl-target-name-override"] = s[len(s)-1]
			}
		}

	} else {
		//else, replace url with sslTargetOverrideUrlSubstitutionExp if it doesnt have any variable declarations like $
		if !strings.Contains(ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp, "$") {
			ordererConfig.GRPCOptions["ssl-target-name-override"] = ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp
		} else {
			//if the sslTargetOverrideUrlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with eventsubstituionexp pattern
			ordererConfig.GRPCOptions["ssl-target-name-override"] = v.ReplaceAllString(ordererName, ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp)
		}

	}
	return &ordererConfig
}

func (c *EndpointConfig) tryMatchingChannelConfig(channelName string) (*fab.ChannelNetworkConfig, string, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, "", err
	}

	// get the mapped channel Name
	mappedChannelName, err := c.mappedChannelName(channelName)
	if err != nil {
		return nil, "", err
	}

	//Get the channelConfig from mappedChannelName
	channelConfig, ok := networkConfig.Channels[strings.ToLower(mappedChannelName)]
	if !ok {
		return nil, "", errors.New("failed to load config from matched Channel")
	}

	return &channelConfig, mappedChannelName, nil
}

func copyPropertiesMap(origMap map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{}, len(origMap))
	for k, v := range origMap {
		newMap[k] = v
	}
	return newMap
}

func (c *EndpointConfig) findMatchingPeer(peerName string) (string, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}
	//Return if no peerMatchers are configured
	if len(c.peerMatchers) == 0 {
		return "", errors.New("no Peer entityMatchers are found")
	}

	//sort the keys
	var keys []int
	for k := range c.peerMatchers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//loop over peerentityMatchers to find the matching peer
	for _, k := range keys {
		v := c.peerMatchers[k]
		if v.MatchString(peerName) {
			// get the matching matchConfig from the index number
			peerMatchConfig := networkConfig.EntityMatchers["peer"][k]
			return peerMatchConfig.MappedHost, nil
		}
	}

	return "", errors.WithStack(status.New(status.ClientStatus, status.NoMatchingPeerEntity.ToInt32(), "no matching peer config found", nil))
}

func (c *EndpointConfig) compileMatchers() error {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return err
	}
	//return no error if entityMatchers is not configured
	if networkConfig.EntityMatchers == nil {
		return nil
	}

	err = c.compilePeerMatcher(networkConfig)
	if err != nil {
		return err
	}
	err = c.compileOrdererMatcher(networkConfig)
	if err != nil {
		return err
	}

	err = c.compileCertificateAuthorityMatcher(networkConfig)
	if err != nil {
		return err
	}

	err = c.compileChannelMatcher(networkConfig)
	return err
}

func (c *EndpointConfig) compileChannelMatcher(networkConfig *fab.NetworkConfig) error {
	var err error
	if networkConfig.EntityMatchers["channel"] != nil {
		channelMatchers := networkConfig.EntityMatchers["channel"]
		for i, matcher := range channelMatchers {
			if matcher.Pattern != "" {
				c.channelMatchers[i], err = regexp.Compile(matcher.Pattern)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *EndpointConfig) compileCertificateAuthorityMatcher(networkConfig *fab.NetworkConfig) error {
	var err error
	if networkConfig.EntityMatchers["certificateauthority"] != nil {
		certMatchersConfig := networkConfig.EntityMatchers["certificateauthority"]
		for i := 0; i < len(certMatchersConfig); i++ {
			if certMatchersConfig[i].Pattern != "" {
				c.caMatchers[i], err = regexp.Compile(certMatchersConfig[i].Pattern)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *EndpointConfig) compileOrdererMatcher(networkConfig *fab.NetworkConfig) error {
	var err error
	if networkConfig.EntityMatchers["orderer"] != nil {
		ordererMatchersConfig := networkConfig.EntityMatchers["orderer"]
		for i := 0; i < len(ordererMatchersConfig); i++ {
			if ordererMatchersConfig[i].Pattern != "" {
				c.ordererMatchers[i], err = regexp.Compile(ordererMatchersConfig[i].Pattern)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *EndpointConfig) compilePeerMatcher(networkConfig *fab.NetworkConfig) error {
	var err error
	if networkConfig.EntityMatchers["peer"] != nil {
		peerMatchersConfig := networkConfig.EntityMatchers["peer"]
		for i := 0; i < len(peerMatchersConfig); i++ {
			if peerMatchersConfig[i].Pattern != "" {
				c.peerMatchers[i], err = regexp.Compile(peerMatchersConfig[i].Pattern)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *EndpointConfig) verifyPeerConfig(p fab.PeerConfig, peerName string, tlsEnabled bool) error {
	if p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && len(p.TLSCACerts.Pem) == 0 && p.TLSCACerts.Path == "" && !c.backend.GetBool("client.tlsCerts.systemCertPool") {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

func (c *EndpointConfig) loadTLSCerts() ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	errs := multi.Errors{}

	orderers, err := c.OrderersConfig()
	if err != nil {
		errs = append(errs, err)
	}
	peers, err := c.NetworkPeers()
	if err != nil {
		errs = append(errs, err)
	}
	for _, peer := range peers {
		cert, err := peer.TLSCACerts.TLSCert()
		if err != nil {
			errs = append(errs, errors.WithMessage(err, "for peer: "+peer.URL))
			continue
		}
		certs = append(certs, cert)
	}
	for _, orderer := range orderers {
		cert, err := orderer.TLSCACerts.TLSCert()
		if err != nil {
			errs = append(errs, errors.WithMessage(err, "for orderer: "+orderer.URL))
			continue
		}
		certs = append(certs, cert)
	}
	return certs, errs.ToError()
}

// Client returns the Client config
func (c *EndpointConfig) client() (*msp.ClientConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	client := config.Client

	client.Organization = strings.ToLower(client.Organization)
	client.TLSCerts.Path = pathvar.Subst(client.TLSCerts.Path)
	client.TLSCerts.Client.Key.Path = pathvar.Subst(client.TLSCerts.Client.Key.Path)
	client.TLSCerts.Client.Cert.Path = pathvar.Subst(client.TLSCerts.Client.Cert.Path)

	return &client, nil
}

//Backend returns config lookup of endpoint config
func (c *EndpointConfig) Backend() *lookup.ConfigLookup {
	return c.backend
}

//CAMatchers returns CA matchers of endpoint config
func (c *EndpointConfig) CAMatchers() map[int]*regexp.Regexp {
	return c.caMatchers
}

//ResetNetworkConfig clears network config cache
func (c *EndpointConfig) ResetNetworkConfig() {
	c.networkConfig = nil
	c.networkConfigCached = false
}

func loadByteKeyOrCertFromFile(c *msp.ClientConfig, isKey bool) ([]byte, error) {
	var path string
	a := "key"
	if isKey {
		path = pathvar.Subst(c.TLSCerts.Client.Key.Path)
		c.TLSCerts.Client.Key.Path = path
	} else {
		a = "cert"
		path = pathvar.Subst(c.TLSCerts.Client.Cert.Path)
		c.TLSCerts.Client.Cert.Path = path
	}
	bts, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("Error loading %s file from '%s' err: %v", a, path, err)
	}
	return bts, nil
}

//peerChannelConfigHookFunc returns hook function for unmarshalling 'fab.PeerChannelConfig'
// Rule : default set to 'true' if not provided in config
func peerChannelConfigHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {

		//If target is of type 'fab.PeerChannelConfig', then only hook should work
		if t == reflect.TypeOf(fab.PeerChannelConfig{}) {
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
func setDefault(dataMap map[string]interface{}, key string, defaultVal bool) {
	_, ok := dataMap[key]
	if !ok {
		dataMap[key] = true
	}
}
