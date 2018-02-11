/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabca"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// FabricProvider enables access to fabric objects such as peer and user based on config or context.
type FabricProvider interface {
	CreateChannelClient(user apifabclient.IdentityContext, cfg apifabclient.ChannelCfg) (apifabclient.Channel, error)
	CreateChannelLedger(ic apifabclient.IdentityContext, name string) (apifabclient.ChannelLedger, error)
	CreateChannelConfig(user apifabclient.IdentityContext, name string) (apifabclient.ChannelConfig, error)
	CreateResourceClient(user apifabclient.IdentityContext) (apifabclient.Resource, error)
	CreateEventHub(ic apifabclient.IdentityContext, name string) (apifabclient.EventHub, error)
	CreateCAClient(orgID string) (apifabca.FabricCAClient, error)

	CreatePeerFromConfig(peerCfg *apiconfig.NetworkPeer) (apifabclient.Peer, error)
	CreateOrdererFromConfig(cfg *apiconfig.OrdererConfig) (apifabclient.Orderer, error)
	CreateUser(name string, signingIdentity *apifabclient.SigningIdentity) (apifabclient.User, error)
}
