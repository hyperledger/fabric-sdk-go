/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	reqContext "context"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"

	"github.com/golang/mock/gomock"
	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	mockCore "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/errors/status"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var testOrdererURL = "127.0.0.1:0"

var ordererAddr string
var ordererMockSrv *mocks.MockBroadcastServer

func TestMain(m *testing.M) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	ordererMockSrv, ordererAddr = startMockServer(grpcServer)
	os.Exit(m.Run())
}

func TestSendDeliverHappy(t *testing.T) {
	ordererConfig := getGRPCOpts(ordererAddr, true, false, true)

	orderer, _ := New(mocks.NewMockConfig(), FromOrdererConfig(ordererConfig))
	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 15*time.Second)
	defer cancel()
	blocks, errs := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

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
}

func TestSendDeliverErr(t *testing.T) {
	ordererConfig := getGRPCOpts(ordererAddr, true, false, true)

	orderer, _ := New(mocks.NewMockConfig(), FromOrdererConfig(ordererConfig))
	// Test deliver with deliver error from OS
	testError := errors.New("test error")
	ordererMockSrv.DeliverError = testError
	defer func() { ordererMockSrv.DeliverError = nil }()

	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 5*time.Second)
	defer cancel()
	blocks, errs := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

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
}

func TestSendDeliver(t *testing.T) {
	orderer, _ := New(mocks.NewMockConfig(), WithURL(testOrdererURL+"invalid-test"))

	// Test deliver happy path
	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 5*time.Second)
	defer cancel()
	blocks, errs := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

	select {
	case block := <-blocks:
		t.Fatalf("This usecase was not supposed to receive blocks : %#v", block)
	case err := <-errs:
		t.Logf("There is an error as expected : %s", err)
	case <-time.After(time.Second * 5):
		t.Fatalf("Did not receive error from SendDeliver")
	}

}

func startMockServer(grpcServer *grpc.Server) (*mocks.MockBroadcastServer, string) {
	lis, err := net.Listen("tcp", testOrdererURL)
	addr := lis.Addr().String()

	broadcastServer := new(mocks.MockBroadcastServer)
	ab.RegisterAtomicBroadcastServer(grpcServer, broadcastServer)
	if err != nil {
		panic(fmt.Sprintf("Error starting test server %s", err))
	}
	fmt.Printf("Starting test server on %s\n", addr)
	go grpcServer.Serve(lis)

	return broadcastServer, addr
}

func startCustomizedMockServer(t *testing.T, serverURL string, grpcServer *grpc.Server, broadcastServer *mocks.MockBroadcastServer) string {
	lis, err := net.Listen("tcp", serverURL)
	addr := lis.Addr().String()

	ab.RegisterAtomicBroadcastServer(grpcServer, broadcastServer)
	if err != nil {
		t.Logf("Error starting test server %s", err)
		t.FailNow()
	}
	t.Logf("Starting test customized server on %s\n", addr)
	go grpcServer.Serve(lis)

	return addr
}

func TestNewOrdererWithTLS(t *testing.T) {
	tlsConfig := endpoint.TLSConfig{Path: "../../../test/fixtures/fabricca/tls/ca/ca_root.pem"}

	cert, err := tlsConfig.TLSCert()

	if err != nil {
		t.Fatalf("Testing New with TLS failed, cause [%s]", err)
	}
	orderer, err := New(mocks.NewMockConfigCustomized(true, false, false), WithURL("grpcs://"), WithTLSCert(cert))

	if orderer == nil || err != nil {
		t.Fatalf("Testing New with TLS failed, cause [%s]", err)
	}

	//Negative Test case
	orderer, err = New(mocks.NewMockConfigCustomized(true, false, true), WithURL("grpcs://"))

	if orderer != nil || err == nil {
		t.Fatalf("Testing New with TLS was supposed to fail")
	}
}

func TestNewOrdererWithMutualTLS(t *testing.T) {
	//Positive Test case
	tlsConfig := endpoint.TLSConfig{Path: "../../../test/fixtures/fabricca/tls/ca/ca_root.pem"}

	cert, err := tlsConfig.TLSCert()

	if err != nil {
		t.Fatalf("Testing New with TLS failed, cause [%s]", err)
	}

	orderer, err := New(mocks.NewMockConfigCustomized(true, true, false), WithURL("grpcs://"), WithTLSCert(cert))

	if orderer == nil || err != nil {
		t.Fatalf("Testing New with Mutual TLS failed, cause [%s]", err)
	}
	//Negative Test case
	orderer, err = New(mocks.NewMockConfigCustomized(true, false, false), WithURL("grpcs://"), WithTLSCert(cert))

	if orderer == nil || err != nil {
		t.Fatalf("Testing New with Mutual TLS failed, cause [%s]", err)
	}
}

func TestSendBroadcastHappy(t *testing.T) {

	ordererConfig := getGRPCOpts(ordererAddr, true, false, true)
	orderer, _ := New(mocks.NewMockConfig(), FromOrdererConfig(ordererConfig))

	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})
	assert.Nil(t, err)
}

func TestSendBroadcastTimeout(t *testing.T) {

	ordererConfig := getGRPCOpts(testOrdererURL+"Test", true, false, true)
	orderer, _ := New(mocks.NewMockConfig(), FromOrdererConfig(ordererConfig))
	orderer.dialTimeout = 15

	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})
	if err == nil {
		t.Fatalf("Expected error 'Orderer Client Status 2 context deadline exceeded'")
	}
	statusError, ok := status.FromError(err)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, grpccodes.Unknown, status.ToGRPCStatusCode(statusError.Code))
	assert.Equal(t, status.OrdererClientStatus, statusError.Group)
}

func TestSendDeliverServerBadResponse(t *testing.T) {

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
	orderer, _ := New(mocks.NewMockConfig(), WithURL("grpc://"+addr), WithInsecure())

	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 5*time.Second)
	defer cancel()
	blocks, errors := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

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

func TestSendDeliverServerSuccessResponse(t *testing.T) {

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

	orderer, _ := New(mocks.NewMockConfig(), WithURL("grpc://"+addr), WithInsecure())

	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 5*time.Second)
	defer cancel()
	blocks, errors := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

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

func TestSendDeliverFailure(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		DeliverResponse: &ab.DeliverResponse{},
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := New(mocks.NewMockConfig(), WithURL("grpc://"+addr), WithInsecure())

	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 5*time.Second)
	defer cancel()
	blocks, errors := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

	select {
	case block := <-blocks:
		t.Fatalf("This usecase was not supposed to get valid block %v", block)
	case err := <-errors:
		if err == nil || !strings.HasPrefix(err.Error(), "unknown response type from ordering service") {
			t.Fatalf("Error response is not working as expected : '%s' ", err.Error())
		}
	case <-time.After(time.Second * 5):
		t.Fatal("Did not receive any response or error from SendDeliver")
	}
}

func TestSendBroadcastServerBadResponse(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		BroadcastInternalServerError: true,
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := New(mocks.NewMockConfig(), WithURL("grpc://"+addr), WithInsecure())

	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})

	if err == nil {
		t.Fatalf("Expected error")
	}
	statusError, ok := status.FromError(err)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, common.Status_INTERNAL_SERVER_ERROR, status.ToOrdererStatusCode(statusError.Code))
	assert.Equal(t, status.OrdererServerStatus, statusError.Group)
}

func TestSendBroadcastError(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		BroadcastError: errors.New("just to test error scenario"),
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := New(mocks.NewMockConfig(), WithURL("grpc://"+addr), WithInsecure())

	statusCode, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})

	if err == nil || statusCode != nil {
		t.Fatalf("expected Send Broadcast to fail with error, but got %s", err)
	}
	statusError, ok := status.FromError(err)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, grpccodes.Unknown, status.ToGRPCStatusCode(statusError.Code))
	assert.Equal(t, status.GRPCTransportStatus, statusError.Group)
}

func TestBroadcastBadDial(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mockCore.NewMockConfig(mockCtrl)

	config.EXPECT().TimeoutOrDefault(core.OrdererConnection).Return(time.Second * 1)
	config.EXPECT().TLSCACertPool(gomock.Any()).Return(x509.NewCertPool(), nil).AnyTimes()

	orderer, err := New(config, WithURL("grpc://127.0.0.1:0"))
	assert.Nil(t, err)
	orderer.grpcDialOption = append(orderer.grpcDialOption, grpc.WithBlock())
	orderer.allowInsecure = true
	_, err = orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})
	assert.NotNil(t, err)

	if err == nil || !strings.Contains(err.Error(), "CONNECTION_FAILED") {
		t.Fatal("Expected connection issues, but got ", err)
	}

}

func TestGetKeepAliveOptions(t *testing.T) {
	grpcOpts := make(map[string]interface{})

	grpcOpts["keep-alive-time"] = "s"
	grpcOpts["keep-alive-timeout"] = 2 * time.Second
	grpcOpts["keep-alive-permit"] = false
	//orderer config with GRPC opts
	ordererConfig := &core.OrdererConfig{
		GRPCOptions: grpcOpts,
	}
	kap := getKeepAliveOptions(ordererConfig)
	if kap.Time != 0 {
		t.Fatalf("Expected 0 time for incorrect keep-alive-time")
	}
	if kap.Timeout != 2*time.Second {
		t.Fatalf("Expected 2 seconds for keep-alive-timeout")
	}
	assert.EqualValues(t, kap.Time, 0)
	assert.EqualValues(t, kap.Timeout, 2*time.Second)
	assert.EqualValues(t, kap.PermitWithoutStream, false)

}

func TestFailFast(t *testing.T) {
	grpcOpts := make(map[string]interface{})
	ordererConfig := &core.OrdererConfig{
		GRPCOptions: grpcOpts,
	}
	failFast := getFailFast(ordererConfig)
	assert.EqualValues(t, failFast, true)

	grpcOpts["fail-fast"] = false
	ordererConfig = &core.OrdererConfig{
		GRPCOptions: grpcOpts,
	}
	failFast = getFailFast(ordererConfig)
	assert.EqualValues(t, failFast, false)
}

func getGRPCOpts(addr string, failFast bool, keepAliveOptions bool, allowInSecure bool) *core.OrdererConfig {
	grpcOpts := make(map[string]interface{})
	//fail fast
	grpcOpts["fail-fast"] = failFast

	//keep alive options
	if keepAliveOptions {
		grpcOpts["keep-alive-time"] = 1 * time.Second
		grpcOpts["keep-alive-timeout"] = 2 * time.Second
		grpcOpts["keep-alive-permit"] = false

	}

	//allow in secure
	grpcOpts["allow-insecure"] = allowInSecure

	//orderer config with GRPC opts
	ordererConfig := &core.OrdererConfig{
		URL:         "grpc://" + addr,
		GRPCOptions: grpcOpts,
	}

	return ordererConfig
}

func TestForDeadlineExceeded(t *testing.T) {
	orderer, _ := New(mocks.NewMockConfig(), WithURL(testOrdererURL+"Test"))
	orderer.dialTimeout = 1 * time.Second
	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})
	if err == nil || !strings.HasPrefix(err.Error(), "Orderer Client Status Code") {
		t.Fatalf("Test SendBroadcast was supposed to fail with 'Orderer Client Status Code ...', instead it failed with [%s] error", err)
	}
}

func TestSendDeliverDefaultOpts(t *testing.T) {
	//keep alive option is not set and fail fast is false - invalid URL
	orderer, _ := New(mocks.NewMockConfig(), WithURL("grpc://"+testOrdererURL+"Test"), WithInsecure())
	orderer.dialTimeout = 5 * time.Second
	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})
	if err == nil {
		t.Fatalf("Expected error 'Orderer Client Status 2 context deadline exceeded' %v", err)
	}

	orderer, _ = New(mocks.NewMockConfig(), WithURL("grpc://"+ordererAddr), WithInsecure())
	orderer.dialTimeout = 5 * time.Second
	// Test deliver happy path
	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 5*time.Second)
	defer cancel()
	blocks, errs := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

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

}

/*
func TestForGRPCErrorsWithKeepAliveOptsFailFast(t *testing.T) {
	//keep alive options set and failfast is true
	ordererConfig := getGRPCOpts("grpc://"+testOrdererURL+"Test", true, true)
	orderer, _ := New(mocks.NewMockConfig(), WithURL(testOrdererURL+"Test"), FromOrdererConfig(ordererConfig))
	orderer.dialTimeout = 2 * time.Second
	_, err := orderer.SendBroadcast(&fab.SignedEnvelope{})
	if err == nil {
		t.Fatalf("Expected error 'Orderer Client Status 2 context deadline exceeded'")
	}
	//expect here GRPC unavaialble since fail fast is set to true
	statusError, ok := status.FromError(err)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, status.ConnectionFailed, status.ToOrdererStatusCode(statusError.Code))
	//	assert.EqualValues(t, grpccodes.Unavailable, status.ToGRPCStatusCode(statusError.Code))
	assert.Equal(t, status.OrdererClientStatus, statusError.Group)
}
*/

func TestForGRPCErrorsWithKeepAliveOpts(t *testing.T) {
	//expect here GRPC deadline exceeded since fail fast is set to false
	ordererConfig := getGRPCOpts(testOrdererURL+"Test", false, true, true)
	orderer, _ := New(mocks.NewMockConfig(), FromOrdererConfig(ordererConfig))
	orderer.dialTimeout = 2 * time.Second
	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})
	if err == nil {
		t.Fatalf("Expected error 'Orderer Client Status 2 context deadline exceeded'")
	}
	statusError, ok := status.FromError(err)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, status.ConnectionFailed, status.ToOrdererStatusCode(statusError.Code))
	//assert.EqualValues(t, grpccodes.DeadlineExceeded, status.ToGRPCStatusCode(statusError.Code))
	assert.Equal(t, status.OrdererClientStatus, statusError.Group)
}

func TestNewOrdererFromConfig(t *testing.T) {

	grpcOpts := make(map[string]interface{})
	//fail fast
	grpcOpts["fail-fast"] = true
	grpcOpts["keep-alive-time"] = 1 * time.Second
	grpcOpts["keep-alive-timeout"] = 2 * time.Second
	grpcOpts["keep-alive-permit"] = false
	//orderer config with GRPC opts
	ordererConfig := &core.OrdererConfig{
		URL:         "",
		GRPCOptions: grpcOpts,
	}
	_, err := New(mocks.NewMockConfig(), FromOrdererConfig(ordererConfig))
	if err != nil {
		t.Fatalf("Failed to get new orderer from config%v", err)
	}
}

// TestNewOrdererSecured validates that insecure option
func TestNewOrdererSecured(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockCore.DefaultMockConfig(mockCtrl)
	config.EXPECT().TimeoutOrDefault(core.OrdererConnection).Return(time.Second * 1).AnyTimes()

	//Test grpc URL
	url := "grpc://0.0.0.0:1234"
	_, err := New(config, WithURL(url), WithInsecure())
	if err != nil {
		t.Fatalf("Peer conn should be constructed")
	}

	//Test grpcs URL
	url = "grpcs://0.0.0.0:1234"
	_, err = New(config, WithURL(url), WithInsecure())
	if err != nil {
		t.Fatalf("Peer conn should be constructed")
	}

	//Test URL without protocol
	url = "0.0.0.0:1234"
	_, err = New(config, WithURL(url), WithInsecure())
	if err != nil {
		t.Fatalf("Peer conn should be constructed")
	}

}
