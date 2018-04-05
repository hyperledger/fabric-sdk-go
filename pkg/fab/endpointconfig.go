/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"crypto/tls"
	"crypto/x509"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/cryptoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/pkg/errors"

	"regexp"

	"sync"

	"io/ioutil"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	cs "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
)

var logger = logging.NewLogger("fabsdk/fab")

const (
	defaultEndorserConnectionTimeout      = time.Second * 10
	defaultPeerResponseTimeout            = time.Minute * 3
	defaultDiscoveryGreylistExpiryTimeout = time.Second * 10
	defaultEventHubConnectionTimeout      = time.Second * 15
	defaultEventRegTimeout                = time.Second * 15
	defaultOrdererConnectionTimeout       = time.Second * 15
	defaultOrdererResponseTimeout         = time.Second * 15
	defaultQueryTimeout                   = time.Minute * 3
	defaultExecuteTimeout                 = time.Minute * 3
	defaultResMgmtTimeout                 = time.Minute * 3
	defaultConnIdleInterval               = time.Second * 30
	defaultEventServiceIdleInterval       = time.Minute * 2
	defaultChannelConfigRefreshInterval   = time.Minute * 30
	defaultChannelMemshpRefreshInterval   = time.Second * 30

	defaultCacheSweepInterval = time.Second * 15
)

//ConfigFromBackend returns endpoint config implementation for given backend
func ConfigFromBackend(coreBackend core.ConfigBackend) (fab.EndpointConfig, error) {

	config := &EndpointConfig{backend: lookup.New(coreBackend)}

	if err := config.cacheNetworkConfiguration(); err != nil {
		return nil, errors.WithMessage(err, "network configuration load failed")
	}

	//Compile the entityMatchers
	config.peerMatchers = make(map[int]*regexp.Regexp)
	config.ordererMatchers = make(map[int]*regexp.Regexp)
	config.caMatchers = make(map[int]*regexp.Regexp)

	matchError := config.compileMatchers()
	if matchError != nil {
		return nil, matchError
	}

	return config, nil
}

// EndpointConfig represents the endpoint configuration for the client
type EndpointConfig struct {
	backend             *lookup.ConfigLookup
	tlsCerts            []*x509.Certificate
	networkConfig       *fab.NetworkConfig
	networkConfigCached bool
	peerMatchers        map[int]*regexp.Regexp
	ordererMatchers     map[int]*regexp.Regexp
	caMatchers          map[int]*regexp.Regexp
	certPoolLock        sync.Mutex
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

	for _, orderer := range config.Orderers {

		if orderer.TLSCACerts.Path != "" {
			orderer.TLSCACerts.Path = pathvar.Subst(orderer.TLSCACerts.Path)
		} else if len(orderer.TLSCACerts.Pem) == 0 && c.backend.GetBool("client.tlsCerts.systemCertPool") == false {
			errors.Errorf("Orderer has no certs configured. Make sure TLSCACerts.Pem or TLSCACerts.Path is set for %s", orderer.URL)
		}

		orderers = append(orderers, orderer)
	}

	return orderers, nil
}

// OrdererConfig returns the requested orderer
func (c *EndpointConfig) OrdererConfig(name string) (*fab.OrdererConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	orderer, ok := config.Orderers[strings.ToLower(name)]
	if !ok {
		logger.Debugf("Could not find Orderer for [%s], trying with Entity Matchers", name)
		matchingOrdererConfig, matchErr := c.tryMatchingOrdererConfig(strings.ToLower(name))
		if matchErr != nil {
			return nil, errors.WithMessage(matchErr, "unable to find Orderer Config")
		}
		logger.Debugf("Found matching Orderer Config for [%s]", name)
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
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	peersConfig := config.Organizations[strings.ToLower(org)].Peers
	peers := []fab.PeerConfig{}

	for _, peerName := range peersConfig {
		p := config.Peers[strings.ToLower(peerName)]
		if err = c.verifyPeerConfig(p, peerName, endpoint.IsTLSEnabled(p.URL)); err != nil {
			logger.Debugf("Could not verify Peer for [%s], trying with Entity Matchers", peerName)
			matchingPeerConfig, matchErr := c.tryMatchingPeerConfig(peerName)
			if matchErr != nil {
				return nil, errors.WithMessage(err, "unable to find Peer Config")
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
func (c *EndpointConfig) PeerConfig(org string, name string) (*fab.PeerConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	peersConfig := config.Organizations[strings.ToLower(org)].Peers
	peerInOrg := false
	for _, p := range peersConfig {
		if p == name {
			peerInOrg = true
		}
	}
	if !peerInOrg {
		return nil, errors.Errorf("peer %s is not part of organization %s", name, org)
	}

	peerConfig, ok := config.Peers[strings.ToLower(name)]
	if !ok {
		logger.Debugf("Could not find Peer for [%s], trying with Entity Matchers", name)
		matchingPeerConfig, matchErr := c.tryMatchingPeerConfig(strings.ToLower(name))
		if matchErr != nil {
			return nil, errors.WithMessage(matchErr, "unable to find peer config")
		}
		logger.Debugf("Found MatchingPeerConfig for [%s]", name)
		peerConfig = *matchingPeerConfig
	}

	if peerConfig.TLSCACerts.Path != "" {
		peerConfig.TLSCACerts.Path = pathvar.Subst(peerConfig.TLSCACerts.Path)
	}
	return &peerConfig, nil
}

// PeerConfigByURL retrieves PeerConfig by URL
func (c *EndpointConfig) PeerConfigByURL(url string) (*fab.PeerConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	var matchPeerConfig *fab.PeerConfig
	staticPeers := config.Peers
	for _, staticPeerConfig := range staticPeers {
		if strings.EqualFold(staticPeerConfig.URL, url) {
			matchPeerConfig = &staticPeerConfig
			break
		}
	}

	if matchPeerConfig == nil {
		// try to match from entity matchers
		logger.Debugf("Could not find Peer for url [%s], trying with Entity Matchers", url)
		matchPeerConfig, err = c.tryMatchingPeerConfig(url)
		if err != nil {
			return nil, errors.WithMessage(err, "No Peer found with the url from config")
		}
		logger.Debugf("Found MatchingPeerConfig for url [%s]", url)
	}

	if matchPeerConfig != nil && matchPeerConfig.TLSCACerts.Path != "" {
		matchPeerConfig.TLSCACerts.Path = pathvar.Subst(matchPeerConfig.TLSCACerts.Path)
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

// NetworkPeers returns the network peers configuration
func (c *EndpointConfig) NetworkPeers() ([]fab.NetworkPeer, error) {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	netPeers := []fab.NetworkPeer{}

	for name, p := range netConfig.Peers {

		if err = c.verifyPeerConfig(p, name, endpoint.IsTLSEnabled(p.URL)); err != nil {
			return nil, err
		}

		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = pathvar.Subst(p.TLSCACerts.Path)
		}

		mspID, err := c.PeerMSPID(name)
		if err != nil {
			return nil, errors.Errorf("failed to retrieve msp id for peer %s", name)
		}

		netPeer := fab.NetworkPeer{PeerConfig: p, MSPID: mspID}
		netPeers = append(netPeers, netPeer)
	}

	return netPeers, nil
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
		return nil, nil
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
		return peers, nil
	}

	for peerName, chPeerConfig := range chConfig.Peers {

		// Get generic peer configuration
		p, ok := netConfig.Peers[strings.ToLower(peerName)]
		if !ok {
			logger.Debugf("Could not find Peer for [%s], trying with Entity Matchers", peerName)
			matchingPeerConfig, matchErr := c.tryMatchingPeerConfig(strings.ToLower(peerName))
			if matchErr != nil {
				return nil, errors.Errorf("peer config not found for %s", peerName)
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

		// Assemble channel peer key
		chPeerKey := "channels." + name + ".peers." + peerName

		// Default value for endorsing peer key is true
		endorsingPeerKey := strings.ToLower(chPeerKey + ".endorsingPeer")
		_, ok = c.backend.Lookup(endorsingPeerKey)
		if !ok {
			chPeerConfig.EndorsingPeer = true
		}

		// Default value for chaincode query key is true
		ccQueryKey := strings.ToLower(chPeerKey + ".chaincodeQuery")
		_, ok = c.backend.Lookup(ccQueryKey)
		if !ok {
			chPeerConfig.ChaincodeQuery = true
		}

		// Default value for ledger query key is true
		ledgerQueryKey := strings.ToLower(chPeerKey + ".ledgerQuery")
		_, ok = c.backend.Lookup(ledgerQueryKey)
		if !ok {
			chPeerConfig.LedgerQuery = true
		}

		// Default value for event source key is true
		eventSourceKey := strings.ToLower(chPeerKey + ".eventSource")
		_, ok = c.backend.Lookup(eventSourceKey)
		if !ok {
			chPeerConfig.EventSource = true
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
// is provided, the certficate is added to the pool
func (c *EndpointConfig) TLSCACertPool(certs ...*x509.Certificate) (*x509.CertPool, error) {

	c.certPoolLock.Lock()
	defer c.certPoolLock.Unlock()

	//add cert if it is not nil and doesn't exists already
	for _, newCert := range certs {
		if newCert != nil && !c.containsCert(newCert) {
			c.tlsCerts = append(c.tlsCerts, newCert)
		}
	}

	//get new cert pool
	tlsCertPool, err := c.getCertPool()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create cert pool")
	}

	//add all tls ca certs to cert pool
	for _, cert := range c.tlsCerts {
		tlsCertPool.AddCert(cert)
	}

	return tlsCertPool, nil
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
	var cb, kb []byte
	cb, err = clientConfig.TLSCerts.Client.Cert.Bytes()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load tls client cert")
	}

	if len(cb) == 0 {
		// if no cert found in the config, return empty cert chain
		return []tls.Certificate{clientCerts}, nil
	}

	// Load private key from cert using default crypto suite
	cs := cs.GetDefault()
	pk, err := cryptoutil.GetPrivateKeyFromCert(cb, cs)

	// If CryptoSuite fails to load private key from cert then load private key from config
	if err != nil || pk == nil {
		logger.Debugf("Reading pk from config, unable to retrieve from cert: %s", err)
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

	// private key was retrieved from cert
	clientCerts, err = cryptoutil.X509KeyPair(cb, pk, cs)
	if err != nil {
		return nil, err
	}

	return []tls.Certificate{clientCerts}, nil
}

// CryptoConfigPath ...
func (c *EndpointConfig) CryptoConfigPath() string {
	return pathvar.Subst(c.backend.GetString("client.cryptoconfig.path"))
}

func (c *EndpointConfig) getTimeout(tType fab.TimeoutType) time.Duration {
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

	err = c.backend.UnmarshalKey("channels", &networkConfig.Channels)
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

func (c *EndpointConfig) tryMatchingPeerConfig(peerName string) (*fab.PeerConfig, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	//Return if no peerMatchers are configured
	if len(c.peerMatchers) == 0 {
		return nil, errors.New("no Peer entityMatchers are found")
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
			//Get the peerConfig from mapped host
			peerConfig, ok := networkConfig.Peers[strings.ToLower(peerMatchConfig.MappedHost)]
			if !ok {
				return nil, errors.New("failed to load config from matched Peer")
			}

			// Make a copy of GRPC options (as it is manipulated below)
			peerConfig.GRPCOptions = copyPropertiesMap(peerConfig.GRPCOptions)

			_, isPortPresentInPeerName := c.getPortIfPresent(peerName)
			//if substitution url is empty, use the same network peer url
			if peerMatchConfig.URLSubstitutionExp == "" {
				port, isPortPresent := c.getPortIfPresent(peerConfig.URL)
				peerConfig.URL = peerName
				//append port of matched config
				if isPortPresent && !isPortPresentInPeerName {
					peerConfig.URL += ":" + strconv.Itoa(port)
				}
			} else {
				//else, replace url with urlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(peerMatchConfig.URLSubstitutionExp, "$") < 0 {
					peerConfig.URL = peerMatchConfig.URLSubstitutionExp
				} else {
					//if the urlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with substituionexp pattern
					peerConfig.URL = v.ReplaceAllString(peerName, peerMatchConfig.URLSubstitutionExp)
				}

			}

			//if eventSubstitution url is empty, use the same network peer url
			if peerMatchConfig.EventURLSubstitutionExp == "" {
				port, isPortPresent := c.getPortIfPresent(peerConfig.EventURL)
				peerConfig.EventURL = peerName
				//append port of matched config
				if isPortPresent && !isPortPresentInPeerName {
					peerConfig.EventURL += ":" + strconv.Itoa(port)
				}
			} else {
				//else, replace url with eventUrlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(peerMatchConfig.EventURLSubstitutionExp, "$") < 0 {
					peerConfig.EventURL = peerMatchConfig.EventURLSubstitutionExp
				} else {
					//if the eventUrlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with eventsubstituionexp pattern
					peerConfig.EventURL = v.ReplaceAllString(peerName, peerMatchConfig.EventURLSubstitutionExp)
				}

			}

			//if sslTargetOverrideUrlSubstitutionExp is empty, use the same network peer host
			if peerMatchConfig.SSLTargetOverrideURLSubstitutionExp == "" {
				if strings.Index(peerName, ":") < 0 {
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
				if strings.Index(peerMatchConfig.SSLTargetOverrideURLSubstitutionExp, "$") < 0 {
					peerConfig.GRPCOptions["ssl-target-name-override"] = peerMatchConfig.SSLTargetOverrideURLSubstitutionExp
				} else {
					//if the sslTargetOverrideUrlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with eventsubstituionexp pattern
					peerConfig.GRPCOptions["ssl-target-name-override"] = v.ReplaceAllString(peerName, peerMatchConfig.SSLTargetOverrideURLSubstitutionExp)
				}

			}
			return &peerConfig, nil
		}
	}

	return nil, errors.WithStack(status.New(status.ClientStatus, status.NoMatchingPeerEntity.ToInt32(), "no matching peer config found", nil))
}

func (c *EndpointConfig) tryMatchingOrdererConfig(ordererName string) (*fab.OrdererConfig, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	//Return if no ordererMatchers are configured
	if len(c.ordererMatchers) == 0 {
		return nil, errors.New("no Orderer entityMatchers are found")
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
			// get the matching matchConfig from the index number
			ordererMatchConfig := networkConfig.EntityMatchers["orderer"][k]
			//Get the ordererConfig from mapped host
			ordererConfig, ok := networkConfig.Orderers[strings.ToLower(ordererMatchConfig.MappedHost)]
			if !ok {
				return nil, errors.New("failed to load config from matched Orderer")
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
				if strings.Index(ordererMatchConfig.URLSubstitutionExp, "$") < 0 {
					ordererConfig.URL = ordererMatchConfig.URLSubstitutionExp
				} else {
					//if the urlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with substituionexp pattern
					ordererConfig.URL = v.ReplaceAllString(ordererName, ordererMatchConfig.URLSubstitutionExp)
				}
			}

			//if sslTargetOverrideUrlSubstitutionExp is empty, use the same network peer host
			if ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp == "" {
				if strings.Index(ordererName, ":") < 0 {
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
				if strings.Index(ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp, "$") < 0 {
					ordererConfig.GRPCOptions["ssl-target-name-override"] = ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp
				} else {
					//if the sslTargetOverrideUrlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with eventsubstituionexp pattern
					ordererConfig.GRPCOptions["ssl-target-name-override"] = v.ReplaceAllString(ordererName, ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp)
				}

			}
			return &ordererConfig, nil
		}
	}

	return nil, errors.WithStack(status.New(status.ClientStatus, status.NoMatchingOrdererEntity.ToInt32(), "no matching orderer config found", nil))
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

// PeerConfig Retrieves a specific peer by name
func (c *EndpointConfig) peerConfig(name string) (*fab.PeerConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	peerConfig, ok := config.Peers[strings.ToLower(name)]
	if !ok {
		logger.Debugf("Could not find PeerConfig for [%s], trying with Entity Matchers", name)
		matchingPeerConfig, matchErr := c.tryMatchingPeerConfig(strings.ToLower(name))
		if matchErr != nil {
			return nil, errors.WithMessage(matchErr, "unable to find peer config")
		}
		logger.Debugf("Found MatchingPeerConfig for [%s]", name)
		peerConfig = *matchingPeerConfig
	}

	if peerConfig.TLSCACerts.Path != "" {
		peerConfig.TLSCACerts.Path = pathvar.Subst(peerConfig.TLSCACerts.Path)
	}
	return &peerConfig, nil
}

func (c *EndpointConfig) verifyPeerConfig(p fab.PeerConfig, peerName string, tlsEnabled bool) error {
	if p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && len(p.TLSCACerts.Pem) == 0 && p.TLSCACerts.Path == "" && c.backend.GetBool("client.tlsCerts.systemCertPool") == false {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

func (c *EndpointConfig) containsCert(newCert *x509.Certificate) bool {
	//TODO may need to maintain separate map of {cert.RawSubject, cert} to improve performance on search
	for _, cert := range c.tlsCerts {
		if cert.Equal(newCert) {
			return true
		}
	}
	return false
}

func (c *EndpointConfig) getCertPool() (*x509.CertPool, error) {
	tlsCertPool := x509.NewCertPool()
	if c.backend.GetBool("client.tlsCerts.systemCertPool") == true {
		var err error
		if tlsCertPool, err = x509.SystemCertPool(); err != nil {
			return nil, err
		}
		logger.Debugf("Loaded system cert pool of size: %d", len(tlsCertPool.Subjects()))
	}
	return tlsCertPool, nil
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
