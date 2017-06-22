/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"net"
	"testing"

	"google.golang.org/grpc"

	api "github.com/hyperledger/fabric-sdk-go/api"
	fc "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	peer "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"

	pb "github.com/hyperledger/fabric/protos/peer"
)

var testAddress = "0.0.0.0:5244"

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
-----END CERTIFICATE-----
`

func TestChannelMethods(t *testing.T) {
	client := mocks.NewMockClient()
	channel, err := NewChannel("testChannel", client)
	if err != nil {
		t.Fatalf("NewChannel return error[%s]", err)
	}
	if channel.GetName() != "testChannel" {
		t.Fatalf("NewChannel create wrong channel")
	}

	_, err = NewChannel("", client)
	if err == nil {
		t.Fatalf("NewChannel didn't return error")
	}
	if err.Error() != "failed to create Channel. Missing required 'name' parameter" {
		t.Fatalf("NewChannel didn't return right error")
	}

	_, err = NewChannel("testChannel", nil)
	if err == nil {
		t.Fatalf("NewChannel didn't return error")
	}
	if err.Error() != "failed to create Channel. Missing required 'clientContext' parameter" {
		t.Fatalf("NewChannel didn't return right error")
	}

}

func TestQueryMethods(t *testing.T) {
	channel, _ := setupTestChannel()

	_, err := channel.QueryBlock(-1)
	if err == nil {
		t.Fatalf("Query block cannot be negative number")
	}

	_, err = channel.QueryBlockByHash(nil)
	if err == nil {
		t.Fatalf("Query hash cannot be nil")
	}
	_, err = channel.QueryByChaincode("", []string{"method"}, nil)
	if err == nil {
		t.Fatalf("QueryByChannelcode: name cannot be empty")
	}

	_, err = channel.QueryByChaincode("qscc", nil, nil)
	if err == nil {
		t.Fatalf("QueryByChannelcode: arguments cannot be empty")
	}

	_, err = channel.QueryByChaincode("qscc", []string{"method"}, nil)
	if err == nil {
		t.Fatalf("QueryByChannelcode: targets cannot be empty")
	}

}

func TestTargetPeers(t *testing.T) {

	p := make(map[string]api.Peer)
	channel := &channel{name: "targetChannel", peers: p}

	// Channel has two peers
	peer1 := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	err := channel.AddPeer(&peer1)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}
	peer2 := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil}
	err = channel.AddPeer(&peer2)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}

	// Set target to invalid URL
	invalidChoice := mocks.MockPeer{MockName: "", MockURL: "http://xyz.com", MockRoles: []string{}, MockCert: nil}
	targetPeers, err := channel.getTargetPeers([]api.Peer{&invalidChoice})
	if err == nil {
		t.Fatalf("Target peer didn't fail for an invalid peer")
	}

	// Test target peers default to channel peers if target peers are not provided
	targetPeers, err = channel.getTargetPeers(nil)

	if err != nil || targetPeers == nil || len(targetPeers) != 2 {
		t.Fatalf("Target Peers failed to default")
	}

	// Set target to valid peer 2 URL
	choice := mocks.MockPeer{MockName: "", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil}
	targetPeers, err = channel.getTargetPeers([]api.Peer{&choice})
	if err != nil {
		t.Fatalf("Failed to get valid target peer")
	}

	// Test target equals our choice
	if len(targetPeers) != 1 || targetPeers[0].GetURL() != peer2.GetURL() || targetPeers[0].GetName() != peer2.GetName() {
		t.Fatalf("Primary and our choice are not equal")
	}

}

func TestPrimaryPeer(t *testing.T) {
	channel, _ := setupTestChannel()

	// Channel had one peer
	peer1 := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	err := channel.AddPeer(&peer1)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}

	// Test primary defaults to channel peer
	primary := channel.GetPrimaryPeer()
	if primary.GetURL() != peer1.GetURL() {
		t.Fatalf("Primary Peer failed to default")
	}

	// Channel has two peers
	peer2 := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil}
	err = channel.AddPeer(&peer2)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}

	// Set primary to invalid URL
	invalidChoice := mocks.MockPeer{MockName: "", MockURL: "http://xyz.com", MockRoles: []string{}, MockCert: nil}
	err = channel.SetPrimaryPeer(&invalidChoice)
	if err == nil {
		t.Fatalf("Primary Peer was set to an invalid peer")
	}

	// Set primary to valid peer 2 URL
	choice := mocks.MockPeer{MockName: "", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil}
	err = channel.SetPrimaryPeer(&choice)
	if err != nil {
		t.Fatalf("Failed to set valid primary peer")
	}

	// Test primary equals our choice
	primary = channel.GetPrimaryPeer()
	if primary.GetURL() != peer2.GetURL() || primary.GetName() != peer2.GetName() {
		t.Fatalf("Primary and our choice are not equal")
	}

}

func TestConcurrentPeers(t *testing.T) {
	const numPeers = 10000
	channel, err := setupMassiveTestChannel(numPeers, 0)
	if err != nil {
		t.Fatalf("Failed to create massive channel: %s", err)
	}

	result, err := channel.SendTransactionProposal(&api.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, 1, nil)
	if err != nil {
		t.Fatalf("SendTransactionProposal return error: %s", err)
	}

	if len(result) != numPeers {
		t.Error("SendTransactionProposal returned an unexpected amount of responses")
	}
}

func TestConcurrentOrderers(t *testing.T) {
	const numOrderers = 10000
	channel, err := setupMassiveTestChannel(0, numOrderers)
	if err != nil {
		t.Fatalf("Failed to create massive channel: %s", err)
	}

	txn := api.Transaction{
		Proposal: &api.TransactionProposal{
			Proposal: &pb.Proposal{},
		},
		Transaction: &pb.Transaction{},
	}
	_, err = channel.SendTransaction(&txn)
	if err != nil {
		t.Fatalf("SendTransaction returned error: %s", err)
	}
}

func TestJoinChannel(t *testing.T) {
	var peers []api.Peer
	endorserServer := startEndorserServer(t)
	channel, _ := setupTestChannel()
	peer, _ := peer.NewPeer(testAddress, "", "", mocks.NewMockConfig())
	peers = append(peers, peer)
	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(mocks.NewSimpleMockBlock())
	nonce, _ := fc.GenerateRandomNonce()
	txID, _ := fc.ComputeTxID(nonce, []byte("testID"))

	genesisBlockReqeust := &api.GenesisBlockRequest{
		TxID:  txID,
		Nonce: nonce,
	}
	genesisBlock, err := channel.GetGenesisBlock(genesisBlockReqeust)
	if err == nil {
		t.Fatalf("Should not have been able to get genesis block because of orderer missing")
	}

	err = channel.AddOrderer(orderer)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}
	genesisBlock, err = channel.GetGenesisBlock(genesisBlockReqeust)
	if err != nil {
		t.Fatalf("Error getting genesis block: %v", err)
	}

	err = channel.JoinChannel(nil)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing request parameter")
	}

	request := &api.JoinChannelRequest{
		Targets:      peers,
		GenesisBlock: genesisBlock,
		Nonce:        nonce,
		//TxID:         txID,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing TxID parameter")
	}

	request = &api.JoinChannelRequest{
		Targets:      peers,
		GenesisBlock: genesisBlock,
		//Nonce:        nonce,
		TxID: txID,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing Nonce parameter")
	}

	request = &api.JoinChannelRequest{
		Targets: peers,
		//GenesisBlock: genesisBlock,
		Nonce: nonce,
		TxID:  txID,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing GenesisBlock parameter")
	}

	request = &api.JoinChannelRequest{
		//Targets: peers,
		GenesisBlock: genesisBlock,
		Nonce:        nonce,
		TxID:         txID,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing Targets parameter")
	}

	request = &api.JoinChannelRequest{
		Targets:      peers,
		GenesisBlock: genesisBlock,
		Nonce:        nonce,
		TxID:         txID,
	}
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of invalid targets")
	}

	err = channel.AddPeer(peer)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}
	// Test join channel with valid arguments
	err = channel.JoinChannel(request)
	if err != nil {
		t.Fatalf("Did not expect error from join channel. Got: %s", err)
	}

	// Test failed proposal error handling
	endorserServer.ProposalError = fmt.Errorf("Test Error")
	request = &api.JoinChannelRequest{Targets: peers, Nonce: nonce, TxID: txID}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Expected error")
	}
}

func TestChannelInitializeFromOrderer(t *testing.T) {
	org1MSPID := "ORG1MSP"
	org2MSPID := "ORG2MSP"

	channel, _ := setupTestChannel()
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
			RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}
	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(builder.Build())
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(builder.Build())
	err := channel.AddOrderer(orderer)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}

	err = channel.Initialize(nil)
	if err != nil {
		t.Fatalf("channel Initialize failed : %v", err)
	}
	mspManager := channel.GetMSPManager()
	if mspManager == nil {
		t.Fatalf("nil MSPManager on new channel")
	}
	msps, err := mspManager.GetMSPs()
	if err != nil || len(msps) == 0 {
		t.Fatalf("At least one MSP expected in MSPManager")
	}
	msp, ok := msps[org1MSPID]
	if !ok {
		t.Fatalf("Could not find %s", org1MSPID)
	}
	if identifier, _ := msp.GetIdentifier(); identifier != org1MSPID {
		t.Fatalf("Expecting MSP identifier to be %s but got %s", org1MSPID, identifier)
	}
	msp, ok = msps[org2MSPID]
	if !ok {
		t.Fatalf("Could not find %s", org2MSPID)
	}
	if identifier, _ := msp.GetIdentifier(); identifier != org2MSPID {
		t.Fatalf("Expecting MSP identifier to be %s but got %s", org2MSPID, identifier)
	}
}

func TestOrganizationUnits(t *testing.T) {
	org1MSPID := "ORG1MSP"
	org2MSPID := "ORG2MSP"

	channel, _ := setupTestChannel()
	orgUnits, err := channel.GetOrganizationUnits()
	if len(orgUnits) > 0 {
		t.Fatalf("Returned non configured organizational unit : %v", err)
	}
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				channel.GetName(),
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
			RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}
	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(builder.Build())
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(builder.Build())
	err = channel.AddOrderer(orderer)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}

	err = channel.Initialize(nil)
	if err != nil {
		t.Fatalf("channel Initialize failed : %v", err)
	}
	orgUnits, err = channel.GetOrganizationUnits()
	if err != nil {
		t.Fatalf("CANNOT retrieve organizational units : %v", err)
	}
	if !isValueInList(channel.GetName(), orgUnits) {
		t.Fatalf("Could not find %s in the list of organizations", channel.GetName())
	}
	if !isValueInList(org1MSPID, orgUnits) {
		t.Fatalf("Could not find %s in the list of organizations", org1MSPID)
	}
	if !isValueInList(org2MSPID, orgUnits) {
		t.Fatalf("Could not find %s in the list of organizations", org2MSPID)
	}

}

func isValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func TestChannelInitializeFromUpdate(t *testing.T) {
	org1MSPID := "ORG1MSP"
	org2MSPID := "ORG2MSP"

	channel, _ := setupTestChannel()
	builder := &mocks.MockConfigUpdateEnvelopeBuilder{
		ChannelID: "mychannel",
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
			RootCA:         validRootCA,
		},
	}

	err := channel.Initialize(builder.BuildBytes())
	if err != nil {
		t.Fatalf("channel Initialize failed : %v", err)
	}
	mspManager := channel.GetMSPManager()
	if mspManager == nil {
		t.Fatalf("nil MSPManager on new channel")
	}
}

func setupTestChannel() (api.Channel, error) {
	client := mocks.NewMockClient()
	user := mocks.NewMockUser("test")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetCryptoSuite(cryptoSuite)
	return NewChannel("testChannel", client)
}

func setupMassiveTestChannel(numberOfPeers int, numberOfOrderers int) (api.Channel, error) {
	channel, error := setupTestChannel()
	if error != nil {
		return channel, error
	}

	for i := 0; i < numberOfPeers; i++ {
		peer := mocks.MockPeer{MockName: fmt.Sprintf("MockPeer%d", i), MockURL: fmt.Sprintf("http://mock%d.peers.r.us", i),
			MockRoles: []string{}, MockCert: nil}
		err := channel.AddPeer(&peer)
		if err != nil {
			return nil, fmt.Errorf("Error adding peer: %v", err)
		}
	}

	for i := 0; i < numberOfOrderers; i++ {
		orderer := mocks.NewMockOrderer(fmt.Sprintf("http://mock%d.orderers.r.us", i), nil)
		err := channel.AddOrderer(orderer)
		if err != nil {
			return nil, fmt.Errorf("Error adding orderer: %v", err)
		}
	}

	return channel, error
}

func startEndorserServer(t *testing.T) *mocks.MockEndorserServer {
	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", testAddress)
	endorserServer := &mocks.MockEndorserServer{}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	if err != nil {
		fmt.Printf("Error starting test server %s", err)
		t.FailNow()
	}
	fmt.Printf("Starting test server\n")
	go grpcServer.Serve(lis)
	return endorserServer
}
