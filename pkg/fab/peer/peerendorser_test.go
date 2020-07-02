/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	reqContext "context"
	"crypto/x509"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	grpcCodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockfab"
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

	config := mockfab.DefaultMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"

	conn, err := newPeerEndorser(getPeerEndorserRequest(url, mockfab.GoodCert, "", config, kap, false, false))

	if err != nil {
		t.Fatal("Peer conn should be constructed")
	}

	optInsecure := reflect.ValueOf(grpc.WithInsecure())

	for _, opt := range conn.grpcDialOption {
		optr := reflect.ValueOf(opt)
		if optr.Pointer() == optInsecure.Pointer() {
			t.Fatal("TLS enabled - insecure not allowed")
		}
	}
}

func TestNewPeerEndorserMutualTLS(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.DefaultMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"
	conn, err := newPeerEndorser(getPeerEndorserRequest(url, mockfab.GoodCert, "", config, kap, false, false))

	if err != nil {
		t.Fatalf("Peer conn should be constructed: %s", err)
	}

	optInsecure := reflect.ValueOf(grpc.WithInsecure())

	for _, opt := range conn.grpcDialOption {
		optr := reflect.ValueOf(opt)
		if optr.Pointer() == optInsecure.Pointer() {
			t.Fatal("TLS enabled - insecure not allowed")
		}
	}
}

func TestNewPeerEndorserMutualTLSNoClientCerts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.DefaultMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"
	_, err := newPeerEndorser(getPeerEndorserRequest(url, mockfab.GoodCert, "", config, kap, false, false))
	if err != nil {
		t.Fatalf("Peer conn should be constructed: %s", err)
	}
}

// TestNewPeerEndorserTLSBadPool validates that a client configured with TLS
// with a bad cert pool fails gracefully.
func TestNewPeerEndorserTLSBadPool(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.BadTLSClientMockConfig(mockCtrl)

	url := "grpcs://0.0.0.0:1234"
	_, err := newPeerEndorser(getPeerEndorserRequest(url, mockfab.BadCert, "", config, kap, false, false))
	if err == nil {
		t.Fatal("Peer conn construction should have failed")
	}
}

// TestNewPeerEndorserSecured validates that secured and allowinsecure options
func TestNewPeerEndorserSecured(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.DefaultMockConfig(mockCtrl)

	url := "grpc://0.0.0.0:1234"
	_, err := newPeerEndorser(getPeerEndorserRequest(url, nil, "", config, kap, false, false))
	if err != nil {
		t.Fatal("Peer conn should be constructed")
	}

	url = "grpcs://0.0.0.0:1234"

	_, err = newPeerEndorser(getPeerEndorserRequest(url, nil, "", config, kap, false, true))
	if err != nil {
		t.Fatal("Peer conn should be constructed")
	}

	url = "0.0.0.0:1234"

	_, err = newPeerEndorser(getPeerEndorserRequest(url, nil, "", config, kap, false, true))
	if err != nil {
		t.Fatal("Peer conn should be constructed")
	}

}

// TestNewPeerEndorserBadParams validates that a client configured without
// params fails
func TestNewPeerEndorserBadParams(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.DefaultMockConfig(mockCtrl)

	url := ""
	_, err := newPeerEndorser(getPeerEndorserRequest(url, nil, "", config, kap, false, false))
	if err == nil {
		t.Fatal("Peer conn should not be constructed - bad params")
	}
}

// TestNewPeerEndorserTLSBad validates that a client configured without
// the cert pool fails
func TestNewPeerEndorserTLSBad(t *testing.T) {
	config := mocks.NewMockEndpointConfigCustomized(true, false, true)
	url := "grpcs://0.0.0.0:1234"

	_, err := newPeerEndorser(getPeerEndorserRequest(url, nil, "", config, kap, false, false))
	if err == nil {
		t.Fatal("Peer conn should not be constructed - bad cert pool")
	}
}

// TestProcessProposalBadDial validates that a down
// endorser fails gracefully.
func TestProcessProposalBadDial(t *testing.T) {
	_, err := testProcessProposal(t, "grpc://"+testAddress)
	if err == nil {
		t.Fatal("Process proposal should have failed")
	}
}

// TestProcessProposalGoodDial validates that an up
// endorser connects.
func TestProcessProposalGoodDial(t *testing.T) {
	srv := &mocks.MockEndorserServer{}
	addr := srv.Start(testAddress)
	defer srv.Stop()

	_, err := testProcessProposal(t, "grpc://"+addr)
	if err != nil {
		t.Fatalf("Process proposal failed (%s)", err)
	}
}

func testProcessProposal(t *testing.T, url string) (*fab.TransactionProposalResponse, error) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mockfab.DefaultMockConfig(mockCtrl)
	config.EXPECT().Timeout(gomock.Any()).Return(time.Second * 1).AnyTimes()

	conn, err := newPeerEndorser(getPeerEndorserRequest(url, nil, "", config, kap, false, true))
	if err != nil {
		t.Fatalf("Peer conn construction error (%s)", err)
	}

	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), normalTimeout)
	defer cancel()
	return conn.ProcessTransactionProposal(ctx, mockProcessProposalRequest())
}

func getPeerEndorserRequest(url string, cert *x509.Certificate, serverHostOverride string,
	config fab.EndpointConfig, kap keepalive.ClientParameters, failFast bool, allowInsecure bool) *peerEndorserRequest {
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

	srv := &mocks.MockEndorserServer{ProposalError: fmt.Errorf(testErrorMessage)}
	addr := srv.Start(testAddress)
	defer srv.Stop()

	_, err := testProcessProposal(t, "grpc://"+addr)
	statusError, ok := status.FromError(err)
	assert.True(t, ok, "Expected status error on failed connection")
	assert.Equal(t, status.GRPCTransportStatus, statusError.Group)
	assert.Equal(t, testErrorMessage, statusError.Message)

	grpcCode := status.ToGRPCStatusCode(statusError.Code)
	assert.Equal(t, grpcCodes.Unknown, grpcCode)
}

func TestChaincodeStatusFromResponse(t *testing.T) {
	//For error response
	response := &pb.ProposalResponse{
		Response: &pb.Response{Status: 500, Payload: []byte("Unknown function"), Message: "Chaincode error"},
	}
	err := extractChaincodeErrorFromResponse(response)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, "Chaincode error", s.Message)
	assert.Equal(t, int32(500), s.Code)
	assert.Equal(t, status.ChaincodeStatus, s.Group)
	assert.Equal(t, []byte("Unknown function"), s.Details[1])

	//For successful response 200
	response = &pb.ProposalResponse{
		Response: &pb.Response{Status: 200, Payload: []byte("TEST"), Message: "Success"},
	}
	err = extractChaincodeErrorFromResponse(response)
	assert.True(t, ok)
	assert.Nil(t, err)

	//For successful response 201
	response = &pb.ProposalResponse{
		Response: &pb.Response{Status: 201, Payload: []byte("TEST"), Message: "Success"},
	}
	err = extractChaincodeErrorFromResponse(response)
	assert.True(t, ok)
	assert.Nil(t, err)

	//For error response - chaincode name not found (v1)
	response = &pb.ProposalResponse{
		Response: &pb.Response{Status: 500, Payload: []byte("Unknown Description"), Message: "make sure the chaincode uq7q9y7lu7 has been successfully instantiated and try again: getccdata mychannel/uq7q9y7lu7 responded with error: could not find chaincode with name 'uq7q9y7lu7'"},
	}
	err = extractChaincodeErrorFromResponse(response)
	s, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, "make sure the chaincode uq7q9y7lu7 has been successfully instantiated and try again: getccdata mychannel/uq7q9y7lu7 responded with error: could not find chaincode with name 'uq7q9y7lu7'", s.Message)
	assert.Equal(t, int32(status.ChaincodeNameNotFound), s.Code)
	assert.Equal(t, status.EndorserClientStatus, s.Group)

	//For error response - chaincode name not found (v2)
	response = &pb.ProposalResponse{
		Response: &pb.Response{Status: 500, Payload: []byte("Unknown Description"), Message: "make sure the chaincode exampleCC2 has been successfully defined on channel orgchannel and try again: chaincode exampleCC2 not found"},
	}
	err = extractChaincodeErrorFromResponse(response)
	s, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, "make sure the chaincode exampleCC2 has been successfully defined on channel orgchannel and try again: chaincode exampleCC2 not found", s.Message)
	assert.Equal(t, int32(status.ChaincodeNameNotFound), s.Code)
	assert.Equal(t, status.EndorserClientStatus, s.Group)

	//For error response - chaincode package not found
	response = &pb.ProposalResponse{
		Response: &pb.Response{Status: 500, Payload: []byte("Unknown Description"), Message: "cannot get package for chaincode (vl5knffa37:v0)"},
	}
	err = extractChaincodeErrorFromResponse(response)
	s, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, "cannot get package for chaincode (vl5knffa37:v0)", s.Message)
	assert.Equal(t, int32(status.ChaincodeNameNotFound), s.Code)
	assert.Equal(t, status.EndorserClientStatus, s.Group)
}
