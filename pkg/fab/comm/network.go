/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

// NetworkPeerConfigFromURL fetches the peer configuration based on a URL.
func NetworkPeerConfigFromURL(cfg fab.EndpointConfig, url string) (*fab.NetworkPeer, error) {
	peerCfg, err := cfg.PeerConfigByURL(url)
	if err != nil {
		return nil, errors.WithMessage(err, "peer not found")
	}
	if peerCfg == nil {
		return nil, errors.New("peer not found")
	}

	// find MSP ID
	networkPeers, err := cfg.NetworkPeers()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to load network peer config")
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
