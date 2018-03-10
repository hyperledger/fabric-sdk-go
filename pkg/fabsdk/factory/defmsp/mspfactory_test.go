/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defmsp

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
)

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

	stateStore, err := kvs.New(
		&kvs.FileKeyValueStoreOptions{
			Path: config.CredentialStorePath(),
		})
	if err != nil {
		t.Fatalf("creating a user store failed: %v", err)
	}

	factory := NewProviderFactory()

	mgr, err := factory.CreateIdentityManager("Org1", stateStore, cryptosuite, config)
	if err != nil {
		t.Fatalf("Unexpected error creating signing manager %v", err)
	}

	_, ok := mgr.(*msp.IdentityManager)
	if !ok {
		t.Fatalf("Unexpected signing manager created")
	}
}
