/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	mspctx "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
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
	defer sdk.Close()

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	ctxProvider := sdk.Context()

	// Test with the default org CA
	testRegisterEnrollWithCAInstance(t, ctxProvider, "")

	// Test with the second org CA instance
	testRegisterEnrollWithCAInstance(t, ctxProvider, "tlsca.org1.example.com")

}

func createMspClient(t *testing.T, ctxProvider context.ClientProvider, caInstance string) (*msp.Client, error) {
	// Get the Client.
	// Without WithOrg option, uses default client organization.
	if caInstance == "" {
		return msp.New(ctxProvider)
	} else {
		return msp.New(ctxProvider, msp.WithCAInstance(caInstance))
	}
}
func testRegisterEnrollWithCAInstance(t *testing.T, ctxProvider context.ClientProvider, caInstance string) {

	mspClient, err := createMspClient(t, ctxProvider, caInstance)
	if err != nil {
		t.Fatalf("failed to create CA client: %s", err)
	}

	// As this integration test spawns a fresh CA instance,
	// we have to enroll the CA registrar first. Otherwise,
	// CA operations that require the registrar's identity
	// will be rejected by the CA.
	registrarEnrollID, registrarEnrollSecret := getRegistrarEnrollmentCredentialsWithCAInstance(t, ctxProvider, caInstance)
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

	revokeResponse, err := mspClient.Revoke(&msp.RevocationRequest{Name: username, GenCRL: true})
	if err != nil {
		t.Fatalf("Revoke return error %s", err)
	}
	if revokeResponse.CRL == nil {
		t.Fatal("Couldn't retrieve CRL")
	}
	ok, err := isInCRL(si.EnrollmentCertificate(), revokeResponse.CRL)
	if err != nil {
		t.Fatalf("Couldn't check if certificate is in CRL %s", err)
	}
	if !ok {
		t.Fatal("Certificate is not in CRL")
	}

}

func isInCRL(certBytes, crlBytes []byte) (bool, error) {
	decoded, _ := pem.Decode(certBytes)
	if decoded == nil {
		return false, errors.New("Failed cert decoding")
	}
	cert, err := x509.ParseCertificate(decoded.Bytes)
	if err != nil {
		return false, err
	}
	crl, err := x509.ParseCRL(crlBytes)
	if err != nil {
		return false, err
	}
	for _, revokedCert := range crl.TBSCertList.RevokedCertificates {
		if cert.SerialNumber.Cmp(revokedCert.SerialNumber) == 0 {
			return true, nil
		}
	}
	return false, nil
}

func TestEnrollWithOptions(t *testing.T) {
	// Instantiate the SDK
	sdk, err := fabsdk.New(integration.ConfigBackend)
	if err != nil {
		t.Fatalf("SDK init failed: %s", err)
	}
	defer sdk.Close()

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

	err = mspClient.Enroll(username, msp.WithSecret(enrollmentSecret), msp.WithType("idemix"))
	if err == nil {
		t.Fatal("Enroll should failed: idemix is not supported")
	}

	attrReqs := []*msp.AttributeRequest{{Name: testAttributes[0].Name, Optional: true}}
	err = mspClient.Enroll(username, msp.WithSecret(enrollmentSecret), msp.WithAttributeRequests(attrReqs))
	if err != nil {
		t.Fatalf("Enroll failed: %s", err)
	}

	// Get the new user's signing identity
	si, err := mspClient.GetSigningIdentity(username)
	if err != nil {
		t.Fatalf("GetSigningIdentity failed: %s", err)
	}

	attrs, err := getCertAttributes(si.EnrollmentCertificate())
	require.NoError(t, err)

	if attrs.Contains(testAttributes[1].Name) {
		t.Fatalf("attribute '%s' shouldn't be found in in certificate", testAttributes[1].Name)
	}

	v, ok, err := attrs.Value(testAttributes[0].Name)
	require.NoError(t, err)
	require.True(t, ok, "attribute '%s' was not found", testAttributes[0].Name)
	require.True(t, v == testAttributes[0].Value, "incorrect value for '%s'; expected '%s' but found '%s'", testAttributes[0].Name, testAttributes[0].Value, v)
}

func TestEnrollWithProfile(t *testing.T) {
	// Instantiate the SDK
	sdk, err := fabsdk.New(integration.ConfigBackend)
	if err != nil {
		t.Fatalf("SDK init failed: %s", err)
	}
	defer sdk.Close()

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

	// Generate a random user name
	username := integration.GenerateRandomID()

	// Register the new user
	enrollmentSecret, err := mspClient.Register(&msp.RegistrationRequest{
		Name: username,
		Type: IdentityTypeUser,
		// Affiliation is mandatory. "org1" and "org2" are hardcoded as CA defaults
		// See https://github.com/hyperledger/fabric-ca/blob/release/cmd/fabric-ca-server/config.go
		Affiliation: "org2",
	})
	if err != nil {
		t.Fatalf("Registration failed: %s", err)
	}

	err = mspClient.Enroll(username, msp.WithSecret(enrollmentSecret), msp.WithProfile("tls"))
	if err != nil {
		t.Fatalf("Enroll failed: %s", err)
	}

}

func TestEnrollWithCSR(t *testing.T) {
	// Instantiate the SDK
	sdk, err := fabsdk.New(integration.ConfigBackend)
	if err != nil {
		t.Fatalf("SDK init failed: %s", err)
	}
	defer sdk.Close()

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

	// Generate a random user name
	username := integration.GenerateRandomID()

	// Register the new user
	enrollmentSecret, err := mspClient.Register(&msp.RegistrationRequest{
		Name: username,
		Type: IdentityTypeUser,
		// Affiliation is mandatory. "org1" and "org2" are hardcoded as CA defaults
		// See https://github.com/hyperledger/fabric-ca/blob/release/cmd/fabric-ca-server/config.go
		Affiliation: "org2",
	})
	if err != nil {
		t.Fatalf("Registration failed: %s", err)
	}

	extraHosts := []string{"localhost", "example.com", "127.0.0.1"}
	csr := &msp.CSRInfo{
		CN:    username,
		Hosts: extraHosts,
	}

	err = mspClient.Enroll(username, msp.WithSecret(enrollmentSecret), msp.WithCSR(csr))
	if err != nil {
		t.Fatalf("Enroll failed: %s", err)
	}

	// Get the new user's signing identity
	si, err := mspClient.GetSigningIdentity(username)
	if err != nil {
		t.Fatalf("GetSigningIdentity failed: %s", err)
	}

	has, err := hasHosts(si, extraHosts)
	if err != nil {
		t.Fatalf("Could not check for host in Signing Identity: %s", err)
	}
	if !has {
		t.Fatalf("Missing host [%s] in Signing Identity", extraHosts[0])
	}

	err = mspClient.Reenroll(username, msp.WithSecret(enrollmentSecret), msp.WithCSR(csr))
	if err != nil {
		t.Fatalf("Reenroll failed: %s", err)
	}

	// Get the new user's signing identity
	si, err = mspClient.GetSigningIdentity(username)
	if err != nil {
		t.Fatalf("GetSigningIdentity failed: %s", err)
	}

	has, err = hasHosts(si, extraHosts)
	if err != nil {
		t.Fatalf("Could not check for host in Signing Identity at reenroll: %s", err)
	}
	if !has {
		t.Fatalf("Missing host [%s] in Signing Identity at reenroll", extraHosts[0])
	}

}

func hasHosts(si mspctx.SigningIdentity, hosts []string) (bool, error) {
	block, _ := pem.Decode(si.EnrollmentCertificate())
	if block == nil || block.Type != "CERTIFICATE" {
		return false, errors.New("Public cert invalid, cannot decode")
	}

	pub, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, errors.New("Could not decode Signing Identity certificate")
	}

	certHosts := make(map[string]struct{}, len(pub.DNSNames)+len(pub.IPAddresses))
	for _, host := range pub.DNSNames {
		certHosts[host] = struct{}{}
	}
	for _, host := range pub.IPAddresses {
		certHosts[host.String()] = struct{}{}
	}

	hasExtraHost := true
	for _, requestedHost := range hosts {
		if _, ok := certHosts[requestedHost]; !ok {
			hasExtraHost = false
			break
		}
	}

	return hasExtraHost, nil
}
func getRegistrarEnrollmentCredentials(t *testing.T, ctxProvider context.ClientProvider) (string, string) {

	return getRegistrarEnrollmentCredentialsWithCAInstance(t, ctxProvider, "")
}

func getRegistrarEnrollmentCredentialsWithCAInstance(t *testing.T, ctxProvider context.ClientProvider, caID string) (string, string) {

	ctx, err := ctxProvider()
	if err != nil {
		t.Fatalf("failed to get context: %s", err)
	}

	myOrg := ctx.IdentityConfig().Client().Organization

	if caID == "" {
		caID = ctx.EndpointConfig().NetworkConfig().Organizations[myOrg].CertificateAuthorities[0]
	}

	caConfig, ok := ctx.IdentityConfig().CAConfig(caID)
	if !ok {
		t.Fatal("CAConfig failed")
	}

	return caConfig.Registrar.EnrollID, caConfig.Registrar.EnrollSecret
}
