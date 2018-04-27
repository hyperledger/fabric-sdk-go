/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package configless

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
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
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/pkg/errors"
)

// endpointconfig_override_test.go is an example of programmatically configuring the sdk by injecting instances that implement EndpointConfig's functions (representing the sdk's configs)
// for the sake of overriding EndpointConfig integration tests, the structure variables below are similar to what is found in /test/fixtures/config/config_test.yaml
// application developers can fully override these functions to load configs in any way that suit their application need

// NOTE: 1. to support test local (flag: TEST_LOCAL=true to use localhost:* URLs for peers, orderers, CAs everywhere), new...() constructor functions where created to test if this flag
//       is enabled using verifyIsLocal...() function calls. These calls will basically switch config URLs (peers, orderers or CA configs) / EventURLs (peer configs) into "localhost:..."
//       Make sure your local /etc/hosts file does not have any ip-dns mapping entries for peers/orderers/CAs
//
//       2. the test assumes the use of the default channel block used in the remaining regular integration tests (for example look at Orderer.Addresses value in
//       test/fixtures/fabric/..specific target fabric release../config/configtx.yaml to see the URL value assigned to the orderer for a specific channel).
//       So Even if the below interfaces will override orderers to localhost for TEST_LOCAL=true, the SDK will still try
//       to create an orderer with the URL found in the channel block mentioned above. You can either create another channel block for your channels,
//       or if you want to use an existing channel block but still want to change the orderer URL, then you can implement EntityMatchers logic for your orderers
//       which is commented out in the code below for reference. Using EntityMatchers will allow the configs to be able to find mapped Orderers/Peers/CA URLs.

var (
	localhostRep = "localhost:"
	dnsMatchRegX = ".*:"
	clientConfig = msp.ClientConfig{
		Organization:    "org1",
		Logging:         api.LoggingType{Level: "info"},
		CryptoConfig:    msp.CCType{Path: "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/${CRYPTOCONFIG_FIXTURES_PATH}"},
		CredentialStore: msp.CredentialStoreType{Path: "/tmp/msp"},
		TLSCerts: endpoint.MutualTLSConfig{Client: endpoint.TLSKeyPair{
			Key:  endpoint.TLSConfig{Path: pathvar.Subst("${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/config/mutual_tls/client_sdk_go-key.pem")},
			Cert: endpoint.TLSConfig{Path: pathvar.Subst("${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/config/mutual_tls/client_sdk_go.pem")}}},
	}

	channelsConfig = map[string]fab.ChannelNetworkConfig{
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
			MSPID:      "Org1MSP",
			CryptoPath: "peerOrganizations/org1.example.com/users/{username}@org1.example.com/msp",
			Peers:      []string{"peer0.org1.example.com"},
			CertificateAuthorities: []string{"ca.org1.example.com"},
		},
		"org2": {
			MSPID:      "Org2MSP",
			CryptoPath: "peerOrganizations/org1.example.com/users/{username}@org2.example.com/msp",
			Peers:      []string{"peer0.org2.example.com"},
			CertificateAuthorities: []string{"ca.org2.example.com"},
		},
		"ordererorg": {
			MSPID:      "OrdererOrg",
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
			TLSCACerts: endpoint.TLSConfig{
				Path: "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/${CRYPTOCONFIG_FIXTURES_PATH}/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem",
			},
		},
	}

	peersConfig = map[string]fab.PeerConfig{
		"peer0.org1.example.com": {
			URL:      "peer0.org1.example.com:7051",
			EventURL: "peer0.org1.example.com:7053",
			GRPCOptions: map[string]interface{}{
				"ssl-target-name-override": "peer0.org1.example.com",
				"keep-alive-time":          0 * time.Second,
				"keep-alive-timeout":       20 * time.Second,
				"keep-alive-permit":        false,
				"fail-fast":                false,
				"allow-insecure":           false,
			},
			TLSCACerts: endpoint.TLSConfig{
				Path: "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem",
			},
		},
		"peer0.org2.example.com": {
			URL:      "peer0.org2.example.com:8051",
			EventURL: "peer0.org2.example.com:8053",
			GRPCOptions: map[string]interface{}{
				"ssl-target-name-override": "peer0.org2.example.com",
				"keep-alive-time":          0 * time.Second,
				"keep-alive-timeout":       20 * time.Second,
				"keep-alive-permit":        false,
				"fail-fast":                false,
				"allow-insecure":           false,
			},
			TLSCACerts: endpoint.TLSConfig{
				Path: "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/${CRYPTOCONFIG_FIXTURES_PATH}/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem",
			},
		},
	}

	caConfig = map[string]msp.CAConfig{
		"ca.org1.example.com": {
			URL: "https://ca.org1.example.com:7054",
			TLSCACerts: endpoint.MutualTLSConfig{
				Path: "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabricca/tls/certs/ca_root.pem",
				Client: endpoint.TLSKeyPair{
					Key: endpoint.TLSConfig{
						Path: "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabricca/tls/certs/client/client_fabric_client-key.pem",
					},
					Cert: endpoint.TLSConfig{
						Path: "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabricca/tls/certs/client/client_fabric_client.pem",
					},
				},
			},
			Registrar: msp.EnrollCredentials{
				EnrollID:     "admin",
				EnrollSecret: "adminpw",
			},
			CAName: "ca.org1.example.com",
		},
		"ca.org2.example.com": {
			URL: "https://ca.org2.example.com:8054",
			TLSCACerts: endpoint.MutualTLSConfig{
				Path: "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabricca/tls/certs/ca_root.pem",
				Client: endpoint.TLSKeyPair{
					Key: endpoint.TLSConfig{
						Path: "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabricca/tls/certs/client/client_fabric_client-key.pem",
					},
					Cert: endpoint.TLSConfig{
						Path: "${GOPATH}/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/fabricca/tls/certs/client/client_fabric_client.pem",
					},
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
		Name:          "config-overridden network",
		Description:   "This config structure is an example of overriding the sdk config by injecting interfaces instead of using a config file",
		Client:        clientConfig,
		Channels:      channelsConfig,
		Organizations: orgsConfig,
		Orderers:      newOrderersConfig(),
		Peers:         newPeersConfig(),
		CertificateAuthorities: newCAsConfig(),
		// EntityMatchers are not used in this implementation
		//EntityMatchers: entityMatchers,
	}

	// creating instances of each interface to be referenced in the integration tests:
	timeoutImpl          = &exampleTimeout{}
	mspIDImpl            = &exampleMSPID{}
	peerMSPIDImpl        = &examplePeerMSPID{}
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
	eventServiceTypeImpl = &exampleEventServiceType{}
	tlsClientCertsImpl   = &exampleTLSClientCerts{}
	cryptoConfigPathImpl = &exampleCryptoConfigPath{}
	endpointConfigImpls  = []interface{}{
		timeoutImpl,
		mspIDImpl,
		peerMSPIDImpl,
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
		eventServiceTypeImpl,
		tlsClientCertsImpl,
		cryptoConfigPathImpl,
	}
)

type exampleTimeout struct{}

var defaultTypes = map[fab.TimeoutType]time.Duration{
	fab.EndorserConnection:       time.Second * 10,
	fab.PeerResponse:             time.Minute * 3,
	fab.DiscoveryGreylistExpiry:  time.Second * 10,
	fab.EventHubConnection:       time.Second * 15,
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

type exampleMSPID struct{}

//MSPID overrides EndpointConfig's MSPID function which returns the mspID for the given org name in the arg
func (m *exampleMSPID) MSPID(org string) (string, error) {
	//lowercase org name to make it case insensitive, depends on application preference, for the sake of this example, make it case in-sensitive
	mspID := orgsConfig[strings.ToLower(org)].MSPID
	if mspID == "" {
		return "", errors.Errorf("MSP ID is empty for org: %s", org)
	}

	return mspID, nil
}

type examplePeerMSPID struct{}

//PeerMSPID overrides EndpointConfig's PeerMSPID function which returns the mspID for the given org name in the arg
func (m *examplePeerMSPID) PeerMSPID(name string) (string, error) {
	var mspID string

	// Find organisation/msp that peer belongs to
	for _, org := range orgsConfig {
		for i := 0; i < len(org.Peers); i++ {
			if strings.EqualFold(org.Peers[i], name) {
				// peer belongs to this org add org msp
				mspID = org.MSPID
				break
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

	return mspID, nil
}

func verifyIsLocalCAsURLs(caConfigs map[string]msp.CAConfig) map[string]msp.CAConfig {
	re := regexp.MustCompile(dnsMatchRegX)
	var newCfg = make(map[string]msp.CAConfig)
	// for local integration tests, replace all urls DNS to localhost:
	if integration.IsLocal() {
		for k, caCfg := range caConfigs {
			caCfg.URL = re.ReplaceAllString(caCfg.URL, localhostRep)
			newCfg[k] = caCfg
		}
	}
	return newCfg
}

func newCAsConfig() map[string]msp.CAConfig {
	c := verifyIsLocalCAsURLs(caConfig)
	caConfig = c
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
func (m *exampleOrderersConfig) OrderersConfig() ([]fab.OrdererConfig, error) {
	orderers := []fab.OrdererConfig{}

	for _, orderer := range orderersConfig {

		if orderer.TLSCACerts.Path != "" {
			orderer.TLSCACerts.Path = pathvar.Subst(orderer.TLSCACerts.Path)
		} else if len(orderer.TLSCACerts.Pem) == 0 && !m.isSystemCertPool {
			return nil, errors.Errorf("Orderer has no certs configured. Make sure TLSCACerts.Pem or TLSCACerts.Path is set for %s", orderer.URL)
		}

		orderers = append(orderers, orderer)
	}

	return orderers, nil
}

type exampleOrdererConfig struct{}

//OrdererConfig overrides EndpointConfig's OrdererConfig function which returns the ordererConfig instance for the name/URL arg
func (m *exampleOrdererConfig) OrdererConfig(ordererNameOrURL string) (*fab.OrdererConfig, error) {
	orderer, ok := networkConfig.Orderers[strings.ToLower(ordererNameOrURL)]
	if !ok {
		// EntityMatchers are not used in this implementation, below is an example of how to use them if needed, see default implementation for live example
		//matchingOrdererConfig := m.tryMatchingOrdererConfig(networkConfig, strings.ToLower(ordererNameOrURL))
		//if matchingOrdererConfig == nil {
		//	return nil, errors.WithStack(status.New(status.ClientStatus, status.NoMatchingOrdererEntity.ToInt32(), "no matching orderer config found", nil))
		//}
		//orderer = *matchingOrdererConfig
		return nil, errors.Errorf("orderer '%s' not found in the configs", ordererNameOrURL)
	}

	if orderer.TLSCACerts.Path != "" {
		orderer.TLSCACerts.Path = pathvar.Subst(orderer.TLSCACerts.Path)
	}

	return &orderer, nil
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
			peer.EventURL = re.ReplaceAllString(peer.EventURL, localhostRep)
			newConfigs[k] = peer
		}
	}

	if len(newConfigs) == 0 {
		return pConfig
	}
	return newConfigs
}

//newPeersConfigImpl will create a new examplePeersConfig instance with proper peers URLs and EventURLs (local vs normal) tests
// local tests use localhost urls, while the remaining tests use default values as set in peersConfig var
func newPeersConfigImpl() *examplePeersConfig {
	pConfig := verifyIsLocalPeersURLs(peersConfig)
	peersConfig = pConfig
	p := &examplePeersConfig{}
	return p
}

//PeersConfig overrides EndpointConfig's PeersConfig function which returns the peersConfig list
func (m *examplePeersConfig) PeersConfig(org string) ([]fab.PeerConfig, error) {
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
			return nil, err
		}
		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = pathvar.Subst(p.TLSCACerts.Path)
		}

		peers = append(peers, p)
	}
	return peers, nil
}

func (m *examplePeersConfig) verifyPeerConfig(p fab.PeerConfig, peerName string, tlsEnabled bool) error {
	if p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && len(p.TLSCACerts.Pem) == 0 && p.TLSCACerts.Path == "" && !m.isSystemCertPool {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

type examplePeerConfig struct{}

// PeerConfig overrides EndpointConfig's PeerConfig function which returns the peerConfig instance for the name/URL arg
func (m *examplePeerConfig) PeerConfig(nameOrURL string) (*fab.PeerConfig, error) {
	pcfg, ok := peersConfig[nameOrURL]
	if ok {
		return &pcfg, nil
	}
	if pcfg.TLSCACerts.Path != "" {
		pcfg.TLSCACerts.Path = pathvar.Subst(pcfg.TLSCACerts.Path)
	}
	// EntityMatchers are not used in this implementation
	// see default implementation (pkg/fab/endpointconfig.go) to see how they're used

	return nil, errors.Errorf("peer '%s' not found in the configs", nameOrURL)
}

type exampleNetworkConfig struct{}

// NetworkConfig overrides EndpointConfig's NetworkConfig function which returns the full network Config instance
func (m *exampleNetworkConfig) NetworkConfig() (*fab.NetworkConfig, error) {
	return &networkConfig, nil
}

type exampleNetworkPeers struct {
	isSystemCertPool bool
}

//NetworkPeers overrides EndpointConfig's NetworkPeers function which returns the networkPeers list
func (m *exampleNetworkPeers) NetworkPeers() ([]fab.NetworkPeer, error) {
	netPeers := []fab.NetworkPeer{}
	// referencing another interface to call PeerMSPID to match config yaml content
	peerMSPID := &examplePeerMSPID{}

	for name, p := range networkConfig.Peers {

		if err := m.verifyPeerConfig(p, name, endpoint.IsTLSEnabled(p.URL)); err != nil {
			return nil, err
		}

		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = pathvar.Subst(p.TLSCACerts.Path)
		}

		mspID, err := peerMSPID.PeerMSPID(name)
		if err != nil {
			return nil, errors.Errorf("failed to retrieve msp id for peer %s", name)
		}

		netPeer := fab.NetworkPeer{PeerConfig: p, MSPID: mspID}
		netPeers = append(netPeers, netPeer)
	}

	return netPeers, nil
}
func (m *exampleNetworkPeers) verifyPeerConfig(p fab.PeerConfig, peerName string, tlsEnabled bool) error {
	if p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && len(p.TLSCACerts.Pem) == 0 && p.TLSCACerts.Path == "" && !m.isSystemCertPool {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

type exampleChannelConfig struct{}

// ChannelConfig overrides EndpointConfig's ChannelConfig function which returns the channelConfig instance for the channel name arg
func (m *exampleChannelConfig) ChannelConfig(channelName string) (*fab.ChannelNetworkConfig, error) {
	ch, ok := channelsConfig[strings.ToLower(channelName)]
	if !ok {
		// EntityMatchers are not used in this implementation, below is an example of how to use them if needed
		//matchingChannel, _, matchErr := m.tryMatchingChannelConfig(channelName)
		//if matchErr != nil {
		//	return nil, errors.WithMessage(matchErr, "channel config not found")
		//}
		//return matchingChannel, nil
		return nil, errors.Errorf("No channel found for '%s'", channelName)
	}

	return &ch, nil
}

type exampleChannelPeers struct {
	isSystemCertPool bool
}

// ChannelPeers overrides EndpointConfig's ChannelPeers function which returns the list of peers for the channel name arg
func (m *exampleChannelPeers) ChannelPeers(channelName string) ([]fab.ChannelPeer, error) {
	peers := []fab.ChannelPeer{}
	// referencing another interface to call PeerMSPID to match config yaml content
	peerMSPID := &examplePeerMSPID{}

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
		return nil, errors.Errorf("No channel found for '%s'", channelName)
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
			return nil, errors.Errorf("No peer found '%s'", peerName)
		}

		if err := m.verifyPeerConfig(p, peerName, endpoint.IsTLSEnabled(p.URL)); err != nil {
			return nil, err
		}

		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = pathvar.Subst(p.TLSCACerts.Path)
		}

		mspID, err := peerMSPID.PeerMSPID(peerName)
		if err != nil {
			return nil, errors.Errorf("failed to retrieve msp id for peer %s", peerName)
		}

		networkPeer := fab.NetworkPeer{PeerConfig: p, MSPID: mspID}

		peer := fab.ChannelPeer{PeerChannelConfig: chPeerConfig, NetworkPeer: networkPeer}

		peers = append(peers, peer)
	}

	return peers, nil

}
func (m *exampleChannelPeers) verifyPeerConfig(p fab.PeerConfig, peerName string, tlsEnabled bool) error {
	if p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && len(p.TLSCACerts.Pem) == 0 && p.TLSCACerts.Path == "" && !m.isSystemCertPool {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

type exampleChannelOrderers struct{}

// ChannelOrderers overrides EndpointConfig's ChannelOrderers function which returns the list of orderers for the channel name arg
func (m *exampleChannelOrderers) ChannelOrderers(channelName string) ([]fab.OrdererConfig, error) {
	// referencing other interfaces to call ChannelConfig and OrdererConfig to match config yaml content
	chCfg := &exampleChannelConfig{}
	oCfg := &exampleOrdererConfig{}

	orderers := []fab.OrdererConfig{}
	channel, err := chCfg.ChannelConfig(channelName)
	if err != nil || channel == nil {
		return nil, errors.Errorf("Unable to retrieve channel config: %s", err)
	}

	for _, chOrderer := range channel.Orderers {
		orderer, err := oCfg.OrdererConfig(chOrderer)
		if err != nil || orderer == nil {
			return nil, errors.Errorf("unable to retrieve orderer config: %s", err)
		}

		orderers = append(orderers, *orderer)
	}

	return orderers, nil
}

type exampleTLSCACertPool struct {
	tlsCertPool commtls.CertPool
}

//newTLSCACertPool will create a new exampleTLSCACertPool instance with useSystemCertPool bool flag
func newTLSCACertPool(useSystemCertPool bool) *exampleTLSCACertPool {
	m := &exampleTLSCACertPool{}
	m.tlsCertPool = commtls.NewCertPool(useSystemCertPool)
	return m
}

// TLSCACertPool overrides EndpointConfig's TLSCACertPool function which will add the list of cert args to the cert pool and return it
func (m *exampleTLSCACertPool) TLSCACertPool(certs ...*x509.Certificate) (*x509.CertPool, error) {
	return m.tlsCertPool.Get(certs...)
}

type exampleEventServiceType struct{}

func (m *exampleEventServiceType) EventServiceType() fab.EventServiceType {
	// if this test is run for the previous release (1.0) then update the config with EVENT_HUB as it doesn't support deliveryService
	if os.Getenv("FABRIC_SDK_CLIENT_EVENTSERVICE_TYPE") == "eventhub" {
		return fab.EventHubEventServiceType
	}
	return fab.DeliverEventServiceType
	//or for EventHub service type, but most configs use Delivery Service starting release 1.1
	//return fab.EventHubEventServiceType
}

type exampleTLSClientCerts struct {
	RWLock *sync.RWMutex
}

// TLSClientCerts overrides EndpointConfig's TLSClientCerts function which will return the list of configured client certs
func (m *exampleTLSClientCerts) TLSClientCerts() ([]tls.Certificate, error) {
	if m.RWLock == nil {
		m.RWLock = &sync.RWMutex{}
	}
	var clientCerts tls.Certificate
	var cb []byte
	cb, err := clientConfig.TLSCerts.Client.Cert.Bytes()
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
		m.RWLock.Lock()
		defer m.RWLock.Unlock()
		return m.loadPrivateKeyFromConfig(&clientConfig, clientCerts, cb)
	}

	// private key was retrieved from cert
	clientCerts, err = cryptoutil.X509KeyPair(cb, pk, cs)
	if err != nil {
		return nil, err
	}

	return []tls.Certificate{clientCerts}, nil
}
func (m *exampleTLSClientCerts) loadPrivateKeyFromConfig(clientConfig *msp.ClientConfig, clientCerts tls.Certificate, cb []byte) ([]tls.Certificate, error) {
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

	return []tls.Certificate{clientCerts}, nil
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

type exampleCryptoConfigPath struct{}

func (m *exampleCryptoConfigPath) CryptoConfigPath() string {
	return pathvar.Subst(clientConfig.CryptoConfig.Path)
}
