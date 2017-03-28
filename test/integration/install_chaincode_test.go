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
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
)

var chain fabricClient.Chain

// Test chaincode install using chaincodePath to create chaincodePackage
func TestChaincodeInstallUsingChaincodePath(t *testing.T) {
	testSetup := BaseSetupImpl{}

	chainCodeVersion := getRandomCCVersion()
	err := testSetup.InstallCC(chain, "install", chainCodePath, chainCodeVersion, nil, nil)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	//Install same chaincode again, should fail
	err = testSetup.InstallCC(chain, "install", chainCodePath, chainCodeVersion, nil, nil)
	if err == nil {
		t.Fatalf("install same chaincode didn't return error")
	}
	fmt.Println(err.Error())
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error")
	}

}

// Test chaincode install using chaincodePackage[byte]
func TestChaincodeInstallUsingChaincodePackage(t *testing.T) {
	testSetup := BaseSetupImpl{}

	chainCodeVersion := getRandomCCVersion()
	testSetup.ChangeGOPATHToDeploy()
	chaincodePackage, err := fabricClient.PackageCC(chainCodePath, "")
	if err != nil {
		t.Fatalf("PackageCC return error: %s", err)
	}
	testSetup.ResetGOPATH()

	err = testSetup.InstallCC(chain, "install", "github.com/example_cc_pkg", chainCodeVersion, chaincodePackage, nil)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}
	//Install same chaincode again, should fail
	err = testSetup.InstallCC(chain, "install", chainCodePath, chainCodeVersion, chaincodePackage, nil)
	if err == nil {
		t.Fatalf("install same chaincode didn't return error")
	}
	fmt.Println(err.Error())
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error")
	}
}

func TestMain(m *testing.M) {
	testSetup := BaseSetupImpl{}

	testSetup.InitConfig()
	var err error
	chain, err = testSetup.GetChain()
	if err != nil {
		fmt.Printf("error from GetChains %v", err)
		os.Exit(-1)
	}
	code := m.Run()
	os.Exit(code)
}

func getRandomCCVersion() string {
	rand.Seed(time.Now().UnixNano())
	return "v0" + strconv.Itoa(rand.Intn(10000000))
}
