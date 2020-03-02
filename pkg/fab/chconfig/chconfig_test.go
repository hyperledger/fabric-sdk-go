/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package chconfig

import (
	reqContext "context"
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"

	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/orderer"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/pkg/errors"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	channelID      = "testChannel"
	configTestFile = "config_test.yaml"
)

func TestChannelConfigWithPeer(t *testing.T) {

	ctx := setupTestContext()
	peer := getPeerWithConfigBlockPayload(t, "http://peer1.com")

	channelConfig, err := New(channelID, WithPeers([]fab.Peer{peer}), WithMinResponses(1), WithMaxTargets(1))
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()

	block, err := channelConfig.QueryBlock(reqCtx)
	if err != nil {
		t.Fatalf(err.Error())
	}
	checkConfigBlock(t, block)

	cfg, err := channelConfig.Query(reqCtx)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if cfg.ID() != channelID {
		t.Fatalf("Channel name error. Expecting %s, got %s ", channelID, cfg.ID())
	}
}

func checkConfigBlock(t *testing.T, block *common.Block) {
	if block.Header == nil {
		t.Fatal("expected header in block")
	}

	_, err := resource.CreateConfigEnvelope(block.Data.Data[0])
	if err != nil {
		t.Fatal("expected envelope in block")
	}
}

func TestChannelConfigWithPeerWithRetries(t *testing.T) {

	numberOfAttempts := 7
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)

	defRetryOpts := retry.DefaultOpts
	defRetryOpts.Attempts = numberOfAttempts
	defRetryOpts.InitialBackoff = 5 * time.Millisecond
	defRetryOpts.BackoffFactor = 1.0

	chConfig := &fab.ChannelEndpointConfig{
		Policies: fab.ChannelPolicies{QueryChannelConfig: fab.QueryChannelConfigPolicy{
			MinResponses: 2,
			MaxTargets:   1, //Ignored since we pass targets
			RetryOpts:    defRetryOpts,
		}},
	}

	mockConfig := &customMockConfig{MockConfig: &mocks.MockConfig{}, chConfig: chConfig}
	ctx.SetEndpointConfig(mockConfig)

	peer1 := getPeerWithConfigBlockPayload(t, "http://peer1.com")
	peer2 := getPeerWithConfigBlockPayload(t, "http://peer2.com")

	channelConfig, err := New(channelID, WithPeers([]fab.Peer{peer1, peer2}))
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Test QueryBlock
	// ---------------

	//Set custom retry handler for tracking number of attempts
	queryBlockRetryHandler := retry.New(defRetryOpts)
	overrideRetryHandler = &customRetryHandler{handler: queryBlockRetryHandler, retries: 0}

	queryBlockReqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(100*time.Second))
	defer cancel()

	_, err = channelConfig.QueryBlock(queryBlockReqCtx)
	if err == nil || !strings.Contains(err.Error(), "ENDORSEMENT_MISMATCH") {
		t.Fatal("Supposed to fail with ENDORSEMENT_MISMATCH. Description: payloads for config block do not match")
	}

	assert.True(t, overrideRetryHandler.(*customRetryHandler).retries-1 == numberOfAttempts, "number of attempts missmatching")

	// Test Query
	// ----------

	//Set custom retry handler for tracking number of attempts
	retryHandler := retry.New(defRetryOpts)
	overrideRetryHandler = &customRetryHandler{handler: retryHandler, retries: 0}

	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(100*time.Second))
	defer cancel()

	_, err = channelConfig.Query(reqCtx)
	if err == nil || !strings.Contains(err.Error(), "ENDORSEMENT_MISMATCH") {
		t.Fatal("Supposed to fail with ENDORSEMENT_MISMATCH. Description: payloads for config block do not match")
	}

	assert.True(t, overrideRetryHandler.(*customRetryHandler).retries-1 == numberOfAttempts, "number of attempts missmatching")
}

func TestChannelConfigWithPeerError(t *testing.T) {

	ctx := setupTestContext()
	peer := getPeerWithConfigBlockPayload(t, "http://peer1.com")

	channelConfig, err := New(channelID, WithPeers([]fab.Peer{peer}), WithMinResponses(2))
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()

	_, err = channelConfig.QueryBlock(reqCtx)
	if err == nil {
		t.Fatal("Should have failed with since there's one endorser and at least two are required")
	}

	_, err = channelConfig.Query(reqCtx)
	if err == nil {
		t.Fatal("Should have failed with since there's one endorser and at least two are required")
	}
}

func TestChannelConfigWithOrdererError(t *testing.T) {

	ctx := setupTestContext()
	o, err := orderer.New(ctx.EndpointConfig(), orderer.WithURL("localhost:9999"))
	assert.Nil(t, err)
	channelConfig, err := New(channelID, WithOrderer(o))
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(1*time.Second))
	defer cancel()

	// Expecting error since orderer is not setup
	_, err = channelConfig.QueryBlock(reqCtx)
	if err == nil {
		t.Fatal("Should have failed since orderer is not available")
	}

	// Expecting error since orderer is not setup
	_, err = channelConfig.Query(reqCtx)
	if err == nil {
		t.Fatal("Should have failed since orderer is not available")
	}

}

func TestRandomMaxTargetsSelections(t *testing.T) {

	testTargets := []fab.ProposalProcessor{
		&mockProposalProcessor{"ONE"}, &mockProposalProcessor{"TWO"}, &mockProposalProcessor{"THREE"},
		&mockProposalProcessor{"FOUR"}, &mockProposalProcessor{"FIVE"}, &mockProposalProcessor{"SIX"},
		&mockProposalProcessor{"SEVEN"}, &mockProposalProcessor{"EIGHT"}, &mockProposalProcessor{"NINE"},
	}

	max := 3
	before := ""
	for _, v := range testTargets[:max] {
		before = before + v.(*mockProposalProcessor).name
	}

	responseTargets := randomMaxTargets(testTargets, max)
	assert.True(t, responseTargets != nil && len(responseTargets) == max, "response target not as expected")

	after := ""
	for _, v := range responseTargets {
		after = after + v.(*mockProposalProcessor).name
	}
	//make sure it is random
	assert.False(t, before == after, "response targets are not random")

	max = 0 //when zero minimum supplied, result should be empty
	responseTargets = randomMaxTargets(testTargets, max)
	assert.True(t, responseTargets != nil && len(responseTargets) == max, "response target not as expected")

	max = 12 //greater than targets length
	responseTargets = randomMaxTargets(testTargets, max)
	assert.True(t, responseTargets != nil && len(responseTargets) == len(testTargets), "response target not as expected")

}

func TestResolveOptsFromConfig(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)

	defRetryOpts := retry.DefaultOpts

	chConfig := &fab.ChannelEndpointConfig{
		Policies: fab.ChannelPolicies{QueryChannelConfig: fab.QueryChannelConfigPolicy{
			MinResponses: 8,
			MaxTargets:   9,
			RetryOpts:    defRetryOpts,
		}},
	}

	mockConfig := &customMockConfig{MockConfig: &mocks.MockConfig{}, chConfig: chConfig}
	ctx.SetEndpointConfig(mockConfig)

	channelConfig, err := New(channelID, WithPeers([]fab.Peer{}))
	if err != nil {
		t.Fatal("Failed to create channel config")
	}

	assert.True(t, channelConfig.opts.MaxTargets == 0, "supposed to be zero when not resolved")
	assert.True(t, channelConfig.opts.MinResponses == 0, "supposed to be zero when not resolved")
	assert.True(t, channelConfig.opts.RetryOpts.RetryableCodes == nil, "supposed to be nil when not resolved")

	channelConfig, err = New(channelID, WithPeers([]fab.Peer{}), WithMinResponses(2))
	if err != nil {
		t.Fatal("Failed to create channel config")
	}

	assert.True(t, channelConfig.opts.MaxTargets == 0, "supposed to be zero when not resolved")
	assert.True(t, channelConfig.opts.MinResponses == 2, "supposed to be loaded with options")
	assert.True(t, channelConfig.opts.RetryOpts.RetryableCodes == nil, "supposed to be nil when not resolved")

	mockConfig.called = false
	channelConfig.resolveOptsFromConfig(ctx)

	assert.True(t, channelConfig.opts.MaxTargets == 9, "supposed to be loaded once opts resolved from config")
	assert.True(t, channelConfig.opts.MinResponses == 2, "supposed to be updated once loaded with non zero value")
	assert.True(t, channelConfig.opts.RetryOpts.RetryableCodes != nil, "supposed to be loaded once opts resolved from config")
	assert.True(t, mockConfig.called, "config.ChannelConfig() not used by resolve opts function")

	//Try again, opts shouldnt get reloaded from config once loaded
	mockConfig.called = false
	channelConfig.resolveOptsFromConfig(ctx)
	assert.False(t, mockConfig.called, "config.ChannelConfig() should not be used by resolve opts function once opts are loaded")
}

func TestResolveOptsDefaultValues(t *testing.T) {
	testResolveOptsDefaultValues(t, channelID)
}

func TestResolveOptsDefaultValuesWithInvalidChannel(t *testing.T) {
	//Should be successful even with invalid channel id
	testResolveOptsDefaultValues(t, "INVALID-CHANNEL-ID")
}

func TestCapabilities(t *testing.T) {
	pvtExpCapability := "V1_1_PVTDATA_EXPERIMENTAL"
	resourceTreeExpCapability := "V1_1_RESOURCETREE_EXPERIMENTAL"
	v1_12Capability := "V1_12"
	V1_4Capability := "V1_4"
	V1_4_2Capability := "V1_4_2"
	v2_0Capability := "V2_0"
	v2_1Capability := "V2_1"
	V3Capability := "V3"
	V3_1Capability := "V3_1"

	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				"Org1MSP",
				"Org2MSP",
			},
			OrdererAddress:          "localhost:9999",
			RootCA:                  validRootCA,
			ChannelCapabilities:     []string{fab.V1_1Capability},
			OrdererCapabilities:     []string{fab.V1_1Capability, v2_0Capability},
			ApplicationCapabilities: []string{fab.V1_2Capability, pvtExpCapability, V3_1Capability, V1_4_2Capability},
		},
		Index:           0,
		LastConfigIndex: 0,
	}

	chConfig, err := extractConfig("mychannel", builder.Build())
	require.NoError(t, err)

	assert.Truef(t, chConfig.HasCapability(fab.ChannelGroupKey, fab.V1_1Capability), "expecting channel capability [%s]", fab.V1_1Capability)
	assert.Truef(t, chConfig.HasCapability(fab.OrdererGroupKey, fab.V1_1Capability), "expecting orderer capability [%s]", fab.V1_1Capability)
	assert.Truef(t, chConfig.HasCapability(fab.OrdererGroupKey, v1_12Capability), "expecting orderer capability [%s] since [%s] is supported", v1_12Capability, v2_0Capability)
	assert.Truef(t, chConfig.HasCapability(fab.OrdererGroupKey, v2_0Capability), "expecting orderer capability [%s]", v2_0Capability)
	assert.Falsef(t, chConfig.HasCapability(fab.OrdererGroupKey, v2_1Capability), "not expecting orderer capability", v2_1Capability)
	assert.Truef(t, chConfig.HasCapability(fab.ApplicationGroupKey, fab.V1_2Capability), "expecting application capability [%s]", fab.V1_2Capability)
	assert.Truef(t, chConfig.HasCapability(fab.ApplicationGroupKey, fab.V1_1Capability), "expecting application capability [%s] since [%s] is supported", fab.V1_1Capability, fab.V1_2Capability)
	assert.Truef(t, chConfig.HasCapability(fab.ApplicationGroupKey, pvtExpCapability), "expecting application capability [%s]", pvtExpCapability)
	assert.Falsef(t, chConfig.HasCapability(fab.ApplicationGroupKey, resourceTreeExpCapability), "not expecting application capability [%s]", resourceTreeExpCapability)
	assert.Truef(t, chConfig.HasCapability(fab.ApplicationGroupKey, V3Capability), "expecting application capability [%s]", V3Capability)
	assert.Truef(t, chConfig.HasCapability(fab.ApplicationGroupKey, V1_4Capability), "expecting application capability [%s]", V1_4Capability)
}

func testResolveOptsDefaultValues(t *testing.T, channelID string) {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)

	configPath := filepath.Join(metadata.GetProjectPath(), "pkg", "core", "config", "testdata", configTestFile)
	backends, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatal("supposed to get valid backends")
	}
	endpointCfg, err := fabImpl.ConfigFromBackend(backends...)
	if err != nil {
		t.Fatal("supposed to get valid endpoint config")
	}
	ctx.SetEndpointConfig(endpointCfg)

	channelConfig, err := New(channelID, WithPeers([]fab.Peer{}))
	if err != nil {
		t.Fatal("Failed to create channel config")
	}

	channelConfig.resolveOptsFromConfig(ctx)
	assert.True(t, channelConfig.opts.MaxTargets == 2, "supposed to be loaded once opts resolved from config")
	assert.True(t, channelConfig.opts.MinResponses == 1, "supposed to be loaded once opts resolved from config")
	assert.True(t, channelConfig.opts.RetryOpts.RetryableCodes != nil, "supposed to be loaded once opts resolved from config")
}

func setupTestContext() context.Client {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)
	ctx.SetEndpointConfig(mocks.NewMockEndpointConfig())
	return ctx
}

func getPeerWithConfigBlockPayload(t *testing.T, peerURL string) fab.Peer {

	// create config block builder in order to create valid payload
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				"Org1MSP",
				"Org2MSP",
			},
			OrdererAddress: "localhost:7054",
			RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}

	payload, err := proto.Marshal(builder.Build())
	if err != nil {
		t.Fatal("Failed to marshal mock block")
	}

	// peer with valid config block payload
	peer := &mocks.MockPeer{MockName: "Peer1", MockURL: peerURL, MockRoles: []string{}, MockCert: nil, Payload: payload, Status: 200}

	return peer
}

//mockProposalProcessor to mock proposal processor for random max target test
type mockProposalProcessor struct {
	name string
}

func (pp *mockProposalProcessor) ProcessTransactionProposal(reqCtx reqContext.Context, request fab.ProcessProposalRequest) (*fab.TransactionProposalResponse, error) {
	return nil, errors.New("not implemented, just mock")
}

//customMockConfig to mock config to override channel configuration options
type customMockConfig struct {
	*mocks.MockConfig
	chConfig *fab.ChannelEndpointConfig
	called   bool
}

func (c *customMockConfig) ChannelConfig(name string) *fab.ChannelEndpointConfig {
	c.called = true
	return c.chConfig
}

//customRetryHandler is wrapper around retry handler which keeps count of attempts for unit-testing
type customRetryHandler struct {
	handler retry.Handler
	retries int
}

func (c *customRetryHandler) Required(err error) bool {
	c.retries++
	return c.handler.Required(err)
}

var validRootCA = `-----BEGIN CERTIFICATE-----
MIICYjCCAgmgAwIBAgIUB3CTDOU47sUC5K4kn/Caqnh114YwCgYIKoZIzj0EAwIw
fzELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xHzAdBgNVBAoTFkludGVybmV0IFdpZGdldHMsIEluYy4xDDAK
BgNVBAsTA1dXVzEUMBIGA1UEAxMLZXhhbXBsZS5jb20wHhcNMTYxMDEyMTkzMTAw
WhcNMjExMDExMTkzMTAwWjB/MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZv
cm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEfMB0GA1UEChMWSW50ZXJuZXQg
V2lkZ2V0cywgSW5jLjEMMAoGA1UECxMDV1dXMRQwEgYDVQQDEwtleGFtcGxlLmNv
bTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABKIH5b2JaSmqiQXHyqC+cmknICcF
i5AddVjsQizDV6uZ4v6s+PWiJyzfA/rTtMvYAPq/yeEHpBUB1j053mxnpMujYzBh
MA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQXZ0I9
qp6CP8TFHZ9bw5nRtZxIEDAfBgNVHSMEGDAWgBQXZ0I9qp6CP8TFHZ9bw5nRtZxI
EDAKBggqhkjOPQQDAgNHADBEAiAHp5Rbp9Em1G/UmKn8WsCbqDfWecVbZPQj3RK4
oG5kQQIgQAe4OOKYhJdh3f7URaKfGTf492/nmRmtK+ySKjpHSrU=
-----END CERTIFICATE-----
`
