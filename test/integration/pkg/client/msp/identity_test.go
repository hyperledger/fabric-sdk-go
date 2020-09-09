/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/assert"
)

func TestIdentity(t *testing.T) {

	mspClient, sdk := setupClient(t)
	defer integration.CleanupUserData(t, sdk)

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

	req := &msp.IdentityRequest{
		ID:          username,
		Affiliation: "org2",
		Type:        IdentityTypeUser,
		Attributes:  testAttributes,
		// Affiliation and ID are mandatory. "org1" and "org2" are hardcoded as CA defaults
		// See https://github.com/hyperledger/fabric-ca/blob/release/cmd/fabric-ca-server/config.go
	}

	// Create new identity
	newIdentity, err := mspClient.CreateIdentity(req)
	if err != nil {
		t.Fatalf("Create identity failed: %s", err)
	}

	if newIdentity.Secret == "" {
		t.Fatal("Secret should have been generated")
	}

	identity, err := mspClient.GetIdentity(username)
	if err != nil {
		t.Fatalf("get identity failed: %s", err)
	}

	t.Logf("Get Identity: [%v]:", identity)

	if !verifyIdentity(req, identity) {
		t.Fatalf("verify identity failed req=[%v]; resp=[%v] ", req, identity)
	}

	// Enroll the new user
	err = mspClient.Enroll(username, msp.WithSecret(newIdentity.Secret))
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

func TestUpdateIdentity(t *testing.T) {

	mspClient, sdk := setupClient(t)
	defer integration.CleanupUserData(t, sdk)

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

	req := &msp.IdentityRequest{
		ID:          username,
		Affiliation: "org2",
		Type:        IdentityTypeUser,
		Attributes:  testAttributes,
		// Affiliation and ID are mandatory. "org1" and "org2" are hardcoded as CA defaults
		// See https://github.com/hyperledger/fabric-ca/blob/release/cmd/fabric-ca-server/config.go
	}

	// Create new identity
	newIdentity, err := mspClient.CreateIdentity(req)
	if err != nil {
		t.Fatalf("Create identity failed: %s", err)
	}

	// Update secret
	req.Secret = "top-secret"

	identity, err := mspClient.ModifyIdentity(req)
	if err != nil {
		t.Fatalf("modify identity failed: %s", err)
	}

	if identity.Secret != "top-secret" {
		t.Fatalf("update identity failed: %s", err)
	}

	// Enroll the new user with old secret
	err = mspClient.Enroll(username, msp.WithSecret(newIdentity.Secret))
	if err == nil {
		t.Fatal("Enroll should have failed since secret has been updated")
	}

	// Enroll the new user with updated secret
	err = mspClient.Enroll(username, msp.WithSecret(identity.Secret))
	if err != nil {
		t.Fatalf("Enroll failed: %s", err)
	}

	removed, err := mspClient.RemoveIdentity(&msp.RemoveIdentityRequest{ID: username})
	if err != nil {
		t.Fatalf("remove identity failed: %s", err)
	}

	t.Logf("Removed identity [%v]", removed)

	// Test enroll with deleted identity
	err = mspClient.Enroll(username, msp.WithSecret(identity.Secret))
	if err == nil {
		t.Fatal("Enroll should have failed since identity has been deleted")
	}
}
func TestGetAllIdentities(t *testing.T) {

	mspClient, sdk := setupClient(t)
	defer integration.CleanupUserData(t, sdk)

	testAttributes := []msp.Attribute{
		{
			Name:  integration.GenerateRandomID(),
			Value: fmt.Sprintf("%s:ecert", integration.GenerateRandomID()),
			ECert: true,
		},
	}

	req1 := &msp.IdentityRequest{
		ID:          integration.GenerateRandomID(),
		Affiliation: "org2",
		Type:        "user",
		Attributes:  testAttributes,
	}

	// Create first identity
	identity, err := mspClient.CreateIdentity(req1)
	if err != nil {
		t.Fatalf("Create identity failed: %s", err)
	}
	t.Logf("First identity created: [%v]", identity)

	// Create second identity
	req2 := &msp.IdentityRequest{
		ID:          integration.GenerateRandomID(),
		Affiliation: "org2",
		Type:        "peer",
	}
	identity, err = mspClient.CreateIdentity(req2)
	if err != nil {
		t.Fatalf("Create identity failed: %s", err)
	}
	t.Logf("Second identity created: [%v]", identity)

	identities, err := mspClient.GetAllIdentities()
	if err != nil {
		t.Fatalf("Retrieve identities failed: %s", err)
	}

	for _, id := range identities {
		t.Logf("Identity: %v", id)
	}

	if !containsIdentities(identities, req1, req2) {
		t.Fatal("Unable to retrieve newly created identities")
	}

	_, err = mspClient.GetAllIdentities(msp.WithCA("invalid"))
	if err == nil {
		t.Fatal("Should have failed for invalid CA")
	}
}

func TestSigningIdentityPrivateKey(t *testing.T) {
	mspClient, sdk := setupClient(t)
	defer integration.CleanupUserData(t, sdk)

	// Generate a random user name
	username := integration.GenerateRandomID()

	req := &msp.IdentityRequest{
		ID:          username,
		Affiliation: "org2",
		Type:        IdentityTypeUser,
	}

	// Create new identity
	newIdentity, err := mspClient.CreateIdentity(req)
	if err != nil {
		t.Fatalf("Create identity failed: %s", err)
	}

	if newIdentity.Secret == "" {
		t.Fatal("Secret should have been generated")
	}

	// Enroll the new user
	err = mspClient.Enroll(username, msp.WithSecret(newIdentity.Secret))
	if err != nil {
		t.Fatalf("Enroll failed: %s", err)
	}

	// Get the new user's signing identity
	si, err := mspClient.GetSigningIdentity(username)
	if err != nil {
		t.Fatalf("GetSigningIdentity failed: %s", err)
	}
	// Get the bytes of the private key
	pk, err := si.PrivateKey().Bytes()
	if err != nil {
		t.Fatalf("Get PrivateKey Bytes should not throw error: %s", err)
	}
	// Test that we have a valid ECPrivateKey
	p, _ := pem.Decode(pk)
	_, err = x509.ParseECPrivateKey(p.Bytes)
	if err != nil {
		t.Fatalf("Get ParseECPrivateKey should not throw error: %s", err)
	}
	// Create a new Identity with the bytes
	identity := gateway.NewX509Identity("org2", string(si.EnrollmentCertificate()), string(pk))
	if identity == nil {
		t.Fatalf("Should return a valid identity")
	}
}

func containsIdentities(identities []*msp.IdentityResponse, requests ...*msp.IdentityRequest) bool {
	for _, request := range requests {
		if !containsIdentity(identities, request) {
			return false
		}
	}
	return true
}

func containsIdentity(identities []*msp.IdentityResponse, request *msp.IdentityRequest) bool {
	for _, identity := range identities {
		if verifyIdentity(request, identity) {
			return true
		}
	}

	return false
}

type setupOptions struct {
	configProvider core.ConfigProvider
}

type setupOption func(*setupOptions) error

func withConfigProvider(configProvider core.ConfigProvider) setupOption {
	return func(s *setupOptions) error {
		s.configProvider = configProvider
		return nil
	}
}

func setupClient(t *testing.T, opts ...setupOption) (*msp.Client, *fabsdk.FabricSDK) {

	o := setupOptions{
		configProvider: integration.ConfigBackend,
	}
	for _, param := range opts {
		err := param(&o)
		if err != nil {
			t.Fatalf("failed to create setup: %s", err)
		}
	}

	// Instantiate the SDK
	sdk, err := fabsdk.New(o.configProvider)
	if err != nil {
		t.Fatalf("SDK init failed: %s", err)
	}
	defer sdk.Close()

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)

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

	return mspClient, sdk

}

func verifyIdentity(req *msp.IdentityRequest, res *msp.IdentityResponse) bool {
	if req.ID != res.ID || req.Affiliation != res.Affiliation || req.Type != res.Type {
		return false
	}

	for _, att := range req.Attributes {
		if !containsAttribute(att, res.Attributes) {
			return false
		}
	}

	return true
}

func containsAttribute(att msp.Attribute, attributes []msp.Attribute) bool {
	for _, a := range attributes {
		if a == att {
			return true
		}
	}
	return false
}

type emptyCredentialStorePathBackend struct {
}

func (c *emptyCredentialStorePathBackend) Lookup(key string) (interface{}, bool) {
	if key == "client.credentialStore.path" {
		return "", true
	}
	return nil, false
}

func TestCreateSDKWithoutCredentialStorePath(t *testing.T) {

	integrationConfigProvider, err := integration.ConfigBackend()
	assert.Nil(t, err)

	emptyCredentialStorePathConfigProvider := func() ([]core.ConfigBackend, error) {
		emptyCredentialStorePathBackendSlice := []core.ConfigBackend{
			&emptyCredentialStorePathBackend{},
		}
		return append(emptyCredentialStorePathBackendSlice, integrationConfigProvider...), nil
	}

	client, sdk := setupClient(t, withConfigProvider(emptyCredentialStorePathConfigProvider))
	defer integration.CleanupUserData(t, sdk)

	ctxProvider := sdk.Context()
	registrarEnrollID, _ := getRegistrarEnrollmentCredentials(t, ctxProvider)

	// CA registrar should have been enrolled after the SDK was created
	sid, err := client.GetSigningIdentity(registrarEnrollID)
	assert.Nil(t, err)
	assert.NotNil(t, sid)
}
