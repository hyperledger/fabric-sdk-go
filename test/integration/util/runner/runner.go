/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package runner

import (
	"fmt"
	"os"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

const (
	org1Name      = "Org1"
	org2Name      = "Org2"
	org1AdminUser = "Admin"
	org2AdminUser = "Admin"
	org1User      = "User1"
	org2User      = "User1"
	channelID     = "mychannel"
	ccPath        = "github.com/example_cc"
)

// Runner provides common data for running integration tests.
type Runner struct {
	Org1Name           string
	Org2Name           string
	Org1AdminUser      string
	Org2AdminUser      string
	Org1User           string
	Org2User           string
	ChannelID          string
	CCPath             string
	sdk                *fabsdk.FabricSDK
	testSetup          *integration.BaseSetupImpl
	installExampleCC   bool
	exampleChaincodeID string
}

// New constructs a Runner instance using defaults.
func New() *Runner {
	r := Runner{
		Org1Name:      org1Name,
		Org2Name:      org2Name,
		Org1AdminUser: org1AdminUser,
		Org2AdminUser: org2AdminUser,
		Org1User:      org1User,
		Org2User:      org2User,
		ChannelID:     channelID,
		CCPath:        ccPath,
	}

	return &r
}

// NewWithExampleCC constructs a Runner instance using defaults and configures to install example CC.
func NewWithExampleCC() *Runner {
	r := New()
	r.installExampleCC = true

	return r
}

// Run executes the test suite against ExampleCC.
func (r *Runner) Run(m *testing.M) {
	gr := m.Run()
	r.teardown()
	os.Exit(gr)
}

// SDK returns the instantiated SDK instance. Panics if nil.
func (r *Runner) SDK() *fabsdk.FabricSDK {
	if r.sdk == nil {
		panic("SDK not instantiated")
	}

	return r.sdk
}

// TestSetup returns the integration test setup.
func (r *Runner) TestSetup() *integration.BaseSetupImpl {
	return r.testSetup
}

// ExampleChaincodeID returns the generated chaincode ID for example CC.
func (r *Runner) ExampleChaincodeID() string {
	return r.exampleChaincodeID
}

// Initialize prepares for the test run.
func (r *Runner) Initialize() {
	r.testSetup = &integration.BaseSetupImpl{
		ChannelID:           r.ChannelID,
		OrgID:               r.Org1Name,
		ChannelConfigTxFile: integration.GetChannelConfigTxPath(r.ChannelID + ".tx"),
	}

	sdk, err := fabsdk.New(integration.ConfigBackend)
	if err != nil {
		panic(fmt.Sprintf("Failed to create new SDK: %s", err))
	}
	r.sdk = sdk

	// Delete all private keys from the crypto suite store
	// and users from the user store
	integration.CleanupUserData(nil, sdk)

	if err := r.testSetup.Initialize(sdk); err != nil {
		panic(err.Error())
	}

	if r.installExampleCC {
		r.exampleChaincodeID = integration.GenerateExampleID(false)
		if err := integration.PrepareExampleCC(sdk, fabsdk.WithUser("Admin"), r.testSetup.OrgID, r.exampleChaincodeID); err != nil {
			panic(fmt.Sprintf("PrepareExampleCC return error: %s", err))
		}
	}
}

func (r *Runner) teardown() {
	integration.CleanupUserData(nil, r.sdk)
	r.sdk.Close()
}
