/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apicore

import (
	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// FabricProvider allows overriding of fabric objects such as peer and user
type FabricProvider interface {
	NewClient(user apifabclient.User) (apifabclient.FabricClient, error)
	NewPeer(url string, certificate *x509.Certificate, serverHostOverride string) (apifabclient.Peer, error)
	NewPeerFromConfig(peerCfg *apiconfig.NetworkPeer) (apifabclient.Peer, error)
	// EnrollUser(orgID, name, pwd string) (apifabca.User, error)
	NewUser(name string, signingIdentity *apifabclient.SigningIdentity) (apifabclient.User, error)
}
