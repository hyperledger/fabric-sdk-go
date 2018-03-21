/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	coreMocks "github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
)

type mockClientContext struct {
	context.Providers
	msp.SigningIdentity
}

func TestCreateInfraProvider(t *testing.T) {
	newInfraProvider(t)
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
	p := newInfraProvider(t)

	url := "grpc://localhost:9999"

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
	p := newInfraProvider(t)
	ctx := mocks.NewMockProviderContext()
	user := mspmocks.NewMockSigningIdentity("user", "user")
	clientCtx := &mockClientContext{
		Providers:       ctx,
		SigningIdentity: user,
	}

	m, err := p.CreateChannelMembership(clientCtx, "test")
	assert.Nil(t, err)
	assert.NotNil(t, m)
}

func newInfraProvider(t *testing.T) *InfraProvider {
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

	ctx := mocks.NewMockProviderContextCustom(cfg, cryptoSuite, coreMocks.NewMockSigningManager(), &mspmocks.MockUserStore{}, im)
	ip := New(cfg)
	ip.Initialize(ctx)

	return ip
}
