/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"fmt"
	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/mocks"
)

const (
	caServerURL = "http://localhost:8090"
	configPath  = "testdata/config_test.yaml"
)

// TestMSP is a unit test for Client enrollment and re-enrollment scenarios
func TestMSP(t *testing.T) {

	f := textFixture{}
	sdk := f.setup()
	defer f.close()

	ctxProvider := sdk.Context()

	// Get the Client.
	// Without WithOrg option, it uses default client organization.
	msp, err := New(ctxProvider)
	if err != nil {
		t.Fatalf("failed to create CA client: %v", err)
	}

	// Empty enrollment ID
	err = msp.Enroll("", WithSecret("user1"))
	if err == nil {
		t.Fatalf("Enroll should return error for empty enrollment ID")
	}

	// Empty enrollment secret
	err = msp.Enroll("enrolledUsername", WithSecret(""))
	if err == nil {
		t.Fatalf("Enroll should return error for empty enrollment secret")
	}

	// Successful enrollment scenario

	enrollUsername := randomUsername()

	_, err = msp.GetSigningIdentity(enrollUsername)
	if err != ErrUserNotFound {
		t.Fatalf("Expected to not find user")
	}

	err = msp.Enroll(enrollUsername, WithSecret("enrollmentSecret"))
	if err != nil {
		t.Fatalf("Enroll return error %v", err)
	}

	_, err = msp.GetSigningIdentity(enrollUsername)
	if err != nil {
		t.Fatalf("Expected to find user")
	}

	enrolledUser, err := msp.GetSigningIdentity(enrollUsername)
	if err != nil {
		t.Fatalf("Expected to find user")
	}

	if enrolledUser.Identifier().ID != enrollUsername {
		t.Fatalf("Enrolled user name doesn't match")
	}

	if enrolledUser.Identifier().MSPID != "Org1MSP" {
		t.Fatalf("Enrolled user mspID doesn't match")
	}

	// Reenroll with empty user
	err = msp.Reenroll("")
	if err == nil {
		t.Fatalf("Expected error with enpty user")
	}
	if err.Error() != "user name missing" {
		t.Fatalf("Expected error user required. Got: %s", err.Error())
	}

	// Reenroll with appropriate user
	err = msp.Reenroll(enrolledUser.Identifier().ID)
	if err != nil {
		t.Fatalf("Reenroll return error %v", err)
	}

	// Try with a non-default org
	msp, err = New(ctxProvider, WithOrg("Org2"))
	if err != nil {
		t.Fatalf("failed to create CA client: %v", err)
	}

	org2lUsername := randomUsername()

	err = msp.Enroll(org2lUsername, WithSecret("enrollmentSecret"))
	if err != nil {
		t.Fatalf("Enroll return error %v", err)
	}

	org2EnrolledUser, err := msp.GetSigningIdentity(org2lUsername)
	if err != nil {
		t.Fatalf("Expected to find user")
	}

	if org2EnrolledUser.Identifier().ID != org2lUsername {
		t.Fatalf("Enrolled user name doesn't match")
	}

	if org2EnrolledUser.Identifier().MSPID != "Org2MSP" {
		t.Fatalf("Enrolled user mspID doesn't match")
	}

}

type textFixture struct {
	config core.Config
}

var caServer = &mocks.MockFabricCAServer{}

func (f *textFixture) setup() *fabsdk.FabricSDK {

	configProvider := config.FromFile(configPath)

	// Instantiate the SDK
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		panic(fmt.Sprintf("SDK init failed: %v", err))
	}

	f.config, err = config.FromFile(configPath)()
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %v", err))
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store
	cleanup(f.config.KeyStorePath())
	cleanup(f.config.CredentialStorePath())

	ctxProvider := sdk.Context()
	ctx, err := ctxProvider()
	if err != nil {
		panic(fmt.Sprintf("Failed to init context: %v", err))
	}
	// Start Http Server if it's not running
	caServer.Start(strings.TrimPrefix(caServerURL, "http://"), ctx.CryptoSuite())

	return sdk
}

func (f *textFixture) close() {
	cleanup(f.config.CredentialStorePath())
	cleanup(f.config.KeyStorePath())
}

func cleanup(storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to remove dir %s: %v\n", storePath, err))
	}
}

func myMSPID(t *testing.T, c core.Config) string {

	clientConfig, err := c.Client()
	if err != nil {
		t.Fatalf("network config retrieval failed: %v", err)
	}

	netConfig, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("network config retrieval failed: %v", err)
	}

	orgConfig, ok := netConfig.Organizations[strings.ToLower(clientConfig.Organization)]
	if !ok {
		t.Fatalf("org config retrieval failed: %v", err)
	}
	return orgConfig.MSPID
}

func randomUsername() string {
	return "user" + strconv.Itoa(rand.Intn(500000))
}
