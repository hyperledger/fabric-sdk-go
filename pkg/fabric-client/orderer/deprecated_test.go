/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/pkg/errors"
)

var kap keepalive.ClientParameters

func TestDeprecatedSendDeliver(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	mockServer, addr := startMockServer(t, grpcServer)

	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig(), kap)
	// Test deliver happy path
	blocks, errs := orderer.SendDeliver(&fab.SignedEnvelope{})
	select {
	case block := <-blocks:
		if string(block.Data.Data[0]) != "test" {
			t.Fatalf("Expected test block got: %#v", block)
		}
	case err := <-errs:
		t.Fatalf("Unexpected error from SendDeliver(): %s", err)
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive block or error from SendDeliver")
	}

	// Test deliver without valid envelope
	blocks, errs = orderer.SendDeliver(nil)
	select {
	case block := <-blocks:
		t.Fatalf("Expected error got block: %#v", block)
	case err := <-errs:
		if err == nil {
			t.Fatalf("Expected error with nil envelope")
		}
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive block or error from SendDeliver")
	}

	// Test deliver with deliver error from OS
	testError := errors.New("test error")
	mockServer.DeliverError = testError
	blocks, errs = orderer.SendDeliver(&fab.SignedEnvelope{})
	select {
	case block := <-blocks:
		t.Fatalf("Expected error got block: %#v", block)
	case err := <-errs:
		if err == nil {
			t.Fatalf("Expected test error when OS Recv() fails, got: %s", err)
		}
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive block or error from SendDeliver")
	}

	orderer, _ = NewOrderer(testOrdererURL+"invalid-test", "", "", mocks.NewMockConfig(), kap)
	// Test deliver happy path
	blocks, errs = orderer.SendDeliver(&fab.SignedEnvelope{})
	select {
	case block := <-blocks:
		t.Fatalf("This usecase was not supposed to receive blocks : %#v", block)
	case err := <-errs:
		t.Logf("There is an error as expected : %s", err)
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive error from SendDeliver")
	}

}

func TestDeprecatedNewOrdererWithTLS(t *testing.T) {
	//Positive Test case
	orderer, err := NewOrderer("grpcs://", "../../../test/fixtures/fabricca/tls/ca/ca_root.pem", "", mocks.NewMockConfigCustomized(true, false, false), kap)
	if orderer == nil || err != nil {
		t.Fatalf("Testing NewOrderer with TLS failed, cause [%s]", err)
	}

	//Negative Test case
	orderer, err = NewOrderer("grpcs://", "", "", mocks.NewMockConfigCustomized(true, false, true), kap)
	if orderer != nil || err == nil {
		t.Fatalf("Testing NewOrderer with TLS was supposed to fail")
	}
}

func TestDeprecatedNewOrdererWithMutualTLS(t *testing.T) {
	//Positive Test case
	orderer, err := NewOrderer("grpcs://", "../../../test/fixtures/fabricca/tls/ca/ca_root.pem", "", mocks.NewMockConfigCustomized(true, true, false), kap)
	if orderer == nil || err != nil {
		t.Fatalf("Testing NewOrderer with Mutual TLS failed, cause [%s]", err)
	}
	//Negative Test case
	orderer, err = NewOrderer("grpcs://", "../../../test/fixtures/fabricca/tls/ca/ca_root.pem", "", mocks.NewMockConfigCustomized(true, false, false), kap)
	if orderer == nil || err != nil {
		t.Fatalf("Testing NewOrderer with Mutual TLS failed, cause [%s]", err)
	}
}

func TestDeprecatedSendBroadcast(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := startMockServer(t, grpcServer)

	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig(), kap)
	_, err := orderer.SendBroadcast(&fab.SignedEnvelope{})

	if err != nil {
		t.Fatalf("Test SendBroadcast was not supposed to fail")
	}

	orderer, _ = NewOrderer(testOrdererURL+"Test", "", "", mocks.NewMockConfig(), kap)
	_, err = orderer.SendBroadcast(&fab.SignedEnvelope{})

	if err == nil || !strings.HasPrefix(err.Error(), "NewAtomicBroadcastClient") {
		t.Fatalf("Test SendBroadcast was supposed to fail with expected error, instead it fail with [%s] error", err)
	}

}

func TestDeprecatedSendDeliverServerBadResponse(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		DeliverResponse: &ab.DeliverResponse{
			Type: &ab.DeliverResponse_Status{
				Status: common.Status_BAD_REQUEST,
			},
		},
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig(), kap)

	blocks, errors := orderer.SendDeliver(&fab.SignedEnvelope{})

	select {
	case block := <-blocks:
		t.Fatalf("This usecase was not supposed to receive blocks : %#v", block)
	case err := <-errors:
		if err.Error() != "error status from ordering service BAD_REQUEST" {
			t.Fatalf("Ordering service error is not received as expected, %s", err)
		}
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive error from SendDeliver")
	}
}

func TestDeprecatedSendDeliverServerSuccessResponse(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		DeliverResponse: &ab.DeliverResponse{
			Type: &ab.DeliverResponse_Status{
				Status: common.Status_SUCCESS,
			},
		},
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)

	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig(), kap)

	blocks, errors := orderer.SendDeliver(&fab.SignedEnvelope{})

	select {
	case block := <-blocks:
		if block != nil {
			t.Fatalf("This usecase was not supposed to get valid block")
		}
	case err := <-errors:
		t.Fatalf("This usecase was not supposed to get error : %s ", err.Error())
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive block from SendDeliver")
	}
}

func TestDeprecatedSendDeliverFailure(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		DeliverResponse: &ab.DeliverResponse{},
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig(), kap)

	blocks, errors := orderer.SendDeliver(&fab.SignedEnvelope{})

	select {
	case block := <-blocks:
		t.Fatalf("This usecase was not supposed to get valid block %v", block)
	case err := <-errors:
		if err == nil || !strings.HasPrefix(err.Error(), "unknown response from ordering service") {
			t.Fatalf("Error response is not working as expected : '%s' ", err.Error())
		}
	case <-time.After(time.Second * 5):
		t.Fatal("Did not receive any response or error from SendDeliver")
	}
}

func TestDeprecatedSendBroadcastServerBadResponse(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		BroadcastInternalServerError: true,
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig(), kap)

	_, err := orderer.SendBroadcast(&fab.SignedEnvelope{})

	if err == nil {
		t.Fatalf("Expected error")
	}
	statusError, ok := status.FromError(err)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, common.Status_INTERNAL_SERVER_ERROR, status.ToOrdererStatusCode(statusError.Code))
	assert.Equal(t, status.OrdererServerStatus, statusError.Group)
}

func TestDeprecatedSendBroadcastError(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		BroadcastError: errors.New("just to test error scenario"),
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := NewOrderer(addr, "", "", mocks.NewMockConfig(), kap)

	status, err := orderer.SendBroadcast(&fab.SignedEnvelope{})

	if err == nil || status != nil {
		t.Fatalf("expected Send Broadcast to fail with error, but got %s", err)
	}

}

func TestDeprecatedInterfaces(t *testing.T) {
	var apiOrderer fab.Orderer
	var orderer Orderer

	apiOrderer = &orderer
	if apiOrderer == nil {
		t.Fatalf("this shouldn't happen.")
	}
}
