/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/stretchr/testify/assert"
)

func TestQueryMethods(t *testing.T) {
	channel, _ := setupTestLedger()
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}

	_, err := channel.QueryBlock(-1, []fab.ProposalProcessor{&peer})
	if err == nil {
		t.Fatalf("Query block cannot be negative number")
	}

	_, err = channel.QueryBlockByHash(nil, []fab.ProposalProcessor{&peer})
	if err == nil {
		t.Fatalf("Query hash cannot be nil")
	}
}

func TestChannelQueryBlock(t *testing.T) {

	channel, _ := setupTestLedger()
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200}

	_, err := channel.QueryBlock(1, []fab.ProposalProcessor{&peer})

	if err != nil {
		t.Fatalf("Test channel query block failed: %s", err)
	}

	_, err = channel.QueryBlockByHash([]byte(""), []fab.ProposalProcessor{&peer})

	if err != nil {
		t.Fatal("Test channel query block by hash failed,")
	}

}

func TestQueryInstantiatedChaincodes(t *testing.T) {
	channel, _ := setupTestLedger()
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200}

	res, err := channel.QueryInstantiatedChaincodes([]fab.ProposalProcessor{&peer})

	if err != nil || res == nil {
		t.Fatalf("Test QueryInstatiated chaincode failed: %v", err)
	}

}

func TestQueryTransaction(t *testing.T) {
	channel, _ := setupTestLedger()
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200}

	res, err := channel.QueryTransaction("txid", []fab.ProposalProcessor{&peer})

	if err != nil || res == nil {
		t.Fatal("Test QueryTransaction failed")
	}
}

func TestQueryInfo(t *testing.T) {
	channel, _ := setupTestLedger()
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200}

	res, err := channel.QueryInfo([]fab.ProposalProcessor{&peer})

	if err != nil || res == nil {
		t.Fatalf("Test QueryInfo failed: %v", err)
	}
}

func TestQueryConfig(t *testing.T) {
	channel, _ := setupTestLedger()

	// empty targets
	_, err := channel.QueryConfigBlock([]fab.Peer{}, 1)
	if err == nil {
		t.Fatalf("Should have failed due to empty targets")
	}

	// min endorsers <= 0
	_, err = channel.QueryConfigBlock([]fab.Peer{mocks.NewMockPeer("Peer1", "http://peer1.com")}, 0)
	if err == nil {
		t.Fatalf("Should have failed due to empty targets")
	}

	// peer without payload
	_, err = channel.QueryConfigBlock([]fab.Peer{mocks.NewMockPeer("Peer1", "http://peer1.com")}, 1)
	if err == nil {
		t.Fatalf("Should have failed due to nil block metadata")
	}

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
		t.Fatalf("Failed to marshal mock block")
	}

	// peer with valid config block payload
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: payload, Status: 200}

	// fail with min endorsers
	res, err := channel.QueryConfigBlock([]fab.Peer{&peer}, 2)
	if err == nil {
		t.Fatalf("Should have failed with since there's one endorser and at least two are required")
	}

	// success with one endorser
	res, err = channel.QueryConfigBlock([]fab.Peer{&peer}, 1)
	if err != nil || res == nil {
		t.Fatalf("Test QueryConfig failed: %v", err)
	}

	// create second endorser with same payload
	peer2 := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, Payload: payload, Status: 200}

	// success with two endorsers
	res, err = channel.QueryConfigBlock([]fab.Peer{&peer, &peer2}, 2)
	if err != nil || res == nil {
		t.Fatalf("Test QueryConfig failed: %v", err)
	}

	// Create different config block payload
	builder2 := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				"Org1MSP",
			},
			OrdererAddress: "builder2:7054",
			RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}

	payload2, err := proto.Marshal(builder2.Build())
	if err != nil {
		t.Fatalf("Failed to marshal mock block 2")
	}

	// peer 2 now had different payload; query config block should fail
	peer2.Payload = payload2
	res, err = channel.QueryConfigBlock([]fab.Peer{&peer, &peer2}, 2)
	if err == nil {
		t.Fatalf("Should have failed for different block payloads")
	}

}

func TestFilterResponses(t *testing.T) {
	tprs := []*fab.TransactionProposalResponse{}
	err := fmt.Errorf("test")
	for i := 0; i <= 100; i++ {
		var s int
		if i%2 == 0 {
			s = 200
		} else {
			s = 500
		}
		tprs = append(tprs, &fab.TransactionProposalResponse{Status: int32(s)})
	}
	f, errs := filterResponses(tprs, err)
	assert.Len(t, f, 51)
	assert.Len(t, errs.(multi.Errors), 51)
}

func setupTestLedger() (*Ledger, error) {
	return setupLedger("testChannel")
}

func setupLedger(channelID string) (*Ledger, error) {
	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)
	return NewLedger(ctx, channelID)
}
