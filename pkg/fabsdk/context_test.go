/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"testing"

	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
)

const (
	identityOptConfigFile = "testdata/test.yaml"
	identityValidOptUser  = "User1"
	identityValidOptOrg   = "Org2"
)

func TestWithUserValid(t *testing.T) {
	sdk, err := New(configImpl.FromFile(identityOptConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}
	defer sdk.Close()

	opts := identityOptions{}
	opt := WithUser(identityValidOptUser)
	err = opt(&opts)
	if err != nil {
		t.Fatalf("Expected no error from opt, but got %v", err)
	}
}

func TestWithIdentity(t *testing.T) {
	sdk, err := New(configImpl.FromFile(identityOptConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}
	defer sdk.Close()

	identityManager, ok := sdk.provider.IdentityManager(identityValidOptOrg)
	if !ok {
		t.Fatalf("Invalid organization: %s", identityValidOptOrg)
	}
	identity, err := identityManager.GetSigningIdentity(identityValidOptUser)
	if err != nil {
		t.Fatalf("Unexpected error loading identity: %v", err)
	}

	opts := identityOptions{}
	opt := WithIdentity(identity)
	err = opt(&opts)
	if err != nil {
		t.Fatalf("Expected no error from opt, but got %v", err)
	}
	if opts.signingIdentity != identity {
		t.Fatal("Expected identity to be populated")
	}
}

func TestFabricSDKContext(t *testing.T) {

	sdk, err := New(configImpl.FromFile(identityOptConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}
	defer sdk.Close()

	ctxProvider := sdk.Context()

	if ctxProvider == nil {
		t.Fatal("context client supposed to be not empty")
	}

	// Anonymous
	ctx, err := ctxProvider()

	if err != nil {
		t.Fatalf("expected to create anonymous context, err: %v", err)
	}

	if ctx == nil {
		t.Fatal("context client will have providers even with anonymous context")
	}

	// Invalid user, invalid org
	ctxProvider = sdk.Context(WithUser("INVALID_USER"), WithOrg("INVALID_ORG_NAME"))

	ctx, err = ctxProvider()

	if err == nil || err.Error() != "invalid options to create identity, invalid org name" {
		t.Fatalf("getting context client supposed to fail with idenity error, err: %v", err)
	}

	if ctx == nil {
		t.Fatal("context client will have providers even if idenity fails")
	}

	// Valid user, invalid org
	ctxProvider = sdk.Context(WithUser(identityValidOptUser), WithOrg("INVALID_ORG_NAME"))

	ctx, err = ctxProvider()

	if err == nil || err.Error() != "invalid options to create identity, invalid org name" {
		t.Fatalf("getting context client supposed to fail with idenity error, err: %v", err)
	}

	if ctx == nil {
		t.Fatal("context client will have providers even if idenity fails")
	}

	// Valid user and org
	ctxProvider = sdk.Context(WithUser(identityValidOptUser), WithOrg(identityValidOptOrg))

	_, err = ctxProvider()
	if err != nil {
		t.Fatalf("getting context supposed to succeed")
	}

	ctxProvider = sdk.Context(WithUser(identityValidOptUser))

	ctx, err = ctxProvider()

	if err != nil {
		t.Fatalf("getting context supposed to succeed")
	}

	if ctx == nil || ctx.Identifier().MSPID == "" {
		t.Fatalf("supposed to get valid context")
	}

}
