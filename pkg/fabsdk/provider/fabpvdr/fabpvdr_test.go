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
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/stretchr/testify/assert"
)

func TestCreateInfraProvider(t *testing.T) {
	newMockInfraProvider(t)
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
	p := newMockInfraProvider(t)

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

func TestCreateMembership(t *testing.T) {
	p := newMockInfraProvider(t)
	m, err := p.CreateChannelMembership(mocks.NewMockChannelCfg(""))
	assert.Nil(t, err)
	assert.NotNil(t, m)
}

func newMockInfraProvider(t *testing.T) *InfraProvider {
	cfg, err := config.FromFile("../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf("config.FromFile failed: %v", err)
	}
	cryptoSuite, err := sw.GetSuiteByConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("cryptosuiteimpl.GetSuiteByConfig: %v", err))
	}
	im := make(map[string]msp.IdentityManager)
	im[""] = &mocks.MockIdentityManager{}

	ctx := mocks.NewMockProviderContextCustom(cfg, cryptoSuite, mocks.NewMockSigningManager(), mocks.NewMockStateStore(), im)
	ip := New(cfg)
	ip.Initialize(ctx)

	return ip
}
