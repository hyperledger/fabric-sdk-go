/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package ledger

import (
	"strings"
	"testing"

	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	channelID = "testChannel"
)

func TestQueryBlock(t *testing.T) {

	peer1 := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}
	peer2 := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}

	lc := setupLedgerClient([]fab.Peer{&peer1, &peer2}, t)

	_, err := lc.QueryBlock(1)
	if err != nil {
		t.Fatalf("Test ledger query block failed: %s", err)
	}

	_, err = lc.QueryBlock(1, WithMaxTargets(3))
	if err != nil {
		t.Fatalf("Test ledger query block failed: %s", err)
	}

	_, err = lc.QueryBlock(1, WithMinTargets(2))
	if err != nil {
		t.Fatalf("Test ledger query block failed: %s", err)
	}

	_, err = lc.QueryBlock(1, WithMinTargets(3))
	expected := "Error getting minimum number of targets"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query block should have failed with '%s'", expected)
	}

	_, err = lc.QueryBlock(1, WithTargets(&peer1))
	if err != nil {
		t.Fatalf("Test ledger query block failed: %s", err)
	}

	_, err = lc.QueryBlock(1, WithTargets(&peer1), WithMinTargets(2))
	expected = "Error getting minimum number of targets"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query block should have failed with '%s'", expected)
	}

}

func TestQueryBlockWithNilTargets(t *testing.T) {

	peer1 := &mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}
	lc := setupLedgerClient([]fab.Peer{peer1}, t)

	_, err := lc.QueryBlock(1, WithTargets(peer1, nil))
	if err == nil || !strings.Contains(err.Error(), "target is nil") {
		t.Fatal("Should have failed due to nil target")
	}
}

func TestQueryBlockDiscoveryError(t *testing.T) {
	peer1 := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}
	peer2 := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}

	expected := "Discovery Error"
	lc := setupLedgerClientWithError(errors.New(expected), nil, []fab.Peer{&peer1, &peer2}, t)
	_, err := lc.QueryBlock(1)
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query block should have failed with '%s'", expected)
	}
}

func TestQueryBlockNegative(t *testing.T) {
	peer1 := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}
	peer2 := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}

	// Discovery service unable to discover peers
	lc := setupLedgerClient([]fab.Peer{}, t)
	expected := "no targets available"
	_, err := lc.QueryBlock(1)
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query block should have failed with '%s'", expected)
	}

	_, err = lc.QueryBlock(1, WithTargets(&peer1), WithTargetFilter(&mspFilter{mspID: "test"}))
	expected = "If targets are provided, filter cannot be provided"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query block should have failed with '%s'", expected)
	}

	expected = "QueryBlock failed"
	lc = setupLedgerClientWithError(nil, errors.New(expected), []fab.Peer{&peer1}, t)
	_, err = lc.QueryBlock(1)
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query block should have failed with '%s'", expected)
	}

	// Test bad peer response
	peer3 := mocks.MockPeer{MockName: "Peer3", MockURL: "http://peer3.com", MockRoles: []string{}, MockCert: nil, Status: 250, MockMSP: "test"}
	lc = setupLedgerClient([]fab.Peer{&peer1, &peer2, &peer3}, t)
	_, err = lc.QueryBlock(1, WithMinTargets(3))
	expected = "is less than MinTargets"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query block should have failed with '%s'", expected)
	}
}

func TestQueryBlockByHash(t *testing.T) {

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}
	lc := setupLedgerClient([]fab.Peer{&peer}, t)

	_, err := lc.QueryBlockByHash([]byte("hash"))
	if err != nil {
		t.Fatalf("Test ledger query block failed: %s", err)
	}

	_, err = lc.QueryBlockByHash([]byte("hash"), WithTargets(&peer))
	if err != nil {
		t.Fatalf("Test ledger query block failed: %s", err)
	}

	_, err = lc.QueryBlockByHash([]byte("hash"), WithTargets(&peer), WithTargetFilter(&mspFilter{mspID: "test"}))
	expected := "If targets are provided, filter cannot be provided"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query block should have failed with '%s'", expected)
	}

	expected = "QueryBlockByHash failed"
	lc = setupLedgerClientWithError(nil, errors.New(expected), []fab.Peer{&peer}, t)
	_, err = lc.QueryBlockByHash([]byte("hash"))
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query block should have failed with '%s'", expected)
	}
}

func TestQueryBlockByTxID(t *testing.T) {

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}
	lc := setupLedgerClient([]fab.Peer{&peer}, t)

	_, err := lc.QueryBlockByTxID("txID")
	if err != nil {
		t.Fatalf("Test ledger query block failed: %s", err)
	}

	_, err = lc.QueryBlockByTxID("txID", WithTargets(&peer))
	if err != nil {
		t.Fatalf("Test ledger query block failed: %s", err)
	}

	_, err = lc.QueryBlockByTxID("txID", WithTargets(&peer), WithTargetFilter(&mspFilter{mspID: "test"}))
	expected := "If targets are provided, filter cannot be provided"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query block should have failed with '%s'", expected)
	}

	expected = "QueryBlockByTxID failed"
	lc = setupLedgerClientWithError(nil, errors.New(expected), []fab.Peer{&peer}, t)
	_, err = lc.QueryBlockByTxID("txID")
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query block should have failed with '%s'", expected)
	}
}

func TestQueryInfo(t *testing.T) {

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}
	lc := setupLedgerClient([]fab.Peer{&peer}, t)

	_, err := lc.QueryInfo()
	if err != nil {
		t.Fatalf("Test ledger query info failed: %s", err)
	}

	_, err = lc.QueryInfo(WithTargets(&peer))
	if err != nil {
		t.Fatalf("Test ledger query info failed: %s", err)
	}

	// Test bad response from peer
	peer2 := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, Status: 405, MockMSP: "test"}
	lc = setupLedgerClient([]fab.Peer{&peer, &peer2}, t)

	_, err = lc.QueryInfo(WithMinTargets(2))
	expected := "is less than MinTargets"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query info should have failed with '%s'", expected)
	}
	_, err = lc.QueryInfo(WithTargets(&peer), WithTargetFilter(&mspFilter{mspID: "test"}))
	expected = "If targets are provided, filter cannot be provided"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query info should have failed with '%s'", expected)
	}

	expected = "QueryInfo failed"
	lc = setupLedgerClientWithError(nil, errors.New(expected), []fab.Peer{&peer}, t)
	_, err = lc.QueryInfo()
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query info should have failed with '%s'", expected)
	}
}

func TestQueryTransaction(t *testing.T) {

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}
	lc := setupLedgerClient([]fab.Peer{&peer}, t)

	_, err := lc.QueryTransaction("1234")
	if err != nil {
		t.Fatalf("Test ledger query transaction failed: %s", err)
	}

	_, err = lc.QueryTransaction("1234", WithTargets(&peer))
	if err != nil {
		t.Fatalf("Test ledger query transaction failed: %s", err)
	}

	// Test bad response from peer
	peer2 := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, Status: 405, MockMSP: "test"}
	lc = setupLedgerClient([]fab.Peer{&peer, &peer2}, t)

	_, err = lc.QueryTransaction("1234", WithMinTargets(2))
	expected := "is less than MinTargets"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query transaction should have failed with '%s'", expected)
	}

	_, err = lc.QueryTransaction("1234", WithTargets(&peer), WithTargetFilter(&mspFilter{mspID: "test"}))
	expected = "If targets are provided, filter cannot be provided"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query transaction should have failed with '%s'", expected)
	}

	expected = "QueryTransaction failed"
	lc = setupLedgerClientWithError(nil, errors.New(expected), []fab.Peer{&peer}, t)
	_, err = lc.QueryTransaction("1234")
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query transaction should have failed with '%s'", expected)
	}
}

func TestQueryConfig(t *testing.T) {
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}
	lc := setupLedgerClient([]fab.Peer{&peer}, t)

	_, err := lc.QueryConfig(WithTargets(&peer), WithTargetFilter(&mspFilter{mspID: "test"}))
	expected := "If targets are provided, filter cannot be provided"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query config should have failed with '%s'", expected)
	}

	_, err = lc.QueryConfig()
	expected = "config block data is nil"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query config should have failed with '%s'", expected)
	}

}

func TestQueryConfigBlock(t *testing.T) {
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200, MockMSP: "test"}
	lc := setupLedgerClient([]fab.Peer{&peer}, t)

	_, err := lc.QueryConfigBlock(WithTargets(&peer), WithTargetFilter(&mspFilter{mspID: "test"}))
	expected := "If targets are provided, filter cannot be provided"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Test ledger query config should have failed with '%s'", expected)
	}

	block, err := lc.QueryConfigBlock()
	assert.NoError(t, err)
	assert.NotNil(t, block)
}

func setupTestChannelService(ctx context.Client, orderers []fab.Orderer) (fab.ChannelService, error) {
	chProvider, err := fcmocks.NewMockChannelProvider(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "mock channel provider creation failed")
	}

	chService, err := chProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "mock channel service creation failed")
	}

	return chService, nil
}

func setupCustomTestContext(t *testing.T, discoveryService fab.DiscoveryService, orderers []fab.Orderer) context.ClientProvider {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := fcmocks.NewMockContext(user)

	if orderers == nil {
		orderer := fcmocks.NewMockOrderer("", nil)
		orderers = []fab.Orderer{orderer}
	}

	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: channelID,
		Orderers:  orderers,
	}

	testChannelSvc, err := setupTestChannelService(ctx, orderers)
	assert.Nil(t, err, "Got error %s", err)
	testChannelSvc.(*fcmocks.MockChannelService).SetTransactor(&transactor)
	testChannelSvc.(*fcmocks.MockChannelService).SetDiscovery(discoveryService)

	channelProvider := ctx.MockProviderContext.ChannelProvider()
	channelProvider.(*fcmocks.MockChannelProvider).SetCustomChannelService(testChannelSvc)

	return createClientContext(ctx)
}

func setupLedgerClient(peers []fab.Peer, t *testing.T) *Client {

	return setupLedgerClientWithError(nil, nil, peers, t)
}

func setupLedgerClientWithError(discErr error, verifyErr error, peers []fab.Peer, t *testing.T) *Client {

	fabCtx := setupCustomTestContext(t, txnmocks.NewMockDiscoveryService(discErr, peers...), nil)

	ctx := createChannelContext(fabCtx, channelID)

	lc, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	lc.verifier = &TestVerifier{verifyErr: verifyErr}

	return lc
}

func createChannelContext(clientContext context.ClientProvider, channelID string) context.ChannelProvider {

	channelProvider := func() (context.Channel, error) {
		return contextImpl.NewChannel(clientContext, channelID)
	}

	return channelProvider
}

func createClientContext(client context.Client) context.ClientProvider {
	return func() (context.Client, error) {
		return client, nil
	}
}

type TestVerifier struct {
	verifyErr error
	matchErr  error
}

// Verify checks transaction proposal response
func (tv *TestVerifier) Verify(response *fab.TransactionProposalResponse) error {
	return tv.verifyErr
}

// Match matches transaction proposal responses
func (tv *TestVerifier) Match(response []*fab.TransactionProposalResponse) error {
	return tv.matchErr
}
