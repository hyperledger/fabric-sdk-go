/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helpers

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/hyperledger/fabric-sdk-go/config"

	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/util"
	"github.com/op/go-logging"
)

var origGoPath = os.Getenv("GOPATH")
var logger = logging.MustGetLogger("fabric_sdk_go")

// GetChain initializes and returns a chain based on config
func GetChain(client fabricClient.Client, chainID string) (fabricClient.Chain, error) {

	chain, err := client.NewChain(chainID)
	if err != nil {
		return nil, fmt.Errorf("NewChain return error: %v", err)
	}
	orderer, err := fabricClient.NewOrderer(fmt.Sprintf("%s:%s", config.GetOrdererHost(), config.GetOrdererPort()),
		config.GetOrdererTLSCertificate(), config.GetOrdererTLSServerHostOverride())
	if err != nil {
		return nil, fmt.Errorf("NewOrderer return error: %v", err)
	}
	chain.AddOrderer(orderer)

	peerConfig, err := config.GetPeersConfig()
	if err != nil {
		return nil, fmt.Errorf("Error reading peer config: %v", err)
	}
	for _, p := range peerConfig {
		endorser, err := fabricClient.NewPeer(fmt.Sprintf("%s:%d", p.Host, p.Port),
			p.TLS.Certificate, p.TLS.ServerHostOverride)
		if err != nil {
			return nil, fmt.Errorf("NewPeer return error: %v", err)
		}
		chain.AddPeer(endorser)
		if p.Primary {
			chain.SetPrimaryPeer(endorser)
		}
	}

	return chain, nil
}

// SendInstallCC  Sends an install proposal to one or more endorsing peers.
func SendInstallCC(client fabricClient.Client, chain fabricClient.Chain, chainCodeID string, chainCodePath string, chainCodeVersion string, chaincodePackage []byte, targets []fabricClient.Peer, deployPath string) error {
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
func SendInstantiateCC(chain fabricClient.Chain, chainCodeID string, chainID string, args []string, chaincodePath string, chaincodeVersion string, targets []fabricClient.Peer, eventHub events.EventHub) error {

	transactionProposalResponse, txID, err := chain.SendInstantiateProposal(chainCodeID, chainID, args, chaincodePath, chaincodeVersion, targets)
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

	if _, err = CreateAndSendTransaction(chain, transactionProposalResponse); err != nil {
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

// CreateAndJoinChannel creates the channel represented by this chain
// and makes the primary peer join it. It reads channel configuration from tx channelConfig file
func CreateAndJoinChannel(client fabricClient.Client, chain fabricClient.Chain, channelConfig string) error {
	// Check if primary peer has joined this channel
	var foundChannel bool
	primaryPeer := chain.GetPrimaryPeer()
	response, err := client.QueryChannels(primaryPeer)
	if err != nil {
		return fmt.Errorf("Error querying channels for primary peer: %s", err)
	}
	for _, channel := range response.Channels {
		if channel.ChannelId == chain.GetName() {
			foundChannel = true
		}
	}

	if foundChannel {
		// There's no need to create a channel, initialize the chain from the orderer and return
		if err := chain.Initialize(nil); err != nil {
			return fmt.Errorf("Error initializing chain: %v", err)
		}
		return nil
	}

	logger.Infof("***** Creating and Joining channel: %s *****\n", chain.GetName())

	configTx, err := ioutil.ReadFile(channelConfig)
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	request := fabricClient.CreateChannelRequest{
		Name:     chain.GetName(),
		Orderer:  chain.GetOrderers()[0],
		Envelope: configTx,
	}
	newChain, err := client.CreateChannel(&request)
	if err != nil {
		return err
	}
	if newChain == nil {
		return fmt.Errorf("CreateChannel returned nil chain")
	}

	// Wait for orderer to process channel metadata
	time.Sleep(time.Second * 2)
	// Test join channel
	creator, err := GetCreatorID(client)
	if err != nil {
		return fmt.Errorf("Could not generate creator ID: %v", err)
	}
	nonce, err := util.GenerateRandomNonce()
	if err != nil {
		return fmt.Errorf("Could not compute nonce: %s", err)
	}
	txID, err := util.ComputeTxID(nonce, creator)
	if err != nil {
		return fmt.Errorf("Could not compute TxID: %s", err)
	}

	req := &fabricClient.JoinChannelRequest{
		Targets: []fabricClient.Peer{chain.GetPrimaryPeer()}, TxID: txID, Nonce: nonce}
	if err = chain.JoinChannel(req); err != nil {
		return err
	}

	logger.Infof("***** Created and Joined channel: %s *****\n", chain.GetName())

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
