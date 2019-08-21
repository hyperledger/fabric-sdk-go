/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	reqContext "context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/stretchr/testify/assert"
)

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

func TestQueryMethods(t *testing.T) {
	channel, _ := setupTestLedger()
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}

	reqCtx, cancel := context.NewRequest(setupContext(), context.WithTimeout(10*time.Second))
	defer cancel()

	_, err := channel.QueryBlockByHash(reqCtx, nil, []fab.ProposalProcessor{&peer}, nil)
	if err == nil {
		t.Fatal("Query hash cannot be nil")
	}

	_, err = channel.QueryBlockByTxID(reqCtx, "", []fab.ProposalProcessor{&peer}, nil)
	if err == nil || !strings.Contains(err.Error(), "txID is required") {
		t.Fatal("Tx ID cannot be nil")
	}
}

func TestQueryBlock(t *testing.T) {

	channel, _ := setupTestLedger()
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200}

	reqCtx, cancel := context.NewRequest(setupContext(), context.WithTimeout(10*time.Second))
	defer cancel()

	_, err := channel.QueryBlock(reqCtx, 1, []fab.ProposalProcessor{&peer}, nil)
	if err != nil {
		t.Fatalf("Test channel query block failed: %s", err)
	}

	_, err = channel.QueryBlockByHash(reqCtx, []byte("hash"), []fab.ProposalProcessor{&peer}, nil)
	if err != nil {
		t.Fatalf("Test channel query block by hash failed: %s", err)
	}

	_, err = channel.QueryBlockByTxID(reqCtx, "1234", []fab.ProposalProcessor{&peer}, nil)
	if err != nil {
		t.Fatalf("Test channel query block by tx ID failed: %s", err)
	}

}

func TestQueryInstantiatedChaincodes(t *testing.T) {
	channel, _ := setupTestLedger()
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200}

	reqCtx, cancel := context.NewRequest(setupContext(), context.WithTimeout(10*time.Second))
	defer cancel()

	res, err := channel.QueryInstantiatedChaincodes(reqCtx, []fab.ProposalProcessor{&peer}, nil)

	if err != nil || res == nil {
		t.Fatalf("Test QueryInstatiated chaincode failed: %s", err)
	}

}

func TestQueryTransaction(t *testing.T) {
	channel, _ := setupTestLedger()
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200}

	reqCtx, cancel := context.NewRequest(setupContext(), context.WithTimeout(10*time.Second))
	defer cancel()

	res, err := channel.QueryTransaction(reqCtx, "txid", []fab.ProposalProcessor{&peer}, nil)

	if err != nil || res == nil {
		t.Fatal("Test QueryTransaction failed")
	}
}

func TestQueryInfo(t *testing.T) {
	channel, _ := setupTestLedger()
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200}

	reqCtx, cancel := context.NewRequest(setupContext(), context.WithTimeout(10*time.Second))
	defer cancel()

	res, err := channel.QueryInfo(reqCtx, []fab.ProposalProcessor{&peer}, nil)

	if err != nil || res == nil {
		t.Fatalf("Test QueryInfo failed: %s", err)
	}
}

func TestQueryConfig(t *testing.T) {
	channel, _ := setupTestLedger()

	reqCtx, cancel := context.NewRequest(setupContext(), context.WithTimeout(10*time.Second))
	defer cancel()

	// empty targets
	_, err := channel.QueryConfigBlock(reqCtx, []fab.ProposalProcessor{}, nil)
	if err == nil {
		t.Fatal("Should have failed due to empty targets")
	}

	// min endorsers <= 0
	_, err = channel.QueryConfigBlock(reqCtx, []fab.ProposalProcessor{mocks.NewMockPeer("Peer1", "http://peer1.com")}, &TransactionProposalResponseVerifier{})
	if err == nil {
		t.Fatal("Should have failed due to empty targets")
	}

	// peer without payload
	_, err = channel.QueryConfigBlock(reqCtx, []fab.ProposalProcessor{mocks.NewMockPeer("Peer1", "http://peer1.com")}, &TransactionProposalResponseVerifier{MinResponses: 1})
	if err == nil {
		t.Fatal("Should have failed due to nil block metadata")
	}

	// create config block builder in order to create valid payload
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				"Org1MSP",
				"Org2MSP",
			},
			OrdererAddress: "localhost:9999",
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
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: payload, Status: 200}

	// fail with min endorsers
	_, err = channel.QueryConfigBlock(reqCtx, []fab.ProposalProcessor{&peer}, &TransactionProposalResponseVerifier{MinResponses: 2})
	if err == nil {
		t.Fatal("Should have failed with since there's one endorser and at least two are required")
	}

	// success with one endorser
	res, err := channel.QueryConfigBlock(reqCtx, []fab.ProposalProcessor{&peer}, &TransactionProposalResponseVerifier{MinResponses: 1})
	if err != nil || res == nil {
		t.Fatalf("Test QueryConfig failed: %s", err)
	}

	// create second endorser with same payload
	peer2 := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, Payload: payload, Status: 200}

	// success with two endorsers
	res, err = channel.QueryConfigBlock(reqCtx, []fab.ProposalProcessor{&peer, &peer2}, &TransactionProposalResponseVerifier{MinResponses: 2})
	if err != nil || res == nil {
		t.Fatalf("Test QueryConfig failed: %s", err)
	}

	// Create different config block payload
	createDifferentConfigBlockPayload(t, peer2, channel, reqCtx, peer)

}

func createDifferentConfigBlockPayload(t *testing.T, peer2 mocks.MockPeer, channel *Ledger, reqCtx reqContext.Context, peer mocks.MockPeer) {
	builder2 := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				"Org1MSP",
			},
			OrdererAddress: "builder2:9999",
			RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}
	payload2, err := proto.Marshal(builder2.Build())
	if err != nil {
		t.Fatal("Failed to marshal mock block 2")
	}
	// peer 2 now had different payload; query config block should fail
	peer2.Payload = payload2
	_, err = channel.QueryConfigBlock(reqCtx, []fab.ProposalProcessor{&peer, &peer2}, &TransactionProposalResponseVerifier{MinResponses: 2})
	if err == nil {
		t.Fatal("Should have failed for different block payloads")
	}
}

func TestQueryConfigBlockDifferentMetadata(t *testing.T) {
	channel, _ := setupTestLedger()
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				"Org1MSP",
				"Org2MSP",
			},
			OrdererAddress: "localhost:9999",
			RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}

	b := builder.Build()
	b.Metadata = &common.BlockMetadata{Metadata: [][]byte{[]byte("test1")}}

	payload1, err := proto.Marshal(b)
	assert.Nil(t, err, "Failed to marshal mock block")

	b.Metadata = &common.BlockMetadata{Metadata: [][]byte{[]byte("test2")}}
	payload2, err := proto.Marshal(b)
	assert.Nil(t, err, "Failed to marshal mock block")

	peer1 := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: payload1, Status: 200}
	peer2 := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, Payload: payload2, Status: 200}

	reqCtx, cancel := context.NewRequest(setupContext(), context.WithTimeout(10*time.Second))
	defer cancel()

	_, err = channel.QueryConfigBlock(reqCtx, []fab.ProposalProcessor{&peer1, &peer2}, &TransactionProposalResponseVerifier{MinResponses: 2})
	assert.Nil(t, err, "Expected success querying blocks with identical block data payloads")
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
	f, errs := filterResponses(tprs, err, &TestVerifier{})
	assert.Len(t, f, 51)
	assert.Len(t, errs.(multi.Errors), 51)
}

func TestFilterResponsesWithVerifyError(t *testing.T) {
	tprs := []*fab.TransactionProposalResponse{}
	err := fmt.Errorf("test")
	tprs = append(tprs, &fab.TransactionProposalResponse{Status: 200})
	f, errs := filterResponses(tprs, err, &TestVerifier{verifyErr: errors.New("error")})
	assert.Len(t, f, 0)
	assert.Len(t, errs.(multi.Errors), 2)
}

func setupTestLedger() (*Ledger, error) {
	return setupLedger("testChannel")
}

func setupLedger(channelID string) (*Ledger, error) {
	return NewLedger(channelID)
}

func setupContext() contextApi.Client {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := mocks.NewMockContext(user)
	return ctx
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
