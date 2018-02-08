// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/kvstore"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	chImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chclient"
	chmgmtImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chmgmtclient"
	resmgmtImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/resmgmtclient"
	mockapisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/pkg/errors"
)

func TestNewChannelMgmtClient(t *testing.T) {
	p := newMockProviders(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSDK := mockapisdk.NewMockProviders(mockCtrl)

	mockSDK.EXPECT().FabricProvider().Return(p.FabricProvider)

	factory := NewSessionClientFactory()
	session := newMockSession()

	client, err := factory.NewChannelMgmtClient(mockSDK, session)
	if err != nil {
		t.Fatalf("Unexpected error creating system client %v", err)
	}

	_, ok := client.(*chmgmtImpl.ChannelMgmtClient)
	if !ok {
		t.Fatalf("Unexpected client created")
	}
}

func TestNewResourceMgmtClient(t *testing.T) {
	p := newMockProviders(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSDK := mockapisdk.NewMockProviders(mockCtrl)

	mockSDK.EXPECT().ChannelProvider().Return(p.ChannelProvider)
	mockSDK.EXPECT().FabricProvider().Return(p.FabricProvider)
	mockSDK.EXPECT().DiscoveryProvider().Return(p.DiscoveryProvider)

	factory := NewSessionClientFactory()
	session := newMockSession()

	client, err := factory.NewResourceMgmtClient(mockSDK, session, nil)
	if err != nil {
		t.Fatalf("Unexpected error creating system client %v", err)
	}

	_, ok := client.(*resmgmtImpl.ResourceMgmtClient)
	if !ok {
		t.Fatalf("Unexpected client created")
	}
}

func TestNewChannelClient(t *testing.T) {
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

	client, err := factory.NewChannelClient(mockSDK, session, "mychannel", nil)
	if err != nil {
		t.Fatalf("Unexpected error creating channel client: %v", err)
	}

	_, ok := client.(*chImpl.ChannelClient)
	if !ok {
		t.Fatalf("Unexpected client created")
	}
}

func TestNewChannelClientBadChannel(t *testing.T) {
	p := newMockProviders(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSDK := mockapisdk.NewMockProviders(mockCtrl)

	mockSDK.EXPECT().ChannelProvider().Return(p.ChannelProvider)

	factory := NewSessionClientFactory()
	session := newMockSession()

	_, err := factory.NewChannelClient(mockSDK, session, "badchannel", nil)
	if err == nil {
		t.Fatalf("Expected error creating channel client")
	}
}

type mockProviders struct {
	CryptoSuite       apicryptosuite.CryptoSuite
	StateStore        kvstore.KVStore
	Config            apiconfig.Config
	SigningManager    apifabclient.SigningManager
	FabricProvider    api.FabricProvider
	DiscoveryProvider apifabclient.DiscoveryProvider
	SelectionProvider apifabclient.SelectionProvider
	ChannelProvider   apifabclient.ChannelProvider
}

func newMockProviders(t *testing.T) *mockProviders {
	coreFactory := defcore.NewProviderFactory()
	svcFactory := defsvc.NewProviderFactory()

	config, err := config.FromFile("../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}
	cryptosuite, err := coreFactory.NewCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	stateStore, err := coreFactory.NewStateStoreProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	signer, err := coreFactory.NewSigningManager(cryptosuite, config)
	if err != nil {
		t.Fatalf("Unexpected error creating signing manager %v", err)
	}

	ctx := fabmocks.NewMockProviderContextCustom(config, cryptosuite, signer)
	fabricProvider, err := coreFactory.NewFabricProvider(ctx)
	if err != nil {
		t.Fatalf("Unexpected error creating fabric provider %v", err)
	}

	dp, err := svcFactory.NewDiscoveryProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating discovery provider %v", err)
	}

	sp, err := svcFactory.NewSelectionProvider(config)
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
	apifabclient.IdentityContext
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

func (s *mockSession) Channel(channelID string) (apifabclient.Channel, error) {
	if s.IsChError {
		return nil, errors.New("error")
	}
	return nil, nil
}

func (s *mockSession) EventHub(channelID string) (apifabclient.EventHub, error) {
	if s.IsEHError {
		return nil, errors.New("error")
	}
	return nil, nil
}
