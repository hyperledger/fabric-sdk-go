/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"os"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/def/factory/defclient"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

func defPkgSuite() SDKOption {
	pkgSuite := apisdk.PkgSuite{
		Core:    defcore.NewProviderFactory(),
		Service: defsvc.NewProviderFactory(),
		Context: defclient.NewOrgClientFactory(),
		Session: defclient.NewSessionClientFactory(),
		Logger:  deflogger.LoggerProvider(),
	}
	return PkgSuiteAsOpt(pkgSuite)
}

func TestNewGoodOpt(t *testing.T) {
	_, err := New(ConfigFile("../../test/fixtures/config/config_test.yaml"), goodOpt(), defPkgSuite())
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}
}

func goodOpt() SDKOption {
	return func(sdk *FabricSDK) (*FabricSDK, error) {
		return sdk, nil
	}
}

func TestNewBadOpt(t *testing.T) {
	_, err := New(ConfigFile("../../test/fixtures/config/config_test.yaml"), badOpt(), defPkgSuite())
	if err == nil {
		t.Fatalf("Expected error from New")
	}
}

func badOpt() SDKOption {
	return func(sdk *FabricSDK) (*FabricSDK, error) {
		return sdk, errors.New("Bad Opt")
	}
}
func TestNewDefaultSDK(t *testing.T) {
	// Test new SDK with invalid config file
	_, err := New(ConfigFile("../../test/fixtures/config/invalid.yaml"), StateStorePath("/tmp/state"), defPkgSuite())
	if err == nil {
		t.Fatalf("Should have failed for invalid config file")
	}

	// Test New SDK with valid config file
	sdk, err := New(ConfigFile("../../test/fixtures/config/config_test.yaml"), StateStorePath("/tmp/state"), defPkgSuite())
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Default channel client (uses organisation from client configuration)
	_, err = sdk.NewChannelClient("mychannel", "User1")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Test configuration failure for channel client (mychannel does't have event source configured for Org2)
	_, err = sdk.NewChannelClientWithOpts("mychannel", "User1", &ChannelClientOpts{OrgName: "Org2"})
	if err == nil {
		t.Fatalf("Should have failed to create channel client since event source not configured for Org2")
	}

	// Test new channel client with options
	_, err = sdk.NewChannelClientWithOpts("orgchannel", "User1", &ChannelClientOpts{OrgName: "Org2"})
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

}

func TestNewChannelMgmtClient(t *testing.T) {

	sdk, err := New(ConfigFile("../../test/fixtures/config/config_test.yaml"), StateStorePath("/tmp/state"), defPkgSuite())
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Test configuration failure for channel management client (invalid user/default organisation)
	_, err = sdk.NewChannelMgmtClient("Invalid")
	if err == nil {
		t.Fatalf("Should have failed to create channel client due to invalid user")
	}

	// Test valid configuration for channel management client
	_, err = sdk.NewChannelMgmtClient("Admin")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Test configuration failure for new channel management client with options (invalid org)
	_, err = sdk.NewChannelMgmtClientWithOpts("Admin", &ChannelMgmtClientOpts{OrgName: "Invalid"})
	if err == nil {
		t.Fatalf("Should have failed to create channel client due to invalid organisation")
	}

	// Test new channel management client with options (orderer admin configuration)
	_, err = sdk.NewChannelMgmtClientWithOpts("Admin", &ChannelMgmtClientOpts{OrgName: "ordererorg"})
	if err != nil {
		t.Fatalf("Failed to create new channel client with opts: %s", err)
	}

}

func TestNewResourceMgmtClient(t *testing.T) {

	sdk, err := New(ConfigFile("../../test/fixtures/config/config_test.yaml"), StateStorePath("/tmp/state"), defPkgSuite())
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Test configuration failure for resource management client (invalid user/default organisation)
	_, err = sdk.NewResourceMgmtClient("Invalid")
	if err == nil {
		t.Fatalf("Should have failed to create resource management client due to invalid user")
	}

	// Test valid configuration for resource management client
	_, err = sdk.NewResourceMgmtClient("Admin")
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Test configuration failure for new resource management client with options (invalid org)
	_, err = sdk.NewResourceMgmtClientWithOpts("Admin", &ResourceMgmtClientOpts{OrgName: "Invalid"})
	if err == nil {
		t.Fatalf("Should have failed to create resource management client due to invalid organization")
	}

	// Test new resource management client with options (Org2 configuration)
	_, err = sdk.NewResourceMgmtClientWithOpts("Admin", &ResourceMgmtClientOpts{OrgName: "Org2"})
	if err != nil {
		t.Fatalf("Failed to create new resource management client with opts: %s", err)
	}
}

func TestNewDefaultTwoValidSDK(t *testing.T) {
	sdk1, err := New(ConfigFile("../../test/fixtures/config/config_test.yaml"), StateStorePath("/tmp/state"), defPkgSuite())
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	sdk2, err := New(ConfigFile("./testdata/test.yaml"), StateStorePath("/tmp/state"), defPkgSuite())
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Default sdk with two channels
	client1, err := sdk1.configProvider.Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client1.Organization != "Org1" {
		t.Fatalf("Unexpected org in config: %s", client1.Organization)
	}

	client2, err := sdk2.configProvider.Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client2.Organization != "Org2" {
		t.Fatalf("Unexpected org in config: %s", client1.Organization)
	}

	// Test SDK1 channel clients ('mychannel', 'orgchannel')
	_, err = sdk1.NewChannelClient("mychannel", "User1")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	_, err = sdk1.NewChannelClient("orgchannel", "User1")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// SDK 2 doesn't have 'mychannel' configured
	_, err = sdk2.NewChannelClient("mychannel", "User1")
	if err == nil {
		t.Fatalf("Should have failed to create channel that is not configured")
	}

	// SDK 2 has 'orgchannel' configured
	_, err = sdk2.NewChannelClient("orgchannel", "User1")
	if err != nil {
		t.Fatalf("Failed to create new 'orgchannel' channel client: %s", err)
	}
}

func TestNewDefaultSDKFromByte(t *testing.T) {
	cBytes, err := loadConfigBytesFromFile(t, "../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Failed to load sample bytes from File. Error: %s", err)
	}

	sdk, err := New(ConfigBytes(cBytes, "yaml"), StateStorePath("/tmp/state"), defPkgSuite())
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	if sdk == nil {
		t.Fatalf("SDK should not be empty when initialized")
	}

	// new SDK expected to panic due to wrong config type which didn't load the configs
	_, err = New(ConfigBytes(cBytes, "json"), StateStorePath("/tmp/state"), defPkgSuite())
	if err == nil {
		t.Fatalf("NewSDK should have returned error due to bad config")
	}

}

func loadConfigBytesFromFile(t *testing.T, filePath string) ([]byte, error) {
	// read test config file into bytes array
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to read config file. Error: %s", err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("Failed to read config file stat. Error: %s", err)
	}
	s := fi.Size()
	cBytes := make([]byte, s, s)
	n, err := f.Read(cBytes)
	if err != nil {
		t.Fatalf("Failed to read test config for bytes array testing. Error: %s", err)
	}
	if n == 0 {
		t.Fatalf("Failed to read test config for bytes array testing. Mock bytes array is empty")
	}
	return cBytes, err
}
