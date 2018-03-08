/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	mockapisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/mocks"
	"github.com/pkg/errors"
)

const (
	sdkConfigFile      = "../../test/fixtures/config/config_test.yaml"
	sdkValidClientUser = "User1"
	sdkValidClientOrg1 = "Org1"
	sdkValidClientOrg2 = "Org2"
)

func TestNewGoodOpt(t *testing.T) {
	sdk, err := New(configImpl.FromFile(sdkConfigFile),
		goodOpt())
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}
	sdk.Close()
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

func TestDoubleClose(t *testing.T) {
	sdk, err := New(configImpl.FromFile(sdkConfigFile),
		goodOpt())
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}
	sdk.Close()
	sdk.Close()
}

func TestWithCorePkg(t *testing.T) {
	// Test New SDK with valid config file
	c, err := configImpl.FromFile(sdkConfigFile)()
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(WithConfig(c))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
	defer sdk.Close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockCoreProviderFactory(mockCtrl)

	factory.EXPECT().CreateCryptoSuiteProvider(c).Return(nil, nil)
	factory.EXPECT().CreateStateStoreProvider(c).Return(nil, nil)
	factory.EXPECT().CreateSigningManager(nil, c).Return(nil, nil)
	factory.EXPECT().CreateIdentityManager(gomock.Any(), gomock.Any(), nil, c).Return(nil, nil).AnyTimes()
	factory.EXPECT().CreateInfraProvider(gomock.Any()).Return(nil, nil)

	_, err = New(WithConfig(c), WithCorePkg(factory))
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

	sdk, err := New(WithConfig(c))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
	defer sdk.Close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockServiceProviderFactory(mockCtrl)

	factory.EXPECT().CreateDiscoveryProvider(c, gomock.Any()).Return(nil, nil)
	factory.EXPECT().CreateSelectionProvider(c).Return(nil, nil)

	_, err = New(WithConfig(c), WithServicePkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
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

	sdk, err := New(WithConfig(c))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
	sdk.Close()

	sdk, err = New(WithConfig(c), WithCorePkg(core))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
	defer sdk.Close()

	// Get resource management
	ctx := sdk.Context(WithUser(sdkValidClientUser), WithOrg(sdkValidClientOrg1))
	_, err = resmgmt.New(ctx)
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

	ps.errOnLogger = true
	_, err = fromPkgSuite(c, &ps)
	if err == nil {
		t.Fatalf("Expected error initializing SDK")
	}
	ps.errOnLogger = false
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
	sdk.Close()
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
	defer sdk.Close()

	client1, err := sdk.Config().Client()
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
