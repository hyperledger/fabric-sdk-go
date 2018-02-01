/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/pkg/errors"
)

const (
	clientConfigFile     = "testdata/test.yaml"
	clientValidAdmin     = "Admin"
	clientValidUser      = "User1"
	clientValidExtraOrg  = "OrgX"
	clientValidExtraUser = "OrgXUser"
)

func TestNewGoodClientOpt(t *testing.T) {
	sdk, err := New(configImpl.FromFile(clientConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser(clientValidUser), goodClientOpt()).ChannelMgmt()
	if err != nil {
		t.Fatalf("Expected no error from Client, but got %v", err)
	}
}

func TestFromConfigGoodClientOpt(t *testing.T) {
	c, err := configImpl.FromFile(clientConfigFile)()
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(WithConfig(c))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser(clientValidUser), goodClientOpt()).ChannelMgmt()
	if err != nil {
		t.Fatalf("Expected no error from Client, but got %v", err)
	}
}

func goodClientOpt() ContextOption {
	return func(opts *contextOptions) error {
		return nil
	}
}

func TestNewBadClientOpt(t *testing.T) {
	sdk, err := New(configImpl.FromFile(clientConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser(clientValidUser), badClientOpt()).ChannelMgmt()
	if err == nil {
		t.Fatal("Expected error from Client")
	}
}

func badClientOpt() ContextOption {
	return func(opts *contextOptions) error {
		return errors.New("Bad Opt")
	}
}

func TestClient(t *testing.T) {
	sdk, err := New(configImpl.FromFile(clientConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser(clientValidUser)).ChannelMgmt()
	if err != nil {
		t.Fatalf("Expected no error from Client, but got %v", err)
	}
}

func TestWithOrg(t *testing.T) {
	sdk, err := New(configImpl.FromFile(clientConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser("notarealuser"), WithOrg(clientValidExtraOrg)).ChannelMgmt()
	if err == nil {
		t.Fatal("Expected error from Client")
	}

	_, err = sdk.NewClient(WithUser(clientValidExtraUser), WithOrg(clientValidExtraOrg)).ChannelMgmt()
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
	c, err := configImpl.FromFile(clientConfigFile)()
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}
	opt := withConfig(c)

	opts := contextOptions{}
	err = opt(&opts)
	if err != nil {
		t.Fatalf("Expected no error from option, but got %v", err)
	}

	if opts.config != c {
		t.Fatalf("Expected config to be set in opts")
	}
}

func TestNoIdentity(t *testing.T) {
	sdk, err := New(configImpl.FromFile(clientConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(noopIdentityOpt(), goodClientOpt()).ChannelMgmt()
	if err == nil {
		t.Fatal("Expected error from Client")
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
