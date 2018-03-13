/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msppvdr

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/mocks"
)

func TestCreateMSPProvider(t *testing.T) {

	coreFactory := defcore.NewProviderFactory()

	config, err := config.FromFile("../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	cryptosuite, err := coreFactory.CreateCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	userStore := &mocks.MockUserStore{}

	provider, err := New(config, cryptosuite, userStore)

	if provider.UserStore() != userStore {
		t.Fatalf("Invalid user store returned")
	}

	mgr, ok := provider.IdentityManager("Org1")
	if !ok {
		t.Fatalf("Expected to return idnetity manager")
	}

	_, ok = mgr.(*msp.IdentityManager)
	if !ok {
		t.Fatalf("Unexpected signing manager created")
	}
}
