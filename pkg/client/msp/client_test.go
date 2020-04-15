/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	mspctx "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	mspImpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	caServerURLListen = "http://localhost:0"
	configFile        = "config_test.yaml"
)

var caServerURL string

type nwConfig struct {
	CertificateAuthorities map[string]mspImpl.CAConfig
}

// TestEnroll is a unit test for Client enrollment and re-enrollment scenarios
func TestEnroll(t *testing.T) {

	f := testFixture{}
	sdk := f.setup()
	defer f.close()

	ctxProvider := sdk.Context()

	// Get the Client.
	// Without WithOrg option, it uses default client organization.
	msp, err := New(ctxProvider)
	if err != nil {
		t.Fatalf("failed to create CA client: %s", err)
	}

	// Empty enrollment ID
	err = msp.Enroll("", WithSecret("user1"))
	if err == nil {
		t.Fatal("Enroll should return error for empty enrollment ID")
	}

	// Empty enrollment secret
	err = msp.Enroll("enrolledUsername", WithSecret(""))
	if err == nil {
		t.Fatal("Enroll should return error for empty enrollment secret")
	}

	enrolledUser := getEnrolledUser(t, msp)

	// Reenroll with empty user
	err = msp.Reenroll("")
	if err == nil {
		t.Fatal("Expected error with enpty user")
	}
	if err.Error() != "user name missing" {
		t.Fatalf("Expected error user required. Got: %s", err)
	}

	// Reenroll with appropriate user
	err = msp.Reenroll(enrolledUser.Identifier().ID)
	if err != nil {
		t.Fatalf("Reenroll return error %s", err)
	}

	// Try with a non-default org
	testWithOrg2(t, ctxProvider)

	// Try with another CA instance
	testWithOrg1TLSCAInstance(t, ctxProvider)

}

func TestEnrollWithProfile(t *testing.T) {
	f := testFixture{}
	sdk := f.setup()
	defer sdk.Close()

	ctxProvider := sdk.Context()
	msp, err := New(ctxProvider)
	require.NoError(t, err)

	enrollUsername := randomUsername()
	_, err = msp.GetSigningIdentity(enrollUsername)
	if err != ErrUserNotFound {
		t.Fatal("Expected to not find user")
	}

	err = msp.Enroll(enrollUsername, WithSecret("enrollmentSecret"), WithProfile("tls"))
	require.NoError(t, err)

	enrolledUser, err := msp.GetSigningIdentity(enrollUsername)
	require.NoError(t, err)

	assert.Equal(t, enrollUsername, enrolledUser.Identifier().ID)
	assert.Equal(t, "Org1MSP", enrolledUser.Identifier().MSPID)

	err = msp.Reenroll(enrolledUser.Identifier().ID, WithProfile("tls"))
	if err != nil {
		t.Fatalf("Reenroll return error %s", err)
	}
}

func TestEnrollWithType(t *testing.T) {
	f := testFixture{}
	sdk := f.setup()
	defer sdk.Close()

	ctxProvider := sdk.Context()
	msp, err := New(ctxProvider)
	require.NoError(t, err)

	enrollUsername := randomUsername()
	_, err = msp.GetSigningIdentity(enrollUsername)
	if err != ErrUserNotFound {
		t.Fatal("Expected to not find user")
	}

	err = msp.Enroll(enrollUsername, WithSecret("enrollmentSecret"), WithType("idemix"))
	if err == nil {
		t.Fatal("idemix enroll not supported")
	}

	err = msp.Reenroll(enrollUsername, WithType("idemix"))
	if err == nil {
		t.Fatal("idemix enroll not supported")
	}
}

func TestEnrollWithLabel(t *testing.T) {
	f := testFixture{}
	sdk := f.setup()
	defer sdk.Close()

	ctxProvider := sdk.Context()
	msp, err := New(ctxProvider)
	require.NoError(t, err)

	enrollUsername := randomUsername()
	_, err = msp.GetSigningIdentity(enrollUsername)
	if err != ErrUserNotFound {
		t.Fatal("Expected to not find user")
	}

	err = msp.Enroll(enrollUsername, WithSecret("enrollmentSecret"), WithLabel("ForFabric"))
	require.NoError(t, err)

	enrolledUser, err := msp.GetSigningIdentity(enrollUsername)
	require.NoError(t, err)

	assert.Equal(t, enrollUsername, enrolledUser.Identifier().ID)
	assert.Equal(t, "Org1MSP", enrolledUser.Identifier().MSPID)

	err = msp.Reenroll(enrolledUser.Identifier().ID, WithLabel("ForFabric"))
	if err != nil {
		t.Fatalf("Reenroll return error %s", err)
	}
}

func TestEnrollWithAttributeRequests(t *testing.T) {
	f := testFixture{}
	sdk := f.setup()
	defer sdk.Close()

	ctxProvider := sdk.Context()
	msp, err := New(ctxProvider)
	require.NoError(t, err)

	enrollUsername := randomUsername()
	_, err = msp.GetSigningIdentity(enrollUsername)
	if err != ErrUserNotFound {
		t.Fatal("Expected to not find user")
	}

	attrReqs := []*AttributeRequest{{Name: "name1", Optional: true}}
	err = msp.Enroll(enrollUsername, WithSecret("enrollmentSecret"), WithAttributeRequests(attrReqs))
	require.NoError(t, err)

	enrolledUser, err := msp.GetSigningIdentity(enrollUsername)
	require.NoError(t, err)

	assert.Equal(t, enrollUsername, enrolledUser.Identifier().ID)
	assert.Equal(t, "Org1MSP", enrolledUser.Identifier().MSPID)

	err = msp.Reenroll(enrolledUser.Identifier().ID, WithAttributeRequests(attrReqs))
	if err != nil {
		t.Fatalf("Reenroll return error %s", err)
	}
}

func TestNewWithNonExistentOrganization(t *testing.T) {
	// Instantiate the SDK
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, configFile)
	sdk, err := fabsdk.New(config.FromFile(configPath))
	if err != nil {
		t.Fatalf("SDK init failed: %s", err)
	}
	_, err = New(sdk.Context(), WithOrg("nonExistentOrg"))
	if err == nil {
		t.Fatal("Should have failed for non-existing organization")
	}

}

func TestNewWithOrg1CAInstance(t *testing.T) {
	// Instantiate the SDK
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, configFile)
	sdk, err := fabsdk.New(config.FromFile(configPath))
	if err != nil {
		t.Fatalf("SDK init failed: %s", err)
	}
	_, err = New(sdk.Context(), WithOrg("Org1"), WithCAInstance("tlsca.org1.example.com"))
	if err != nil {
		t.Fatal("Should not have failed with existing CA instance")
	}

}

func TestNewWithOrg2CAInstance(t *testing.T) {
	// Instantiate the SDK
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, configFile)
	sdk, err := fabsdk.New(config.FromFile(configPath))
	if err != nil {
		t.Fatalf("SDK init failed: %s", err)
	}
	_, err = New(sdk.Context(), WithOrg("Org2"), WithCAInstance("ca.org2.example.com"))
	if err != nil {
		t.Fatal("Should not have failed with existing CA instance")
	}

}

func TestNewWithCAInstanceFromAnotherOrg(t *testing.T) {
	// Instantiate the SDK
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, configFile)
	sdk, err := fabsdk.New(config.FromFile(configPath))
	if err != nil {
		t.Fatalf("SDK init failed: %s", err)
	}
	_, err = New(sdk.Context(), WithOrg("Org1"), WithCAInstance("ca.org2.example.com"))
	if err == nil {
		t.Fatal("Should have failed with CA instance from another org")
	}

}

// TestNewWithProviderError tests New with provider error
func TestNewWithProviderError(t *testing.T) {

	// Create msp client with client provider error
	_, err := New(mockContextWithProviderError())
	if err == nil {
		t.Fatal("Should have failed due to provider failure")
	}
}

func mockContextWithProviderError() contextApi.ClientProvider {
	return func() (contextApi.Client, error) {
		return nil, errors.New("Test Error")
	}
}

// TestNewWithProviderError tests error in client option
func TestNewWithClientOptionError(t *testing.T) {
	_, err := New(mockClientProvider(), WithClientOptionError())
	if err == nil {
		t.Fatal("Should have failed due to client option failure")
	}
}

// WithClientOptionError client option that generates error
func WithClientOptionError() ClientOption {
	return func(o *clientOptions) error {
		return errors.New("Client option error")
	}
}

func TestRegister(t *testing.T) {
	f := testFixture{}
	sdk := f.setup()
	defer f.close()

	ctxProvider := sdk.Context()

	// Get the Client.
	// Without WithOrg option, it uses default client organization.
	msp, err := New(ctxProvider)
	if err != nil {
		t.Fatalf("failed to create CA client: %s", err)
	}

	_, err = msp.Register(&RegistrationRequest{Name: "testuser"})
	if err != nil {
		t.Fatalf("Register return error %s", err)
	}

}

func TestRevoke(t *testing.T) {
	f := testFixture{}
	sdk := f.setup()
	defer f.close()

	ctxProvider := sdk.Context()

	// Get the Client.
	// Without WithOrg option, it uses default client organization.
	msp, err := New(ctxProvider)
	if err != nil {
		t.Fatalf("failed to create CA client: %s", err)
	}

	_, err = msp.Revoke(&RevocationRequest{Name: "testuser"})
	if err != nil {
		t.Fatalf("Revoke return error %s", err)
	}

}

// TestCreateIdentityFailure tests failures in CreateIdentity
func TestCreateIdentityFailure(t *testing.T) {

	// Create msp client
	c, err := New(mockClientProvider())
	if err != nil {
		t.Fatalf("failed to create CA client: %s", err)
	}

	// Missing required ID
	_, err = c.CreateIdentity(&IdentityRequest{})
	if err == nil || !strings.Contains(err.Error(), "ID is required") {
		t.Fatalf("Should have failed to create identity due to missing ID: %s", err)
	}

}

// TestModifyIdentityFailure tests failures in ModifyIdentity
func TestModifyIdentityFailure(t *testing.T) {

	// Create msp client
	c, err := New(mockClientProvider())
	if err != nil {
		t.Fatalf("failed to create CA client: %s", err)
	}

	// Missing required ID
	_, err = c.ModifyIdentity(&IdentityRequest{Affiliation: "org2", Secret: "top-secret", Attributes: []Attribute{{Name: "attName1", Value: "attValue1"}}})
	if err == nil || !strings.Contains(err.Error(), "ID is required") {
		t.Fatalf("Should have failed to update identity due to missing id: %s", err)
	}

}

// TestRemoveIdentityFailure tests different failures in RemoveIdentity
func TestRemoveIdentityFailure(t *testing.T) {

	// Create msp client
	c, err := New(mockClientProvider())
	if err != nil {
		t.Fatalf("failed to create CA client: %s", err)
	}

	// Missing required ID
	_, err = c.RemoveIdentity(&RemoveIdentityRequest{})
	if err == nil || !strings.Contains(err.Error(), "ID is required") {
		t.Fatalf("Should have failed to create identity due to missing id: %s", err)
	}

}

// TestGetIdentityFailure tests failures in GetIdentity
func TestGetIdentityFailure(t *testing.T) {

	// Create msp client
	c, err := New(mockClientProvider())
	if err != nil {
		t.Fatalf("failed to create CA client: %s", err)
	}

	_, err = c.GetIdentity("")
	if err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Fatalf("Should have failed to get identity due to missing id: %s", err)
	}

	_, err = c.GetIdentity("123", withOptionError())
	if err == nil {
		t.Fatal("Should have failed due to error in opton")
	}
}

// TestGetAllIdentitiesFailure tests failures in GetAllIdentities
func TestGetAllIdentitiesFailure(t *testing.T) {

	// Create msp client
	c, err := New(mockClientProvider())
	if err != nil {
		t.Fatalf("failed to create CA client: %s", err)
	}

	_, err = c.GetAllIdentities(withOptionError())
	if err == nil {
		t.Fatal("Should have failed due to error in opton")
	}

}

// withOptionError is request option that generates error
func withOptionError() RequestOption {
	return func(o *requestOptions) error {
		return errors.New("Option Error")
	}
}

func testWithOrg2(t *testing.T, ctxProvider contextApi.ClientProvider) {
	msp, err := New(ctxProvider, WithOrg("Org2"))
	if err != nil {
		t.Fatalf("failed to create CA client: %s", err)
	}

	org2lUsername := randomUsername()

	err = msp.Enroll(org2lUsername, WithSecret("enrollmentSecret"))
	if err != nil {
		t.Fatalf("Enroll return error %s", err)
	}

	org2EnrolledUser, err := msp.GetSigningIdentity(org2lUsername)
	if err != nil {
		t.Fatal("Expected to find user")
	}

	if org2EnrolledUser.Identifier().ID != org2lUsername {
		t.Fatal("Enrolled user name doesn't match")
	}

	if org2EnrolledUser.Identifier().MSPID != "Org2MSP" {
		t.Fatal("Enrolled user mspID doesn't match")
	}
}

func testWithOrg1TLSCAInstance(t *testing.T, ctxProvider contextApi.ClientProvider) {
	msp, err := New(ctxProvider, WithOrg("Org1"), WithCAInstance("tlsca.org1.example.com"))
	if err != nil {
		t.Fatalf("failed to create CA client: %s", err)
	}

	org1lUsername := randomUsername()

	err = msp.Enroll(org1lUsername, WithSecret("enrollmentSecret"))
	if err != nil {
		t.Fatalf("Enroll return error %s", err)
	}

	org2EnrolledUser, err := msp.GetSigningIdentity(org1lUsername)
	if err != nil {
		t.Fatal("Expected to find user")
	}

	if org2EnrolledUser.Identifier().ID != org1lUsername {
		t.Fatal("Enrolled user name doesn't match")
	}

	if org2EnrolledUser.Identifier().MSPID != "Org1MSP" {
		t.Fatal("Enrolled user mspID doesn't match")
	}
}

func getEnrolledUser(t *testing.T, msp *Client) mspctx.SigningIdentity {
	// Successful enrollment scenario

	enrollUsername := randomUsername()

	_, err := msp.GetSigningIdentity(enrollUsername)
	if err != ErrUserNotFound {
		t.Fatal("Expected to not find user")
	}

	err = msp.Enroll(enrollUsername, WithSecret("enrollmentSecret"))
	if err != nil {
		t.Fatalf("Enroll return error %s", err)
	}

	enrolledUser, err := msp.GetSigningIdentity(enrollUsername)
	if err != nil {
		t.Fatal("Expected to find user")
	}

	if enrolledUser.Identifier().ID != enrollUsername {
		t.Fatal("Enrolled user name doesn't match")
	}

	if enrolledUser.Identifier().MSPID != "Org1MSP" {
		t.Fatal("Enrolled user mspID doesn't match")
	}
	return enrolledUser
}

type testFixture struct {
	cryptoSuiteConfig core.CryptoSuiteConfig
	identityConfig    msp.IdentityConfig
}

var caServer = &mockmsp.MockFabricCAServer{}

func (f *testFixture) setup() *fabsdk.FabricSDK {

	var lis net.Listener
	var err error
	if !caServer.Running() {
		lis, err = net.Listen("tcp", strings.TrimPrefix(caServerURLListen, "http://"))
		if err != nil {
			panic(fmt.Sprintf("Error starting CA Server %s", err))
		}

		caServerURL = "http://" + lis.Addr().String()
	}

	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, configFile)
	backend, err := config.FromFile(configPath)()
	if err != nil {
		panic(err)
	}

	//Override ca matchers for this test
	customBackend := getCustomBackend(backend...)

	configProvider := func() ([]core.ConfigBackend, error) {
		return customBackend, nil
	}

	// Instantiate the SDK
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		panic(fmt.Sprintf("SDK init failed: %s", err))
	}

	configBackend, err := sdk.Config()
	if err != nil {
		panic(fmt.Sprintf("Failed to get config: %s", err))
	}

	f.cryptoSuiteConfig = cryptosuite.ConfigFromBackend(configBackend)
	f.identityConfig, err = mspImpl.ConfigFromBackend(configBackend)
	if err != nil {
		panic(fmt.Sprintf("Failed to get identity config: %s", err))
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store
	cleanup(f.cryptoSuiteConfig.KeyStorePath())
	cleanup(f.identityConfig.CredentialStorePath())

	ctxProvider := sdk.Context()
	ctx, err := ctxProvider()
	if err != nil {
		panic(fmt.Sprintf("Failed to init context: %s", err))
	}

	// Start Http Server if it's not running
	if !caServer.Running() {
		caServer.Start(lis, ctx.CryptoSuite())
	}

	return sdk
}

func (f *testFixture) close() {
	cleanup(f.identityConfig.CredentialStorePath())
	cleanup(f.cryptoSuiteConfig.KeyStorePath())
}

func cleanup(storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to remove dir %s: %s\n", storePath, err))
	}
	// Recreate the directory only
	if err := os.MkdirAll(storePath, os.FileMode(os.ModePerm)); err != nil {
		panic(fmt.Sprintf("Failed to recreate dir %s: %s\n", storePath, err))
	}
}

func randomUsername() string {
	return "user" + strconv.Itoa(rand.Intn(500000))
}

func getCustomBackend(currentBackends ...core.ConfigBackend) []core.ConfigBackend {
	backendMap := make(map[string]interface{})

	//Custom URLs for ca configs
	networkConfig := nwConfig{}
	configLookup := lookup.New(currentBackends...)
	configLookup.UnmarshalKey("certificateAuthorities", &networkConfig.CertificateAuthorities)

	ca1Config := networkConfig.CertificateAuthorities["ca.org1.example.com"]
	ca1Config.URL = caServerURL
	tlsca1Config := networkConfig.CertificateAuthorities["tlsca.org1.example.com"]
	tlsca1Config.URL = caServerURL
	ca2Config := networkConfig.CertificateAuthorities["ca.org2.example.com"]
	ca2Config.URL = caServerURL

	networkConfig.CertificateAuthorities["ca.org1.example.com"] = ca1Config
	networkConfig.CertificateAuthorities["tlsca.org1.example.com"] = tlsca1Config
	networkConfig.CertificateAuthorities["ca.org2.example.com"] = ca2Config
	backendMap["certificateAuthorities"] = networkConfig.CertificateAuthorities

	backends := append([]core.ConfigBackend{}, &mocks.MockConfigBackend{KeyValueMap: backendMap})
	return append(backends, currentBackends...)

}
