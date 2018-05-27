/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type localDiscoveryService struct {
	config fab.EndpointConfig
	peers  []fab.Peer
}

// GetPeers is used to get local peers
func (ds *localDiscoveryService) GetPeers() ([]fab.Peer, error) {
	return ds.peers, nil
}
