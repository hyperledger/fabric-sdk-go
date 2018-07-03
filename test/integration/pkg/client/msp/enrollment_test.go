/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

const (
	IdentityTypeUser = "user"
)

func TestRegisterEnroll(t *testing.T) {

	// Instantiate the SDK
	sdk, err := fabsdk.New(integration.ConfigBackend)

	if err != nil {
		t.Fatalf("SDK init failed: %s", err)
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	ctxProvider := sdk.Context()

	// Get the Client.
	// Without WithOrg option, uses default client organization.
	mspClient, err := msp.New(ctxProvider)
	if err != nil {
		t.Fatalf("failed to create CA client: %s", err)
	}

	// As this integration test spawns a fresh CA instance,
	// we have to enroll the CA registrar first. Otherwise,
	// CA operations that require the registrar's identity
	// will be rejected by the CA.
	registrarEnrollID, registrarEnrollSecret := getRegistrarEnrollmentCredentials(t, ctxProvider)
	err = mspClient.Enroll(registrarEnrollID, msp.WithSecret(registrarEnrollSecret))
	if err != nil {
		t.Fatalf("Enroll failed: %s", err)
	}

	// The enrollment process generates a new private key and
	// enrollment certificate for the user. The private key
	// is stored in the SDK crypto provider's key store, while the
	// enrollment certificate is stored in the SKD's user store
	// (state store). The CAClient will lookup the
	// registrar's identity information in these stores.

	// Generate a random user name
	username := integration.GenerateRandomID()

	testAttributes := []msp.Attribute{
		{
			Name:  integration.GenerateRandomID(),
			Value: fmt.Sprintf("%s:ecert", integration.GenerateRandomID()),
			ECert: true,
		},
		{
			Name:  integration.GenerateRandomID(),
			Value: fmt.Sprintf("%s:ecert", integration.GenerateRandomID()),
			ECert: true,
		},
	}

	// Register the new user
	enrollmentSecret, err := mspClient.Register(&msp.RegistrationRequest{
		Name:       username,
		Type:       IdentityTypeUser,
		Attributes: testAttributes,
		// Affiliation is mandatory. "org1" and "org2" are hardcoded as CA defaults
		// See https://github.com/hyperledger/fabric-ca/blob/release/cmd/fabric-ca-server/config.go
		Affiliation: "org2",
	})
	if err != nil {
		t.Fatalf("Registration failed: %s", err)
	}

	// Enroll the new user
	err = mspClient.Enroll(username, msp.WithSecret(enrollmentSecret))
	if err != nil {
		t.Fatalf("Enroll failed: %s", err)
	}

	// Get the new user's signing identity
	si, err := mspClient.GetSigningIdentity(username)
	if err != nil {
		t.Fatalf("GetSigningIdentity failed: %s", err)
	}

	checkCertAttributes(t, si.EnrollmentCertificate(), testAttributes)

}

func getRegistrarEnrollmentCredentials(t *testing.T, ctxProvider context.ClientProvider) (string, string) {

	ctx, err := ctxProvider()
	if err != nil {
		t.Fatalf("failed to get context: %s", err)
	}

	myOrg := ctx.IdentityConfig().Client().Organization

	caConfig, ok := ctx.IdentityConfig().CAConfig(myOrg)
	if !ok {
		t.Fatal("CAConfig failed")
	}

	return caConfig.Registrar.EnrollID, caConfig.Registrar.EnrollSecret
}
