/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	reqContext "context"
	"crypto/x509"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	grpcCodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	grpcstatus "google.golang.org/grpc/status"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	mockCore "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

const (
	peer1URL    = "localhost:0"
	peer2URL    = "localhost:0"
	testAddress = "127.0.0.1:0"
)

var kap keepalive.ClientParameters

// TestNewPeerEndorserTLS validates that a client configured with TLS
// creates the correct dial options.
func TestNewPeerEndorserTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockCore.DefaultMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"

	conn, err := newPeerEndorser(getPeerEndorserRequest(url, mockCore.GoodCert, "", config, kap, false, false))

	if err != nil {
		t.Fatalf("Peer conn should be constructed")
	}

	optInsecure := reflect.ValueOf(grpc.WithInsecure())

	for _, opt := range conn.grpcDialOption {
		optr := reflect.ValueOf(opt)
		if optr.Pointer() == optInsecure.Pointer() {
			t.Fatalf("TLS enabled - insecure not allowed")
		}
	}
}

func TestNewPeerEndorserMutualTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockCore.DefaultMockConfig(mockCtrl)

	//mutualTLSCerts := apiconfig.MutualTLSConfig{
	//	Client: struct {
	//		KeyPem   string
	//		Keyfile  string
	//		CertPem  string
	//		Certfile string
	//	}{KeyPem: "", Keyfile: "../../../test/fixtures/config/mutual_tls/client_sdk_go-key.pem", CertPem: "", Certfile: "../../../test/fixtures/config/mutual_tls/client_sdk_go.pem"},
	//}

	url := "grpcs://0.0.0.0:1234"
	conn, err := newPeerEndorser(getPeerEndorserRequest(url, mockCore.GoodCert, "", config, kap, false, false))

	if err != nil {
		t.Fatalf("Peer conn should be constructed: %v", err)
	}

	optInsecure := reflect.ValueOf(grpc.WithInsecure())

	for _, opt := range conn.grpcDialOption {
		optr := reflect.ValueOf(opt)
		if optr.Pointer() == optInsecure.Pointer() {
			t.Fatalf("TLS enabled - insecure not allowed")
		}
	}
}

func TestNewPeerEndorserMutualTLSNoClientCerts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockCore.DefaultMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"
	_, err := newPeerEndorser(getPeerEndorserRequest(url, mockCore.GoodCert, "", config, kap, false, false))
	if err != nil {
		t.Fatalf("Peer conn should be constructed: %v", err)
	}
}

// TestNewPeerEndorserTLSBadPool validates that a client configured with TLS
// with a bad cert pool fails gracefully.
func TestNewPeerEndorserTLSBadPool(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockCore.DefaultMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"
	_, err := newPeerEndorser(getPeerEndorserRequest(url, mockCore.BadCert, "", config, kap, false, false))
	if err == nil {
		t.Fatalf("Peer conn construction should have failed")
	}
}

// TestNewPeerEndorserSecured validates that secured and allowinsecure options
func TestNewPeerEndorserSecured(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockCore.DefaultMockConfig(mockCtrl)

	url := "grpc://0.0.0.0:1234"
	_, err := newPeerEndorser(getPeerEndorserRequest(url, nil, "", config, kap, false, false))
	if err != nil {
		t.Fatalf("Peer conn should be constructed")
	}

	url = "grpcs://0.0.0.0:1234"

	_, err = newPeerEndorser(getPeerEndorserRequest(url, nil, "", config, kap, false, true))
	if err != nil {
		t.Fatalf("Peer conn should be constructed")
	}

	url = "0.0.0.0:1234"

	_, err = newPeerEndorser(getPeerEndorserRequest(url, nil, "", config, kap, false, true))
	if err != nil {
		t.Fatalf("Peer conn should be constructed")
	}

}

// TestNewPeerEndorserBadParams validates that a client configured without
// params fails
func TestNewPeerEndorserBadParams(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockCore.DefaultMockConfig(mockCtrl)

	url := ""
	_, err := newPeerEndorser(getPeerEndorserRequest(url, nil, "", config, kap, false, false))
	if err == nil {
		t.Fatalf("Peer conn should not be constructed - bad params")
	}
}

// TestNewPeerEndorserTLSBad validates that a client configured without
// the cert pool fails
func TestNewPeerEndorserTLSBad(t *testing.T) {
	config := mocks.NewMockConfigCustomized(true, false, true)
	url := "grpcs://0.0.0.0:1234"

	_, err := newPeerEndorser(getPeerEndorserRequest(url, nil, "", config, kap, false, false))
	if err == nil {
		t.Fatalf("Peer conn should not be constructed - bad cert pool")
	}
}

// TestProcessProposalBadDial validates that a down
// endorser fails gracefully.
func TestProcessProposalBadDial(t *testing.T) {
	_, err := testProcessProposal(t, "grpc://"+testAddress)
	if err == nil {
		t.Fatalf("Process proposal should have failed")
	}
}

// TestProcessProposalGoodDial validates that an up
// endorser connects.
func TestProcessProposalGoodDial(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := startEndorserServer(t, grpcServer)

	_, err := testProcessProposal(t, "grpc://"+addr)
	if err != nil {
		t.Fatalf("Process proposal failed (%v)", err)
	}
}

func testProcessProposal(t *testing.T, url string) (*fab.TransactionProposalResponse, error) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mockCore.DefaultMockConfig(mockCtrl)
	config.EXPECT().TimeoutOrDefault(gomock.Any()).Return(time.Second * 1).AnyTimes()

	conn, err := newPeerEndorser(getPeerEndorserRequest(url, nil, "", config, kap, false, true))
	if err != nil {
		t.Fatalf("Peer conn construction error (%v)", err)
	}

	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), normalTimeout)
	defer cancel()
	return conn.ProcessTransactionProposal(ctx, mockProcessProposalRequest())
}

func getPeerEndorserRequest(url string, cert *x509.Certificate, serverHostOverride string,
	config core.Config, kap keepalive.ClientParameters, failFast bool, allowInsecure bool) *peerEndorserRequest {
	return &peerEndorserRequest{
		target:             url,
		certificate:        cert,
		serverHostOverride: serverHostOverride,
		config:             config,
		kap:                kap,
		failFast:           false,
		allowInsecure:      allowInsecure,
		commManager:        &defCommManager{},
	}

}

func mockProcessProposalRequest() fab.ProcessProposalRequest {
	return fab.ProcessProposalRequest{
		SignedProposal: &pb.SignedProposal{},
	}
}

func startEndorserServer(t *testing.T, grpcServer *grpc.Server) (*mocks.MockEndorserServer, string) {
	return startEndorserServerWithError(t, grpcServer, nil)
}

func startEndorserServerWithError(t *testing.T, grpcServer *grpc.Server, testErr error) (*mocks.MockEndorserServer, string) {
	lis, err := net.Listen("tcp", testAddress)
	addr := lis.Addr().String()

	endorserServer := &mocks.MockEndorserServer{ProposalError: testErr}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	if err != nil {
		t.Logf("Error starting test server %s", err)
		t.FailNow()
	}
	t.Logf("Starting test server (endorser server in peerendorser_test) on %s", addr)
	go grpcServer.Serve(lis)
	return endorserServer, addr
}

func TestEndorserConnectionError(t *testing.T) {
	_, err := testProcessProposal(t, "grpc://"+testAddress)
	assert.NotNil(t, err, "Expected connection error without server running")

	statusError, ok := status.FromError(err)
	assert.True(t, ok, "Expected status error on failed connection")
	assert.Equal(t, status.EndorserClientStatus, statusError.Group)
	assert.Equal(t, int32(status.ConnectionFailed), statusError.Code)
}

func TestEndorserRPCError(t *testing.T) {
	testErrorMessage := "RPC error condition"

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := startEndorserServerWithError(t, grpcServer, fmt.Errorf(testErrorMessage))

	_, err := testProcessProposal(t, "grpc://"+addr)
	statusError, ok := status.FromError(err)
	assert.True(t, ok, "Expected status error on failed connection")
	assert.Equal(t, status.GRPCTransportStatus, statusError.Group)
	assert.Equal(t, testErrorMessage, statusError.Message)

	grpcCode := status.ToGRPCStatusCode(statusError.Code)
	assert.Equal(t, grpcCodes.Unknown, grpcCode)
}

func TestExtractChainCodeError(t *testing.T) {
	expectedMsg := "Chaincode error(status: 500, message: Invalid function (dummy) call)"
	error := grpcstatus.New(grpcCodes.Unknown, expectedMsg)
	code, message, _ := extractChaincodeError(error)
	if code != 500 {
		t.Fatalf("Expected code to be 500")
	}
	if message != "Invalid function (dummy) call" {
		t.Fatalf("Expected message not found")
	}
}
