/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

// DiscoveryProvider is used to discover peers on the network
type DiscoveryProvider interface {
	NewDiscoveryService(channel Channel) (DiscoveryService, error)
}

// DiscoveryService is used to discover eligible peers on specific channel
type DiscoveryService interface {
	GetPeers(chaincodeID string) ([]Peer, error)
}
