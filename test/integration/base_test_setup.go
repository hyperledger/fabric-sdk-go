/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"os"
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

// BaseSetupImpl implementation of BaseTestSetup
type BaseSetupImpl struct {
	SDK             *fabsdk.FabricSDK
	Identity        fab.IdentityContext
	Client          fab.Resource
	Transactor      fab.Transactor
	Targets         []fab.ProposalProcessor
	Channel         fab.Channel
	EventHub        fab.EventHub
	ConnectEventHub bool
	ConfigFile      string
	OrgID           string
	ChannelID       string
	Initialized     bool
	ChannelConfig   string
}

// Initial B values for ExampleCC
const (
	ExampleCCInitB    = "200"
	ExampleCCUpgradeB = "400"
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

// Initialize reads configuration from file and sets up client, channel and event hub
func (setup *BaseSetupImpl) Initialize() error {
	// Create SDK setup for the integration tests
	sdk, err := fabsdk.New(config.FromFile(setup.ConfigFile))
	if err != nil {
		return errors.WithMessage(err, "SDK init failed")
	}
	setup.SDK = sdk

	client := sdk.NewClient(fabsdk.WithUser("Admin"), fabsdk.WithOrg(setup.OrgID))

	session, err := client.Session()
	if err != nil {
		return errors.WithMessage(err, "failed getting admin user session for org")
	}
	setup.Identity = session

	rc, err := sdk.FabricProvider().CreateResourceClient(setup.Identity)
	if err != nil {
		return errors.WithMessage(err, "NewResourceClient failed")
	}
	setup.Client = rc

	targets, err := getOrgTargets(sdk.Config(), setup.OrgID)
	if err != nil {
		return errors.Wrapf(err, "loading target peers from config failed")
	}
	setup.Targets = targets

	// Create channel for tests
	req := chmgmt.SaveChannelRequest{ChannelID: setup.ChannelID, ChannelConfig: setup.ChannelConfig, SigningIdentity: session}
	InitializeChannel(sdk, setup.OrgID, req, targets)

	// Create the channel transactor
	chService, err := client.ChannelService(setup.ChannelID)
	if err != nil {
		return errors.WithMessage(err, "channel service creation failed")
	}
	transactor, err := chService.Transactor()
	if err != nil {
		return errors.WithMessage(err, "transactor client creation failed")
	}
	setup.Transactor = transactor

	channel, err := chService.Channel()
	if err != nil {
		return errors.WithMessage(err, "channel client creation failed")
	}
	setup.Channel = channel

	eventHub, err := chService.EventHub()
	if err != nil {
		return errors.WithMessage(err, "eventhub client creation failed")
	}
	if setup.ConnectEventHub {
		if err := eventHub.Connect(); err != nil {
			return errors.WithMessage(err, "eventHub connect failed")
		}
	}
	setup.EventHub = eventHub

	setup.Initialized = true

	return nil
}

func getOrgTargets(config apiconfig.Config, org string) ([]fab.ProposalProcessor, error) {
	targets := []fab.ProposalProcessor{}

	peerConfig, err := config.PeersConfig(org)
	if err != nil {
		return nil, errors.WithMessage(err, "reading peer config failed")
	}
	for _, p := range peerConfig {
		target, err := peer.New(config, peer.FromPeerConfig(&apiconfig.NetworkPeer{PeerConfig: p}))
		if err != nil {
			return nil, errors.WithMessage(err, "NewPeer failed")
		}
		targets = append(targets, target)
	}
	return targets, nil
}

// InitConfig ...
func (setup *BaseSetupImpl) InitConfig() apiconfig.ConfigProvider {
	return config.FromFile(setup.ConfigFile)
}

// InstallCC use low level client to install chaincode
func (setup *BaseSetupImpl) InstallCC(name string, path string, version string, ccPackage *fab.CCPackage, targets []fab.ProposalProcessor) error {

	icr := fab.InstallChaincodeRequest{Name: name, Path: path, Version: version, Package: ccPackage, Targets: targets}

	_, _, err := setup.Client.InstallChaincode(icr)
	if err != nil {
		return errors.WithMessage(err, "InstallChaincode failed")
	}

	return nil
}

// GetDeployPath ..
func GetDeployPath() string {
	pwd, _ := os.Getwd()
	return path.Join(pwd, "../../fixtures/testdata")
}

// InstallAndInstantiateExampleCC install and instantiate using resource management client
func InstallAndInstantiateExampleCC(sdk *fabsdk.FabricSDK, user fabsdk.IdentityOption, orgName string, chainCodeID string) error {
	return InstallAndInstantiateCC(sdk, user, orgName, chainCodeID, "github.com/example_cc", "v0", GetDeployPath(), initArgs)
}

// InstallAndInstantiateCC install and instantiate using resource management client
func InstallAndInstantiateCC(sdk *fabsdk.FabricSDK, user fabsdk.IdentityOption, orgName string, ccName, ccPath, ccVersion, goPath string, ccArgs [][]byte) error {

	ccPkg, err := packager.NewCCPackage(ccPath, goPath)
	if err != nil {
		return errors.WithMessage(err, "creating chaincode package failed")
	}

	mspID, err := sdk.Config().MspID(orgName)
	if err != nil {
		return errors.WithMessage(err, "looking up MSP ID failed")
	}

	// Resource management client is responsible for managing resources (joining channels, install/instantiate/upgrade chaincodes)
	resMgmtClient, err := sdk.NewClient(user, fabsdk.WithOrg(orgName)).ResourceMgmt()
	if err != nil {
		return errors.WithMessage(err, "Failed to create new resource management client")
	}

	_, err = resMgmtClient.InstallCC(resmgmt.InstallCCRequest{Name: ccName, Path: ccPath, Version: ccVersion, Package: ccPkg})
	if err != nil {
		return err
	}

	ccPolicy := cauthdsl.SignedByMspMember(mspID)
	return resMgmtClient.InstantiateCC("mychannel", resmgmt.InstantiateCCRequest{Name: ccName, Path: ccPath, Version: ccVersion, Args: ccArgs, Policy: ccPolicy})
}

// CreateAndSendTransactionProposal ... TODO duplicate
func CreateAndSendTransactionProposal(transactor fab.ProposalSender, chainCodeID string,
	fcn string, args [][]byte, targets []fab.ProposalProcessor, transientData map[string][]byte) ([]*fab.TransactionProposalResponse, *fab.TransactionProposal, error) {

	propReq := fab.ChaincodeInvokeRequest{
		Fcn:          fcn,
		Args:         args,
		TransientMap: transientData,
		ChaincodeID:  chainCodeID,
	}

	tp, err := transactor.CreateChaincodeInvokeProposal(propReq)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "creating transaction proposal failed")
	}

	tpr, err := transactor.SendTransactionProposal(tp, targets)
	return tpr, tp, err
}

// CreateAndSendTransaction ...
func CreateAndSendTransaction(transactor fab.Sender, proposal *fab.TransactionProposal, resps []*fab.TransactionProposalResponse) (*fab.TransactionResponse, error) {

	txRequest := fab.TransactionRequest{
		Proposal:          proposal,
		ProposalResponses: resps,
	}
	tx, err := transactor.CreateTransaction(txRequest)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateTransaction failed")
	}

	transactionResponse, err := transactor.SendTransaction(tx)
	if err != nil {
		return nil, errors.WithMessage(err, "SendTransaction failed")

	}

	if transactionResponse.Err != nil {
		return nil, errors.Wrapf(transactionResponse.Err, "orderer %s failed", transactionResponse.Orderer)
	}

	return transactionResponse, nil
}

// RegisterTxEvent registers on the given eventhub for the give transaction
// returns a boolean channel which receives true when the event is complete
// and an error channel for errors
// TODO - Duplicate
func RegisterTxEvent(t *testing.T, txID fab.TransactionID, eventHub fab.EventHub) (chan bool, chan error) {
	done := make(chan bool)
	fail := make(chan error)

	eventHub.RegisterTxEvent(txID, func(txId string, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			t.Logf("Received error event for txid(%s)", txId)
			fail <- err
		} else {
			t.Logf("Received success event for txid(%s)", txId)
			done <- true
		}
	})

	return done, fail
}
