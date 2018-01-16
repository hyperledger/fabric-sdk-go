/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"reflect"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/sw"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	identityImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

func TestNewFabricProvider(t *testing.T) {
	newMockFabricProvider(t)
}

func TestNewClient(t *testing.T) {
	p := newMockFabricProvider(t)

	user := mocks.NewMockUser("user")
	client, err := p.NewClient(user)
	if err != nil {
		t.Fatalf("Unexpected error creating client %v", err)
	}

	_, ok := client.(*clientImpl.Client)
	if !ok {
		t.Fatalf("Unexpected client impl created")
	}

	// Brittle tests follow (may need to be removed when we minimize client interface)
	if !reflect.DeepEqual(client.StateStore(), p.stateStore) {
		t.Fatalf("Unexpected keyvalue store")
	}
	if !reflect.DeepEqual(client.CryptoSuite(), p.cryptoSuite) {
		t.Fatalf("Unexpected cryptosuite")
	}
	if !reflect.DeepEqual(client.SigningManager(), p.signer) {
		t.Fatalf("Unexpected signing manager")
	}
	if !reflect.DeepEqual(client.Config(), p.config) {
		t.Fatalf("Unexpected config")
	}
}

func TestNewCAClient(t *testing.T) {
	p := newMockFabricProvider(t)

	org := "org1"

	client, err := p.NewCAClient(org)
	if err != nil {
		t.Fatalf("Unexpected error creating client %v", err)
	}

	_, ok := client.(*fabricCAClient.FabricCA)
	if !ok {
		t.Fatalf("Unexpected client impl created")
	}

	conf, err := p.config.CAConfig(org)
	if err != nil {
		t.Fatalf("Unexpected error getting CA config %v", err)
	}

	// Brittle tests follow
	e := conf.CAName
	a := client.CAName()

	if a != e {
		t.Fatalf("Unexpected CA name %s", a)
	}
}

func TestNewPeer(t *testing.T) {
	p := newMockFabricProvider(t)

	url := "grpcs://localhost:8080"

	peer, err := p.NewPeer(url, mock_apiconfig.GoodCert, "")
	if err != nil {
		t.Fatalf("Unexpected error creating peer %v", err)
	}

	verifyPeer(t, peer, url)
}

func verifyPeer(t *testing.T, peer apifabclient.Peer, url string) {
	_, ok := peer.(*peerImpl.Peer)
	if !ok {
		t.Fatalf("Unexpected peer impl created")
	}

	// Brittle tests follow
	a := peer.URL()

	if a != url {
		t.Fatalf("Unexpected URL %s", a)
	}
}

func TestNewPeerFromConfig(t *testing.T) {
	p := newMockFabricProvider(t)

	url := "grpc://localhost:8080"

	peerCfg := apiconfig.NetworkPeer{
		PeerConfig: apiconfig.PeerConfig{
			URL: url,
		},
	}

	peer, err := p.NewPeerFromConfig(&peerCfg)

	if err != nil {
		t.Fatalf("Unexpected error creating peer %v", err)
	}

	verifyPeer(t, peer, url)
}

func TestNewUser(t *testing.T) {
	org := "org1"

	p := newMockFabricProvider(t)
	cm, err := mocks.NewMockCredentialManager(org, p.config, p.cryptoSuite)
	if err != nil {
		t.Fatalf("Unexpected error creating credential manager %v", err)
	}

	signingIdentity, err := cm.GetSigningIdentity("user")
	if err != nil {
		t.Fatalf("Unexpected error getting signing identity %v", err)
	}

	user, err := p.NewUser("user", signingIdentity)
	if err != nil {
		t.Fatalf("Unexpected error getting user %v", err)
	}

	_, ok := user.(*identityImpl.User)
	if !ok {
		t.Fatalf("Unexpected peer impl created")
	}
}

func newMockFabricProvider(t *testing.T) *FabricProvider {
	config := mocks.NewMockConfig()
	kv, err := keyvaluestore.CreateNewFileKeyValueStore("/tmp/fabsdktest")
	if err != nil {
		t.Fatalf("Unexpected error getting keyvalue store %v", err)
	}
	cryptosuite, err := sw.GetSuiteWithDefaultEphemeral()
	if err != nil {
		t.Fatalf("Unexpected error getting cryptosuite %v", err)
	}
	signer := mocks.NewMockSigningManager()

	return NewFabricProvider(config, kv, cryptosuite, signer)
}
