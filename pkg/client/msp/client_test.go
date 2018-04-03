/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"io/ioutil"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"

	"fmt"
	"os"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	mspctx "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

const (
	caServerURLListen = "http://localhost:0"
	configPath        = "testdata/config_test.yaml"
)

var caServerURL string

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

	enrolledUser := getEnrolledUser(t, msp)

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
	testWithOrg2(t, ctxProvider)

}

func testWithOrg2(t *testing.T, ctxProvider contextApi.ClientProvider) {
	msp, err := New(ctxProvider, WithOrg("Org2"))
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

func getEnrolledUser(t *testing.T, msp *Client) mspctx.SigningIdentity {
	// Successful enrollment scenario

	enrollUsername := randomUsername()

	_, err := msp.GetSigningIdentity(enrollUsername)
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
	return enrolledUser
}

type textFixture struct {
	cryptoSuiteConfig core.CryptoSuiteConfig
	identityConfig    msp.IdentityConfig
}

var caServer = &mockmsp.MockFabricCAServer{}

func (f *textFixture) setup() *fabsdk.FabricSDK {

	var lis net.Listener
	var err error
	if !caServer.Running() {
		lis, err = net.Listen("tcp", strings.TrimPrefix(caServerURLListen, "http://"))
		if err != nil {
			panic(fmt.Sprintf("Error starting CA Server %s", err))
		}

		caServerURL = "http://" + lis.Addr().String()
	}

	cfgRaw := readConfigWithReplacement(configPath, "http://localhost:8050", caServerURL)
	configProvider := config.FromRaw(cfgRaw, "yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %v", err))
	}

	// Instantiate the SDK
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		panic(fmt.Sprintf("SDK init failed: %v", err))
	}

	f.cryptoSuiteConfig, _, f.identityConfig, err = sdk.Config()()
	if err != nil {
		panic(fmt.Sprintf("Failed to get config: %v", err))
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store
	cleanup(f.cryptoSuiteConfig.KeyStorePath())
	cleanup(f.identityConfig.CredentialStorePath())

	ctxProvider := sdk.Context()
	ctx, err := ctxProvider()
	if err != nil {
		panic(fmt.Sprintf("Failed to init context: %v", err))
	}

	// Start Http Server if it's not running
	if !caServer.Running() {
		caServer.Start(lis, ctx.CryptoSuite())
	}

	return sdk
}

func (f *textFixture) close() {
	cleanup(f.identityConfig.CredentialStorePath())
	cleanup(f.cryptoSuiteConfig.KeyStorePath())
}

func readConfigWithReplacement(path string, origURL, newURL string) []byte {
	cfgRaw, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config [%s]", err))
	}

	updatedCfg := strings.Replace(string(cfgRaw), origURL, newURL, -1)
	return []byte(updatedCfg)
}

func cleanup(storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to remove dir %s: %v\n", storePath, err))
	}
}

func randomUsername() string {
	return "user" + strconv.Itoa(rand.Intn(500000))
}
