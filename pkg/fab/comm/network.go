/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

// NetworkPeerConfig fetches the peer configuration based on a key (name or URL).
func NetworkPeerConfig(cfg fab.EndpointConfig, key string) (*fab.NetworkPeer, error) {
	peerCfg, ok := cfg.PeerConfig(key)
	if !ok {
		return nil, errors.Errorf("peer not found")
	}

	// find MSP ID
	networkPeers, ok := cfg.NetworkPeers()
	if !ok {
		return nil, errors.New("unable to load network peer config")
	}

	var mspID string
	for _, peer := range networkPeers {
		if peer.URL == peerCfg.URL { // need to use the looked-up URL due to matching
			mspID = peer.MSPID
			break
		}
	}

	np := fab.NetworkPeer{
		PeerConfig: *peerCfg,
		MSPID:      mspID,
	}

	return &np, nil
}

// SearchPeerConfigFromURL searches for the peer configuration based on a URL.
func SearchPeerConfigFromURL(cfg fab.EndpointConfig, url string) (*fab.PeerConfig, error) {
	peerCfg, ok := cfg.PeerConfig(url)

	if ok {
		return peerCfg, nil
	}

	//If the given url is already parsed URL through entity matcher, then 'cfg.PeerConfig()'
	//may return NoMatchingPeerEntity error. So retry with network peer URLs
	networkPeers, ok := cfg.NetworkPeers()
	if !ok {
		return nil, errors.New("unable to load network peer config")
	}

	for _, peer := range networkPeers {
		if peer.URL == url {
			return &peer.PeerConfig, nil
		}
	}

	return nil, errors.Errorf("unable to get peerconfig for given url : %s", url)
}

// MSPID returns the MSP ID for the requested organization
func MSPID(cfg fab.EndpointConfig, org string) (string, bool) {
	networkConfig, ok := cfg.NetworkConfig()
	if !ok {
		return "", false
	}
	// viper lowercases all key maps, org is lower case
	mspID := networkConfig.Organizations[strings.ToLower(org)].MSPID
	if mspID == "" {
		return "", false
	}

	return mspID, true
}
