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
	"os"
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/config"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/fabric-ca-client"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/events"
	"github.com/hyperledger/fabric/bccsp"

	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
	kvs "github.com/hyperledger/fabric-sdk-go/fabric-client/keyvaluestore"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// BaseTestSetup is an interface used by the integration tests
// it performs setup activities like user enrollment, chain creation,
// crypto suite selection, and event hub initialization
type BaseTestSetup interface {
	GetChains(t *testing.T) (*fabricClient.Chain, *fabricClient.Chain)
	GetEventHub(t *testing.T, interestedEvents []*pb.Interest) *events.EventHub
}

// BaseSetupImpl implementation of BaseTestSetup
type BaseSetupImpl struct {
}

// GetChains initializes and returns a query chain and invoke chain
func (setup *BaseSetupImpl) GetChains(t *testing.T) (fabricClient.Chain, fabricClient.Chain, fabricClient.Chain) {
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
		t.Fatalf("Failed getting ephemeral software-based BCCSP [%s]", err)
	}
	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
	stateStore, err := kvs.CreateNewFileKeyValueStore("/tmp/enroll_user")
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	client.SetStateStore(stateStore)
	user, err := client.GetUserContext("admin")
	if err != nil {
		t.Fatalf("client.GetUserContext return error: %v", err)
	}
	if user == nil {
		fabricCAClient, err1 := fabricCAClient.NewFabricCAClient()
		if err1 != nil {
			t.Fatalf("NewFabricCAClient return error: %v", err)
		}
		key, cert, err1 := fabricCAClient.Enroll("admin", "adminpw")
		keyPem, _ := pem.Decode(key)
		if err1 != nil {
			t.Fatalf("Enroll return error: %v", err1)
		}
		user := fabricClient.NewUser("admin")
		k, err1 := client.GetCryptoSuite().KeyImport(keyPem.Bytes, &bccsp.ECDSAPrivateKeyImportOpts{Temporary: false})
		if err1 != nil {
			t.Fatalf("KeyImport return error: %v", err)
		}
		user.SetPrivateKey(k)
		user.SetEnrollmentCertificate(cert)
		err = client.SetUserContext(user, false)
		if err != nil {
			t.Fatalf("client.SetUserContext return error: %v", err)
		}
	}

	querychain, err := client.NewChain("mychannel")
	if err != nil {
		t.Fatalf("NewChain return error: %v", err)
	}

	invokechain, err := client.NewChain("invokechain")
	if err != nil {
		t.Fatalf("NewChain return error: %v", err)
	}
	orderer, err := fabricClient.CreateNewOrderer(fmt.Sprintf("%s:%s", config.GetOrdererHost(), config.GetOrdererPort()),
		config.GetOrdererTLSCertificate(), config.GetOrdererTLSServerHostOverride())
	if err != nil {
		t.Fatalf("CreateNewOrderer return error: %v", err)
	}
	invokechain.AddOrderer(orderer)

	for _, p := range config.GetPeersConfig() {
		endorser, err := fabricClient.CreateNewPeer(fmt.Sprintf("%s:%s", p.Host, p.Port), p.TLSCertificate, p.TLSServerHostOverride)
		if err != nil {
			t.Fatalf("CreateNewPeer return error: %v", err)
		}
		querychain.AddPeer(endorser)
		invokechain.AddPeer(endorser)
		break
	}

	deploychain, err := client.NewChain("deploychain")
	if err != nil {
		t.Fatalf("NewChain return error: %v", err)
	}
	orderer, err = fabricClient.CreateNewOrderer(fmt.Sprintf("%s:%s", config.GetOrdererHost(), config.GetOrdererPort()),
		config.GetOrdererTLSCertificate(), config.GetOrdererTLSServerHostOverride())
	if err != nil {
		t.Fatalf("CreateNewOrderer return error: %v", err)
	}
	deploychain.AddOrderer(orderer)

	for _, p := range config.GetPeersConfig() {
		endorser, err := fabricClient.CreateNewPeer(fmt.Sprintf("%s:%s", p.Host, p.Port), p.TLSCertificate, p.TLSServerHostOverride)
		if err != nil {
			t.Fatalf("CreateNewPeer return error: %v", err)
		}
		deploychain.AddPeer(endorser)
	}

	return querychain, invokechain, deploychain

}

// GetEventHub initilizes the event hub
func (setup *BaseSetupImpl) GetEventHub(t *testing.T,
	interestedEvents []*pb.Interest) events.EventHub {
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
		t.Fatalf("No EventHub configuration found")
	}

	// TODO: this is coming back in some other form
	/*if interestedEvents != nil {
		eventHub.SetInterestedEvents(interestedEvents)
	}*/
	if err := eventHub.Connect(); err != nil {
		t.Fatalf("Failed eventHub.Connect() [%s]", err)
	}

	return eventHub
}

// SetupChaincodeDeploy set up environment
func (setup *BaseSetupImpl) SetupChaincodeDeploy() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Setenv("GOPATH", path.Join(pwd, "../fixtures"))
}
