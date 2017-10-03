/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"fmt"
	"testing"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

var testAddress = "127.0.0.1:0"

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
	if channel.Name() != "testChannel" {
		t.Fatalf("NewChannel create wrong channel")
	}

	_, err = NewChannel("", client)
	if err == nil {
		t.Fatalf("NewChannel didn't return error")
	}
	if err.Error() != "name is required" {
		t.Fatalf("NewChannel didn't return right error")
	}

	_, err = NewChannel("testChannel", nil)
	if err == nil {
		t.Fatalf("NewChannel didn't return error")
	}
	if err.Error() != "client is required" {
		t.Fatalf("NewChannel didn't return right error")
	}

}

func TestInterfaces(t *testing.T) {
	var apiChannel fab.Channel
	var channel Channel

	apiChannel = &channel
	if apiChannel == nil {
		t.Fatalf("this shouldn't happen.")
	}
}

func TestAddRemoveOrderer(t *testing.T) {

	//Setup channel
	channel, _ := setupTestChannel()

	//Create mock orderer
	orderer := mocks.NewMockOrderer("", nil)

	//Add an orderer
	channel.AddOrderer(orderer)

	//Check if orderer is being added successfully
	if len(channel.Orderers()) != 1 {
		t.Fatal("Adding orderers to channel failed")
	}

	//Remove the orderer now
	channel.RemoveOrderer(orderer)

	//Check if list of orderers is empty now
	if len(channel.Orderers()) != 0 {
		t.Fatal("Removing orderers from channel failed")
	}
}

func TestAnchorAndRemovePeers(t *testing.T) {
	//Setup channel
	channel, _ := setupTestChannel()

	//Add a Peer
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	//Remove and Test
	channel.RemovePeer(&peer)
	if len(channel.Peers()) != 0 {
		t.Fatal("Remove Peer failed")
	}

	//Add the Peer again
	channel.AddPeer(&peer)

	channel.Initialize(nil)
	if len(channel.AnchorPeers()) != 0 {
		//Currently testing only for empty anchor list
		t.Fatal("Anchor peer list is incorrect")
	}
}

func TestPrimaryPeer(t *testing.T) {
	channel, _ := setupTestChannel()

	if channel.PrimaryPeer() != nil {
		t.Fatal("Call to Primary peer on empty channel should always return nil")
	}

	// Channel had one peer
	peer1 := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	err := channel.AddPeer(&peer1)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}

	// Test primary defaults to channel peer
	primary := channel.PrimaryPeer()
	if primary.URL() != peer1.URL() {
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
	primary = channel.PrimaryPeer()
	if primary.URL() != peer2.URL() {
		t.Fatalf("Primary and our choice are not equal")
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
	if !channel.IsInitialized() {
		t.Fatalf("channel Initialize failed : channel initialized flag not set")
	}

	mspManager := channel.MSPManager()
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

	channel.SetMSPManager(nil)
	if channel.MSPManager() != nil {
		t.Fatal("Set MSPManager is not working as expected")
	}

}

func TestOrganizationUnits(t *testing.T) {
	org1MSPID := "ORG1MSP"
	org2MSPID := "ORG2MSP"

	channel, _ := setupTestChannel()
	orgUnits, err := channel.OrganizationUnits()

	if len(orgUnits) > 0 {
		t.Fatalf("Returned non configured organizational unit : %v", err)
	}
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				channel.Name(),
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
	orgUnits, err = channel.OrganizationUnits()
	if err != nil {
		t.Fatalf("CANNOT retrieve organizational units : %v", err)
	}
	if !isValueInList(channel.Name(), orgUnits) {
		t.Fatalf("Could not find %s in the list of organizations", channel.Name())
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

func setupTestChannel() (*Channel, error) {
	client := mocks.NewMockClient()
	user := mocks.NewMockUser("test")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetUserContext(user)
	client.SetCryptoSuite(cryptoSuite)
	return NewChannel("testChannel", client)
}

func setupMassiveTestChannel(numberOfPeers int, numberOfOrderers int) (*Channel, error) {
	channel, error := setupTestChannel()
	if error != nil {
		return channel, error
	}

	for i := 0; i < numberOfPeers; i++ {
		peer := mocks.MockPeer{MockName: fmt.Sprintf("MockPeer%d", i), MockURL: fmt.Sprintf("http://mock%d.peers.r.us", i),
			MockRoles: []string{}, MockCert: nil}
		err := channel.AddPeer(&peer)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to add peer")
		}
	}

	for i := 0; i < numberOfOrderers; i++ {
		orderer := mocks.NewMockOrderer(fmt.Sprintf("http://mock%d.orderers.r.us", i), nil)
		err := channel.AddOrderer(orderer)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to add orderer")
		}
	}

	return channel, error
}
