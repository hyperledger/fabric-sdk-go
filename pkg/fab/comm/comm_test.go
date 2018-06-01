/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	eventmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	peerAddress     = "localhost:9999"
	endorserAddress = "127.0.0.1:0"
	peerURL         = "grpc://" + peerAddress
)

func TestMain(m *testing.M) {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	lis, err := net.Listen("tcp", peerAddress)
	if err != nil {
		panic(fmt.Sprintf("Error starting events listener %s", err))
	}

	testServer = eventmocks.NewMockEventhubServer()

	pb.RegisterEventsServer(grpcServer, testServer)

	go grpcServer.Serve(lis)

	srvs, addrs, err := startEndorsers(30, endorserAddress)
	if err != nil {
		panic(fmt.Sprintf("Error starting endorser %s", err))
	}
	for _, srv := range srvs {
		defer srv.Stop()
	}
	endorserAddr = addrs

	time.Sleep(2 * time.Second)
	os.Exit(m.Run())
}

func startEndorsers(count int, address string) ([]*grpc.Server, []string, error) {
	srvs := make([]*grpc.Server, 0, count)
	addrs := make([]string, 0, count)

	for i := 0; i < count; i++ {
		srv := grpc.NewServer()
		_, addr, ok := startEndorserServer(srv, address)
		if !ok {
			return nil, nil, errors.New("unable to start GRPC server")
		}
		srvs = append(srvs, srv)
		addrs = append(addrs, addr)
	}
	return srvs, addrs, nil
}

func startEndorserServer(grpcServer *grpc.Server, address string) (*mocks.MockEndorserServer, string, bool) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Error starting test server %s\n", err)
		return nil, "", false
	}
	addr := lis.Addr().String()

	endorserServer := &mocks.MockEndorserServer{}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	fmt.Printf("Starting test server on %s\n", addr)
	go grpcServer.Serve(lis)
	return endorserServer, addr, true
}
