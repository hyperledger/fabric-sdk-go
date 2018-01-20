/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/def/factory/defsvc"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicore"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	chImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chclient"
	chmgmtImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chmgmtclient"
	resmgmtImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/resmgmtclient"
	mockapisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api/mocks"
)

func TestNewChannelMgmtClient(t *testing.T) {
	p := newMockProviders(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSDK := mockapisdk.NewMockProviders(mockCtrl)

	mockSDK.EXPECT().FabricProvider().Return(p.FabricProvider)

	factory := NewSessionClientFactory()
	session := newMockSession()

	client, err := factory.NewChannelMgmtClient(mockSDK, session, p.ConfigProvider)
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

	mockSDK.EXPECT().FabricProvider().Return(p.FabricProvider)
	mockSDK.EXPECT().DiscoveryProvider().Return(p.DiscoveryProvider)

	factory := NewSessionClientFactory()
	session := newMockSession()

	client, err := factory.NewResourceMgmtClient(mockSDK, session, p.ConfigProvider, nil)
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

	mockSDK.EXPECT().ConfigProvider().Return(p.ConfigProvider)
	mockSDK.EXPECT().CryptoSuiteProvider().Return(p.CryptosuiteProvider)
	mockSDK.EXPECT().StateStoreProvider().Return(p.StateStoreProvider)
	mockSDK.EXPECT().SigningManager().Return(p.SigningManager)
	mockSDK.EXPECT().DiscoveryProvider().Return(p.DiscoveryProvider)
	mockSDK.EXPECT().SelectionProvider().Return(p.SelectionProvider)

	factory := NewSessionClientFactory()
	session := newMockSession()

	client, err := factory.NewChannelClient(mockSDK, session, p.ConfigProvider, "mychannel")
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

	mockSDK.EXPECT().ConfigProvider().Return(p.ConfigProvider)
	mockSDK.EXPECT().CryptoSuiteProvider().Return(p.CryptosuiteProvider)
	mockSDK.EXPECT().StateStoreProvider().Return(p.StateStoreProvider)
	mockSDK.EXPECT().SigningManager().Return(p.SigningManager)

	factory := NewSessionClientFactory()
	session := newMockSession()

	_, err := factory.NewChannelClient(mockSDK, session, p.ConfigProvider, "badchannel")
	if err == nil {
		t.Fatalf("Expected error creating channel client")
	}
}

func TestNewChannelClientBadOrg(t *testing.T) {
	p := newMockProviders(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSDK := mockapisdk.NewMockProviders(mockCtrl)

	mockSDK.EXPECT().ConfigProvider().Return(p.ConfigProvider)
	mockSDK.EXPECT().CryptoSuiteProvider().Return(p.CryptosuiteProvider)
	mockSDK.EXPECT().StateStoreProvider().Return(p.StateStoreProvider)
	mockSDK.EXPECT().SigningManager().Return(p.SigningManager)
	mockSDK.EXPECT().DiscoveryProvider().Return(p.DiscoveryProvider)
	mockSDK.EXPECT().SelectionProvider().Return(p.SelectionProvider)

	factory := NewSessionClientFactory()
	session := newMockSessionWithUser("user1", "BadOrg")

	_, err := factory.NewChannelClient(mockSDK, session, p.ConfigProvider, "mychannel")
	if err == nil {
		t.Fatalf("Expected error creating channel client")
	}
}

func getChannelMock(client apifabclient.Resource, channelID string) (apifabclient.Channel, error) {
	return channel.NewChannel("channel", client)
}

type mockProviders struct {
	CryptosuiteProvider apicryptosuite.CryptoSuite
	StateStoreProvider  apifabclient.KeyValueStore
	ConfigProvider      apiconfig.Config
	SigningManager      apifabclient.SigningManager
	FabricProvider      apicore.FabricProvider
	DiscoveryProvider   apifabclient.DiscoveryProvider
	SelectionProvider   apifabclient.SelectionProvider
}

func newMockProviders(t *testing.T) *mockProviders {
	coreFactory := defcore.NewProviderFactory()
	svcFactory := defsvc.NewProviderFactory()

	config, err := config.FromFile("../../../test/fixtures/config/config_test.yaml")()
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

	fabricProvider, err := coreFactory.NewFabricProvider(config, stateStore, cryptosuite, signer)
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

	providers := mockProviders{
		CryptosuiteProvider: cryptosuite,
		StateStoreProvider:  stateStore,
		ConfigProvider:      config,
		SigningManager:      signer,
		FabricProvider:      fabricProvider,
		DiscoveryProvider:   dp,
		SelectionProvider:   sp,
	}
	return &providers
}

type mockSession struct {
	user apifabclient.IdentityContext
}

func newMockSession() *mockSession {
	return newMockSessionWithUser("user1", "Org1MSP")
}

func newMockSessionWithUser(username, mspID string) *mockSession {
	user := fabmocks.NewMockUserWithMSPID(username, mspID)
	session := mockSession{
		user,
	}
	return &session
}

func (s *mockSession) Identity() apifabclient.IdentityContext {
	return s.user
}
