/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package admin

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	api "github.com/hyperledger/fabric-sdk-go/api"
	internal "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/internal"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")
var origGoPath = os.Getenv("GOPATH")

// SendInstallCC  Sends an install proposal to one or more endorsing peers.
func SendInstallCC(client api.FabricClient, chainCodeID string, chainCodePath string,
	chainCodeVersion string, chaincodePackage []byte, targets []api.Peer, deployPath string) error {

	changeGOPATHToDeploy(deployPath)
	transactionProposalResponse, _, err := client.InstallChaincode(chainCodeID, chainCodePath, chainCodeVersion, chaincodePackage, targets)
	resetGOPATH()
	if err != nil {
		return fmt.Errorf("InstallChaincode returned error: %v", err)
	}
	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("InstallChaincode Endorser %s returned error: %v", v.Endorser, v.Err)
		}
		logger.Debugf("InstallChaincode Endorser '%s' returned ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	return nil
}

// SendInstantiateCC Sends instantiate CC proposal to one or more endorsing peers
func SendInstantiateCC(channel api.Channel, chainCodeID string, channelID string, args []string,
	chaincodePath string, chaincodeVersion string, targets []api.Peer, eventHub api.EventHub) error {

	transactionProposalResponse, txID, err := channel.SendInstantiateProposal(chainCodeID,
		channelID, args, chaincodePath, chaincodeVersion, targets)
	if err != nil {
		return fmt.Errorf("SendInstantiateProposal returned error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("SendInstantiateProposal Endorser %s returned error: %v", v.Endorser, v.Err)
		}
		logger.Debug("SendInstantiateProposal Endorser '%s' returned ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	// Register for commit event
	done, fail := internal.RegisterTxEvent(txID, eventHub)

	if _, err = internal.CreateAndSendTransaction(channel, transactionProposalResponse); err != nil {
		return fmt.Errorf("CreateTransaction returned error: %v", err)
	}

	select {
	case <-done:
	case <-fail:
		return fmt.Errorf("instantiateCC Error received from eventhub for txid(%s) error(%v)", txID, fail)
	case <-time.After(time.Second * 30):
		return fmt.Errorf("instantiateCC Didn't receive block event for txid(%s)", txID)
	}
	return nil
}

// CreateOrUpdateChannel creates a channel if it does not exist or updates a channel
// if it does and a different channelConfig is used
func CreateOrUpdateChannel(client api.FabricClient, ordererUser api.User, orgUser api.User, channel api.Channel, channelConfig string) error {
	logger.Debugf("***** Creating or updating channel: %s *****\n", channel.Name())

	currentUser := client.GetUserContext()
	defer client.SetUserContext(currentUser)

	client.SetUserContext(orgUser)

	configTx, err := ioutil.ReadFile(channelConfig)
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	config, err := client.ExtractChannelConfig(configTx)
	if err != nil {
		return fmt.Errorf("Error extracting channel config: %v", err)
	}

	configSignature, err := client.SignChannelConfig(config)
	if err != nil {
		return fmt.Errorf("Error signing configuration: %v", err)
	}

	var configSignatures []*common.ConfigSignature
	configSignatures = append(configSignatures, configSignature)

	creator, err := client.GetIdentity()
	if err != nil {
		return fmt.Errorf("Error getting creator: %v", err)
	}
	nonce, err := internal.GenerateRandomNonce()
	if err != nil {
		return fmt.Errorf("Could not compute nonce: %s", err)
	}
	txID, err := internal.ComputeTxID(nonce, creator)
	if err != nil {
		return fmt.Errorf("Could not compute TxID: %s", err)
	}

	request := api.CreateChannelRequest{
		Name:       channel.Name(),
		Orderer:    channel.Orderers()[0],
		Config:     config,
		Signatures: configSignatures,
		TxID:       txID,
		Nonce:      nonce,
	}

	client.SetUserContext(ordererUser)
	err = client.CreateChannel(&request)
	if err != nil {
		return fmt.Errorf("CreateChannel returned error: %v", err)
	}

	return nil
}

// JoinChannel joins a channel that has already been created
func JoinChannel(client api.FabricClient, orgUser api.User, channel api.Channel) error {
	currentUser := client.GetUserContext()
	defer client.SetUserContext(currentUser)

	client.SetUserContext(orgUser)

	creator, err := client.GetIdentity()
	if err != nil {
		return fmt.Errorf("Error getting creator: %v", err)
	}

	nonce, err := internal.GenerateRandomNonce()
	if err != nil {
		return fmt.Errorf("Could not compute nonce: %s", err)
	}
	txID, err := internal.ComputeTxID(nonce, creator)
	if err != nil {
		return fmt.Errorf("Could not compute TxID: %s", err)
	}

	genesisBlockRequest := &api.GenesisBlockRequest{
		TxID:  txID,
		Nonce: nonce,
	}
	genesisBlock, err := channel.GenesisBlock(genesisBlockRequest)
	if err != nil {
		return fmt.Errorf("Error getting genesis block: %v", err)
	}

	nonce, err = internal.GenerateRandomNonce()
	if err != nil {
		return fmt.Errorf("Could not compute nonce: %s", err)
	}
	txID, err = internal.ComputeTxID(nonce, creator)
	if err != nil {
		return fmt.Errorf("Could not compute TxID: %s", err)
	}
	joinChannelRequest := &api.JoinChannelRequest{
		Targets:      channel.Peers(),
		GenesisBlock: genesisBlock,
		TxID:         txID,
		Nonce:        nonce,
	}

	err = channel.JoinChannel(joinChannelRequest)
	if err != nil {
		return fmt.Errorf("Error joining channel: %s", err)
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
