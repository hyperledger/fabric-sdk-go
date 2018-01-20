/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"testing"

	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
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

	opts := identityOptions{}
	opt := WithUser(identityValidOptUser)
	err = opt(&opts, sdk, identityValidOptOrg)
	if err != nil {
		t.Fatalf("Expected no error from opt, but got %v", err)
	}
	if opts.identity == nil {
		t.Fatal("Expected identity to be populated")
	}
}

func TestWithUserInvalid(t *testing.T) {
	sdk, err := New(configImpl.FromFile(identityOptConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	opts := identityOptions{}
	opt := WithUser("notarealuser")
	err = opt(&opts, sdk, identityValidOptOrg)
	if err == nil {
		t.Fatal("Expected error from opt")
	}
	if opts.identity != nil {
		t.Fatal("Expected identity to not be populated")
	}
}

func TestWithIdentity(t *testing.T) {
	sdk, err := New(configImpl.FromFile(identityOptConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	identity, err := sdk.newUser(identityValidOptOrg, identityValidOptUser)
	if err != nil {
		t.Fatalf("Unexpected error loading identity: %v", err)
	}

	opts := identityOptions{}
	opt := WithIdentity(identity)
	err = opt(&opts, sdk, identityValidOptOrg)
	if err != nil {
		t.Fatalf("Expected no error from opt, but got %v", err)
	}
	if opts.identity != identity {
		t.Fatal("Expected identity to be populated")
	}
}
