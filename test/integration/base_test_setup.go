/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"os"
	"path"
	"time"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	fabAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
)

// BaseSetupImpl implementation of BaseTestSetup
type BaseSetupImpl struct {
	Identity          msp.Identity
	Targets           []string
	ConfigFile        string
	OrgID             string
	ChannelID         string
	ChannelConfigFile string
}

// Initial B values for ExampleCC
const (
	ExampleCCInitB    = "200"
	ExampleCCUpgradeB = "400"
	AdminUser         = "Admin"
	OrdererOrgName    = "ordererorg"
)

// ExampleCC query and transaction arguments
var queryArgs = [][]byte{[]byte("query"), []byte("b")}
var txArgs = [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}

// ExampleCC init and upgrade args
var initArgs = [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte(ExampleCCInitB)}
var upgradeArgs = [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte(ExampleCCUpgradeB)}

// ExampleCCQueryArgs returns example cc query args
func ExampleCCQueryArgs() [][]byte {
	return queryArgs
}

// ExampleCCTxArgs returns example cc move funds args
func ExampleCCTxArgs() [][]byte {
	return txArgs
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

// Initialize reads configuration from file and sets up client, channel and event hub
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

	r, err := os.Open(setup.ChannelConfigFile)
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

// GetDeployPath ..
func GetDeployPath() string {
	pwd, _ := os.Getwd()
	return path.Join(pwd, "../../fixtures/testdata")
}

// InstallAndInstantiateExampleCC install and instantiate using resource management client
func InstallAndInstantiateExampleCC(sdk *fabsdk.FabricSDK, user fabsdk.ContextOption, orgName string, chainCodeID string) (resmgmt.InstantiateCCResponse, error) {
	return InstallAndInstantiateCC(sdk, user, orgName, chainCodeID, "github.com/example_cc", "v0", GetDeployPath(), initArgs)
}

// InstallAndInstantiateCC install and instantiate using resource management client
func InstallAndInstantiateCC(sdk *fabsdk.FabricSDK, user fabsdk.ContextOption, orgName string, ccName, ccPath, ccVersion, goPath string, ccArgs [][]byte) (resmgmt.InstantiateCCResponse, error) {

	ccPkg, err := packager.NewCCPackage(ccPath, goPath)
	if err != nil {
		return resmgmt.InstantiateCCResponse{}, errors.WithMessage(err, "creating chaincode package failed")
	}

	configBackend, err := sdk.Config()
	if err != nil {
		return resmgmt.InstantiateCCResponse{}, errors.WithMessage(err, "failed to get config backend")
	}

	endpointConfig, err := fab.ConfigFromBackend(configBackend)
	if err != nil {
		return resmgmt.InstantiateCCResponse{}, errors.WithMessage(err, "failed to get endpoint config")
	}

	mspID, ok := comm.MSPID(endpointConfig, orgName)
	if !ok {
		return resmgmt.InstantiateCCResponse{}, errors.New("looking up MSP ID failed")
	}

	//prepare context
	clientContext := sdk.Context(user, fabsdk.WithOrg(orgName))

	// Resource management client is responsible for managing resources (joining channels, install/instantiate/upgrade chaincodes)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return resmgmt.InstantiateCCResponse{}, errors.WithMessage(err, "Failed to create new resource management client")
	}

	_, err = resMgmtClient.InstallCC(resmgmt.InstallCCRequest{Name: ccName, Path: ccPath, Version: ccVersion, Package: ccPkg}, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return resmgmt.InstantiateCCResponse{}, err
	}

	ccPolicy := cauthdsl.SignedByMspMember(mspID)
	return resMgmtClient.InstantiateCC("mychannel", resmgmt.InstantiateCCRequest{Name: ccName, Path: ccPath, Version: ccVersion, Args: ccArgs, Policy: ccPolicy}, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
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
func CreateChannelAndUpdateAnchorPeers(sdk *fabsdk.FabricSDK, channelID string, channelConfigFile string, orgsContext []*OrgContext) error {
	ordererCtx := sdk.Context(fabsdk.WithUser(AdminUser), fabsdk.WithOrg(OrdererOrgName))

	// Channel management client is responsible for managing channels (create/update channel)
	chMgmtClient, err := resmgmt.New(ordererCtx)
	if err != nil {
		return errors.New("failed to get a new resmgmt client for orderer")
	}

	var signingIdentities []msp.SigningIdentity
	for _, orgCtx := range orgsContext {
		signingIdentities = append(signingIdentities, orgCtx.SigningIdentity)
	}

	req := resmgmt.SaveChannelRequest{
		ChannelID:         channelID,
		ChannelConfigPath: path.Join("../../../", metadata.ChannelConfigPath, channelConfigFile),
		SigningIdentities: signingIdentities,
	}
	_, err = chMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	if err != nil {
		return err
	}

	for _, orgCtx := range orgsContext {
		req := resmgmt.SaveChannelRequest{
			ChannelID:         channelID,
			ChannelConfigPath: path.Join("../../../", metadata.ChannelConfigPath, orgCtx.AnchorPeerConfigFile),
			SigningIdentities: []msp.SigningIdentity{orgCtx.SigningIdentity},
		}
		if _, err := orgCtx.ResMgmt.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
			return err
		}
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

// InstallAndInstantiateChaincode installs the given chaincode to all peers in the given orgs and instantiates it on the given channel
func InstallAndInstantiateChaincode(channelID string, ccPkg *resource.CCPackage, ccID, ccVersion, ccPolicy string, orgs []*OrgContext, collConfigs ...*cb.CollectionConfig) error {
	for _, orgCtx := range orgs {
		if err := InstallChaincode(orgCtx.ResMgmt, orgCtx.CtxProvider, ccPkg, ccID, ccVersion, orgCtx.Peers); err != nil {
			return errors.Wrapf(err, "failed to install chaincode to peers in org [%s]", orgCtx.OrgID)
		}
	}
	_, err := InstantiateChaincode(orgs[0].ResMgmt, channelID, ccID, ccVersion, ccPolicy, collConfigs...)
	return err
}

// InstallChaincode installs the given chaincode to the given peers
func InstallChaincode(resMgmt *resmgmt.Client, ctxProvider contextAPI.ClientProvider, ccPkg *resource.CCPackage, ccName, ccVersion string, localPeers []fabAPI.Peer) error {
	installCCReq := resmgmt.InstallCCRequest{Name: ccName, Path: "github.com/example_cc", Version: ccVersion, Package: ccPkg}
	_, err := resMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	return err
}

// InstantiateChaincode instantiates the given chaincode to the given channel
func InstantiateChaincode(resMgmt *resmgmt.Client, channelID, ccName, ccVersion string, ccPolicyStr string, collConfigs ...*cb.CollectionConfig) (resmgmt.InstantiateCCResponse, error) {
	ccPolicy, err := cauthdsl.FromString(ccPolicyStr)
	if err != nil {
		return resmgmt.InstantiateCCResponse{}, errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
	}

	return resMgmt.InstantiateCC(
		channelID,
		resmgmt.InstantiateCCRequest{
			Name:       ccName,
			Path:       "github.com/example_cc",
			Version:    ccVersion,
			Args:       ExampleCCInitArgs(),
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

	var peers []fabAPI.Peer
	for i := 0; i < 10; i++ {
		peers, err = ctx.LocalDiscoveryService().GetPeers()
		if err != nil {
			return nil, errors.Wrapf(err, "error getting peers for MSP [%s]", ctx.Identifier().MSPID)
		}
		if len(peers) >= expectedPeers {
			break
		}
		// wait some time to allow the gossip to propagate the peers discovery
		time.Sleep(3 * time.Second)
	}
	if expectedPeers != len(peers) {
		return nil, errors.Errorf("Expecting %d peers but got %d", expectedPeers, len(peers))
	}
	return peers, nil
}
