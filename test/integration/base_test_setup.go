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

package integration

import (
	"encoding/pem"
	"fmt"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/config"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/fabric-ca-client"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/events"
	"github.com/hyperledger/fabric/bccsp"

	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
	kvs "github.com/hyperledger/fabric-sdk-go/fabric-client/keyvaluestore"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var chainCodeID = ""
var chainID = "mychannel"
var chainCodePath = "github.com/example_cc"
var chainCodeVersion = "v0"
var goPath string

// BaseTestSetup is an interface used by the integration tests
// it performs setup activities like user enrollment, chain creation,
// crypto suite selection, and event hub initialization
type BaseTestSetup interface {
	GetChain() (fabricClient.Chain, error)
	GetEventHub(interestedEvents []*pb.Interest) (events.EventHub, error)
	InstallCC(chain fabricClient.Chain, chainCodeID string, chainCodePath string, chainCodeVersion string, chaincodePackage []byte, targets []fabricClient.Peer) error
	InstantiateCC(chain fabricClient.Chain, eventHub events.EventHub) error
	GetQueryValue(t *testing.T, chain fabricClient.Chain) (string, error)
	Invoke(chain fabricClient.Chain, eventHub events.EventHub) (string, error)
	InitConfig()
	ChangeGOPATHToDeploy()
	ResetGOPATH()
	GenerateRandomCCID()
}

// BaseSetupImpl implementation of BaseTestSetup
type BaseSetupImpl struct {
}

// GetChain initializes and returns a chain
func (setup *BaseSetupImpl) GetChain() (fabricClient.Chain, error) {
	client := fabricClient.NewClient()

	err := bccspFactory.InitFactories(&bccspFactory.FactoryOpts{
		ProviderName: "SW",
		SwOpts: &bccspFactory.SwOpts{
			HashFamily: config.GetSecurityAlgorithm(),
			SecLevel:   config.GetSecurityLevel(),
			FileKeystore: &bccspFactory.FileKeystoreOpts{
				KeyStorePath: config.GetKeyStorePath(),
			},
			Ephemeral: false,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Failed getting ephemeral software-based BCCSP [%s]", err)
	}
	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
	stateStore, err := kvs.CreateNewFileKeyValueStore("/tmp/enroll_user")
	if err != nil {
		return nil, fmt.Errorf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	client.SetStateStore(stateStore)
	user, err := client.GetUserContext("admin")
	if err != nil {
		return nil, fmt.Errorf("client.GetUserContext return error: %v", err)
	}
	if user == nil {
		fabricCAClient, err1 := fabricCAClient.NewFabricCAClient()
		if err1 != nil {
			return nil, fmt.Errorf("NewFabricCAClient return error: %v", err)
		}
		key, cert, err1 := fabricCAClient.Enroll("admin", "adminpw")
		keyPem, _ := pem.Decode(key)
		if err1 != nil {
			return nil, fmt.Errorf("Enroll return error: %v", err1)
		}
		user := fabricClient.NewUser("admin")
		k, err1 := client.GetCryptoSuite().KeyImport(keyPem.Bytes, &bccsp.ECDSAPrivateKeyImportOpts{Temporary: false})
		if err1 != nil {
			return nil, fmt.Errorf("KeyImport return error: %v", err)
		}
		user.SetPrivateKey(k)
		user.SetEnrollmentCertificate(cert)
		err = client.SetUserContext(user, false)
		if err != nil {
			return nil, fmt.Errorf("client.SetUserContext return error: %v", err)
		}
	}

	chain, err := client.NewChain(chainID)
	if err != nil {
		return nil, fmt.Errorf("NewChain return error: %v", err)
	}
	orderer, err := fabricClient.CreateNewOrderer(fmt.Sprintf("%s:%s", config.GetOrdererHost(), config.GetOrdererPort()),
		config.GetOrdererTLSCertificate(), config.GetOrdererTLSServerHostOverride())
	if err != nil {
		return nil, fmt.Errorf("CreateNewOrderer return error: %v", err)
	}
	chain.AddOrderer(orderer)

	for _, p := range config.GetPeersConfig() {
		endorser, err := fabricClient.CreateNewPeer(fmt.Sprintf("%s:%s", p.Host, p.Port), p.TLSCertificate, p.TLSServerHostOverride)
		if err != nil {
			return nil, fmt.Errorf("CreateNewPeer return error: %v", err)
		}
		chain.AddPeer(endorser)
		if p.Port == "7051" {
			chain.SetPrimaryPeer(endorser)
		}
	}

	return chain, nil

}

// GetEventHub initilizes the event hub
func (setup *BaseSetupImpl) GetEventHub(interestedEvents []*pb.Interest) (events.EventHub, error) {
	eventHub := events.NewEventHub()
	foundEventHub := false
	for _, p := range config.GetPeersConfig() {
		if p.EventHost != "" && p.EventPort != "" {
			eventHub.SetPeerAddr(fmt.Sprintf("%s:%s", p.EventHost, p.EventPort), p.TLSCertificate, p.TLSServerHostOverride)
			foundEventHub = true
			break
		}
	}

	if !foundEventHub {
		return nil, fmt.Errorf("No EventHub configuration found")
	}

	// TODO: this is coming back in some other form
	/*if interestedEvents != nil {
		eventHub.SetInterestedEvents(interestedEvents)
	}*/
	if err := eventHub.Connect(); err != nil {
		return nil, fmt.Errorf("Failed eventHub.Connect() [%s]", err)
	}

	return eventHub, nil
}

// InstallCC ...
func (setup *BaseSetupImpl) InstallCC(chain fabricClient.Chain, chainCodeID string, chainCodePath string, chainCodeVersion string, chaincodePackage []byte, targets []fabricClient.Peer) error {
	setup.ChangeGOPATHToDeploy()
	transactionProposalResponse, _, err := chain.SendInstallProposal(chainCodeID, chainCodePath, chainCodeVersion, chaincodePackage, targets)
	if err != nil {
		return fmt.Errorf("SendInstallProposal return error: %v", err)
	}
	setup.ResetGOPATH()

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("SendInstallProposal Endorser %s return error: %v", v.Endorser, v.Err)
		}
		fmt.Printf("SendInstallProposal Endorser '%s' return ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	return nil

}

// InstantiateCC ...
func (setup *BaseSetupImpl) InstantiateCC(chain fabricClient.Chain, eventHub events.EventHub) error {

	var args []string
	args = append(args, "init")
	args = append(args, "a")
	args = append(args, "100")
	args = append(args, "b")
	args = append(args, "200")

	transactionProposalResponse, txID, err := chain.SendInstantiateProposal(chainCodeID, chainID, args, chainCodePath, chainCodeVersion, nil)
	if err != nil {
		return fmt.Errorf("SendInstantiateProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("SendInstantiateProposal Endorser %s return error: %v", v.Endorser, v.Err)
		}
		fmt.Printf("SendInstantiateProposal Endorser '%s' return ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	tx, err := chain.CreateTransaction(transactionProposalResponse)
	if err != nil {
		return fmt.Errorf("CreateTransaction return error: %v", err)

	}
	transactionResponse, err := chain.SendTransaction(tx)
	if err != nil {
		return fmt.Errorf("SendTransaction return error: %v", err)

	}
	for _, v := range transactionResponse {
		if v.Err != nil {
			return fmt.Errorf("Orderer %s return error: %v", v.Orderer, v.Err)
		}
	}
	done := make(chan bool)
	fail := make(chan error)

	eventHub.RegisterTxEvent(txID, func(txId string, err error) {
		if err != nil {
			fail <- err
		} else {
			fmt.Printf("instantiateCC receive success event for txid(%s)\n", txId)
			done <- true
		}

	})

	select {
	case <-done:
	case <-fail:
		return fmt.Errorf("instantiateCC Error received from eventhub for txid(%s) error(%v)", txID, fail)
	case <-time.After(time.Second * 30):
		return fmt.Errorf("instantiateCC Didn't receive block event for txid(%s)", txID)
	}
	return nil

}

// GetQueryValue ...
func (setup *BaseSetupImpl) GetQueryValue(t *testing.T, chain fabricClient.Chain) (string, error) {

	var args []string
	args = append(args, "invoke")
	args = append(args, "query")
	args = append(args, "b")

	signedProposal, err := chain.CreateTransactionProposal(chainCodeID, chainID, args, true, nil)
	if err != nil {
		return "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}
	transactionProposalResponses, err := chain.SendTransactionProposal(signedProposal, 0, []fabricClient.Peer{chain.GetPrimaryPeer()})
	if err != nil {
		return "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			return "", fmt.Errorf("query Endorser %s return error: %v", v.Endorser, v.Err)
		}
		return string(v.GetResponsePayload()), nil
	}
	return "", nil
}

// Invoke ...
func (setup *BaseSetupImpl) Invoke(chain fabricClient.Chain, eventHub events.EventHub) (string, error) {

	var args []string
	args = append(args, "invoke")
	args = append(args, "move")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "1")

	signedProposal, err := chain.CreateTransactionProposal(chainCodeID, chainID, args, true, nil)
	if err != nil {
		return "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}
	transactionProposalResponse, err := chain.SendTransactionProposal(signedProposal, 0, []fabricClient.Peer{chain.GetPrimaryPeer()})
	if err != nil {
		return "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return "", fmt.Errorf("invoke Endorser %s return error: %v", v.Endorser, v.Err)
		}
		fmt.Printf("invoke Endorser '%s' return ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	tx, err := chain.CreateTransaction(transactionProposalResponse)
	if err != nil {
		return "", fmt.Errorf("CreateTransaction return error: %v", err)

	}
	transactionResponse, err := chain.SendTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction return error: %v", err)

	}
	for _, v := range transactionResponse {
		if v.Err != nil {
			return "", fmt.Errorf("Orderer %s return error: %v", v.Orderer, v.Err)
		}
	}
	done := make(chan bool)
	fail := make(chan error)
	eventHub.RegisterTxEvent(signedProposal.TransactionID, func(txId string, err error) {
		if err != nil {
			fail <- err
		} else {
			fmt.Printf("invoke receive success event for txid(%s)\n", txId)
			done <- true
		}
	})

	select {
	case <-done:
	case <-fail:
		return "", fmt.Errorf("invoke Error received from eventhub for txid(%s) error(%v)", signedProposal.TransactionID, fail)
	case <-time.After(time.Second * 30):
		return "", fmt.Errorf("invoke Didn't receive block event for txid(%s)", signedProposal.TransactionID)
	}
	return signedProposal.TransactionID, nil

}

func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// ChangeGOPATHToDeploy ...
func (setup *BaseSetupImpl) ChangeGOPATHToDeploy() {
	goPath = os.Getenv("GOPATH")
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Setenv("GOPATH", path.Join(pwd, "../fixtures"))
}

// ResetGOPATH ...
func (setup *BaseSetupImpl) ResetGOPATH() {
	os.Setenv("GOPATH", goPath)
}

// InitConfig ...
func (setup *BaseSetupImpl) InitConfig() {
	err := config.InitConfig("../fixtures/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
	setup.GenerateRandomCCID()
}

// GenerateRandomCCID ...
func (setup *BaseSetupImpl) GenerateRandomCCID() {
	rand.Seed(time.Now().UnixNano())
	chainCodeID = randomString(10)
}
