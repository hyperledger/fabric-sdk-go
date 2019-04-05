/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/discovery"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const (
	peerAddress  = "localhost:9999"
	peer2Address = "localhost:9998"
	peer3Address = "localhost:9997"
)

func TestDiscoveryClient(t *testing.T) {
	channelID := "mychannel"
	clientCtx := newMockContext()

	client, err := New(clientCtx)
	assert.NoError(t, err)

	req := discclient.NewRequest().AddLocalPeersQuery().OfChannel(channelID).AddPeersQuery()

	grpcOptions := map[string]interface{}{
		"allow-insecure": true,
	}
	target1 := fab.PeerConfig{
		URL:         peerAddress,
		GRPCOptions: grpcOptions,
	}
	target2 := fab.PeerConfig{
		URL:         peer2Address,
		GRPCOptions: grpcOptions,
	}
	target3 := fab.PeerConfig{
		URL:         peer3Address,
		GRPCOptions: grpcOptions,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	responses, err := client.Send(ctx, req, target1, target2, target3)
	cancel()

	assert.Error(t, err)
	errs, ok := err.(multi.Errors)
	assert.True(t, ok)

	if len(responses) != 1 {
		t.Fatalf("expecting 1 response but got %d", len(responses))
	}

	response := responses[0]
	assert.Equal(t, peerAddress, response.Target())
	locResp := response.ForLocal()
	peers, err := locResp.Peers()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(peers))
	t.Logf("Got success response for local query from [%s]: Num Peers: %d", response.Target(), len(peers))

	chResp := response.ForChannel(channelID)
	peers, err = chResp.Peers()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(peers))
	t.Logf("Got success response from channel query [%s]: Num Peers: %d", response.Target(), len(peers))

	assert.Equal(t, 2, len(errs))
	t.Logf("Got error responses: %s", errs)
}

var discoverServer *discmocks.MockDiscoveryServer

func TestMain(m *testing.M) {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	lis, err := net.Listen("tcp", peerAddress)
	if err != nil {
		panic(fmt.Sprintf("Error starting events listener %s", err))
	}

	discoverServer = discmocks.NewServer(
		discmocks.WithLocalPeers(
			&discmocks.MockDiscoveryPeerEndpoint{
				MSPID:        "Org1MSP",
				Endpoint:     peerAddress,
				LedgerHeight: 26,
			},
		),
		discmocks.WithPeers(
			&discmocks.MockDiscoveryPeerEndpoint{
				MSPID:        "Org1MSP",
				Endpoint:     peerAddress,
				LedgerHeight: 26,
			},
			&discmocks.MockDiscoveryPeerEndpoint{
				MSPID:        "Org2MSP",
				Endpoint:     peer2Address,
				LedgerHeight: 25,
			},
		),
	)

	discovery.RegisterDiscoveryServer(grpcServer, discoverServer)

	go grpcServer.Serve(lis)

	time.Sleep(2 * time.Second)
	os.Exit(m.Run())
}

func newMockContext() *mocks.MockContext {
	context := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("user1", "test"))
	context.SetCustomInfraProvider(comm.NewMockInfraProvider())
	return context
}
