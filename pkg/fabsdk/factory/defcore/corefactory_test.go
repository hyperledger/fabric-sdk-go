/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defcore

import (
	"testing"

	cryptosuitewrapper "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/wrapper"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/modlog"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	signingMgr "github.com/hyperledger/fabric-sdk-go/pkg/fab/signingmgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
)

func TestCreateCryptoSuiteProvider(t *testing.T) {
	factory := NewProviderFactory()
	config := mocks.NewMockCryptoConfig()

	cryptosuite, err := factory.CreateCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %s", err)
	}

	_, ok := cryptosuite.(*cryptosuitewrapper.CryptoSuite)
	if !ok {
		t.Fatal("Unexpected cryptosuite provider created")
	}
}

func TestCreateSigningManager(t *testing.T) {
	factory := NewProviderFactory()
	config := mocks.NewMockCryptoConfig()

	cryptosuite, err := factory.CreateCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %s", err)
	}

	signer, err := factory.CreateSigningManager(cryptosuite)
	if err != nil {
		t.Fatalf("Unexpected error creating signing manager %s", err)
	}

	_, ok := signer.(*signingMgr.SigningManager)
	if !ok {
		t.Fatal("Unexpected signing manager created")
	}
}

func TestNewFactoryInfraProvider(t *testing.T) {
	factory := NewProviderFactory()
	ctx := mocks.NewMockProviderContext()

	infraProvider, err := factory.CreateInfraProvider(ctx.EndpointConfig())
	if err != nil {
		t.Fatalf("Unexpected error creating fabric provider %s", err)
	}

	_, ok := infraProvider.(*fabpvdr.InfraProvider)
	if !ok {
		t.Fatal("Unexpected fabric provider created")
	}
}

func TestNewLoggingProvider(t *testing.T) {
	logger := NewLoggerProvider()

	_, ok := logger.(*modlog.Provider)
	if !ok {
		t.Fatal("Unexpected logger provider created")
	}
}
