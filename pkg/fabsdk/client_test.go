/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

const (
	clientConfigFile     = "testdata/test.yaml"
	clientValidAdmin     = "Admin"
	clientValidUser      = "User1"
	clientValidExtraOrg  = "OrgX"
	clientValidExtraUser = "OrgXUser"
)

func TestNewGoodClientOpt(t *testing.T) {
	c, err := configImpl.FromFile(clientConfigFile)
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser(clientValidUser), goodClientOpt())
	if err != nil {
		t.Fatalf("Expected no error from Client, but got %v", err)
	}
}

func goodClientOpt() ClientOption {
	return func(opts *clientOptions) error {
		return nil
	}
}

func TestNewBadClientOpt(t *testing.T) {
	c, err := configImpl.FromFile(clientConfigFile)
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser(clientValidUser), badClientOpt())
	if err == nil {
		t.Fatal("Expected error from Client")
	}
}

func badClientOpt() ClientOption {
	return func(opts *clientOptions) error {
		return errors.New("Bad Opt")
	}
}

func TestClient(t *testing.T) {
	c, err := configImpl.FromFile(clientConfigFile)
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser(clientValidUser))
	if err != nil {
		t.Fatalf("Expected no error from Client, but got %v", err)
	}
}

func TestWithOrg(t *testing.T) {
	c, err := configImpl.FromFile(clientConfigFile)
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser("notarealuser"), WithOrg(clientValidExtraOrg))
	if err == nil {
		t.Fatal("Expected error from Client")
	}

	_, err = sdk.NewClient(WithUser(clientValidExtraUser), WithOrg(clientValidExtraOrg))
	if err != nil {
		t.Fatalf("Expected no error from Client, but got %v", err)
	}
}

func TestWithFilter(t *testing.T) {
	tf := mockTargetFilter{}
	opt := WithTargetFilter(&tf)

	opts := clientOptions{}
	err := opt(&opts)
	if err != nil {
		t.Fatalf("Expected no error from option, but got %v", err)
	}

	if opts.targetFilter != &tf {
		t.Fatalf("Expected target filter to be set in opts")
	}
}

func TestWithConfig(t *testing.T) {
	c, err := configImpl.FromFile(clientConfigFile)
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}
	opt := withConfig(c)

	opts := clientOptions{}
	err = opt(&opts)
	if err != nil {
		t.Fatalf("Expected no error from option, but got %v", err)
	}

	if opts.configProvider != c {
		t.Fatalf("Expected config to be set in opts")
	}
}

func TestNoIdentity(t *testing.T) {
	c, err := configImpl.FromFile(clientConfigFile)
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(noopIdentityOpt(), goodClientOpt())
	if err == nil {
		t.Fatal("Expected error from Client")
	}
}

func TestNewChannelMgmtClient(t *testing.T) {
	c, err := configImpl.FromFile("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Test configuration failure for channel management client (invalid user/default organisation)
	_, err = sdk.NewClientChannelMgmt(WithUser("Invalid"))
	if err == nil {
		t.Fatalf("Should have failed to create channel client due to invalid user")
	}

	// Test valid configuration for channel management client
	_, err = sdk.NewClientChannelMgmt(WithUser(clientValidAdmin))
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}
}

func TestNewResourceMgmtClient(t *testing.T) {
	c, err := configImpl.FromFile("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Test configuration failure for resource management client (invalid user/default organisation)
	_, err = sdk.NewClientResourceMgmt(WithUser("Invalid"))
	if err == nil {
		t.Fatalf("Should have failed to create resource management client due to invalid user")
	}

	// Test valid configuration for resource management client
	_, err = sdk.NewClientResourceMgmt(WithUser(clientValidAdmin))
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}
}

func noopIdentityOpt() IdentityOption {
	return func(o *identityOptions, sdk *FabricSDK, orgName string) error {
		return nil
	}
}

type mockTargetFilter struct{}

func (f *mockTargetFilter) Accept(peer apifabclient.Peer) bool {
	return false
}
