/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package admin

import (
	"io/ioutil"
	"os"
	"time"

	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	internal "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/internal"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")
var origGoPath = os.Getenv("GOPATH")

// SendInstallCC  Sends an install proposal to one or more endorsing peers.
func SendInstallCC(client fab.FabricClient, chainCodeID string, chainCodePath string,
	chainCodeVersion string, chaincodePackage []byte, targets []fab.Peer, deployPath string) error {

	changeGOPATHToDeploy(deployPath)
	transactionProposalResponse, _, err := client.InstallChaincode(chainCodeID, chainCodePath, chainCodeVersion, chaincodePackage, targets)
	resetGOPATH()
	if err != nil {
		return errors.WithMessage(err, "InstallChaincode failed")
	}
	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			logger.Debugf("InstallChaincode endorser %s returned error", v.Endorser)
			return errors.WithMessage(v.Err, "InstallChaincode endorser failed")
		}
		logger.Debugf("InstallChaincode endorser '%s' returned ProposalResponse status:%v", v.Endorser, v.Status)
	}

	return nil
}

// SendInstantiateCC Sends instantiate CC proposal to one or more endorsing peers
func SendInstantiateCC(channel fab.Channel, chainCodeID string, args []string,
	chaincodePath string, chaincodeVersion string, chaincodePolicy *common.SignaturePolicyEnvelope, targets []apitxn.ProposalProcessor, eventHub fab.EventHub) error {

	transactionProposalResponse, txID, err := channel.SendInstantiateProposal(chainCodeID,
		args, chaincodePath, chaincodeVersion, chaincodePolicy, targets)
	if err != nil {
		return errors.WithMessage(err, "SendInstantiateProposal failed")
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			logger.Debugf("SendInstantiateProposal endorser %s returned error", v.Endorser)
			return errors.WithMessage(v.Err, "SendInstantiateProposal endorser failed")
		}
		logger.Debug("SendInstantiateProposal endorser '%s' returned ProposalResponse status:%v", v.Endorser, v.Status)
	}

	// Register for commit event
	chcode := internal.RegisterTxEvent(txID, eventHub)

	if _, err = internal.CreateAndSendTransaction(channel, transactionProposalResponse); err != nil {
		return errors.WithMessage(err, "CreateAndSendTransaction failed")
	}

	select {
	case code := <-chcode:
		if code == peer.TxValidationCode_VALID {
			return nil
		}
		logger.Debugf("instantiateCC error received from eventhub for txid(%s), code(%s)", txID, code)
		return errors.Errorf("instantiateCC with code %s", code)
	case <-time.After(time.Second * 30):
		logger.Debugf("instantiateCC didn't receive block event for txid(%s)", txID)
		return errors.New("instantiateCC timeout")
	}
}

// SendUpgradeCC Sends upgrade CC proposal to one or more endorsing peers
func SendUpgradeCC(channel fab.Channel, chainCodeID string, args []string,
	chaincodePath string, chaincodeVersion string, chaincodePolicy *common.SignaturePolicyEnvelope, targets []apitxn.ProposalProcessor, eventHub fab.EventHub) error {

	transactionProposalResponse, txID, err := channel.SendUpgradeProposal(chainCodeID,
		args, chaincodePath, chaincodeVersion, chaincodePolicy, targets)
	if err != nil {
		return errors.WithMessage(err, "SendUpgradeProposal failed")
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			logger.Debugf("SendUpgradeProposal endorser %s failed", v.Endorser)
			return errors.WithMessage(v.Err, "SendUpgradeProposal endorser failed")
		}
		logger.Debug("SendUpgradeProposal Endorser '%s' returned ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	// Register for commit event
	chcode := internal.RegisterTxEvent(txID, eventHub)

	if _, err = internal.CreateAndSendTransaction(channel, transactionProposalResponse); err != nil {
		return errors.WithMessage(err, "CreateAndSendTransaction failed")
	}

	select {
	case code := <-chcode:
		if code == peer.TxValidationCode_VALID {
			return nil
		}
		logger.Debugf("upgradeCC Error received from eventhub for txid(%s) code(%s)", txID, code)
		return errors.Errorf("upgradeCC failed with code %s", code)
	case <-time.After(time.Second * 30):
		logger.Debugf("instantiateCC didn't receive block event for txid(%s)", txID)
		return errors.New("upgradeCC timeout")
	}
}

// CreateOrUpdateChannel creates a channel if it does not exist or updates a channel
// if it does and a different channelConfig is used
func CreateOrUpdateChannel(client fab.FabricClient, ordererUser ca.User, orgUser ca.User, channel fab.Channel, channelConfig string) error {
	logger.Debugf("***** Creating or updating channel: %s *****\n", channel.Name())

	currentUser := client.UserContext()
	defer client.SetUserContext(currentUser)

	client.SetUserContext(orgUser)

	configTx, err := ioutil.ReadFile(channelConfig)
	if err != nil {
		return errors.Wrap(err, "reading config file failed")
	}

	config, err := client.ExtractChannelConfig(configTx)
	if err != nil {
		return errors.WithMessage(err, "extracting channel config failed")
	}

	configSignature, err := client.SignChannelConfig(config)
	if err != nil {
		return errors.WithMessage(err, "signing configuration failed")
	}

	var configSignatures []*common.ConfigSignature
	configSignatures = append(configSignatures, configSignature)

	request := fab.CreateChannelRequest{
		Name:       channel.Name(),
		Orderer:    channel.Orderers()[0],
		Config:     config,
		Signatures: configSignatures,
	}

	client.SetUserContext(ordererUser)
	_, err = client.CreateChannel(request)
	if err != nil {
		return errors.WithMessage(err, "CreateChannel failed")
	}

	return nil
}

// JoinChannel joins a channel that has already been created
func JoinChannel(client fab.FabricClient, orgUser ca.User, channel fab.Channel) error {
	currentUser := client.UserContext()
	defer client.SetUserContext(currentUser)

	client.SetUserContext(orgUser)

	txnid, err := client.NewTxnID()
	if err != nil {
		return errors.WithMessage(err, "NewTxnID failed")
	}

	genesisBlockRequest := &fab.GenesisBlockRequest{
		TxnID: txnid,
	}
	genesisBlock, err := channel.GenesisBlock(genesisBlockRequest)
	if err != nil {
		return errors.WithMessage(err, "genesis block retrieval failed")
	}

	txnid2, err := client.NewTxnID()
	if err != nil {
		return errors.WithMessage(err, "NewTxnID failed")
	}

	joinChannelRequest := &fab.JoinChannelRequest{
		Targets:      channel.Peers(),
		GenesisBlock: genesisBlock,
		TxnID:        txnid2,
	}

	err = channel.JoinChannel(joinChannelRequest)
	if err != nil {
		return errors.WithMessage(err, "join channel failed")
	}

	return nil
}

// ChangeGOPATHToDeploy changes go path to fixtures folder
func changeGOPATHToDeploy(deployPath string) {
	os.Setenv("GOPATH", deployPath)
}

// ResetGOPATH resets go path to original
func resetGOPATH() {
	os.Setenv("GOPATH", origGoPath)
}
