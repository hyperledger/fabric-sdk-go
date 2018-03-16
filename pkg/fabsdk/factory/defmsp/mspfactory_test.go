/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defmsp

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	mockCore "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	mspimpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
)

func TestCreateUserStore(t *testing.T) {
	factory := NewProviderFactory()

	config := mocks.NewMockConfig()

	userStore, err := factory.CreateUserStore(config)
	if err != nil {
		t.Fatalf("Unexpected error creating state store %v", err)
	}

	_, ok := userStore.(*mspimpl.CertFileUserStore)
	if !ok {
		t.Fatalf("Unexpected state store created")
	}
}

func newMockUserStore(t *testing.T) msp.UserStore {
	factory := NewProviderFactory()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mockCore.NewMockConfig(mockCtrl)

	mockClientConfig := core.ClientConfig{
		CredentialStore: core.CredentialStoreType{
			Path: "/tmp/fabsdkgo_test/store",
		},
	}
	mockConfig.EXPECT().Client().Return(&mockClientConfig, nil)

	userStore, err := factory.CreateUserStore(mockConfig)
	if err != nil {
		t.Fatalf("Unexpected error creating user store %v", err)
	}
	return userStore
}
func TestCreateUserStoreByConfig(t *testing.T) {
	userStore := newMockUserStore(t)

	_, ok := userStore.(*mspimpl.CertFileUserStore)
	if !ok {
		t.Fatalf("Unexpected user store created")
	}
}

func TestCreateUserStoreEmptyConfig(t *testing.T) {
	factory := NewProviderFactory()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mockCore.NewMockConfig(mockCtrl)

	mockClientConfig := core.ClientConfig{}
	mockConfig.EXPECT().Client().Return(&mockClientConfig, nil)

	_, err := factory.CreateUserStore(mockConfig)
	if err == nil {
		t.Fatal("Expected error creating user store")
	}
}

func TestCreateUserStoreFailConfig(t *testing.T) {
	factory := NewProviderFactory()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mockCore.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Client().Return(nil, errors.New("error"))

	_, err := factory.CreateUserStore(mockConfig)
	if err == nil {
		t.Fatal("Expected error creating user store")
	}
}

func TestCreateIdentityManager(t *testing.T) {

	coreFactory := defcore.NewProviderFactory()

	config, err := config.FromFile("../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	cryptosuite, err := coreFactory.CreateCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	factory := NewProviderFactory()
	userStore, err := factory.CreateUserStore(config)
	if err != nil {
		t.Fatalf("Unexpected error creating user store %v", err)
	}

	provider, err := factory.CreateIdentityManagerProvider(config, cryptosuite, userStore)
	if err != nil {
		t.Fatalf("Unexpected error creating provider %v", err)
	}

	mgr, ok := provider.IdentityManager("Org1")
	if !ok {
		t.Fatalf("Unexpected error creating identity manager %v", err)
	}

	_, ok = mgr.(msp.IdentityManager)
	if !ok {
		t.Fatalf("Unexpected signing manager created")
	}
}
