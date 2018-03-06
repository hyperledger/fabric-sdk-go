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

	identityManager, ok := sdk.Context().IdentityManager(identityValidOptOrg)
	if !ok {
		t.Fatalf("Invalid organization: %s", identityValidOptOrg)
	}
	identity, err := identityManager.GetUser(identityValidOptUser)
	if err != nil {
		t.Fatalf("Unexpected error loading identity: %v", err)
	}

	opts := identityOptions{}
	opt := WithIdentity(identity)
	err = opt(&opts)
	if err != nil {
		t.Fatalf("Expected no error from opt, but got %v", err)
	}
	if opts.identity != identity {
		t.Fatal("Expected identity to be populated")
	}
}

func TestFabricSDKContext(t *testing.T) {

	sdk, err := New(configImpl.FromFile(identityOptConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}
	defer sdk.Close()

	client := sdk.Context()

	if client == nil {
		t.Fatal("context client supposed to be not empty")
	}

	client = sdk.Context(WithUser("INVALID_USER"), WithOrgName("INVALID_ORG_NAME"))

	if client == nil {
		t.Fatal("context client supposed to be not empty, even with invalid identity options")
	}

}
