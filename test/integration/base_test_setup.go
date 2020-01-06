/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	fabAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// BaseSetupImpl implementation of BaseTestSetup
type BaseSetupImpl struct {
	Identity            msp.Identity
	Targets             []string
	ConfigFile          string
	OrgID               string
	ChannelID           string
	ChannelConfigTxFile string
}

// Initial B values for ExampleCC
const (
	ExampleCCInitB    = "200"
	ExampleCCUpgradeB = "400"
	AdminUser         = "Admin"
	OrdererOrgName    = "OrdererOrg"
	keyExp            = "key-%s-%s"
)

// ExampleCC query and transaction arguments
var defaultQueryArgs = [][]byte{[]byte("query"), []byte("b")}
var defaultTxArgs = [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}

// ExampleCC init and upgrade args
var initArgs = [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte(ExampleCCInitB)}
var upgradeArgs = [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte(ExampleCCUpgradeB)}
var resetArgs = [][]byte{[]byte("a"), []byte("100"), []byte("b"), []byte(ExampleCCInitB)}

// ExampleCCDefaultQueryArgs returns example cc query args
func ExampleCCDefaultQueryArgs() [][]byte {
	return defaultQueryArgs
}

// ExampleCCQueryArgs returns example cc query args
func ExampleCCQueryArgs(key string) [][]byte {
	return [][]byte{[]byte("query"), []byte(key)}
}

// ExampleCCTxArgs returns example cc query args
func ExampleCCTxArgs(from, to, val string) [][]byte {
	return [][]byte{[]byte("move"), []byte(from), []byte(to), []byte(val)}
}

// ExampleCCDefaultTxArgs returns example cc move funds args
func ExampleCCDefaultTxArgs() [][]byte {
	return defaultTxArgs
}

// ExampleCCTxRandomSetArgs returns example cc set args with random key-value pairs
func ExampleCCTxRandomSetArgs() [][]byte {
	return [][]byte{[]byte("set"), []byte(GenerateRandomID()), []byte(GenerateRandomID())}
}

//ExampleCCTxSetArgs sets the given key value in examplecc
func ExampleCCTxSetArgs(key, value string) [][]byte {
	return [][]byte{[]byte("set"), []byte(key), []byte(value)}
}

//ExampleCCInitArgs returns example cc initialization args
func ExampleCCInitArgs() [][]byte {
	return initArgs
}

//ExampleCCUpgradeArgs returns example cc upgrade args
func ExampleCCUpgradeArgs() [][]byte {
	return upgradeArgs
}

// IsJoinedChannel returns true if the given peer has joined the given channel
func IsJoinedChannel(channelID string, resMgmtClient *resmgmt.Client, peer fabAPI.Peer) (bool, error) {
	resp, err := resMgmtClient.QueryChannels(resmgmt.WithTargets(peer))
	if err != nil {
		return false, err
	}
	for _, chInfo := range resp.Channels {
		if chInfo.ChannelId == channelID {
			return true, nil
		}
	}
	return false, nil
}

// Initialize reads configuration from file and sets up client and channel
func (setup *BaseSetupImpl) Initialize(sdk *fabsdk.FabricSDK) error {

	mspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(setup.OrgID))
	adminIdentity, err := mspClient.GetSigningIdentity(AdminUser)
	if err != nil {
		return errors.WithMessage(err, "failed to get client context")
	}
	setup.Identity = adminIdentity

	var cfgBackends []core.ConfigBackend
	configBackend, err := sdk.Config()
	if err != nil {
		//For some tests SDK may not have backend set, try with config file if backend is missing
		cfgBackends, err = ConfigBackend()
		if err != nil {
			return errors.Wrapf(err, "failed to get config backend from config: %s", err)
		}
	} else {
		cfgBackends = append(cfgBackends, configBackend)
	}

	targets, err := OrgTargetPeers([]string{setup.OrgID}, cfgBackends...)
	if err != nil {
		return errors.Wrapf(err, "loading target peers from config failed")
	}
	setup.Targets = targets

	r, err := os.Open(setup.ChannelConfigTxFile)
	if err != nil {
		return errors.Wrapf(err, "opening channel config file failed")
	}
	defer func() {
		if err = r.Close(); err != nil {
			test.Logf("close error %v", err)
		}

	}()

	// Create channel for tests
	req := resmgmt.SaveChannelRequest{ChannelID: setup.ChannelID, ChannelConfig: r, SigningIdentities: []msp.SigningIdentity{adminIdentity}}
	if err = InitializeChannel(sdk, setup.OrgID, req, targets); err != nil {
		return errors.WithMessage(err, "failed to initialize channel")
	}

	return nil
}

// GetDeployPath returns the path to the chaincode fixtures
func GetDeployPath() string {
	const ccPath = "test/fixtures/testdata/go"
	return filepath.Join(metadata.GetProjectPath(), ccPath)
}

// GetJavaDeployPath returns the path to the java chaincode fixtrues
func GetJavaDeployPath() string {
	const ccPath = "test/fixtures/testdata/java"
	return filepath.Join(metadata.GetProjectPath(), ccPath)
}

// GetNodeDeployPath returns the path to the node chaincode fixtrues
func GetNodeDeployPath() string {
	const ccPath = "test/fixtures/testdata/node"
	return filepath.Join(metadata.GetProjectPath(), ccPath)
}

// GetChannelConfigTxPath returns the path to the named channel config file
func GetChannelConfigTxPath(filename string) string {
	return filepath.Join(metadata.GetProjectPath(), metadata.ChannelConfigPath, filename)
}

// GetConfigPath returns the path to the named config fixture file
func GetConfigPath(filename string) string {
	const configPath = "test/fixtures/config"
	return filepath.Join(metadata.GetProjectPath(), configPath, filename)
}

// GetConfigOverridesPath returns the path to the named config override fixture file
func GetConfigOverridesPath(filename string) string {
	const configPath = "test/fixtures/config"
	return filepath.Join(metadata.GetProjectPath(), configPath, "overrides", filename)
}

// GetCryptoConfigPath returns the path to the named crypto-config override fixture file
func GetCryptoConfigPath(filename string) string {
	const configPath = "test/fixtures/fabric/v1/crypto-config"
	return filepath.Join(metadata.GetProjectPath(), configPath, filename)
}

// OrgContext provides SDK client context for a given org
type OrgContext struct {
	OrgID                string
	CtxProvider          contextAPI.ClientProvider
	SigningIdentity      msp.SigningIdentity
	ResMgmt              *resmgmt.Client
	Peers                []fabAPI.Peer
	AnchorPeerConfigFile string
}

// CreateChannelAndUpdateAnchorPeers creates the channel and updates all of the anchor peers for all orgs
func CreateChannelAndUpdateAnchorPeers(t *testing.T, sdk *fabsdk.FabricSDK, channelID string, channelConfigFile string, orgsContext []*OrgContext) error {
	ordererCtx := sdk.Context(fabsdk.WithUser(AdminUser), fabsdk.WithOrg(OrdererOrgName))

	// Channel management client is responsible for managing channels (create/update channel)
	chMgmtClient, err := resmgmt.New(ordererCtx)
	if err != nil {
		return errors.New("failed to get a new resmgmt client for orderer")
	}

	var lastConfigBlock uint64
	var signingIdentities []msp.SigningIdentity
	for _, orgCtx := range orgsContext {
		signingIdentities = append(signingIdentities, orgCtx.SigningIdentity)
	}

	req := resmgmt.SaveChannelRequest{
		ChannelID:         channelID,
		ChannelConfigPath: GetChannelConfigTxPath(channelConfigFile),
		SigningIdentities: signingIdentities,
	}
	_, err = chMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	if err != nil {
		return err
	}

	lastConfigBlock = WaitForOrdererConfigUpdate(t, orgsContext[0].ResMgmt, channelID, true, lastConfigBlock)

	for _, orgCtx := range orgsContext {
		req := resmgmt.SaveChannelRequest{
			ChannelID:         channelID,
			ChannelConfigPath: GetChannelConfigTxPath(orgCtx.AnchorPeerConfigFile),
			SigningIdentities: []msp.SigningIdentity{orgCtx.SigningIdentity},
		}
		if _, err := orgCtx.ResMgmt.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
			return err
		}

		lastConfigBlock = WaitForOrdererConfigUpdate(t, orgCtx.ResMgmt, channelID, false, lastConfigBlock)
	}

	return nil
}

// JoinPeersToChannel joins all peers in all of the given orgs to the given channel
func JoinPeersToChannel(channelID string, orgsContext []*OrgContext) error {
	for _, orgCtx := range orgsContext {
		err := orgCtx.ResMgmt.JoinChannel(
			channelID,
			resmgmt.WithRetry(retry.DefaultResMgmtOpts),
			resmgmt.WithOrdererEndpoint("orderer.example.com"),
			resmgmt.WithTargets(orgCtx.Peers...),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to join peers in org [%s] to channel [%s]", orgCtx.OrgID, channelID)
		}
	}
	return nil
}

// InstallChaincodeWithOrgContexts installs the given chaincode to orgs
func InstallChaincodeWithOrgContexts(orgs []*OrgContext, ccPkg *resource.CCPackage, ccPath, ccID, ccVersion string) error {
	for _, orgCtx := range orgs {
		if err := InstallChaincode(orgCtx.ResMgmt, ccPkg, ccPath, ccID, ccVersion, orgCtx.Peers); err != nil {
			return errors.Wrapf(err, "failed to install chaincode to peers in org [%s]", orgCtx.OrgID)
		}
	}

	return nil
}

// InstallChaincode installs the given chaincode to the given peers
func InstallChaincode(resMgmt *resmgmt.Client, ccPkg *resource.CCPackage, ccPath, ccName, ccVersion string, localPeers []fabAPI.Peer) error {
	installCCReq := resmgmt.InstallCCRequest{Name: ccName, Path: ccPath, Version: ccVersion, Package: ccPkg}
	_, err := resMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return err
	}

	installed, err := queryInstalledCC(resMgmt, ccName, ccVersion, localPeers)

	if err != nil {
		return err
	}

	if !installed {
		return errors.New("chaincode was not installed on all peers")
	}

	return nil
}

// InstantiateChaincode instantiates the given chaincode to the given channel
func InstantiateChaincode(resMgmt *resmgmt.Client, channelID, ccName, ccPath, ccVersion string, ccPolicyStr string, args [][]byte, collConfigs ...*pb.CollectionConfig) (resmgmt.InstantiateCCResponse, error) {
	ccPolicy, err := cauthdsl.FromString(ccPolicyStr)
	if err != nil {
		return resmgmt.InstantiateCCResponse{}, errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
	}

	return resMgmt.InstantiateCC(
		channelID,
		resmgmt.InstantiateCCRequest{
			Name:       ccName,
			Path:       ccPath,
			Version:    ccVersion,
			Args:       args,
			Policy:     ccPolicy,
			CollConfig: collConfigs,
		},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	)
}

// InstantiateJavaChaincode instantiates the given java chaincode to the given channel
func InstantiateJavaChaincode(resMgmt *resmgmt.Client, channelID, ccName, ccPath, ccVersion string, ccPolicyStr string, args [][]byte, collConfigs ...*pb.CollectionConfig) (resmgmt.InstantiateCCResponse, error) {
	ccPolicy, err := cauthdsl.FromString(ccPolicyStr)
	if err != nil {
		return resmgmt.InstantiateCCResponse{}, errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
	}

	return resMgmt.InstantiateCC(
		channelID,
		resmgmt.InstantiateCCRequest{
			Name:       ccName,
			Path:       ccPath,
			Version:    ccVersion,
			Lang:       pb.ChaincodeSpec_JAVA,
			Args:       args,
			Policy:     ccPolicy,
			CollConfig: collConfigs,
		},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	)
}

// InstantiateNodeChaincode instantiates the given node chaincode to the given channel
func InstantiateNodeChaincode(resMgmt *resmgmt.Client, channelID, ccName, ccPath, ccVersion string, ccPolicyStr string, args [][]byte, collConfigs ...*pb.CollectionConfig) (resmgmt.InstantiateCCResponse, error) {
	ccPolicy, err := cauthdsl.FromString(ccPolicyStr)
	if err != nil {
		return resmgmt.InstantiateCCResponse{}, errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
	}

	return resMgmt.InstantiateCC(
		channelID,
		resmgmt.InstantiateCCRequest{
			Name:       ccName,
			Path:       ccPath,
			Version:    ccVersion,
			Lang:       pb.ChaincodeSpec_NODE,
			Args:       args,
			Policy:     ccPolicy,
			CollConfig: collConfigs,
		},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	)
}

// UpgradeChaincode upgrades the given chaincode on the given channel
func UpgradeChaincode(resMgmt *resmgmt.Client, channelID, ccName, ccPath, ccVersion string, ccPolicyStr string, args [][]byte, collConfigs ...*pb.CollectionConfig) (resmgmt.UpgradeCCResponse, error) {
	ccPolicy, err := cauthdsl.FromString(ccPolicyStr)
	if err != nil {
		return resmgmt.UpgradeCCResponse{}, errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
	}

	return resMgmt.UpgradeCC(
		channelID,
		resmgmt.UpgradeCCRequest{
			Name:       ccName,
			Path:       ccPath,
			Version:    ccVersion,
			Args:       args,
			Policy:     ccPolicy,
			CollConfig: collConfigs,
		},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	)
}

// UpgradeJavaChaincode upgrades the given java chaincode on the given channel
func UpgradeJavaChaincode(resMgmt *resmgmt.Client, channelID, ccName, ccPath, ccVersion string, ccPolicyStr string, args [][]byte, collConfigs ...*pb.CollectionConfig) (resmgmt.UpgradeCCResponse, error) {
	ccPolicy, err := cauthdsl.FromString(ccPolicyStr)
	if err != nil {
		return resmgmt.UpgradeCCResponse{}, errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
	}

	return resMgmt.UpgradeCC(
		channelID,
		resmgmt.UpgradeCCRequest{
			Name:       ccName,
			Path:       ccPath,
			Version:    ccVersion,
			Lang:       pb.ChaincodeSpec_JAVA,
			Args:       args,
			Policy:     ccPolicy,
			CollConfig: collConfigs,
		},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	)
}

// UpgradeNodeChaincode upgrades the given node chaincode on the given channel
func UpgradeNodeChaincode(resMgmt *resmgmt.Client, channelID, ccName, ccPath, ccVersion string, ccPolicyStr string, args [][]byte, collConfigs ...*pb.CollectionConfig) (resmgmt.UpgradeCCResponse, error) {
	ccPolicy, err := cauthdsl.FromString(ccPolicyStr)
	if err != nil {
		return resmgmt.UpgradeCCResponse{}, errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
	}

	return resMgmt.UpgradeCC(
		channelID,
		resmgmt.UpgradeCCRequest{
			Name:       ccName,
			Path:       ccPath,
			Version:    ccVersion,
			Lang:       pb.ChaincodeSpec_NODE,
			Args:       args,
			Policy:     ccPolicy,
			CollConfig: collConfigs,
		},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	)
}

// DiscoverLocalPeers queries the local peers for the given MSP context and returns all of the peers. If
// the number of peers does not match the expected number then an error is returned.
func DiscoverLocalPeers(ctxProvider contextAPI.ClientProvider, expectedPeers int) ([]fabAPI.Peer, error) {
	ctx, err := contextImpl.NewLocal(ctxProvider)
	if err != nil {
		return nil, errors.Wrap(err, "error creating local context")
	}

	discoveredPeers, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			peers, serviceErr := ctx.LocalDiscoveryService().GetPeers()
			if serviceErr != nil {
				return nil, errors.Wrapf(serviceErr, "error getting peers for MSP [%s]", ctx.Identifier().MSPID)
			}
			if len(peers) < expectedPeers {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Expecting %d peers but got %d", expectedPeers, len(peers)), nil)
			}
			return peers, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return discoveredPeers.([]fabAPI.Peer), nil
}

// EnsureChannelCreatedAndPeersJoined creates a channel, joins all peers in the given orgs to the channel and updates the anchor peers of each org.
func EnsureChannelCreatedAndPeersJoined(t *testing.T, sdk *fabsdk.FabricSDK, channelID string, channelTxFile string, orgsContext []*OrgContext) error {
	joined, err := IsJoinedChannel(channelID, orgsContext[0].ResMgmt, orgsContext[0].Peers[0])
	if err != nil {
		return err
	}

	if joined {
		return nil
	}

	// Create the channel and update anchor peers for all orgs
	if err := CreateChannelAndUpdateAnchorPeers(t, sdk, channelID, channelTxFile, orgsContext); err != nil {
		return err
	}

	return JoinPeersToChannel(channelID, orgsContext)
}

// WaitForOrdererConfigUpdate waits until the config block update has been committed.
// In Fabric 1.0 there is a bug that panics the orderer if more than one config update is added to the same block.
// This function may be invoked after each config update as a workaround.
func WaitForOrdererConfigUpdate(t *testing.T, client *resmgmt.Client, channelID string, genesis bool, lastConfigBlock uint64) uint64 {

	blockNum, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			chConfig, err := client.QueryConfigFromOrderer(channelID, resmgmt.WithOrdererEndpoint("orderer.example.com"))
			if err != nil {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), err.Error(), nil)
			}

			currentBlock := chConfig.BlockNumber()
			if currentBlock <= lastConfigBlock && !genesis {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Block number was not incremented [%d, %d]", currentBlock, lastConfigBlock), nil)
			}

			block, err := client.QueryConfigBlockFromOrderer(channelID, resmgmt.WithOrdererEndpoint("orderer.example.com"))
			if err != nil {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), err.Error(), nil)
			}
			if block.Header.Number != currentBlock {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Invalid block number [%d, %d]", block.Header.Number, currentBlock), nil)
			}

			return &currentBlock, nil
		},
	)

	require.NoError(t, err)
	return *blockNum.(*uint64)
}

func queryInstalledCC(resMgmt *resmgmt.Client, ccName, ccVersion string, peers []fabAPI.Peer) (bool, error) {
	installed, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			ok, err := isCCInstalled(resMgmt, ccName, ccVersion, peers)
			if err != nil {
				return &ok, err
			}
			if !ok {
				return &ok, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Chaincode [%s:%s] is not installed on all peers in Org1", ccName, ccVersion), nil)
			}
			return &ok, nil
		},
	)

	if err != nil {
		s, ok := status.FromError(err)
		if ok && s.Code == status.GenericTransient.ToInt32() {
			return false, nil
		}
		return false, errors.WithMessage(err, "isCCInstalled invocation failed")
	}

	return *(installed).(*bool), nil
}

func isCCInstalled(resMgmt *resmgmt.Client, ccName, ccVersion string, peers []fabAPI.Peer) (bool, error) {
	installedOnAllPeers := true
	for _, peer := range peers {
		resp, err := resMgmt.QueryInstalledChaincodes(resmgmt.WithTargets(peer))
		if err != nil {
			return false, errors.WithMessage(err, "querying for installed chaincodes failed")
		}

		found := false
		for _, ccInfo := range resp.Chaincodes {
			if ccInfo.Name == ccName && ccInfo.Version == ccVersion {
				found = true
				break
			}
		}
		if !found {
			installedOnAllPeers = false
		}
	}
	return installedOnAllPeers, nil
}

//GetKeyName creates random key name based on test name
func GetKeyName(t *testing.T) string {
	return fmt.Sprintf(keyExp, t.Name(), GenerateRandomID())
}

//ResetKeys resets given set of keys in example cc to given value
func ResetKeys(t *testing.T, ctx contextAPI.ChannelProvider, chaincodeID, value string, keys ...string) {
	chClient, err := channel.New(ctx)
	require.NoError(t, err, "Failed to create new channel client for resetting keys")
	for _, key := range keys {
		// Synchronous transaction
		_, e := chClient.Execute(
			channel.Request{
				ChaincodeID: chaincodeID,
				Fcn:         "invoke",
				Args:        ExampleCCTxSetArgs(key, value),
			},
			channel.WithRetry(retry.DefaultChannelOpts))
		require.NoError(t, e, "Failed to reset keys")
	}
}
