/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defcore

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core/mocks"
	cryptosuitewrapper "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/wrapper"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	signingMgr "github.com/hyperledger/fabric-sdk-go/pkg/fab/signingmgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/modlog"
)

func TestCreateStateStoreProvider(t *testing.T) {
	factory := NewProviderFactory()

	config := mocks.NewMockConfig()

	stateStore, err := factory.CreateStateStoreProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating state store provider %v", err)
	}

	_, ok := stateStore.(*kvs.FileKeyValueStore)
	if !ok {
		t.Fatalf("Unexpected state store provider created")
	}
}

func newMockStateStore(t *testing.T) api.KVStore {
	factory := NewProviderFactory()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)

	mockClientConfig := core.ClientConfig{
		CredentialStore: core.CredentialStoreType{
			Path: "/tmp/fabsdkgo_test/store",
		},
	}
	mockConfig.EXPECT().Client().Return(&mockClientConfig, nil)

	stateStore, err := factory.CreateStateStoreProvider(mockConfig)
	if err != nil {
		t.Fatalf("Unexpected error creating state store provider %v", err)
	}
	return stateStore
}
func TestCreateStateStoreProviderByConfig(t *testing.T) {
	stateStore := newMockStateStore(t)

	_, ok := stateStore.(*kvs.FileKeyValueStore)
	if !ok {
		t.Fatalf("Unexpected state store provider created")
	}
}

func TestCreateStateStoreProviderEmptyConfig(t *testing.T) {
	factory := NewProviderFactory()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)

	mockClientConfig := core.ClientConfig{}
	mockConfig.EXPECT().Client().Return(&mockClientConfig, nil)

	_, err := factory.CreateStateStoreProvider(mockConfig)
	if err == nil {
		t.Fatal("Expected error creating state store provider")
	}
}

func TestCreateStateStoreProviderFailConfig(t *testing.T) {
	factory := NewProviderFactory()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Client().Return(nil, errors.New("error"))

	_, err := factory.CreateStateStoreProvider(mockConfig)
	if err == nil {
		t.Fatal("Expected error creating state store provider")
	}
}

func TestCreateCryptoSuiteProvider(t *testing.T) {
	factory := NewProviderFactory()
	config := mocks.NewMockConfig()

	cryptosuite, err := factory.CreateCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	_, ok := cryptosuite.(*cryptosuitewrapper.CryptoSuite)
	if !ok {
		t.Fatalf("Unexpected cryptosuite provider created")
	}
}

func TestCreateSigningManager(t *testing.T) {
	factory := NewProviderFactory()
	config := mocks.NewMockConfig()

	cryptosuite, err := factory.CreateCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	signer, err := factory.CreateSigningManager(cryptosuite, config)
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

	fabricProvider, err := factory.CreateFabricProvider(ctx)
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
