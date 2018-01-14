/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	mockapisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api/mocks"
)

func TestPanicOnNilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Passing nil configuration was supposed to panic")
		}
	}()

	New(nil)
}

func TestNewGoodOpt(t *testing.T) {
	c, err := configImpl.FromFile("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	_, err = New(c, goodOpt())
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}
}

func goodOpt() Option {
	return func(opts *options) error {
		return nil
	}
}

func TestNewBadOpt(t *testing.T) {
	c, err := configImpl.FromFile("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	_, err = New(c, badOpt())
	if err == nil {
		t.Fatalf("Expected error from New")
	}
}

func badOpt() Option {
	return func(opts *options) error {
		return errors.New("Bad Opt")
	}
}
func TestNewDefaultSDK(t *testing.T) {
	// Test New SDK with valid config file
	c, err := configImpl.FromFile("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c)
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

func TestWithCorePkg(t *testing.T) {
	// Test New SDK with valid config file
	c, err := configImpl.FromFile("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	_, err = New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockCoreProviderFactory(mockCtrl)

	factory.EXPECT().NewCryptoSuiteProvider(c).Return(nil, nil)
	factory.EXPECT().NewStateStoreProvider(c).Return(nil, nil)
	factory.EXPECT().NewSigningManager(nil, c).Return(nil, nil)
	factory.EXPECT().NewFabricProvider(c, nil, nil, nil).Return(nil, nil)

	_, err = New(c, WithCorePkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
}

func TestWithServicePkg(t *testing.T) {
	// Test New SDK with valid config file
	c, err := configImpl.FromFile("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	_, err = New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockServiceProviderFactory(mockCtrl)

	factory.EXPECT().NewDiscoveryProvider(c).Return(nil, nil)
	factory.EXPECT().NewSelectionProvider(c).Return(nil, nil)

	_, err = New(c, WithServicePkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
}

func TestWithContextPkg(t *testing.T) {
	// Test New SDK with valid config file
	c, err := configImpl.FromFile("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	core, err := newMockCorePkg(c)
	if err != nil {
		t.Fatalf("Error initializing core factory: %s", err)
	}

	_, err = New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Use real implementation of credential manager to provide in later response
	pkgSuite := defPkgSuite{}
	ctx, err := pkgSuite.Context()
	if err != nil {
		t.Fatalf("Unexpected error getting context: %s", err)
	}

	cm, err := ctx.NewCredentialManager("Org1", c, core.cryptoSuite)
	if err != nil {
		t.Fatalf("Unexpected error getting credential manager: %s", err)
	}

	// Create mock to ensure the provided factory is called.
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockOrgClientFactory(mockCtrl)

	factory.EXPECT().NewCredentialManager("Org1", c, core.cryptoSuite).Return(cm, nil)

	sdk, err := New(c, WithCorePkg(core), WithContextPkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Use a method that invokes credential manager (e.g., new user)
	_, err = sdk.NewPreEnrolledUser("Org1", "User1")
	if err != nil {
		t.Fatalf("Unexpected error getting user: %s", err)
	}
}

func TestWithSessionPkg(t *testing.T) {
	// Test New SDK with valid config file
	c, err := configImpl.FromFile("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	core, err := newMockCorePkg(c)
	if err != nil {
		t.Fatalf("Error initializing core factory: %s", err)
	}

	_, err = New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Create mock to ensure the provided factory is called.
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockSessionClientFactory(mockCtrl)

	sdk, err := New(c, WithCorePkg(core), WithSessionPkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Use real implementation of credential manager to provide in later response
	pkgSuite := defPkgSuite{}
	sessPkg, err := pkgSuite.Session()
	if err != nil {
		t.Fatalf("Unexpected error getting context: %s", err)
	}

	session, err := sdk.NewPreEnrolledUserSession("Org1", "User1")
	if err != nil {
		t.Fatalf("Unexpected error getting session: %s", err)
	}

	cm, err := sessPkg.NewChannelMgmtClient(sdk, session, c)
	if err != nil {
		t.Fatalf("Unexpected error getting credential manager: %s", err)
	}
	factory.EXPECT().NewChannelMgmtClient(sdk, gomock.Any(), c).Return(cm, nil)

	// Use a method that invokes credential manager (e.g., new user)
	_, err = sdk.NewChannelMgmtClient("User1")
	if err != nil {
		t.Fatalf("Unexpected error getting channel management client: %s", err)
	}
}

func TestErrPkgSuite(t *testing.T) {
	ps := mockPkgSuite{}

	c, err := configImpl.FromFile("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	_, err = fromPkgSuite(c, &ps)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	ps.errOnCore = true
	_, err = fromPkgSuite(c, &ps)
	if err == nil {
		t.Fatalf("Expected error initializing SDK")
	}
	ps.errOnCore = false

	ps.errOnService = true
	_, err = fromPkgSuite(c, &ps)
	if err == nil {
		t.Fatalf("Expected error initializing SDK")
	}
	ps.errOnService = false

	ps.errOnContext = true
	_, err = fromPkgSuite(c, &ps)
	if err == nil {
		t.Fatalf("Expected error initializing SDK")
	}
	ps.errOnContext = false

	ps.errOnSession = true
	_, err = fromPkgSuite(c, &ps)
	if err == nil {
		t.Fatalf("Expected error initializing SDK")
	}
	ps.errOnSession = false

	ps.errOnLogger = true
	_, err = fromPkgSuite(c, &ps)
	if err == nil {
		t.Fatalf("Expected error initializing SDK")
	}
	ps.errOnLogger = false
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
	c, err := configImpl.FromFile("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c)
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
	c1, err := configImpl.FromFile("../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk1, err := New(c1)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	c2, err := configImpl.FromFile("./testdata/test.yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk2, err := New(c2)
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

	c1, err := configImpl.FromRaw(cBytes, "yaml")
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c1)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	if sdk == nil {
		t.Fatalf("SDK should not be empty when initialized")
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
