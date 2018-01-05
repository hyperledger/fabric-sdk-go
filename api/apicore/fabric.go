/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apicore

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// FabricProvider allows overriding of fabric objects such as peer and user
type FabricProvider interface {
	NewClient(user apifabclient.User) (apifabclient.FabricClient, error)
	NewPeer(url string, certificate string, serverHostOverride string) (apifabclient.Peer, error)
	NewPeerFromConfig(peerCfg *apiconfig.NetworkPeer) (apifabclient.Peer, error)
	// EnrollUser(orgID, name, pwd string) (apifabca.User, error)
	NewUser(name string, signingIdentity *apifabclient.SigningIdentity) (apifabclient.User, error)
}
