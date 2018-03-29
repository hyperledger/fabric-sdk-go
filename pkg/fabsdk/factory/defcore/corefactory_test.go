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
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	_, ok := cryptosuite.(*cryptosuitewrapper.CryptoSuite)
	if !ok {
		t.Fatalf("Unexpected cryptosuite provider created")
	}
}

func TestCreateSigningManager(t *testing.T) {
	factory := NewProviderFactory()
	config := mocks.NewMockCryptoConfig()

	cryptosuite, err := factory.CreateCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	signer, err := factory.CreateSigningManager(cryptosuite)
	if err != nil {
		t.Fatalf("Unexpected error creating signing manager %v", err)
	}

	_, ok := signer.(*signingMgr.SigningManager)
	if !ok {
		t.Fatalf("Unexpected signing manager created")
	}
}

func TestNewFactoryInfraProvider(t *testing.T) {
	factory := NewProviderFactory()
	ctx := mocks.NewMockProviderContext()

	infraProvider, err := factory.CreateInfraProvider(ctx.EndpointConfig())
	if err != nil {
		t.Fatalf("Unexpected error creating fabric provider %v", err)
	}

	_, ok := infraProvider.(*fabpvdr.InfraProvider)
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
