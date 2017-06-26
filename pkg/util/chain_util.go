/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package util

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	api "github.com/hyperledger/fabric-sdk-go/api"
	"github.com/hyperledger/fabric/common/crypto"

	orderer "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	peer "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"

	"github.com/hyperledger/fabric/protos/common"
	protos_utils "github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"
)

var origGoPath = os.Getenv("GOPATH")
var logger = logging.MustGetLogger("fabric_sdk_go")

// GetChannel initializes and returns a channel based on config
func GetChannel(client api.FabricClient, channelID string) (api.Channel, error) {

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("NewChannel return error: %v", err)
	}
	orderer, err := orderer.NewOrderer(fmt.Sprintf("%s:%s", client.GetConfig().GetOrdererHost(), client.GetConfig().GetOrdererPort()),
		client.GetConfig().GetOrdererTLSCertificate(), client.GetConfig().GetOrdererTLSServerHostOverride(), client.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("NewOrderer return error: %v", err)
	}
	err = channel.AddOrderer(orderer)
	if err != nil {
		return nil, fmt.Errorf("Error adding orderer: %v", err)
	}

	peerConfig, err := client.GetConfig().GetPeersConfig()
	if err != nil {
		return nil, fmt.Errorf("Error reading peer config: %v", err)
	}
	for _, p := range peerConfig {
		endorser, err := peer.NewPeerTLSFromCert(fmt.Sprintf("%s:%d", p.Host, p.Port),
			p.TLS.Certificate, p.TLS.ServerHostOverride, client.GetConfig())
		if err != nil {
			return nil, fmt.Errorf("NewPeer return error: %v", err)
		}
		err = channel.AddPeer(endorser)
		if err != nil {
			return nil, fmt.Errorf("Error adding peer: %v", err)
		}
		if p.Primary {
			channel.SetPrimaryPeer(endorser)
		}
	}

	return channel, nil
}

// SendInstallCC  Sends an install proposal to one or more endorsing peers.
func SendInstallCC(client api.FabricClient, channel api.Channel, chainCodeID string, chainCodePath string, chainCodeVersion string, chaincodePackage []byte, targets []api.Peer, deployPath string) error {
	ChangeGOPATHToDeploy(deployPath)
	transactionProposalResponse, _, err := client.InstallChaincode(chainCodeID, chainCodePath, chainCodeVersion, chaincodePackage, targets)
	ResetGOPATH()
	if err != nil {
		return fmt.Errorf("InstallChaincode return error: %v", err)
	}
	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("InstallChaincode Endorser %s return error: %v", v.Endorser, v.Err)
		}
		logger.Debugf("InstallChaincode Endorser '%s' return ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	return nil

}

// SendInstantiateCC Sends instantiate CC proposal to one or more endorsing peers
func SendInstantiateCC(channel api.Channel, chainCodeID string, channelID string, args []string, chaincodePath string, chaincodeVersion string, targets []api.Peer, eventHub api.EventHub) error {

	transactionProposalResponse, txID, err := channel.SendInstantiateProposal(chainCodeID, channelID, args, chaincodePath, chaincodeVersion, targets)
	if err != nil {
		return fmt.Errorf("SendInstantiateProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("SendInstantiateProposal Endorser %s return error: %v", v.Endorser, v.Err)
		}
		logger.Debug("SendInstantiateProposal Endorser '%s' return ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	// Register for commit event
	done, fail := RegisterTxEvent(txID, eventHub)

	if _, err = CreateAndSendTransaction(channel, transactionProposalResponse); err != nil {
		return fmt.Errorf("CreateTransaction return error: %v", err)
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

// CreateAndJoinChannel creates the channel represented by this channel
// and makes the primary peer join it. It reads channel configuration from tx channelConfig file
func CreateAndJoinChannel(client api.FabricClient, ordererUser api.User, orgUser api.User, channel api.Channel, channelConfig string) error {
	// Check if primary peer has joined this channel
	var foundChannel bool
	primaryPeer := channel.GetPrimaryPeer()
	client.SetUserContext(orgUser)
	response, err := client.QueryChannels(primaryPeer)
	if err != nil {
		return fmt.Errorf("Error querying channels for primary peer: %s", err)
	}
	for _, responseChannel := range response.Channels {
		if responseChannel.ChannelId == channel.GetName() {
			foundChannel = true
		}
	}

	if foundChannel {
		// There's no need to create a channel, initialize the channel from the orderer and return
		if err := channel.Initialize(nil); err != nil {
			return fmt.Errorf("Error initializing channel: %v", err)
		}
		return nil
	}

	logger.Infof("***** Creating and Joining channel: %s *****\n", channel.GetName())

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
	nonce, err := generateRandomNonce()
	if err != nil {
		return fmt.Errorf("Could not compute nonce: %s", err)
	}
	txID, err := computeTxID(nonce, creator)
	if err != nil {
		return fmt.Errorf("Could not compute TxID: %s", err)
	}

	request := api.CreateChannelRequest{
		Name:       channel.GetName(),
		Orderer:    channel.GetOrderers()[0],
		Config:     config,
		Signatures: configSignatures,
		TxID:       txID,
		Nonce:      nonce,
	}

	client.SetUserContext(ordererUser)
	err = client.CreateChannel(&request)
	if err != nil {
		return fmt.Errorf("CreateChannel returned error")
	}

	// Wait for orderer to process channel metadata
	time.Sleep(time.Second * 3)

	client.SetUserContext(orgUser)

	nonce, err = generateRandomNonce()
	if err != nil {
		return fmt.Errorf("Could not compute nonce: %s", err)
	}
	txID, err = computeTxID(nonce, creator)
	if err != nil {
		return fmt.Errorf("Could not compute TxID: %s", err)
	}

	genesisBlockRequest := &api.GenesisBlockRequest{
		TxID:  txID,
		Nonce: nonce,
	}
	genesisBlock, err := channel.GetGenesisBlock(genesisBlockRequest)
	if err != nil {
		return fmt.Errorf("Error getting genesis block: %v", err)
	}

	nonce, err = generateRandomNonce()
	if err != nil {
		return fmt.Errorf("Could not compute nonce: %s", err)
	}
	txID, err = computeTxID(nonce, creator)
	if err != nil {
		return fmt.Errorf("Could not compute TxID: %s", err)
	}
	joinChannelRequest := &api.JoinChannelRequest{
		Targets:      channel.GetPeers(),
		GenesisBlock: genesisBlock,
		TxID:         txID,
		Nonce:        nonce,
	}
	err = channel.JoinChannel(joinChannelRequest)
	if err != nil {
		return fmt.Errorf("Error joining channel: %s", err)
	}

	logger.Infof("***** Created and Joined channel: %s *****\n", channel.GetName())

	return nil
}

// Utility to create random string of strlen length
func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// ChangeGOPATHToDeploy changes go path to fixtures folder
func ChangeGOPATHToDeploy(deployPath string) {
	os.Setenv("GOPATH", deployPath)
}

// ResetGOPATH resets go path to original
func ResetGOPATH() {
	os.Setenv("GOPATH", origGoPath)
}

// GenerateRandomID generates random ID
func GenerateRandomID() string {
	rand.Seed(time.Now().UnixNano())
	return randomString(10)
}

// generateRandomNonce generates a random nonce
func generateRandomNonce() ([]byte, error) {
	return crypto.GetRandomNonce()
}

// computeTxID computes a transaction ID from a given nonce and creator ID
func computeTxID(nonce []byte, creator []byte) (string, error) {
	return protos_utils.ComputeProposalTxID(nonce, creator)
}
