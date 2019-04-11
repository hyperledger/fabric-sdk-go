/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

const (
	identityOptConfigFile = "config_test.yaml"
	identityValidOptUser  = "User1"
	identityValidOptOrg   = "Org2"
)

func TestWithUserValid(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, identityOptConfigFile)
	sdk, err := New(config.FromFile(configPath))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %s", err)
	}
	defer sdk.Close()

	opts := identityOptions{}
	opt := WithUser(identityValidOptUser)
	err = opt(&opts)
	if err != nil {
		t.Fatalf("Expected no error from opt, but got %s", err)
	}
}

func TestWithIdentity(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, identityOptConfigFile)
	sdk, err := New(config.FromFile(configPath))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %s", err)
	}
	defer sdk.Close()

	identityManager, ok := sdk.provider.IdentityManager(identityValidOptOrg)
	if !ok {
		t.Fatalf("Invalid organization: %s", identityValidOptOrg)
	}
	identity, err := identityManager.GetSigningIdentity(identityValidOptUser)
	if err != nil {
		t.Fatalf("Unexpected error loading identity: %s", err)
	}

	opts := identityOptions{}
	opt := WithIdentity(identity)
	err = opt(&opts)
	if err != nil {
		t.Fatalf("Expected no error from opt, but got %s", err)
	}
	if opts.signingIdentity != identity {
		t.Fatal("Expected identity to be populated")
	}
}

func TestFabricSDKContext(t *testing.T) {

	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, identityOptConfigFile)
	sdk, err := New(config.FromFile(configPath))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %s", err)
	}
	defer sdk.Close()

	ctxProvider := sdk.Context()

	if ctxProvider == nil {
		t.Fatal("context client supposed to be not empty")
	}

	// Anonymous
	ctx, err := ctxProvider()

	if err != nil {
		t.Fatalf("expected to create anonymous context, err: %s", err)
	}

	if ctx == nil {
		t.Fatal("context client will have providers even with anonymous context")
	}

	// Invalid user, invalid org
	ctxProvider = sdk.Context(WithUser("INVALID_USER"), WithOrg("INVALID_ORG_NAME"))

	ctx, err = ctxProvider()

	if err == nil || err.Error() != "invalid options to create identity, invalid org name" {
		t.Fatalf("getting context client supposed to fail with idenity error, err: %s", err)
	}

	if ctx == nil {
		t.Fatal("context client will have providers even if idenity fails")
	}

	// Valid user, invalid org
	checkValidUserAndInvalidOrg(sdk, t)

	// Valid user and org
	checkValidUserAndOrg(sdk, t)

}

func checkValidUserAndInvalidOrg(sdk *FabricSDK, t *testing.T) {
	ctxProvider := sdk.Context(WithUser(identityValidOptUser), WithOrg("INVALID_ORG_NAME"))
	ctx, err := ctxProvider()
	if err == nil || err.Error() != "invalid options to create identity, invalid org name" {
		t.Fatalf("getting context client supposed to fail with idenity error, err: %s", err)
	}
	if ctx == nil {
		t.Fatal("context client will have providers even if idenity fails")
	}
}

func checkValidUserAndOrg(sdk *FabricSDK, t *testing.T) {
	ctxProvider := sdk.Context(WithUser(identityValidOptUser), WithOrg(identityValidOptOrg))
	_, err := ctxProvider()
	if err != nil {
		t.Fatal("getting context supposed to succeed")
	}
	ctxProvider = sdk.Context(WithUser(identityValidOptUser))
	ctx, err := ctxProvider()
	if err != nil {
		t.Fatal("getting context supposed to succeed")
	}
	if ctx == nil || ctx.Identifier().MSPID == "" {
		t.Fatal("supposed to get valid context")
	}
}
