/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// discoveryService implements discovery service
type discoveryService struct {
	config fab.EndpointConfig
	peers  []fab.Peer
}

// GetPeers is used to get peers
func (ds *discoveryService) GetPeers() ([]fab.Peer, error) {

	return ds.peers, nil
}
