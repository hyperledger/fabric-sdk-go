/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	reqContext "context"
	"crypto/x509"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go/common"
	ab "github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockfab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var testOrdererURL = "127.0.0.1:0"
var testOrdererMainURL = "127.0.0.1:0"

var ordererAddr string
var ordererMockSrv *mocks.MockBroadcastServer

func TestMain(m *testing.M) {
	ordererMockSrv = &mocks.MockBroadcastServer{}
	ordererAddr = ordererMockSrv.Start(testOrdererMainURL)
	defer ordererMockSrv.Stop()

	os.Exit(m.Run())
}

func TestSendDeliverHappy(t *testing.T) {
	ordererConfig := getGRPCOpts(ordererAddr, true, false, true)

	orderer, _ := New(mocks.NewMockEndpointConfig(), FromOrdererConfig(ordererConfig))
	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 15*time.Second)
	defer cancel()
	blocks, errs := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

	select {
	case block, ok := <-blocks:
		if !ok {
			t.Fatalf("Expected test block but got nothing")
		}
		if string(block.Data.Data[0]) != "test" {
			t.Fatalf("Expected test block got: %#v", block)
		}
	case err := <-errs:
		t.Fatalf("Unexpected error from SendDeliver(): %s", err)
	case <-time.After(time.Second * 5):
		t.Fatal("Did not receive block or error from SendDeliver")
	}

	assert.Len(t, errs, 0, "not supposed to get error")
}

func TestSendDeliverErr(t *testing.T) {
	ordererConfig := getGRPCOpts(ordererAddr, true, false, true)

	orderer, _ := New(mocks.NewMockEndpointConfig(), FromOrdererConfig(ordererConfig))
	// Test deliver with deliver error from OS
	testError := errors.New("test error")
	ordererMockSrv.DeliverError = testError
	defer func() { ordererMockSrv.DeliverError = nil }()

	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 5*time.Second)
	defer cancel()
	blocks, errs := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

read:
	for {
		select {
		case block, ok := <-blocks:
			if ok {
				t.Fatalf("Expected error got block: %#v", block)
				break read
			}
		case err := <-errs:
			if err == nil {
				t.Fatal("Expected test error when OS Recv() fails, got nil")
			} else {
				t.Logf("There is an error as expected : %s", err)
			}
			break read
		case <-time.After(time.Second * 5):
			t.Fatal("Did not receive block or error from SendDeliver")
			break read
		}
	}
}

func TestSendDeliverConnFailed(t *testing.T) {
	orderer, _ := New(mocks.NewMockEndpointConfig(), WithURL(testOrdererURL+"invalid-test"))
	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 5*time.Second)
	defer cancel()

	blocks, errs := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

	select {
	case block, ok := <-blocks:
		if ok {
			t.Fatalf("This usecase was not supposed to receive blocks : %#v", block)
		}
	case <-time.After(time.Second * 5):
		t.Fatal("Did not receive closed response from SendDeliver")
	}

	select {
	case err := <-errs:
		t.Logf("There is an error as expected : %s", err)
	case <-time.After(time.Second * 5):
		t.Fatal("Did not receive error from SendDeliver")
	}

}

func TestNewOrdererWithTLS(t *testing.T) {
	tlsConfig := endpoint.TLSConfig{Path: filepath.Join("testdata", "ca.crt")}
	err := tlsConfig.LoadBytes()
	if err != nil {
		t.Fatalf("tlsConfig.LoadBytes() failed, cause [%s]", err)
	}

	cert, ok, err := tlsConfig.TLSCert()
	if err != nil || !ok {
		t.Fatalf("Testing New with TLS failed, cause [%s]", err)
	}
	orderer, err := New(mocks.NewMockEndpointConfigCustomized(true, false, false), WithURL("grpcs://"), WithTLSCert(cert))

	if orderer == nil || err != nil {
		t.Fatalf("Testing New with TLS failed, cause [%s]", err)
	}

	//Negative Test case
	orderer, err = New(mocks.NewMockEndpointConfigCustomized(true, false, true), WithURL("grpcs://"))

	if orderer != nil || err == nil {
		t.Fatal("Testing New with TLS was supposed to fail")
	}
}

func TestNewOrdererWithMutualTLS(t *testing.T) {
	//Positive Test case
	tlsConfig := endpoint.TLSConfig{Path: filepath.Join("testdata", "ca.crt")}
	err := tlsConfig.LoadBytes()
	if err != nil {
		t.Fatalf("tlsConfig.LoadBytes() failed, cause [%s]", err)
	}

	cert, ok, err := tlsConfig.TLSCert()

	if err != nil || !ok {
		t.Fatalf("Testing New with TLS failed, cause [%s]", err)
	}

	orderer, err := New(mocks.NewMockEndpointConfigCustomized(true, true, false), WithURL("grpcs://"), WithTLSCert(cert))

	if orderer == nil || err != nil {
		t.Fatalf("Testing New with Mutual TLS failed, cause [%s]", err)
	}
	//Negative Test case
	orderer, err = New(mocks.NewMockEndpointConfigCustomized(true, false, false), WithURL("grpcs://"), WithTLSCert(cert))

	if orderer == nil || err != nil {
		t.Fatalf("Testing New with Mutual TLS failed, cause [%s]", err)
	}
}

func TestSendBroadcastHappy(t *testing.T) {

	ordererConfig := getGRPCOpts(ordererAddr, true, false, true)
	orderer, _ := New(mocks.NewMockEndpointConfig(), FromOrdererConfig(ordererConfig))

	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})
	assert.Nil(t, err)
}

func TestSendBroadcastTimeout(t *testing.T) {

	ordererConfig := getGRPCOpts(testOrdererURL+"Test", true, false, true)
	orderer, _ := New(mocks.NewMockEndpointConfig(), FromOrdererConfig(ordererConfig))
	orderer.dialTimeout = 15

	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})
	if err == nil {
		t.Fatal("Expected error 'Orderer Client Status 2 context deadline exceeded'")
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

	addr := broadcastServer.Start(testOrdererURL)
	defer broadcastServer.Stop()

	orderer, _ := New(mocks.NewMockEndpointConfig(), WithURL("grpc://"+addr), WithInsecure())

	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 5*time.Second)
	defer cancel()
	blocks, errors := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

read:
	for {
		select {
		case block, ok := <-blocks:
			if ok {
				t.Fatalf("This usecase was not supposed to receive blocks : %#v", block)
				break read
			}
		case err := <-errors:
			if !strings.Contains(err.Error(), "BAD_REQUEST") {
				t.Fatalf("Ordering service error is not received as expected, %s", err)
			}
			break read
		case <-time.After(time.Second * 5):
			t.Fatal("Did not receive error from SendDeliver")
			break read
		}
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

	addr := broadcastServer.Start(testOrdererURL)
	defer broadcastServer.Stop()

	orderer, _ := New(mocks.NewMockEndpointConfig(), WithURL("grpc://"+addr), WithInsecure())

	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 5*time.Second)
	defer cancel()
	blocks, errors := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

	select {
	case _, ok := <-blocks:
		if ok {
			t.Fatal("This usecase was not supposed to get valid block")
		}
	case <-time.After(time.Second * 5):
		t.Fatal("Did not receive block from SendDeliver")
	}

	assert.Len(t, errors, 0, "not supposed to get error")
}

func TestSendDeliverFailure(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		DeliverResponse: &ab.DeliverResponse{},
		DeliverError:    errors.New("fail me"),
	}

	addr := broadcastServer.Start(testOrdererURL)
	defer broadcastServer.Stop()

	orderer, _ := New(mocks.NewMockEndpointConfig(), WithURL("grpc://"+addr), WithInsecure())

	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 5*time.Second)
	defer cancel()
	blocks, errs := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

read:
	for {
		select {
		case block, ok := <-blocks:
			if ok {
				t.Fatalf("This usecase was not supposed to get valid block %+v", block)
				break read
			}
		case err := <-errs:
			if err == nil || !strings.HasPrefix(err.Error(), "recv from ordering service failed") {
				t.Fatalf("Error response is not working as expected : '%s' ", err)
			}
			break read
		case <-time.After(time.Second * 5):
			t.Fatal("Timeout: did not receive any response or error from SendDeliver")
			break read
		}
	}
}

func TestSendBroadcastServerBadResponse(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		BroadcastInternalServerError: true,
	}

	addr := broadcastServer.Start(testOrdererURL)
	defer broadcastServer.Stop()
	orderer, _ := New(mocks.NewMockEndpointConfig(), WithURL("grpc://"+addr), WithInsecure())

	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})

	if err == nil {
		t.Fatal("Expected error")
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

	addr := broadcastServer.Start(testOrdererURL)
	defer broadcastServer.Stop()

	orderer, _ := New(mocks.NewMockEndpointConfig(), WithURL("grpc://"+addr), WithInsecure())

	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})

	if err == nil {
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
	config := mockfab.NewMockEndpointConfig(mockCtrl)

	config.EXPECT().Timeout(fab.OrdererConnection).Return(time.Second * 1)
	config.EXPECT().TLSCACertPool().Return(&mockfab.MockCertPool{CertPool: x509.NewCertPool()}).AnyTimes()

	orderer, err := New(config, WithURL("grpc://127.0.0.1:0"))
	assert.Nil(t, err)
	orderer.grpcDialOption = append(orderer.grpcDialOption, grpc.WithBlock())
	orderer.allowInsecure = true
	_, err = orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})
	assert.NotNil(t, err)

	if err == nil || !strings.Contains(err.Error(), "CONNECTION_FAILED") {
		t.Fatalf("Expected connection issues, but got %s ", err)
	}

}

func TestGetKeepAliveOptions(t *testing.T) {
	grpcOpts := make(map[string]interface{})

	grpcOpts["keep-alive-time"] = "s"
	grpcOpts["keep-alive-timeout"] = 2 * time.Second
	grpcOpts["keep-alive-permit"] = false
	//orderer config with GRPC opts
	ordererConfig := &fab.OrdererConfig{
		GRPCOptions: grpcOpts,
	}
	kap := getKeepAliveOptions(ordererConfig)
	if kap.Time != 0 {
		t.Fatal("Expected 0 time for incorrect keep-alive-time")
	}
	if kap.Timeout != 2*time.Second {
		t.Fatal("Expected 2 seconds for keep-alive-timeout")
	}
	assert.EqualValues(t, kap.Time, 0)
	assert.EqualValues(t, kap.Timeout, 2*time.Second)
	assert.EqualValues(t, kap.PermitWithoutStream, false)

}

func TestFailFast(t *testing.T) {
	grpcOpts := make(map[string]interface{})
	ordererConfig := &fab.OrdererConfig{
		GRPCOptions: grpcOpts,
	}
	failFast := getFailFast(ordererConfig)
	assert.EqualValues(t, failFast, true)

	grpcOpts["fail-fast"] = false
	ordererConfig = &fab.OrdererConfig{
		GRPCOptions: grpcOpts,
	}
	failFast = getFailFast(ordererConfig)
	assert.EqualValues(t, failFast, false)
}

func getGRPCOpts(addr string, failFast bool, keepAliveOptions bool, allowInSecure bool) *fab.OrdererConfig {
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
	ordererConfig := &fab.OrdererConfig{
		URL:         "grpc://" + addr,
		GRPCOptions: grpcOpts,
	}

	return ordererConfig
}

func TestForDeadlineExceeded(t *testing.T) {
	orderer, _ := New(mocks.NewMockEndpointConfig(), WithURL(testOrdererURL+"Test"))
	orderer.dialTimeout = 1 * time.Second
	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})
	if err == nil || !strings.HasPrefix(err.Error(), "Orderer Client Status Code") {
		t.Fatalf("Test SendBroadcast was supposed to fail with 'Orderer Client Status Code ...', instead it failed with [%s] error", err)
	}
}

func TestSendDeliverDefaultOpts(t *testing.T) {
	//keep alive option is not set and fail fast is false - invalid URL
	orderer, _ := New(mocks.NewMockEndpointConfig(), WithURL("grpc://"+testOrdererURL+"Test"), WithInsecure())
	orderer.dialTimeout = 5 * time.Second
	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})
	if err == nil {
		t.Fatalf("Expected error 'Orderer Client Status 2 context deadline exceeded' %s", err)
	}

	orderer, _ = New(mocks.NewMockEndpointConfig(), WithURL("grpc://"+ordererAddr), WithInsecure())
	orderer.dialTimeout = 5 * time.Second
	// Test deliver happy path
	ctx, cancel := reqContext.WithTimeout(reqContext.Background(), 5*time.Second)
	defer cancel()
	blocks, errs := orderer.SendDeliver(ctx, &fab.SignedEnvelope{})

read:
	for {
		select {
		case block, ok := <-blocks:
			if !ok {
				t.Fatalf("Expected test block got nothing")
			}

			if string(block.Data.Data[0]) != "test" {
				t.Fatalf("Expected test block got: %#v", block)
			}
			break read
		case err := <-errs:
			t.Fatalf("Unexpected error from SendDeliver(): %s", err)
			break read
		case <-time.After(time.Second * 5):
			t.Fatal("Did not receive block or error from SendDeliver")
			break read
		}
	}

	assert.Len(t, errs, 0, "not supposed to get error")
}

/*
func TestForGRPCErrorsWithKeepAliveOptsFailFast(t *testing.T) {
	//keep alive options set and failfast is true
	ordererConfig := getGRPCOpts("grpc://"+testOrdererURL+"Test", true, true)
	orderer, _ := New(mockcore.NewMockConfig(), WithURL(testOrdererURL+"Test"), FromOrdererConfig(ordererConfig))
	orderer.dialTimeout = 2 * time.Second
	_, err := orderer.SendBroadcast(&fab.SignedEnvelope{})
	if err == nil {
		t.Fatal("Expected error 'Orderer Client Status 2 context deadline exceeded'")
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
	orderer, _ := New(mocks.NewMockEndpointConfig(), FromOrdererConfig(ordererConfig))
	orderer.dialTimeout = 2 * time.Second
	_, err := orderer.SendBroadcast(reqContext.Background(), &fab.SignedEnvelope{})
	if err == nil {
		t.Fatal("Expected error 'Orderer Client Status 2 context deadline exceeded'")
	}
	statusError, ok := status.FromError(err)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, status.ConnectionFailed, status.ToOrdererStatusCode(statusError.Code))
	//assert.EqualValues(t, grpccodes.DeadlineExceeded, status.ToGRPCStatusCode(statusError.Code))
	assert.Equal(t, status.OrdererClientStatus, statusError.Group)
}

func TestNewOrdererFromOrdererName(t *testing.T) {
	t.Run("run simple FromOrdererName", func(t *testing.T){
		_, err := New(mocks.NewMockEndpointConfig(), FromOrdererName("orderer"))
		if err != nil {
			t.Fatalf("Failed to get new orderer from name. Error: %s", err)
		}
	})

	t.Run("run FromOrdererName with Ignore orderer in config", func(t *testing.T){
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockEndpoingCfg := mockfab.NewMockEndpointConfig(mockCtrl)

		mockEndpoingCfg.EXPECT().OrdererConfig("orderer").Return(nil, false, true)

		_, err := New(mockEndpoingCfg, FromOrdererName("orderer"))
		if err == nil {
			t.Fatal("Expected error but got nil")
		}
	})

	t.Run("run FromOrdererName with orderer not found in config", func(t *testing.T){
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockEndpoingCfg := mockfab.NewMockEndpointConfig(mockCtrl)

		mockEndpoingCfg.EXPECT().OrdererConfig("orderer").Return(nil, false, false)

		_, err := New(mockEndpoingCfg, FromOrdererName("orderer"))
		if err == nil {
			t.Fatal("Expected error but got nil")
		}
	})
}

func TestNewOrdererFromConfig(t *testing.T) {

	grpcOpts := make(map[string]interface{})
	//fail fast
	grpcOpts["fail-fast"] = true
	grpcOpts["keep-alive-time"] = 1 * time.Second
	grpcOpts["keep-alive-timeout"] = 2 * time.Second
	grpcOpts["keep-alive-permit"] = false
	//orderer config with GRPC opts
	ordererConfig := &fab.OrdererConfig{
		URL:         "",
		GRPCOptions: grpcOpts,
	}
	_, err := New(mocks.NewMockEndpointConfig(), FromOrdererConfig(ordererConfig))
	if err != nil {
		t.Fatalf("Failed to get new orderer from config. Error: %s", err)
	}
}

// TestNewOrdererSecured validates that insecure option
func TestNewOrdererSecured(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockfab.DefaultMockConfig(mockCtrl)
	config.EXPECT().Timeout(fab.OrdererConnection).Return(time.Second * 1).AnyTimes()

	//Test grpc URL
	url := "grpc://0.0.0.0:1234"
	_, err := New(config, WithURL(url), WithInsecure())
	if err != nil {
		t.Fatal("Peer conn should be constructed")
	}

	//Test grpcs URL
	url = "grpcs://0.0.0.0:1234"
	_, err = New(config, WithURL(url), WithInsecure())
	if err != nil {
		t.Fatal("Peer conn should be constructed")
	}

	//Test URL without protocol
	url = "0.0.0.0:1234"
	_, err = New(config, WithURL(url), WithInsecure())
	if err != nil {
		t.Fatal("Peer conn should be constructed")
	}

}
