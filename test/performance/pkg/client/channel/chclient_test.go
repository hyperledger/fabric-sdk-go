/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/common/ccprovider"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

const (
	channelID         = "myChannel"
	peerTLSServerCert = "${FABRIC_SDK_GO_PROJECT_PATH}/test/fixtures/fabric/v1/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt"
	peerTLSServerKey  = "${FABRIC_SDK_GO_PROJECT_PATH}/test/fixtures/fabric/v1/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key"

	ordererTLSServerCert = "${FABRIC_SDK_GO_PROJECT_PATH}/test/fixtures/fabric/v1/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt"
	ordererTLSServerKey  = "${FABRIC_SDK_GO_PROJECT_PATH}/test/fixtures/fabric/v1/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key"

	testhost          = "peer0.org1.example.com"
	testport          = 7051
	testBrodcasthost  = "orderer.example.com"
	testBroadcastport = 7050
)

var sdkClient *fabsdk.FabricSDK
var chClient *channel.Client

var (
	fixture            *testFixture
	ordererMockSrv     *fcmocks.MockBroadcastServer
	mockEndorserServer *MockEndorserServer
	chRq               = channel.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("b")}}
	endorserURL        = fmt.Sprintf("%s:%d", testhost, testport)
	ordererURL         = fmt.Sprintf("%s:%d", testBrodcasthost, testBroadcastport)
)

func BenchmarkCallExecuteTx(b *testing.B) {
	// report memory allocations for this benchmark
	b.ReportAllocs()

	// using channel Client, let's start the benchmark
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := chClient.Execute(chRq)
		require.NoError(b, err, "expected no error for valid channel client Execute invoke")

		//b.Logf("Execute Responses: %s", resp.Responses)
	}
}

func BenchmarkCallQuery(b *testing.B) {
	// report memory allocations for this benchmark
	b.ReportAllocs()

	// using channel Client, let's start the benchmark
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := chClient.Query(chRq)
		require.NoError(b, err, "expected no error for valid channel client Query invoke")

		//b.Logf("Query Responses: %s", resp.Responses)
	}
}

func BenchmarkCallExecuteTxParallel(b *testing.B) {
	// report memory allocations for this benchmark
	b.ReportAllocs()

	// using channel Client, let's start the benchmark
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := chClient.Execute(chRq)
			require.NoError(b, err, "expected no error for valid channel client parallel Execute invoke")

			//b.Logf("Execute Responses: %s", resp.Responses)
		}
	})
}

func BenchmarkCallQueryTxParallel(b *testing.B) {
	// report memory allocations for this benchmark
	b.ReportAllocs()

	// using channel Client, let's start the benchmark
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := chClient.Query(chRq)
			require.NoError(b, err, "expected no error for valid channel client parallel Query invoke")

			//b.Logf("Execute Responses: %s", resp.Responses)
		}
	})
}

func TestMain(m *testing.M) {
	setUp(m)
	r := m.Run()
	teardown()
	os.Exit(r)
}

func setUp(m *testing.M) {
	// do any setup here...
	tlsServerCertFile := testdata.Path(pathvar.Subst(peerTLSServerCert))
	tlsServerKeyFile := testdata.Path(pathvar.Subst(peerTLSServerKey))

	creds, err := credentials.NewServerTLSFromFile(tlsServerCertFile, tlsServerKeyFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to create new peer tls creds from file: %s", err))
	}
	payloadMap := make(map[string][]byte, 2)
	payloadMap["GetConfigBlock"] = getConfigBlockPayload()
	payloadMap["getccdata"] = getCCDataPayload()
	payloadMap["invoke"] = []byte("moved 'b' bytes")
	payloadMap["default"] = []byte("value")

	// setup mocked peer
	mockEndorserServer = &MockEndorserServer{Creds: creds}
	mockEndorserServer.SetMockPeer(&MockPeer{MockName: "Peer1", MockURL: endorserURL, MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200,
		Payload: payloadMap})

	// create a delivery channel for the orderer and the deliveryservice
	d := make(chan *pb.FilteredBlock, 10)

	fmt.Println("***************** Mocked Peer Started: ", mockEndorserServer.Start(endorserURL, d), " ******************************")

	// setup real client sdk and context (no mocks for the client side)
	fixture = &testFixture{}
	var ctx context.Client
	sdkClient, ctx = fixture.setup()

	// setup mocked broadcast server with tls credentials (mocking orderer requests)
	tlsServerCertFile = testdata.Path(pathvar.Subst(ordererTLSServerCert))
	tlsServerKeyFile = testdata.Path(pathvar.Subst(ordererTLSServerKey))

	creds, err = credentials.NewServerTLSFromFile(tlsServerCertFile, tlsServerKeyFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to create new orderer tls creds from file: %s", err))
	}

	// create a mock ordererBroadcastServer with a filteredDeliveries channel that communicates with a delivery server
	ordererMockSrv = &fcmocks.MockBroadcastServer{Creds: creds, FilteredDeliveries: d}
	fmt.Println("***************** Mocked Orderer Started: ", ordererMockSrv.Start(ordererURL), " ******************************")

	chClient = setupChannelClient(fixture.endpointConfig, ctx)
}

func teardown() {
	// do any teardown activities here ..
	sdkClient.Close()
	mockEndorserServer.Stop()
	fixture.close()
	ordererMockSrv.Stop()
}

func getConfigBlockPayload() []byte {
	// create config block builder in order to create valid payload
	builder := &fcmocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: fcmocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				"Org1MSP",
			},
			OrdererAddress:          fmt.Sprintf("grpc://%s:%d", testBrodcasthost, testBroadcastport),
			RootCA:                  rootCA,
			ApplicationCapabilities: []string{fab.V1_1Capability},
		},
		Index:           0,
		LastConfigIndex: 0,
	}

	payload, err := proto.Marshal(builder.Build())
	if err != nil {
		fmt.Println("Error marching ConfigBlockPayload: ", err)
	}

	return payload
}

func getCCDataPayload() []byte {
	ccPolicy := policydsl.SignedByMspMember("Org1MSP")
	pp, err := proto.Marshal(ccPolicy)
	if err != nil {
		panic(fmt.Sprintf("failed to build mock CC Policy: %s", err))
	}

	ccData := &ccprovider.ChaincodeData{
		Name:   "lscc",
		Policy: pp,
	}

	pd, err := proto.Marshal(ccData)
	if err != nil {
		panic(fmt.Sprintf("failed to build mock CC Data: %s", err))
	}

	return pd
}

func setupChannelClient(endpointConfig fab.EndpointConfig, ctx context.Client) *channel.Client {
	clntPvdr := setupCustomTestContext(endpointConfig, ctx)

	chPvdr := createChannelContext(clntPvdr, channelID)

	ch, err := channel.New(chPvdr)

	if err != nil {
		panic(fmt.Sprintf("Failed to create new channel client: %s", err))
	}

	return ch
}

func setupCustomTestContext(endpointConfig fab.EndpointConfig, ctx context.Client) context.ClientProvider {
	_, err := setupTestChannelService(ctx, endpointConfig)
	if err != nil {
		panic(fmt.Sprintf("Got error setting up TestChannelService %s", err))
	}

	return createClientContext(ctx)
}

func setupTestChannelService(ctx context.Client, endpointConfig fab.EndpointConfig) (fab.ChannelService, error) {
	chProvider, err := chpvdr.New(endpointConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "channel provider creation failed")
	}

	chService, err := chProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "channel service creation failed")
	}

	return chService, nil
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

// RootCA ca
var rootCA = `-----BEGIN CERTIFICATE-----
MIIB8TCCAZegAwIBAgIQU59imQ+xl+FmwuiFyUgFezAKBggqhkjOPQQDAjBYMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzENMAsGA1UEChMET3JnMTENMAsGA1UEAxMET3JnMTAeFw0xNzA1MDgw
OTMwMzRaFw0yNzA1MDYwOTMwMzRaMFgxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpD
YWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMQ0wCwYDVQQKEwRPcmcx
MQ0wCwYDVQQDEwRPcmcxMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEFkpP6EqE
87ghFi25UWLvgPatxDiYKYaVSPvpo/XDJ0+9uUmK/C2r5Bvvxx1t8eTROwN77tEK
r+jbJIxX3ZYQMKNDMEEwDgYDVR0PAQH/BAQDAgGmMA8GA1UdJQQIMAYGBFUdJQAw
DwYDVR0TAQH/BAUwAwEB/zANBgNVHQ4EBgQEAQIDBDAKBggqhkjOPQQDAgNIADBF
AiEA1Xkrpq+wrmfVVuY12dJfMQlSx+v0Q3cYce9BE1i2mioCIAzqyduK/lHPI81b
nWiU9JF9dRQ69dEV9dxd/gzamfFU
-----END CERTIFICATE-----`
