/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	channelImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	identityImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identitymgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
)

func TestCreateFabricProvider(t *testing.T) {
	newMockFabricProvider(t)
}

func TestCreateChannelClient(t *testing.T) {
	p := newMockFabricProvider(t)

	user := mocks.NewMockUser("user")
	client, err := p.CreateChannelClient(user, mocks.NewMockChannelCfg("mychannel"))
	if err != nil {
		t.Fatalf("Unexpected error creating client %v", err)
	}

	_, ok := client.(*channelImpl.Channel)
	if !ok {
		t.Fatalf("Unexpected client impl created: %v", client)
	}
}

func TestCreateResourceClient(t *testing.T) {
	p := newMockFabricProvider(t)

	user := mocks.NewMockUser("user")
	client, err := p.CreateResourceClient(user)
	if err != nil {
		t.Fatalf("Unexpected error creating client %v", err)
	}

	_, ok := client.(*resource.Resource)
	if !ok {
		t.Fatalf("Unexpected client impl created")
	}
}

func TestCreateCAClient(t *testing.T) {
	p := newMockFabricProvider(t)

	org := "org1"

	client, err := p.CreateIdentityManager(org)
	if err != nil {
		t.Fatalf("Unexpected error creating client %v", err)
	}

	_, ok := client.(*identitymgr.IdentityManager)
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

func verifyPeer(t *testing.T, peer fab.Peer, url string) {
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

func TestCreatePeerFromConfig(t *testing.T) {
	p := newMockFabricProvider(t)

	url := "grpc://localhost:8080"

	peerCfg := core.NetworkPeer{
		PeerConfig: core.PeerConfig{
			URL: url,
		},
	}

	peer, err := p.CreatePeerFromConfig(&peerCfg)

	if err != nil {
		t.Fatalf("Unexpected error creating peer %v", err)
	}

	verifyPeer(t, peer, url)
}

func TestCreateUser(t *testing.T) {
	org := "org1"

	p := newMockFabricProvider(t)
	cm, err := mocks.NewMockIdentityManager(org, p.providerContext.Config(), p.providerContext.CryptoSuite())
	if err != nil {
		t.Fatalf("Unexpected error creating credential manager %v", err)
	}

	signingIdentity, err := cm.GetSigningIdentity("user")
	if err != nil {
		t.Fatalf("Unexpected error getting signing identity %v", err)
	}

	user, err := p.CreateUser("user", signingIdentity)
	if err != nil {
		t.Fatalf("Unexpected error getting user %v", err)
	}

	_, ok := user.(*identityImpl.User)
	if !ok {
		t.Fatalf("Unexpected peer impl created")
	}
}

func newMockFabricProvider(t *testing.T) *FabricProvider {
	cfg, err := config.FromFile("../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf("config.FromFile failed: %v", err)
	}
	cryptoSuite, err := sw.GetSuiteByConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("cryptosuiteimpl.GetSuiteByConfig: %v", err))
	}
	ctx := mocks.NewMockProviderContextCustom(cfg, cryptoSuite, mocks.NewMockSigningManager())
	return New(ctx)
}
