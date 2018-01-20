/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apicore

import (
	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabca"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// FabricProvider enables access to fabric objects such as peer and user
type FabricProvider interface {
	NewChannelClient(user apifabclient.IdentityContext, name string) (apifabclient.Channel, error)
	NewResourceClient(user apifabclient.IdentityContext) (apifabclient.Resource, error)
	NewCAClient(orgID string) (apifabca.FabricCAClient, error)

	NewPeer(url string, certificate *x509.Certificate, serverHostOverride string) (apifabclient.Peer, error)
	NewPeerFromConfig(peerCfg *apiconfig.NetworkPeer) (apifabclient.Peer, error)
	NewUser(name string, signingIdentity *apifabclient.SigningIdentity) (apifabclient.User, error)
}
