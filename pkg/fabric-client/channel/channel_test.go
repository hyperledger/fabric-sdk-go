/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"fmt"
	"reflect"
	"testing"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/pkg/errors"
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
	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)
	channel, err := New(ctx, mocks.NewMockChannelCfg("testChannel"))
	if err != nil {
		t.Fatalf("New return error[%s]", err)
	}
	if channel.Name() != "testChannel" {
		t.Fatalf("New create wrong channel")
	}

	_, err = New(ctx, mocks.NewMockChannelCfg(""))
	if err != nil {
		t.Fatalf("Got error creating channel with empty channel ID: %s", err)
	}

	_, err = New(nil, mocks.NewMockChannelCfg("testChannel"))
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

func TestAddAndRemovePeers(t *testing.T) {
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
	if len(channel.Peers()) != 1 {
		t.Fatal("Add Peer failed")
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

func TestQueryOnSystemChannel(t *testing.T) {
	channel, _ := setupChannel(systemChannel)
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200}
	err := channel.AddPeer(&peer)
	if err != nil {
		t.Fatalf("Error adding peer to channel: %s", err)
	}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "ccID",
		Fcn:         "method",
		Args:        [][]byte{[]byte("arg")},
	}
	if _, err := channel.QueryByChaincode(request); err != nil {
		t.Fatalf("Error invoking chaincode on system channel: %s", err)
	}
}

func TestQueryBySystemChaincode(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Status: 200}
	channel.AddPeer(&peer)

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	resp, err := channel.QueryBySystemChaincode(request)
	if err != nil {
		t.Fatalf("Failed to query: %s", err)
	}
	expectedResp := []byte("A")

	if !reflect.DeepEqual(resp[0], expectedResp) {
		t.Fatalf("Unexpected transaction proposal response: %v", resp)
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
	return setupChannel("testChannel")
}

func setupChannel(channelID string) (*Channel, error) {
	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)
	return New(ctx, mocks.NewMockChannelCfg(channelID))
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
