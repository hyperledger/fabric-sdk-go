/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	mockapisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/mocksdkapi"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/pkg/errors"
)

const (
	sdkConfigFile      = "../../test/fixtures/config/config_test.yaml"
	sdkValidClientUser = "User1"
	sdkValidClientOrg1 = "org1"
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
	c := configImpl.FromFile(sdkConfigFile)
	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
	defer sdk.Close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockCoreProviderFactory(mockCtrl)

	factory.EXPECT().CreateCryptoSuiteProvider(gomock.Any()).Return(nil, nil)
	factory.EXPECT().CreateSigningManager(nil).Return(nil, nil)
	factory.EXPECT().CreateInfraProvider(gomock.Any()).Return(nil, nil)

	_, err = New(c, WithCorePkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
}

func TestWithMSPPkg(t *testing.T) {
	// Test New SDK with valid config file
	c := configImpl.FromFile(sdkConfigFile)

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
	defer sdk.Close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockMSPProviderFactory(mockCtrl)

	factory.EXPECT().CreateUserStore(gomock.Any()).Return(nil, nil)
	factory.EXPECT().CreateIdentityManagerProvider(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	_, err = New(c, WithMSPPkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
}

func TestWithServicePkg(t *testing.T) {
	// Test New SDK with valid config file
	c := configImpl.FromFile(sdkConfigFile)

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
	defer sdk.Close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockServiceProviderFactory(mockCtrl)

	factory.EXPECT().CreateDiscoveryProvider(gomock.Any()).Return(nil, nil)
	factory.EXPECT().CreateLocalDiscoveryProvider(gomock.Any()).Return(nil, nil)
	factory.EXPECT().CreateSelectionProvider(gomock.Any()).Return(nil, nil)

	_, err = New(c, WithServicePkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
}

func TestWithSessionPkg(t *testing.T) {
	// Test New SDK with valid config file
	c := configImpl.FromFile(sdkConfigFile)

	core, err := newMockCorePkg(c)
	if err != nil {
		t.Fatalf("Error initializing core factory: %s", err)
	}

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
	sdk.Close()

	sdk, err = New(c, WithCorePkg(core))
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

	c := configImpl.FromFile(sdkConfigFile)

	_, err := fromPkgSuite(c, &ps)
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
	cBytes := make([]byte, s)
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

	configBackend, err := sdk.Config()
	if err != nil {
		t.Fatalf("Error getting config backend from sdk: %s", err)
	}

	identityConfig, err := msp.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatalf("Error getting identity config: %s", err)
	}

	client1, err := identityConfig.Client()
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

func TestBadConfigFile(t *testing.T) {
	_, err := New(configImpl.FromFile("../../pkg/core/config/testdata/viper-test.yaml"))
	if err == nil {
		t.Fatalf("Expected error from New with bad config file")
	}
}

func TestWithConfigEndpoint(t *testing.T) {
	// Test New SDK with valid config file
	c := configImpl.FromFile(sdkConfigFile)

	np := &MockNetworkPeers{}
	co := &MockChannelOrderers{}
	// override EndpointConfig's NetworkConfig() function with np's and co's instances
	sdk, err := New(c, WithConfigEndpoint(np, co))
	if err != nil {
		t.Fatalf("Error inializing sdk WithConfigEndpoint: %s", err)
	}

	// TODO: configBackend always uses default EndpointConfig... need to expose the final endpointConfig (with opts/default functions)
	// without necessary fetching a backend as it is not used directly anymore if the user chooses
	// to fully override EndpointConfig ...
	// (ConfigFromBackend() should be hidden):
	//configBackend, err := sdk.Config()
	//if err != nil {
	//	t.Fatalf("Error getting config backend from sdk: %s", err)
	//}

	// it is not safe to assume fabImpl.ConfigFromBackend(configBackend) will return the final
	// EndpointConfig type intended by the user if they wish to override some or all of the interface, therefore:
	//endpointConfig, err := fabImpl.ConfigFromBackend(configBackend)
	//if err != nil {
	//	t.Fatalf("Error getting identity config: %s", err)
	//}
	// will always use the default implementation for the set configBackend
	// for the purpose of this test, we're getting endpointConfig from opts directly as we have overridden
	// some functions by calling WithConfigEndpoint(np, mo) above
	endpointConfig := sdk.opts.endpointConfig

	network, err := endpointConfig.NetworkPeers()
	if err != nil {
		t.Fatalf("Error getting NetworkPeer from config: %s", err)
	}
	expectedNetwork, err := np.NetworkPeers()
	if err != nil {
		t.Fatalf("Error getting extecd NetworkPeer from direct config: %s", err)
	}
	if !reflect.DeepEqual(network, expectedNetwork) {
		t.Fatalf("Expected NetworkPeer was not returned by the sdk's config. Expected: %s, Received: %s", expectedNetwork, network)
	}

	channelOrderers, err := endpointConfig.ChannelOrderers("")
	if err != nil {
		t.Fatalf("Error getting ChannelOrderers from config: %s", err)
	}
	expectedChannelOrderers, err := co.ChannelOrderers("")
	if err != nil {
		t.Fatalf("Error getting extecd ChannelOrderers from direct config: %s", err)
	}
	if !reflect.DeepEqual(channelOrderers, expectedChannelOrderers) {
		t.Fatalf("Expected ChannelOrderers was not returned by the sdk's config. Expected: %s, Received: %s", expectedChannelOrderers, channelOrderers)
	}

}

func TestWithConfigEndpointAndBadOpt(t *testing.T) {
	c := configImpl.FromFile(sdkConfigFile)

	np := &MockNetworkPeers{}
	co := &MockChannelOrderers{}

	var badOpt interface{}
	// test bad opt
	_, err := New(c, WithConfigEndpoint(np, co, badOpt))
	if err == nil {
		t.Fatal("expected empty endpointConfig during inializing sdk WithConfigEndpoint with a bad option but got no error")
	}
}

type MockNetworkPeers struct{}

func (M *MockNetworkPeers) NetworkPeers() ([]fab.NetworkPeer, error) {
	return []fab.NetworkPeer{{PeerConfig: fab.PeerConfig{URL: "p.com", EventURL: "event.p.com", GRPCOptions: nil, TLSCACerts: endpoint.TLSConfig{Path: "", Pem: ""}}, MSPID: ""}}, nil
}

type MockChannelOrderers struct{}

func (M *MockChannelOrderers) ChannelOrderers(name string) ([]fab.OrdererConfig, error) {
	return []fab.OrdererConfig{}, nil
}
