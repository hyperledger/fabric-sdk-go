// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/fabricselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	context2 "github.com/hyperledger/fabric-sdk-go/pkg/context"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	fabDiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	mockapisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/mocksdkapi"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sdkConfigFile      = "config_test.yaml"
	sdkValidClientUser = "User1"
	sdkValidClientOrg1 = "org1"
)

func TestNewGoodOpt(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	sdk, err := New(configImpl.FromFile(configPath),
		goodOpt())
	if err != nil {
		t.Fatalf("Expected no error from New, but got %s", err)
	}
	sdk.Close()
}

func goodOpt() Option {
	return func(opts *options) error {
		return nil
	}
}

func TestNewBadOpt(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	_, err := New(configImpl.FromFile(configPath),
		badOpt())
	if err == nil {
		t.Fatal("Expected error from New")
	}
}

func badOpt() Option {
	return func(opts *options) error {
		return errors.New("Bad Opt")
	}
}

func TestDoubleClose(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	sdk, err := New(configImpl.FromFile(configPath),
		goodOpt())
	if err != nil {
		t.Fatalf("Expected no error from New, but got %s", err)
	}
	sdk.Close()
	sdk.Close()
}

func TestWithCorePkg(t *testing.T) {
	// Test New SDK with valid config file
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	c := configImpl.FromFile(configPath)
	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
	defer sdk.Close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockCoreProviderFactory(mockCtrl)

	factory.EXPECT().CreateCryptoSuiteProvider(gomock.Any()).Return(nil, nil)
	factory.EXPECT().CreateSigningManager(gomock.Any()).Return(nil, nil)
	factory.EXPECT().CreateInfraProvider(gomock.Any()).Return(nil, nil)

	_, err = New(c, WithCorePkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
}

func TestWithMSPPkg(t *testing.T) {
	// Test New SDK with valid config file
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	c := configImpl.FromFile(configPath)

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
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	c := configImpl.FromFile(configPath)

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
	defer sdk.Close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	factory := mockapisdk.NewMockServiceProviderFactory(mockCtrl)

	factory.EXPECT().CreateLocalDiscoveryProvider(gomock.Any()).Return(nil, nil)
	factory.EXPECT().CreateChannelProvider(gomock.Any()).Return(nil, nil)

	_, err = New(c, WithServicePkg(factory))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}
}

func TestWithSessionPkg(t *testing.T) {
	// Test New SDK with valid config file
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	c := configImpl.FromFile(configPath)

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

	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	c := configImpl.FromFile(configPath)

	_, err := fromPkgSuite(c, &ps)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	ps.errOnCore = true
	_, err = fromPkgSuite(c, &ps)
	if err == nil {
		t.Fatal("Expected error initializing SDK")
	}
	ps.errOnCore = false

	ps.errOnService = true
	_, err = fromPkgSuite(c, &ps)
	if err == nil {
		t.Fatal("Expected error initializing SDK")
	}
	ps.errOnService = false

	ps.errOnLogger = true
	_, err = fromPkgSuite(c, &ps)
	if err == nil {
		t.Fatal("Expected error initializing SDK")
	}
	ps.errOnLogger = false
}

func TestNewDefaultSDKFromByte(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	cBytes, err := loadConfigBytesFromFile(t, configPath)
	if err != nil {
		t.Fatalf("Failed to load sample bytes from File. Error: %s", err)
	}

	sdk, err := New(configImpl.FromRaw(cBytes, "yaml"))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	if sdk == nil {
		t.Fatal("SDK should not be empty when initialized")
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
		t.Fatal("Failed to read test config for bytes array testing. Mock bytes array is empty")
	}
	return cBytes, err
}

func TestWithConfigSuccess(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	sdk, err := New(configImpl.FromFile(configPath))
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

	client1 := identityConfig.Client()
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

func TestEmptyConfigFile(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", "viper-test.yaml")
	_, err := New(configImpl.FromFile(configPath))
	assert.Nil(t, err, "New with empty config file should not have failed")
}

func TestWithConfigEndpoint(t *testing.T) {
	// Test New SDK with valid config file
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	c := configImpl.FromFile(configPath)

	np := &MockNetworkPeers{}
	// override EndpointConfig's NetworkConfig() function with np's and co's instances
	sdk, err := New(c, WithEndpointConfig(np))
	if err != nil {
		t.Fatalf("Error inializing sdk WithEndpointConfig: %s", err)
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
	// some functions by calling WithEndpointConfig(np, mo) above
	endpointConfig := sdk.opts.endpointConfig

	network := endpointConfig.NetworkPeers()
	expectedNetwork := np.NetworkPeers()
	if !reflect.DeepEqual(network, expectedNetwork) {
		t.Fatalf("Expected NetworkPeer was not returned by the sdk's config. Expected: %v, Received: %v", expectedNetwork, network)
	}
}

func TestWithConfigEndpointAndBadOpt(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	c := configImpl.FromFile(configPath)

	np := &MockNetworkPeers{}
	co := &MockChannelOrderers{}

	var badOpt interface{}
	// test bad opt
	_, err := New(c, WithEndpointConfig(np, co, badOpt))
	if err == nil {
		t.Fatal("expected empty endpointConfig during inializing sdk WithEndpointConfig with a bad option but got no error")
	}
}

func TestCloseContext(t *testing.T) {
	const channelID = "orgchannel"

	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	c := configImpl.FromFile(configPath)

	core, err := newMockCorePkg(c)
	require.NoError(t, err)

	sdk, err := New(c,
		WithCorePkg(core),
		WithServicePkg(&dynamicDiscoveryProviderFactory{}),
		WithProviderOpts(
			dynamicdiscovery.WithRefreshInterval(3*time.Millisecond),
		),
	)
	require.NoError(t, err)
	defer sdk.Close()

	chCfg := mocks.NewMockChannelCfg(channelID)
	chCfg.MockCapabilities[fab.ApplicationGroupKey][fab.V1_2Capability] = true
	chpvdr.SetChannelConfig(chCfg)

	discClient := fabDiscovery.NewMockDiscoveryClient()
	dynamicdiscovery.SetClientProvider(
		func(ctx context.Client) (fabDiscovery.Client, error) {
			return discClient, nil
		},
	)

	getDiscovery := func(orgID, userID string) (context.Channel, fab.DiscoveryService) {
		chCtxtProvider := sdk.ChannelContext(channelID, WithUser(userID), WithOrg(orgID))
		require.NotNil(t, chCtxtProvider)

		chCtxt, err := chCtxtProvider()
		require.NoError(t, err)
		require.NotNil(t, chCtxt)

		chService := chCtxt.ChannelService()
		require.NotNil(t, chService)

		discovery, err := chService.Discovery()
		require.NoError(t, err)

		return chCtxt, discovery
	}

	chCtxt1, discovery1 := getDiscovery(sdkValidClientOrg1, sdkValidClientUser)
	_, discovery2 := getDiscovery(sdkValidClientOrg2, sdkValidClientUser)

	discClient.SetResponses(
		&fabDiscovery.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
		},
	)

	_, err = discovery1.GetPeers()
	assert.NoError(t, err)

	_, err = discovery2.GetPeers()
	assert.NoError(t, err)

	// Close the first client context
	sdk.CloseContext(chCtxt1)

	// Wait for the cache to refresh
	time.Sleep(10 * time.Millisecond)

	// Subsequent calls on the first service should fail since the service is closed
	_, err = discovery1.GetPeers()
	assert.Error(t, err)
	assert.EqualError(t, err, "Discovery client has been closed")

	// Get the ChannelService from the second context; this one should still be valid
	_, err = discovery2.GetPeers()
	assert.NoError(t, err)
}

func TestErrorHandler(t *testing.T) {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, sdkConfigFile)
	c := configImpl.FromFile(configPath)

	core, err := newMockCorePkg(c)
	require.NoError(t, err)

	discClient := fabDiscovery.NewMockDiscoveryClient()
	dynamicdiscovery.SetClientProvider(func(ctx context.Client) (fabDiscovery.Client, error) {
		return discClient, nil
	})
	fabricselection.SetClientProvider(func(ctx context.Client) (fabricselection.DiscoveryClient, error) {
		return discClient, nil
	})

	var sdk *FabricSDK
	var chService fab.ChannelService
	var localCtxt context.Local
	var mutex sync.RWMutex

	newContext := func(user, org string) {
		getClientCtxt := sdk.Context(WithUser(user), WithOrg(org))
		require.NotNil(t, getClientCtxt)

		chCtxt, err := contextImpl.NewChannel(getClientCtxt, "orgchannel")
		require.NoError(t, err)
		require.NotNil(t, chCtxt)

		s := chCtxt.ChannelService()
		require.NotNil(t, s)

		lc, err := context2.NewLocal(getClientCtxt)
		require.NoError(t, err)

		mutex.Lock()
		defer mutex.Unlock()

		chService = s
		localCtxt = lc
	}

	getChannelService := func() fab.ChannelService {
		mutex.Lock()
		defer mutex.Unlock()
		return chService
	}

	getLocalCtxt := func() context.Local {
		mutex.Lock()
		defer mutex.Unlock()
		return localCtxt
	}

	errHandler := func(ctxt fab.ClientContext, channelID string, err error) {
		//todo this misunderstanding with DiscoveryError will be removed once fabricselection is fixed
		//https://github.com/hyperledger/fabric-sdk-go/pull/62#issuecomment-605343770
		selectionDiscoveryErr, selectionOk := errors.Cause(err).(fabricselection.DiscoveryError)
		dynamicDiscoveryErr, discoveryOk := errors.Cause(err).(dynamicdiscovery.DiscoveryError)

		// Analyse the error to see if it needs handling
		if (selectionOk && !selectionDiscoveryErr.IsAccessDenied()) || (discoveryOk && !dynamicDiscoveryErr.IsAccessDenied()) {
			// Transient error; no handling necessary
			return
		}

		// Need to spawn a new Go routine or else deadlock results when calling CloseContext
		go func() {
			sdk.CloseContext(ctxt)

			// Reset the successful response
			discClient.SetResponses(
				&fabDiscovery.MockDiscoverEndpointResponse{
					PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
				},
			)

			newContext(sdkValidClientUser, sdkValidClientOrg2)
		}()
	}

	sdk, err = New(c,
		WithCorePkg(core),
		WithServicePkg(&dynamicDiscoveryProviderFactory{}),
		WithErrorHandler(errHandler),
		WithProviderOpts(
			dynamicdiscovery.WithRefreshInterval(3*time.Millisecond),
			fabricselection.WithRefreshInterval(3*time.Millisecond),
		),
	)
	require.NoError(t, err)
	defer sdk.Close()

	chCfg := mocks.NewMockChannelCfg("orgchannel")
	chCfg.MockCapabilities[fab.ApplicationGroupKey][fab.V1_2Capability] = true
	chpvdr.SetChannelConfig(chCfg)

	newContext(sdkValidClientUser, sdkValidClientOrg1)

	localDiscovery := getLocalCtxt().LocalDiscoveryService()
	require.NotNil(t, localDiscovery)

	discovery, err := getChannelService().Discovery()
	require.NoError(t, err)

	selection, err := getChannelService().Selection()
	require.NoError(t, err)

	// First set a successful response
	discClient.SetResponses(
		&fabDiscovery.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
		},
	)

	_, err = localDiscovery.GetPeers()
	assert.NoError(t, err)

	_, err = discovery.GetPeers()
	assert.NoError(t, err)

	_, err = selection.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: "cc1"}})
	require.NoError(t, err)

	// Simulate a transient error
	discClient.SetResponses(
		&fabDiscovery.MockDiscoverEndpointResponse{
			Error: errors.New("some transient error"),
		},
	)

	time.Sleep(10 * time.Millisecond)

	_, err = localDiscovery.GetPeers()
	assert.NoError(t, err)

	_, err = discovery.GetPeers()
	assert.NoError(t, err)

	_, err = selection.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: "cc1"}})
	require.NoError(t, err)

	// Simulate an access-denied (could be due to a user being revoked)
	discClient.SetResponses(
		&fabDiscovery.MockDiscoverEndpointResponse{
			Error: errors.New(dynamicdiscovery.AccessDenied),
		},
	)

	time.Sleep(10 * time.Millisecond)

	// Subsequent calls on the old services should fail since the service is closed
	_, err = localDiscovery.GetPeers()
	assert.EqualError(t, err, "Discovery client has been closed")

	_, err = discovery.GetPeers()
	assert.EqualError(t, err, "Discovery client has been closed")

	_, err = selection.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: "cc1"}})
	assert.EqualError(t, err, "Selection service has been closed")

	// Refresh the services with the new context

	localDiscovery = getLocalCtxt().LocalDiscoveryService()
	require.NotNil(t, localDiscovery)

	_, err = localDiscovery.GetPeers()
	assert.NoError(t, err)

	discovery, err = getChannelService().Discovery()
	require.NoError(t, err)

	_, err = discovery.GetPeers()
	assert.NoError(t, err)

	selection, err = getChannelService().Selection()
	require.NoError(t, err)

	_, err = selection.GetEndorsersForChaincode([]*fab.ChaincodeCall{{ID: "cc1"}})
	require.NoError(t, err)
}

type MockNetworkPeers struct{}

func (M *MockNetworkPeers) NetworkPeers() []fab.NetworkPeer {
	return []fab.NetworkPeer{{PeerConfig: fab.PeerConfig{URL: "p.com"}, MSPID: ""}}
}

type MockChannelOrderers struct{}

func (M *MockChannelOrderers) ChannelOrderers(name string) []fab.OrdererConfig {
	return []fab.OrdererConfig{}
}

type dynamicDiscoveryProviderFactory struct {
	defsvc.ProviderFactory
}

// CreateLocalDiscoveryProvider returns a new local dynamic discovery provider
func (f *dynamicDiscoveryProviderFactory) CreateLocalDiscoveryProvider(config fab.EndpointConfig) (fab.LocalDiscoveryProvider, error) {
	return dynamicdiscovery.NewLocalProvider(config), nil
}
