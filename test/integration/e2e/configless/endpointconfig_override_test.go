/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package configless

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	commtls "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm/tls"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/cryptoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	logApi "github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/pkg/errors"
)

// endpointconfig_override_test.go is an example of programmatically configuring the sdk by injecting instances that implement EndpointConfig's functions (representing the sdk's configs)
// for the sake of overriding EndpointConfig integration tests, the structure variables below are similar to what is found in /test/fixtures/config/config_e2e.yaml
// application developers can fully override these functions to load configs in any way that suit their application need

// NOTE: 1. to support test local (flag: TEST_LOCAL=true to use localhost:* URLs for peers, orderers, CAs everywhere), new...() constructor functions where created to test if this flag
//       is enabled using verifyIsLocal...() function calls. These calls will basically switch config URLs (peers, orderers or CA configs) into "localhost:..."
//       Make sure your local /etc/hosts file does not have any ip-dns mapping entries for peers/orderers/CAs
//
//       2. the test assumes the use of the default channel block used in the remaining regular integration tests (for example look at Orderer.Addresses value in
//       test/fixtures/fabric/<specific target fabric release>/config/configtx.yaml to see the URL value assigned to the orderer for a specific channel).
//       So Even if the below interfaces will override orderers to localhost for TEST_LOCAL=true, the SDK will still try
//       to create an orderer with the URL found in the channel block mentioned above. You can either create another channel block for your channels,
//       or if you want to use an existing channel block but still want to change the orderer URL, then you can implement EntityMatchers logic for your orderers
//       which is commented out in the code below for reference. Using EntityMatchers will allow the configs to be able to find mapped Orderers/Peers/CA URLs.

// ClientConfig provides the definition of the client configuration
type clientConfig struct {
	Organization    string
	Logging         logApi.LoggingType
	CryptoConfig    msp.CCType
	TLSCerts        endpoint.MutualTLSConfig
	TLSKey          []byte
	TLSCert         []byte
	CredentialStore msp.CredentialStoreType
}

// caConfig defines a CA configuration in identity config
type caConfig struct {
	ID         string
	URL        string
	TLSCACerts endpoint.MutualTLSConfig
	Registrar  msp.EnrollCredentials
	CAName     string
}

var (
	localhostRep = "localhost:"
	dnsMatchRegX = ".*:"
	client       = clientConfig{
		Organization:    "org1",
		Logging:         api.LoggingType{Level: "info"},
		CryptoConfig:    msp.CCType{Path: pathvar.Subst("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}")},
		CredentialStore: msp.CredentialStoreType{Path: "/tmp/msp"},
		TLSCerts: endpoint.MutualTLSConfig{Client: endpoint.TLSKeyPair{
			Key:  newTLSConfig("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/tls.example.com/users/User1@tls.example.com/tls/client.key"),
			Cert: newTLSConfig("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/tls.example.com/users/User1@tls.example.com/tls/client.crt")}},
	}

	channelsConfig = map[string]fab.ChannelEndpointConfig{
		"mychannel": {
			Orderers: []string{"orderer.example.com"},
			Peers: map[string]fab.PeerChannelConfig{
				"peer0.org1.example.com": {
					EndorsingPeer:  true,
					ChaincodeQuery: true,
					LedgerQuery:    true,
					EventSource:    true,
				},
			},
			Policies: fab.ChannelPolicies{
				QueryChannelConfig: fab.QueryChannelConfigPolicy{
					MinResponses: 1,
					MaxTargets:   1,
					RetryOpts: retry.Opts{
						Attempts:       5,
						InitialBackoff: 500 * time.Millisecond,
						MaxBackoff:     5 * time.Second,
						BackoffFactor:  2.0,
					},
				},
				EventService: fab.EventServicePolicy{
					ResolverStrategy:                 fab.MinBlockHeightStrategy,
					MinBlockHeightResolverMode:       fab.ResolveByThreshold,
					BlockHeightLagThreshold:          5,
					ReconnectBlockHeightLagThreshold: 10,
					PeerMonitorPeriod:                5 * time.Second,
				},
			},
		},
		"orgchannel": {
			Orderers: []string{"orderer.example.com"},
			Peers: map[string]fab.PeerChannelConfig{
				"peer0.org1.example.com": {
					EndorsingPeer:  true,
					ChaincodeQuery: true,
					LedgerQuery:    true,
					EventSource:    true,
				},
				"peer0.org2.example.com": {
					EndorsingPeer:  true,
					ChaincodeQuery: true,
					LedgerQuery:    true,
					EventSource:    true,
				},
			},
			Policies: fab.ChannelPolicies{
				QueryChannelConfig: fab.QueryChannelConfigPolicy{
					MinResponses: 1,
					MaxTargets:   1,
					RetryOpts: retry.Opts{
						Attempts:       5,
						InitialBackoff: 500 * time.Millisecond,
						MaxBackoff:     5 * time.Second,
						BackoffFactor:  2.0,
					},
				},
			},
		},
	}
	orgsConfig = map[string]fab.OrganizationConfig{
		"org1": {
			MSPID:                  "Org1MSP",
			CryptoPath:             "peerOrganizations/org1.example.com/users/{username}@org1.example.com/msp",
			Peers:                  []string{"peer0.org1.example.com"},
			CertificateAuthorities: []string{"ca.org1.example.com"},
		},
		"org2": {
			MSPID:                  "Org2MSP",
			CryptoPath:             "peerOrganizations/org1.example.com/users/{username}@org2.example.com/msp",
			Peers:                  []string{"peer0.org2.example.com"},
			CertificateAuthorities: []string{"ca.org2.example.com"},
		},
		"ordererorg": {
			MSPID:      "OrdererMSP",
			CryptoPath: "ordererOrganizations/example.com/users/{username}@example.com/msp",
		},
	}

	orderersConfig = map[string]fab.OrdererConfig{
		"orderer.example.com": {
			URL: "orderer.example.com:7050",
			GRPCOptions: map[string]interface{}{
				"ssl-target-name-override": "orderer.example.com",
				"keep-alive-time":          0 * time.Second,
				"keep-alive-timeout":       20 * time.Second,
				"keep-alive-permit":        false,
				"fail-fast":                false,
				"allow-insecure":           false,
			},
			TLSCACert: tlsCertByBytes("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem"),
		},
	}

	peersConfig = map[string]fab.PeerConfig{
		"peer0.org1.example.com": {
			URL: "peer0.org1.example.com:7051",
			GRPCOptions: map[string]interface{}{
				"ssl-target-name-override": "peer0.org1.example.com",
				"keep-alive-time":          0 * time.Second,
				"keep-alive-timeout":       20 * time.Second,
				"keep-alive-permit":        false,
				"fail-fast":                false,
				"allow-insecure":           false,
			},
			TLSCACert: tlsCertByBytes("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem"),
		},
		"peer0.org2.example.com": {
			URL: "peer0.org2.example.com:8051",
			GRPCOptions: map[string]interface{}{
				"ssl-target-name-override": "peer0.org2.example.com",
				"keep-alive-time":          0 * time.Second,
				"keep-alive-timeout":       20 * time.Second,
				"keep-alive-permit":        false,
				"fail-fast":                false,
				"allow-insecure":           false,
			},
			TLSCACert: tlsCertByBytes("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem"),
		},
	}

	peersByLocalURL = map[string]fab.PeerConfig{
		"localhost:7051": {
			URL: "localhost:7051",
			GRPCOptions: map[string]interface{}{
				"ssl-target-name-override": "peer0.org1.example.com",
				"keep-alive-time":          0 * time.Second,
				"keep-alive-timeout":       20 * time.Second,
				"keep-alive-permit":        false,
				"fail-fast":                false,
				"allow-insecure":           false,
			},
			TLSCACert: tlsCertByBytes("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem"),
		},
		"localhost:8051": {
			URL: "localhost:8051",
			GRPCOptions: map[string]interface{}{
				"ssl-target-name-override": "peer0.org2.example.com",
				"keep-alive-time":          0 * time.Second,
				"keep-alive-timeout":       20 * time.Second,
				"keep-alive-permit":        false,
				"fail-fast":                false,
				"allow-insecure":           false,
			},
			TLSCACert: tlsCertByBytes("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem"),
		},
	}

	caConfigObj = map[string]caConfig{
		"ca.org1.example.com": {
			ID:  "ca.org1.example.com",
			URL: "https://ca.org1.example.com:7054",
			TLSCACerts: endpoint.MutualTLSConfig{
				Path: pathvar.Subst("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem"),
				Client: endpoint.TLSKeyPair{
					Key:  newTLSConfig("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/tls.example.com/users/User1@tls.example.com/tls/client.key"),
					Cert: newTLSConfig("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/tls.example.com/users/User1@tls.example.com/tls/client.crt"),
				},
			},
			Registrar: msp.EnrollCredentials{
				EnrollID:     "admin",
				EnrollSecret: "adminpw",
			},
			CAName: "ca.org1.example.com",
		},
		"ca.org2.example.com": {
			ID:  "ca.org2.example.com",
			URL: "https://ca.org2.example.com:8054",
			TLSCACerts: endpoint.MutualTLSConfig{
				Path: pathvar.Subst("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem"),
				Client: endpoint.TLSKeyPair{
					Key:  newTLSConfig("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/tls.example.com/users/User1@tls.example.com/tls/client.key"),
					Cert: newTLSConfig("${FABRIC_SDK_GO_PROJECT_PATH}/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/tls.example.com/users/User1@tls.example.com/tls/client.crt"),
				},
			},
			Registrar: msp.EnrollCredentials{
				EnrollID:     "admin",
				EnrollSecret: "adminpw",
			},
			CAName: "ca.org2.example.com",
		},
	}

	networkConfig = fab.NetworkConfig{
		Channels:      channelsConfig,
		Organizations: orgsConfig,
		Orderers:      newOrderersConfig(),
		Peers:         newPeersConfig(),
		// EntityMatchers are not used in this implementation
		//EntityMatchers: entityMatchers,
	}

	// creating instances of each interface to be referenced in the integration tests:
	timeoutImpl          = &exampleTimeout{}
	orderersConfigImpl   = newOrderersConfigImpl()
	ordererConfigImpl    = &exampleOrdererConfig{}
	peersConfigImpl      = newPeersConfigImpl()
	peerConfigImpl       = &examplePeerConfig{}
	networkConfigImpl    = &exampleNetworkConfig{}
	networkPeersImpl     = &exampleNetworkPeers{}
	channelConfigImpl    = &exampleChannelConfig{}
	channelPeersImpl     = &exampleChannelPeers{}
	channelOrderersImpl  = &exampleChannelOrderers{}
	tlsCACertPoolImpl    = newTLSCACertPool(false)
	tlsClientCertsImpl   = &exampleTLSClientCerts{}
	cryptoConfigPathImpl = &exampleCryptoConfigPath{}
	endpointConfigImpls  = []interface{}{
		timeoutImpl,
		orderersConfigImpl,
		ordererConfigImpl,
		peersConfigImpl,
		peerConfigImpl,
		networkConfigImpl,
		networkPeersImpl,
		channelConfigImpl,
		channelPeersImpl,
		channelOrderersImpl,
		tlsCACertPoolImpl,
		tlsClientCertsImpl,
		cryptoConfigPathImpl,
	}
)

type exampleTimeout struct{}

var defaultTypes = map[fab.TimeoutType]time.Duration{
	fab.PeerConnection:           time.Second * 10,
	fab.PeerResponse:             time.Minute * 3,
	fab.DiscoveryGreylistExpiry:  time.Second * 10,
	fab.EventReg:                 time.Second * 15,
	fab.OrdererConnection:        time.Second * 15,
	fab.OrdererResponse:          time.Minute * 2,
	fab.DiscoveryConnection:      time.Second * 15,
	fab.DiscoveryResponse:        time.Second * 15,
	fab.Query:                    time.Minute * 3,
	fab.Execute:                  time.Minute * 3,
	fab.ResMgmt:                  time.Minute * 3,
	fab.ConnectionIdle:           time.Second * 30,
	fab.EventServiceIdle:         time.Minute * 2,
	fab.ChannelConfigRefresh:     time.Minute * 90,
	fab.ChannelMembershipRefresh: time.Second * 60,
	fab.DiscoveryServiceRefresh:  time.Second * 10,
	fab.SelectionServiceRefresh:  time.Minute * 15,
	// EXPERIMENTAL - do we need this to be configurable?
	fab.CacheSweepInterval: time.Second * 15,
}

//Timeout overrides EndpointConfig's Timeout function which returns the timeout for the given timeoutType in the arg
func (m *exampleTimeout) Timeout(tType fab.TimeoutType) time.Duration {
	t, ok := defaultTypes[tType]
	if !ok {
		return time.Second * 30 // general default if type is not found
	}
	return t
}

//PeerMSPID  returns the mspID for the given org name in the arg
func PeerMSPID(name string) (string, bool) {
	// Find organisation/msp that peer belongs to
	for _, org := range orgsConfig {
		for i := 0; i < len(org.Peers); i++ {
			if strings.EqualFold(org.Peers[i], name) {
				// peer belongs to this org add org msp
				return org.MSPID, true
				// EntityMatchers are not used in this implementation, below is an example of how to use them if needed
				//} else {
				//
				//	peer, err := m.findMatchingPeer(org.Peers[i])
				//	if err == nil && strings.EqualFold(peer, name) {
				//		mspID = org.MSPID
				//		break
				//	}
			}
		}
	}

	return "", false
}

func verifyIsLocalCAsURLs(caConfigs map[string]caConfig) map[string]caConfig {
	re := regexp.MustCompile(dnsMatchRegX)
	var newCfg = make(map[string]caConfig)
	// for local integration tests, replace all urls DNS to localhost:
	if integration.IsLocal() {
		for k, caCfg := range caConfigs {
			caCfg.URL = re.ReplaceAllString(caCfg.URL, localhostRep)
			newCfg[k] = caCfg
		}
	}
	return newCfg
}

func newCAsConfig() map[string]caConfig {
	c := verifyIsLocalCAsURLs(caConfigObj)
	caConfigObj = c
	return c
}

func newPeersConfig() map[string]fab.PeerConfig {
	p := verifyIsLocalPeersURLs(peersConfig)
	peersConfig = p
	return p
}

func newOrderersConfig() map[string]fab.OrdererConfig {
	o := verifyIsLocalOrderersURLs(orderersConfig)
	orderersConfig = o
	return o
}

func verifyIsLocalOrderersURLs(oConfig map[string]fab.OrdererConfig) map[string]fab.OrdererConfig {
	re := regexp.MustCompile(dnsMatchRegX)
	var newConfig = make(map[string]fab.OrdererConfig)
	// for local integration tests, replace all urls DNS to localhost:
	if integration.IsLocal() {
		for k, orderer := range oConfig {
			orderer.URL = re.ReplaceAllString(orderer.URL, localhostRep)
			newConfig[k] = orderer
		}
	}

	if len(newConfig) == 0 {
		return oConfig
	}
	return newConfig
}

//newOrderersConfigImpl will create a new exampleOrderersConfig instance with proper ordrerer URLs (local vs normal) tests
// local tests use localhost urls, while the remaining tests use default values as set in orderersConfig var
func newOrderersConfigImpl() *exampleOrderersConfig {
	oConfig := verifyIsLocalOrderersURLs(orderersConfig)
	orderersConfig = oConfig
	o := &exampleOrderersConfig{}
	return o
}

type exampleOrderersConfig struct {
	isSystemCertPool bool
}

//OrderersConfig overrides EndpointConfig's OrderersConfig function which returns the ordererConfigs list
func (m *exampleOrderersConfig) OrderersConfig() []fab.OrdererConfig {
	orderers := []fab.OrdererConfig{}

	for _, orderer := range orderersConfig {

		if orderer.TLSCACert == nil && !m.isSystemCertPool {
			return nil
		}
		orderers = append(orderers, orderer)
	}

	return orderers
}

type exampleOrdererConfig struct{}

//OrdererConfig overrides EndpointConfig's OrdererConfig function which returns the ordererConfig instance for the name/URL arg
func (m *exampleOrdererConfig) OrdererConfig(ordererNameOrURL string) (*fab.OrdererConfig, bool, bool) {
	orderer, ok := networkConfig.Orderers[strings.ToLower(ordererNameOrURL)]
	if !ok {
		// EntityMatchers are not used in this implementation, below is an example of how to use them if needed, see default implementation for live example
		//matchingOrdererConfig := m.tryMatchingOrdererConfig(networkConfig, strings.ToLower(ordererNameOrURL))
		//if matchingOrdererConfig == nil {
		//	return nil, errors.WithStack(status.New(status.ClientStatus, status.NoMatchingOrdererEntity.ToInt32(), "no matching orderer config found", nil))
		//}
		//orderer = *matchingOrdererConfig
		return nil, false, false
	}

	return &orderer, true, false
}

type examplePeersConfig struct {
	isSystemCertPool bool
}

func verifyIsLocalPeersURLs(pConfig map[string]fab.PeerConfig) map[string]fab.PeerConfig {
	re := regexp.MustCompile(dnsMatchRegX)
	var newConfigs = make(map[string]fab.PeerConfig)
	// for local integration tests, replace all urls DNS to localhost:
	if integration.IsLocal() {
		for k, peer := range pConfig {
			peer.URL = re.ReplaceAllString(peer.URL, localhostRep)
			newConfigs[k] = peer
		}
	}

	if len(newConfigs) == 0 {
		return pConfig
	}
	return newConfigs
}

//newPeersConfigImpl will create a new examplePeersConfig instance with proper peers URLs (local vs normal) tests
// local tests use localhost urls, while the remaining tests use default values as set in peersConfig var
func newPeersConfigImpl() *examplePeersConfig {
	pConfig := verifyIsLocalPeersURLs(peersConfig)
	peersConfig = pConfig
	p := &examplePeersConfig{}
	return p
}

//PeersConfig overrides EndpointConfig's PeersConfig function which returns the peersConfig list
func (m *examplePeersConfig) PeersConfig(org string) ([]fab.PeerConfig, bool) {
	orgPeers := orgsConfig[strings.ToLower(org)].Peers
	peers := []fab.PeerConfig{}

	for _, peerName := range orgPeers {
		p := networkConfig.Peers[strings.ToLower(peerName)]
		if err := m.verifyPeerConfig(p, peerName, endpoint.IsTLSEnabled(p.URL)); err != nil {
			// EntityMatchers are not used in this implementation, below is an example of how to use them if needed
			//matchingPeerConfig := m.tryMatchingPeerConfig(networkConfig, peerName)
			//if matchingPeerConfig == nil {
			//	continue
			//}
			//
			//p = *matchingPeerConfig
			return nil, false
		}
		peers = append(peers, p)
	}
	return peers, true
}

func (m *examplePeersConfig) verifyPeerConfig(p fab.PeerConfig, peerName string, tlsEnabled bool) error {
	if p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && p.TLSCACert == nil && !m.isSystemCertPool {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

type examplePeerConfig struct{}

// PeerConfig overrides EndpointConfig's PeerConfig function which returns the peerConfig instance for the name/URL arg
func (m *examplePeerConfig) PeerConfig(nameOrURL string) (*fab.PeerConfig, bool) {
	pcfg, ok := peersConfig[nameOrURL]
	if ok {
		return &pcfg, true
	}

	if integration.IsLocal() {
		pcfg, ok := peersByLocalURL[nameOrURL]
		if ok {
			return &pcfg, true
		}
	}

	i := strings.Index(nameOrURL, ":")
	if i > 0 {
		return m.PeerConfig(nameOrURL[0:i])
	}

	return nil, false
}

type exampleNetworkConfig struct{}

// NetworkConfig overrides EndpointConfig's NetworkConfig function which returns the full network Config instance
func (m *exampleNetworkConfig) NetworkConfig() *fab.NetworkConfig {
	return &networkConfig
}

type exampleNetworkPeers struct {
	isSystemCertPool bool
}

//NetworkPeers overrides EndpointConfig's NetworkPeers function which returns the networkPeers list
func (m *exampleNetworkPeers) NetworkPeers() []fab.NetworkPeer {
	netPeers := []fab.NetworkPeer{}
	// referencing another interface to call PeerMSPID to match config yaml content

	for name, p := range networkConfig.Peers {

		if err := m.verifyPeerConfig(p, name, endpoint.IsTLSEnabled(p.URL)); err != nil {
			return nil
		}

		mspID, ok := PeerMSPID(name)
		if !ok {
			return nil
		}

		netPeer := fab.NetworkPeer{PeerConfig: p, MSPID: mspID}
		netPeers = append(netPeers, netPeer)
	}

	return netPeers
}

func (m *exampleNetworkPeers) verifyPeerConfig(p fab.PeerConfig, peerName string, tlsEnabled bool) error {
	if p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && p.TLSCACert == nil && !m.isSystemCertPool {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

type exampleChannelConfig struct{}

// ChannelConfig overrides EndpointConfig's ChannelConfig function which returns the channelConfig instance for the channel name arg
func (m *exampleChannelConfig) ChannelConfig(channelName string) *fab.ChannelEndpointConfig {
	ch, ok := channelsConfig[strings.ToLower(channelName)]
	if !ok {
		// EntityMatchers are not used in this implementation, below is an example of how to use them if needed
		//matchingChannel, _, matchErr := m.tryMatchingChannelConfig(channelName)
		//if matchErr != nil {
		//	return nil, errors.WithMessage(matchErr, "channel config not found")
		//}
		//return matchingChannel, nil
		return &fab.ChannelEndpointConfig{}
	}

	return &ch
}

type exampleChannelPeers struct {
	isSystemCertPool bool
}

// ChannelPeers overrides EndpointConfig's ChannelPeers function which returns the list of peers for the channel name arg
func (m *exampleChannelPeers) ChannelPeers(channelName string) []fab.ChannelPeer {
	peers := []fab.ChannelPeer{}

	chConfig, ok := channelsConfig[strings.ToLower(channelName)]
	if !ok {
		// EntityMatchers are not used in this implementation, below is an example of how to use them if needed
		//matchingChannel, _, matchErr := m.tryMatchingChannelConfig(channelName)
		//if matchErr != nil {
		//	return peers, nil
		//}
		//
		//// reset 'name' with the mappedChannel as it's referenced further below
		//chConfig = *matchingChannel
		return nil
	}

	for peerName, chPeerConfig := range chConfig.Peers {

		// Get generic peer configuration
		p, ok := peersConfig[strings.ToLower(peerName)]
		if !ok {
			// EntityMatchers are not used in this implementation, below is an example of how to use them if needed
			//matchingPeerConfig := m.tryMatchingPeerConfig(networkConfig, strings.ToLower(peerName))
			//if matchingPeerConfig == nil {
			//	continue
			//}
			//p = *matchingPeerConfig
			return nil
		}

		if err := m.verifyPeerConfig(p, peerName, endpoint.IsTLSEnabled(p.URL)); err != nil {
			return nil
		}

		mspID, ok := PeerMSPID(peerName)
		if !ok {
			return nil
		}

		networkPeer := fab.NetworkPeer{PeerConfig: p, MSPID: mspID}

		peer := fab.ChannelPeer{PeerChannelConfig: chPeerConfig, NetworkPeer: networkPeer}

		peers = append(peers, peer)
	}

	return peers

}

func (m *exampleChannelPeers) verifyPeerConfig(p fab.PeerConfig, peerName string, tlsEnabled bool) error {
	if p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && p.TLSCACert == nil && !m.isSystemCertPool {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

type exampleChannelOrderers struct{}

// ChannelOrderers overrides EndpointConfig's ChannelOrderers function which returns the list of orderers for the channel name arg
func (m *exampleChannelOrderers) ChannelOrderers(channelName string) []fab.OrdererConfig {
	// referencing other interfaces to call ChannelConfig and OrdererConfig to match config yaml content
	chCfg := &exampleChannelConfig{}
	oCfg := &exampleOrdererConfig{}

	orderers := []fab.OrdererConfig{}
	channel := chCfg.ChannelConfig(channelName)

	for _, chOrderer := range channel.Orderers {
		orderer, ok, _ := oCfg.OrdererConfig(chOrderer)
		if !ok || orderer == nil {
			return nil
		}
		orderers = append(orderers, *orderer)
	}

	return orderers
}

type exampleTLSCACertPool struct {
	tlsCertPool commtls.CertPool
}

//newTLSCACertPool will create a new exampleTLSCACertPool instance with useSystemCertPool bool flag
func newTLSCACertPool(useSystemCertPool bool) *exampleTLSCACertPool {
	m := &exampleTLSCACertPool{}
	var err error
	m.tlsCertPool, err = commtls.NewCertPool(useSystemCertPool)
	if err != nil {
		panic(err)
	}
	return m
}

// TLSCACertPool overrides EndpointConfig's TLSCACertPool function which will add the list of cert args to the cert pool and return it
func (m *exampleTLSCACertPool) TLSCACertPool() commtls.CertPool {
	return m.tlsCertPool
}

type exampleTLSClientCerts struct {
	RWLock sync.RWMutex
}

// TLSClientCerts overrides EndpointConfig's TLSClientCerts function which will return the list of configured client certs
func (m *exampleTLSClientCerts) TLSClientCerts() []tls.Certificate {
	var clientCerts tls.Certificate
	cb := client.TLSCerts.Client.Cert.Bytes()

	if len(cb) == 0 {
		// if no cert found in the config, return empty cert chain
		return []tls.Certificate{clientCerts}
	}

	// Load private key from cert using default crypto suite
	cs := cryptosuite.GetDefault()
	pk, err := cryptoutil.GetPrivateKeyFromCert(cb, cs)

	// If CryptoSuite fails to load private key from cert then load private key from config
	if err != nil || pk == nil {
		m.RWLock.Lock()
		defer m.RWLock.Unlock()
		ccs, err := m.loadPrivateKeyFromConfig(&client, clientCerts, cb)
		if err != nil {
			return nil
		}
		return ccs
	}

	// private key was retrieved from cert
	clientCerts, err = cryptoutil.X509KeyPair(cb, pk, cs)
	if err != nil {
		return nil
	}

	return []tls.Certificate{clientCerts}
}
func (m *exampleTLSClientCerts) loadPrivateKeyFromConfig(clientConfig *clientConfig, clientCerts tls.Certificate, cb []byte) ([]tls.Certificate, error) {

	kb := clientConfig.TLSCerts.Client.Key.Bytes()

	// load the key/cert pair from []byte
	clientCerts, err := tls.X509KeyPair(cb, kb)
	if err != nil {
		return nil, errors.Errorf("Error loading cert/key pair as TLS client credentials: %s", err)
	}

	return []tls.Certificate{clientCerts}, nil
}

type exampleCryptoConfigPath struct{}

func (m *exampleCryptoConfigPath) CryptoConfigPath() string {
	return client.CryptoConfig.Path
}

func newTLSConfig(path string) endpoint.TLSConfig {
	config := endpoint.TLSConfig{Path: pathvar.Subst(path)}
	if err := config.LoadBytes(); err != nil {
		panic(fmt.Sprintf("error loading bytes: %s", err))
	}
	return config
}

func tlsCertByBytes(path string) *x509.Certificate {

	bytes, err := ioutil.ReadFile(pathvar.Subst(path))
	if err != nil {
		return nil
	}

	block, _ := pem.Decode(bytes)

	if block != nil {
		pub, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			panic(err)
		}

		return pub
	}

	//no cert found and there is no error
	return nil
}
