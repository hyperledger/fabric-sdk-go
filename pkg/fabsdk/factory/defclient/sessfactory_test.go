// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	mockapisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/pkg/errors"
)

func TestCreateChannelClient(t *testing.T) {
	p := newMockProviders(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSDK := mockapisdk.NewMockProviders(mockCtrl)

	mockSDK.EXPECT().ChannelProvider().Return(p.ChannelProvider)
	mockSDK.EXPECT().DiscoveryProvider().Return(p.DiscoveryProvider)
	mockSDK.EXPECT().SelectionProvider().Return(p.SelectionProvider)
	mockSDK.EXPECT().Config().Return(p.Config)

	factory := NewSessionClientFactory()
	session := newMockSession()

	_, err := factory.CreateChannelClient(mockSDK, session, "mychannel", nil)
	if err != nil {
		t.Fatalf("Unexpected error creating channel client: %v", err)
	}

}

func TestCreateChannelClientBadChannel(t *testing.T) {
	p := newMockProviders(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSDK := mockapisdk.NewMockProviders(mockCtrl)

	mockSDK.EXPECT().ChannelProvider().Return(p.ChannelProvider)

	factory := NewSessionClientFactory()
	session := newMockSession()

	_, err := factory.CreateChannelClient(mockSDK, session, "badchannel", nil)
	if err == nil {
		t.Fatalf("Expected error creating channel client")
	}
}

type mockProviders struct {
	CryptoSuite       core.CryptoSuite
	StateStore        contextApi.KVStore
	Config            core.Config
	SigningManager    contextApi.SigningManager
	FabricProvider    api.FabricProvider
	DiscoveryProvider fab.DiscoveryProvider
	SelectionProvider fab.SelectionProvider
	ChannelProvider   fab.ChannelProvider
}

func newMockProviders(t *testing.T) *mockProviders {
	coreFactory := defcore.NewProviderFactory()
	svcFactory := defsvc.NewProviderFactory()

	config, err := config.FromFile("../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}
	cryptosuite, err := coreFactory.CreateCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	stateStore, err := coreFactory.CreateStateStoreProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	signer, err := coreFactory.CreateSigningManager(cryptosuite, config)
	if err != nil {
		t.Fatalf("Unexpected error creating signing manager %v", err)
	}

	ctx := fabmocks.NewMockProviderContextCustom(config, cryptosuite, signer)
	fabricProvider, err := coreFactory.CreateFabricProvider(ctx)
	if err != nil {
		t.Fatalf("Unexpected error creating fabric provider %v", err)
	}

	dp, err := svcFactory.CreateDiscoveryProvider(config, fabricProvider)
	if err != nil {
		t.Fatalf("Unexpected error creating discovery provider %v", err)
	}

	sp, err := svcFactory.CreateSelectionProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating discovery provider %v", err)
	}

	cp, err := chpvdr.New(fabricProvider)
	if err != nil {
		t.Fatalf("Unexpected error creating discovery provider %v", err)
	}

	cp.SetChannelConfig(fabmocks.NewMockChannelCfg("mychannel"))
	cp.SetChannelConfig(fabmocks.NewMockChannelCfg("orgchannel"))

	providers := mockProviders{
		CryptoSuite:       cryptosuite,
		StateStore:        stateStore,
		Config:            config,
		SigningManager:    signer,
		FabricProvider:    fabricProvider,
		DiscoveryProvider: dp,
		SelectionProvider: sp,
		ChannelProvider:   cp,
	}

	return &providers
}

type mockSession struct {
	context.IdentityContext
	IsChError bool
	IsEHError bool
}

func newMockSession() *mockSession {
	return newMockSessionWithUser("user1", "Org1MSP")
}

func newMockSessionWithUser(username, mspID string) *mockSession {
	ic := fabmocks.NewMockUserWithMSPID(username, mspID)
	session := mockSession{
		IdentityContext: ic,
	}
	return &session
}

func (s *mockSession) EventHub(channelID string) (fab.EventHub, error) {
	if s.IsEHError {
		return nil, errors.New("error")
	}
	return nil, nil
}
