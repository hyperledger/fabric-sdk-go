/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
	channelImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	identityImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/resource"
)

func TestNewFabricProvider(t *testing.T) {
	newMockFabricProvider(t)
}

func TestNewChannelClient(t *testing.T) {
	p := newMockFabricProvider(t)

	user := mocks.NewMockUser("user")
	client, err := p.NewChannelClient(user, "mychannel")
	if err != nil {
		t.Fatalf("Unexpected error creating client %v", err)
	}

	_, ok := client.(*channelImpl.Channel)
	if !ok {
		t.Fatalf("Unexpected client impl created: %v", client)
	}
}

func TestNewResourceClient(t *testing.T) {
	p := newMockFabricProvider(t)

	user := mocks.NewMockUser("user")
	client, err := p.NewResourceClient(user)
	if err != nil {
		t.Fatalf("Unexpected error creating client %v", err)
	}

	_, ok := client.(*resource.Resource)
	if !ok {
		t.Fatalf("Unexpected client impl created")
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

	conf, err := p.providerContext.Config().CAConfig(org)
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
	cm, err := mocks.NewMockCredentialManager(org, p.providerContext.Config(), p.providerContext.CryptoSuite())
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
	ctx := mocks.NewMockProviderContext()
	return New(ctx)
}

/*
config := mocks.NewMockConfig()
cryptosuite, err := sw.GetSuiteWithDefaultEphemeral()
if err != nil {
	t.Fatalf("Unexpected error getting cryptosuite %v", err)
}
signer := mocks.NewMockSigningManager()

ctx := mocks.MockProviderContext{
	Config: mocks.NewMockConfig(),

}
*/
