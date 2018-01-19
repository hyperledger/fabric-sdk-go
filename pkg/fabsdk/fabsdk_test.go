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

const (
	sdkConfigFile      = "../../test/fixtures/config/config_test.yaml"
	sdkValidClientUser = "User1"
	sdkValidClientOrg1 = "Org1"
	sdkValidClientOrg2 = "Org2"
)

func TestNewGoodOpt(t *testing.T) {
	_, err := New(configImpl.FromFile(sdkConfigFile),
		goodOpt())
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
	_, err := New(configImpl.FromFile(sdkConfigFile),
		badOpt())
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
	sdk, err := New(configImpl.FromFile(sdkConfigFile))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	clt1, err := sdk.NewClient(WithUser(sdkValidClientUser))
	if err != nil {
		t.Fatalf("Failed to create client: %s", err)
	}

	// Default channel client (uses organisation from client configuration)
	_, err = clt1.Channel("mychannel")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	clt2, err := sdk.NewClient(WithUser(sdkValidClientUser), WithOrg(sdkValidClientOrg2))
	if err != nil {
		t.Fatalf("Failed to create client: %s", err)
	}

	// Test configuration failure for channel client (mychannel does't have event source configured for Org2)
	_, err = clt2.Channel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to create channel client since event source not configured for Org2")
	}

	// Test new channel client with options
	_, err = clt2.Channel("orgchannel")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

}

func TestWithCorePkg(t *testing.T) {
	// Test New SDK with valid config file
	c, err := configImpl.FromFile(sdkConfigFile)()
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	_, err = fromConfig(c)
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

	_, err = fromConfig(c, WithCorePkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
}

func TestWithServicePkg(t *testing.T) {
	// Test New SDK with valid config file
	c, err := configImpl.FromFile(sdkConfigFile)()
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	_, err = fromConfig(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockServiceProviderFactory(mockCtrl)

	factory.EXPECT().NewDiscoveryProvider(c).Return(nil, nil)
	factory.EXPECT().NewSelectionProvider(c).Return(nil, nil)

	_, err = fromConfig(c, WithServicePkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
}

func TestWithContextPkg(t *testing.T) {
	// Test New SDK with valid config file
	c, err := configImpl.FromFile(sdkConfigFile)()
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	core, err := newMockCorePkg(c)
	if err != nil {
		t.Fatalf("Error initializing core factory: %s", err)
	}

	_, err = fromConfig(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Use real implementation of credential manager to provide in later response
	pkgSuite := defPkgSuite{}
	ctx, err := pkgSuite.Context()
	if err != nil {
		t.Fatalf("Unexpected error getting context: %s", err)
	}

	cm, err := ctx.NewCredentialManager(sdkValidClientOrg1, c, core.cryptoSuite)
	if err != nil {
		t.Fatalf("Unexpected error getting credential manager: %s", err)
	}

	// Create mock to ensure the provided factory is called.
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockOrgClientFactory(mockCtrl)

	factory.EXPECT().NewCredentialManager(sdkValidClientOrg1, c, core.cryptoSuite).Return(cm, nil)

	sdk, err := fromConfig(c, WithCorePkg(core), WithContextPkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Use a method that invokes credential manager (e.g., new user)
	_, err = sdk.NewPreEnrolledUser(sdkValidClientOrg1, sdkValidClientUser)
	if err != nil {
		t.Fatalf("Unexpected error getting user: %s", err)
	}
}

func TestWithSessionPkg(t *testing.T) {
	// Test New SDK with valid config file
	c, err := configImpl.FromFile(sdkConfigFile)()
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	core, err := newMockCorePkg(c)
	if err != nil {
		t.Fatalf("Error initializing core factory: %s", err)
	}

	_, err = fromConfig(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Create mock to ensure the provided factory is called.
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockSessionClientFactory(mockCtrl)

	sdk, err := fromConfig(c, WithCorePkg(core), WithSessionPkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Use real implementation of credential manager to provide in later response
	pkgSuite := defPkgSuite{}
	sessPkg, err := pkgSuite.Session()
	if err != nil {
		t.Fatalf("Unexpected error getting context: %s", err)
	}

	identity, err := sdk.newIdentity(sdkValidClientOrg1, WithUser(sdkValidClientUser))
	if err != nil {
		t.Fatalf("Unexpected error getting identity: %s", err)
	}

	session := newSession(identity)

	cm, err := sessPkg.NewChannelMgmtClient(sdk, session, c)
	if err != nil {
		t.Fatalf("Unexpected error getting credential manager: %s", err)
	}
	factory.EXPECT().NewChannelMgmtClient(sdk, gomock.Any(), c).Return(cm, nil)

	clt, err := sdk.NewClient(WithUser(sdkValidClientUser))
	if err != nil {
		t.Fatalf("Unexpected error getting client: %s", err)
	}

	// Use a method that invokes credential manager (e.g., new user)
	_, err = clt.ChannelMgmt()
	if err != nil {
		t.Fatalf("Unexpected error getting channel management client: %s", err)
	}
}

func TestErrPkgSuite(t *testing.T) {
	ps := mockPkgSuite{}

	c, err := configImpl.FromFile(sdkConfigFile)()
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

func TestNewDefaultTwoValidSDK(t *testing.T) {
	sdk1, err := New(configImpl.FromFile(sdkConfigFile))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	sdk2, err := New(configImpl.FromFile("./testdata/test.yaml"))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Default sdk with two channels
	client1, err := sdk1.configProvider.Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client1.Organization != sdkValidClientOrg1 {
		t.Fatalf("Unexpected org in config: %s", client1.Organization)
	}

	client2, err := sdk2.configProvider.Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client2.Organization != sdkValidClientOrg2 {
		t.Fatalf("Unexpected org in config: %s", client1.Organization)
	}

	clt1, err := sdk1.NewClient(WithUser(sdkValidClientUser))
	if err != nil {
		t.Fatalf("Failed to create client: %s", err)
	}

	// Test SDK1 channel clients ('mychannel', 'orgchannel')
	_, err = clt1.Channel("mychannel")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	_, err = clt1.Channel("orgchannel")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	clt2, err := sdk2.NewClient(WithUser(sdkValidClientUser))
	if err != nil {
		t.Fatalf("Failed to create client: %s", err)
	}

	// SDK 2 doesn't have 'mychannel' configured
	_, err = clt2.Channel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to create channel that is not configured")
	}

	// SDK 2 has 'orgchannel' configured
	_, err = clt2.Channel("orgchannel")
	if err != nil {
		t.Fatalf("Failed to create new 'orgchannel' channel client: %s", err)
	}
}

func TestNewDefaultSDKFromByte(t *testing.T) {
	cBytes, err := loadConfigBytesFromFile(t, sdkConfigFile)
	if err != nil {
		t.Fatalf("Failed to load sample bytes from File. Error: %s", err)
	}

	sdk, err := New(configImpl.FromRaw(cBytes, "yaml"))
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

func TestWithConfigSuccess(t *testing.T) {
	sdk, err := New(configImpl.FromFile(sdkConfigFile))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	client1, err := sdk.configProvider.Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client1.Organization != sdkValidClientOrg1 {
		t.Fatalf("Unexpected org in config: %s", client1.Organization)
	}
}

func TestWithConfigFailure(t *testing.T) {
	_, err := New(configImpl.FromFile("notarealfile"))
	if err == nil {
		t.Fatal("Expected failure due to invalid config")
	}
}
