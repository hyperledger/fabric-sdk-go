/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"

	"github.com/golang/mock/gomock"
	apiconfig "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var testOrdererURL = "127.0.0.1:0"

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
-----END CERTIFICATE-----`

func TestSendDeliver(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	mockServer, addr := startMockServer(t, grpcServer)
	ordererConfig := getGRPCOpts(addr, true, false)

	orderer, _ := New(mocks.NewMockConfig(), WithURL(addr), FromOrdererConfig(ordererConfig))

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

	orderer, _ = New(mocks.NewMockConfig(), WithURL(testOrdererURL+"invalid-test"), FromOrdererConfig(ordererConfig))
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

func startMockServer(t *testing.T, grpcServer *grpc.Server) (*mocks.MockBroadcastServer, string) {
	lis, err := net.Listen("tcp", testOrdererURL)
	addr := lis.Addr().String()

	broadcastServer := new(mocks.MockBroadcastServer)
	ab.RegisterAtomicBroadcastServer(grpcServer, broadcastServer)
	if err != nil {
		t.Logf("Error starting test server %s", err)
		t.FailNow()
	}
	t.Logf("Starting test server on %s", addr)
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
	tlsConfig := apiconfig.TLSConfig{Path: "../../../test/fixtures/fabricca/tls/ca/ca_root.pem"}

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
	tlsConfig := apiconfig.TLSConfig{Path: "../../../test/fixtures/fabricca/tls/ca/ca_root.pem"}

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

func TestSendBroadcast(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := startMockServer(t, grpcServer)
	ordererConfig := getGRPCOpts(addr, true, false)
	orderer, _ := New(mocks.NewMockConfig(), WithURL(addr), FromOrdererConfig(ordererConfig))
	_, err := orderer.SendBroadcast(&fab.SignedEnvelope{})

	if err != nil {
		t.Fatalf("Test SendBroadcast was not supposed to fail")
	}

	orderer, _ = New(mocks.NewMockConfig(), WithURL(testOrdererURL+"Test"), FromOrdererConfig(ordererConfig))
	orderer.dialTimeout = 15
	_, err = orderer.SendBroadcast(&fab.SignedEnvelope{})
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
	orderer, _ := New(mocks.NewMockConfig(), WithURL(addr))

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

	orderer, _ := New(mocks.NewMockConfig(), WithURL(addr))

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

func TestSendDeliverFailure(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		DeliverResponse: &ab.DeliverResponse{},
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := New(mocks.NewMockConfig(), WithURL(addr))

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

func TestSendBroadcastServerBadResponse(t *testing.T) {

	broadcastServer := mocks.MockBroadcastServer{
		BroadcastInternalServerError: true,
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	addr := startCustomizedMockServer(t, testOrdererURL, grpcServer, &broadcastServer)
	orderer, _ := New(mocks.NewMockConfig(), WithURL(addr))

	_, err := orderer.SendBroadcast(&fab.SignedEnvelope{})

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
	orderer, _ := New(mocks.NewMockConfig(), WithURL(addr))

	statusCode, err := orderer.SendBroadcast(&fab.SignedEnvelope{})

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
	config := mock_apiconfig.NewMockConfig(mockCtrl)

	config.EXPECT().TimeoutOrDefault(apiconfig.OrdererConnection).Return(time.Second * 1)

	orderer, _ := NewOrderer("127.0.0.1:0", "", "", config, kap)
	orderer.grpcDialOption = append(orderer.grpcDialOption, grpc.WithBlock())
	_, err := orderer.SendBroadcast(&fab.SignedEnvelope{})
	assert.NotNil(t, err)

	statusError, ok := status.FromError(err)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, status.ConnectionFailed, status.ToSDKStatusCode(statusError.Code))
	assert.Equal(t, status.OrdererClientStatus, statusError.Group)
}

func TestInterfaces(t *testing.T) {
	var apiOrderer fab.Orderer
	var orderer Orderer

	apiOrderer = &orderer
	if apiOrderer == nil {
		t.Fatalf("this shouldn't happen.")
	}
}

func TestGetKeepAliveOptions(t *testing.T) {
	grpcOpts := make(map[string]interface{})

	grpcOpts["keep-alive-time"] = "s"
	grpcOpts["keep-alive-timeout"] = 2 * time.Second
	grpcOpts["keep-alive-permit"] = false
	//orderer config with GRPC opts
	ordererConfig := &apiconfig.OrdererConfig{
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
	ordererConfig := &apiconfig.OrdererConfig{
		GRPCOptions: grpcOpts,
	}
	failFast := getFailFast(ordererConfig)
	assert.EqualValues(t, failFast, true)

	grpcOpts["fail-fast"] = false
	ordererConfig = &apiconfig.OrdererConfig{
		GRPCOptions: grpcOpts,
	}
	failFast = getFailFast(ordererConfig)
	assert.EqualValues(t, failFast, false)
}

func getGRPCOpts(addr string, failFast bool, keepAliveOptions bool) *apiconfig.OrdererConfig {
	grpcOpts := make(map[string]interface{})
	//fail fast
	grpcOpts["fail-fast"] = failFast
	//keep alive options
	if keepAliveOptions {
		grpcOpts["keep-alive-time"] = 1 * time.Second
		grpcOpts["keep-alive-timeout"] = 2 * time.Second
		grpcOpts["keep-alive-permit"] = false
	}
	//orderer config with GRPC opts
	ordererConfig := &apiconfig.OrdererConfig{
		URL:         addr,
		GRPCOptions: grpcOpts,
	}

	return ordererConfig
}

func TestForDeadlineExceeded(t *testing.T) {
	orderer, _ := New(mocks.NewMockConfig(), WithURL(testOrdererURL+"Test"))
	orderer.dialTimeout = 1 * time.Second
	_, err := orderer.SendBroadcast(&fab.SignedEnvelope{})
	if err == nil || !strings.HasPrefix(err.Error(), "NewAtomicBroadcastClient") {
		t.Fatalf("Test SendBroadcast was supposed to fail with 'gRPC Transport Status Code: (4) DeadlineExceeded', instead it failed with [%s] error", err)
	}
}

func TestSendDeliverDefaultOpts(t *testing.T) {
	//keep alive option is not set and fail fast is false - invalid URL
	orderer, _ := New(mocks.NewMockConfig(), WithURL(testOrdererURL+"Test"))
	orderer.dialTimeout = 5 * time.Second
	fmt.Printf("GRPC opts%v \n", orderer.grpcDialOption)
	for i, v := range orderer.grpcDialOption {
		fmt.Printf("%v %v %v\n", i, &v, reflect.TypeOf(v))

	}
	_, err := orderer.SendBroadcast(&fab.SignedEnvelope{})
	if err == nil {
		t.Fatalf("Expected error 'Orderer Client Status 2 context deadline exceeded' %v", err)
	}

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := startMockServer(t, grpcServer)

	orderer, _ = New(mocks.NewMockConfig(), WithURL(addr))
	orderer.dialTimeout = 5 * time.Second
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

}

func TestForGRPCErrorsWithKeepAliveOpts(t *testing.T) {
	//keep alive options set and failfast is true
	ordererConfig := getGRPCOpts(testOrdererURL+"Test", true, true)
	orderer, _ := New(mocks.NewMockConfig(), WithURL(testOrdererURL+"Test"), FromOrdererConfig(ordererConfig))
	orderer.dialTimeout = 5 * time.Second
	_, err := orderer.SendBroadcast(&fab.SignedEnvelope{})
	if err == nil {
		t.Fatalf("Expected error 'Orderer Client Status 2 context deadline exceeded'")
	}
	//expect here GRPC unavaialble since fail fast is set to true
	statusError, ok := status.FromError(err)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, grpccodes.Unavailable, status.ToGRPCStatusCode(statusError.Code))
	assert.Equal(t, status.GRPCTransportStatus, statusError.Group)
	//expect here GRPC deadline exceeded since fail fast is set to false
	ordererConfig = getGRPCOpts(testOrdererURL+"Test", false, true)
	orderer, _ = New(mocks.NewMockConfig(), WithURL(testOrdererURL+"Test"), FromOrdererConfig(ordererConfig))
	orderer.dialTimeout = 5 * time.Second
	_, err = orderer.SendBroadcast(&fab.SignedEnvelope{})
	if err == nil {
		t.Fatalf("Expected error 'Orderer Client Status 2 context deadline exceeded'")
	}
	statusError, ok = status.FromError(err)
	fmt.Printf("%v %v", err, statusError)
	assert.True(t, ok, "Expected status error")
	assert.EqualValues(t, grpccodes.DeadlineExceeded, status.ToGRPCStatusCode(statusError.Code))
	assert.Equal(t, status.GRPCTransportStatus, statusError.Group)

}

func TestNewOrdererFromConfig(t *testing.T) {

	grpcOpts := make(map[string]interface{})
	//fail fast
	grpcOpts["fail-fast"] = true
	grpcOpts["keep-alive-time"] = 1 * time.Second
	grpcOpts["keep-alive-timeout"] = 2 * time.Second
	grpcOpts["keep-alive-permit"] = false
	//orderer config with GRPC opts
	ordererConfig := &apiconfig.OrdererConfig{
		URL:         "",
		GRPCOptions: grpcOpts,
	}
	_, err := NewOrdererFromConfig(ordererConfig, mocks.NewMockConfig())
	if err != nil {
		t.Fatalf("Failed to get new orderer from config %v", err)
	}

}
