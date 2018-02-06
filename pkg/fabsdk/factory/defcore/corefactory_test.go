/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defcore

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	"github.com/hyperledger/fabric-sdk-go/api/kvstore"
	cryptosuitewrapper "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/wrapper"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	signingMgr "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/signingmgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/modlog"
)

func TestNewStateStoreProvider(t *testing.T) {
	factory := NewProviderFactory()

	config := mocks.NewMockConfig()

	stateStore, err := factory.NewStateStoreProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating state store provider %v", err)
	}

	_, ok := stateStore.(*kvs.FileKeyValueStore)
	if !ok {
		t.Fatalf("Unexpected state store provider created")
	}
}

func newMockStateStore(t *testing.T) kvstore.KVStore {
	factory := NewProviderFactory()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)

	mockClientConfig := apiconfig.ClientConfig{
		CredentialStore: apiconfig.CredentialStoreType{
			Path: "/tmp/fabsdkgo_test/store",
		},
	}
	mockConfig.EXPECT().Client().Return(&mockClientConfig, nil)

	stateStore, err := factory.NewStateStoreProvider(mockConfig)
	if err != nil {
		t.Fatalf("Unexpected error creating state store provider %v", err)
	}
	return stateStore
}
func TestNewStateStoreProviderByConfig(t *testing.T) {
	stateStore := newMockStateStore(t)

	_, ok := stateStore.(*kvs.FileKeyValueStore)
	if !ok {
		t.Fatalf("Unexpected state store provider created")
	}
}

func TestNewStateStoreProviderEmptyConfig(t *testing.T) {
	factory := NewProviderFactory()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)

	mockClientConfig := apiconfig.ClientConfig{}
	mockConfig.EXPECT().Client().Return(&mockClientConfig, nil)

	_, err := factory.NewStateStoreProvider(mockConfig)
	if err == nil {
		t.Fatal("Expected error creating state store provider")
	}
}

func TestNewStateStoreProviderFailConfig(t *testing.T) {
	factory := NewProviderFactory()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Client().Return(nil, errors.New("error"))

	_, err := factory.NewStateStoreProvider(mockConfig)
	if err == nil {
		t.Fatal("Expected error creating state store provider")
	}
}

func TestNewCryptoSuiteProvider(t *testing.T) {
	factory := NewProviderFactory()
	config := mocks.NewMockConfig()

	cryptosuite, err := factory.NewCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	_, ok := cryptosuite.(*cryptosuitewrapper.CryptoSuite)
	if !ok {
		t.Fatalf("Unexpected cryptosuite provider created")
	}
}

func TestNewSigningManager(t *testing.T) {
	factory := NewProviderFactory()
	config := mocks.NewMockConfig()

	cryptosuite, err := factory.NewCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	signer, err := factory.NewSigningManager(cryptosuite, config)
	if err != nil {
		t.Fatalf("Unexpected error creating signing manager %v", err)
	}

	_, ok := signer.(*signingMgr.SigningManager)
	if !ok {
		t.Fatalf("Unexpected signing manager created")
	}
}

func TestNewFactoryFabricProvider(t *testing.T) {
	factory := NewProviderFactory()
	ctx := mocks.NewMockProviderContext()

	fabricProvider, err := factory.NewFabricProvider(ctx)
	if err != nil {
		t.Fatalf("Unexpected error creating fabric provider %v", err)
	}

	_, ok := fabricProvider.(*fabpvdr.FabricProvider)
	if !ok {
		t.Fatalf("Unexpected fabric provider created")
	}
}

func TestNewLoggingProvider(t *testing.T) {
	logger := NewLoggerProvider()

	_, ok := logger.(*modlog.Provider)
	if !ok {
		t.Fatalf("Unexpected logger provider created")
	}
}
