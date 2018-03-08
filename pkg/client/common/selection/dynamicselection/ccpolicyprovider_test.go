/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicselection

import (
	"strings"
	"testing"

	"net"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc"
)

func startEndorserServer(t *testing.T, grpcServer *grpc.Server, testAddress string) (*mocks.MockEndorserServer, string) {
	lis, err := net.Listen("tcp", testAddress)
	addr := lis.Addr().String()

	endorserServer := &mocks.MockEndorserServer{}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	if err != nil {
		t.Logf("Error starting test server %s", err)
		t.FailNow()
	}
	t.Logf("Starting test server on %s\n", addr)
	go grpcServer.Serve(lis)
	return endorserServer, addr
}

func TestCCPolicyProvider(t *testing.T) {
	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(config.FromFile("../../../../../test/fixtures/config/config_test.yaml"))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	clientContext := sdk.Context(fabsdk.WithUser("User1"), fabsdk.WithOrg("Org1"))

	context, err := clientContext()
	if err != nil {
		t.Fatal("Failed to create context")
	}

	// Nil sdk
	ccPolicyProvider, err := newCCPolicyProvider(nil, "mychannel", "User1", "Org1")
	if err == nil {
		t.Fatalf("Should have failed for nil sdk")
	}

	// Invalid channelID
	ccPolicyProvider, err = newCCPolicyProvider(context, "", "User1", "Org1")
	if err == nil {
		t.Fatalf("Should have failed for empty channel")
	}

	// Empty user name
	ccPolicyProvider, err = newCCPolicyProvider(context, "mychannel", "", "Prg1")
	if err == nil {
		t.Fatalf("Should have failed for empty user name")
	}

	// Empty org name
	ccPolicyProvider, err = newCCPolicyProvider(context, "mychannel", "User1", "")
	if err == nil {
		t.Fatalf("Should have failed for nil sdk")
	}

	// Invalid channel
	ccPolicyProvider, err = newCCPolicyProvider(context, "non-existent", "User1", "Org1")
	if err == nil {
		t.Fatalf("Should have failed for invalid channel name")
	}

	// All good
	ccPolicyProvider, err = newCCPolicyProvider(context, "mychannel", "User1", "Org1")
	if err != nil {
		t.Fatalf("Failed to setup cc policy provider: %s", err)
	}

	// Empty chaincode ID
	_, err = ccPolicyProvider.GetChaincodePolicy("")
	if err == nil {
		t.Fatalf("Should have failed to retrieve chaincode policy for empty chaincode id")
	}

	// Non-existent chaincode ID
	_, err = ccPolicyProvider.GetChaincodePolicy("abc")
	if err == nil {
		t.Fatalf("Should have failed to retrieve non-existent cc policy")
	}
}

func TestBadClient(t *testing.T) {
	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(config.FromFile("../../../../../test/fixtures/config/config_test.yaml"))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	clientContext := sdk.Context(fabsdk.WithUser("User1"), fabsdk.WithOrg("Org1"))

	context, err := clientContext()
	if err != nil {
		t.Fatal("Failed to create context")
	}

	// Non-existent user
	_, err = newCCPolicyProvider(context, "mychannel", "Invalid", "Org1")
	if !strings.Contains(err.Error(), "user not found") {
		t.Fatalf("Should have failed for invalid user name: %v", err)
	}

	// Invalid org
	_, err = newCCPolicyProvider(context, "mychannel", "User1", "Invalid")
	if !strings.Contains(err.Error(), "invalid org name") {
		t.Fatalf("Should have failed for invalid org name")
	}
}
