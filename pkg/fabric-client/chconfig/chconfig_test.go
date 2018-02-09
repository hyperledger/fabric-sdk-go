/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package chconfig

import (
	"testing"

	"github.com/golang/protobuf/proto"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

const (
	channelID = "testChannel"
)

func TestChannelConfigWithPeer(t *testing.T) {

	ctx := setupTestContext()
	peer := getPeerWithConfigBlockPayload(t)

	channelConfig, err := New(ctx, channelID, WithPeers([]fab.Peer{peer}), WithMinResponses(1))
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	cfg, err := channelConfig.Query()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if cfg.Name() != channelID {
		t.Fatalf("Channel name error. Expecting %s, got %s ", channelID, cfg.Name())
	}
}

func TestChannelConfigWithPeerError(t *testing.T) {

	ctx := setupTestContext()
	peer := getPeerWithConfigBlockPayload(t)

	channelConfig, err := New(ctx, channelID, WithPeers([]fab.Peer{peer}), WithMinResponses(2))
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	_, err = channelConfig.Query()
	if err == nil {
		t.Fatalf("Should have failed with since there's one endorser and at least two are required")
	}
}

func TestChannelConfigWithOrdererError(t *testing.T) {

	ctx := setupTestContext()

	channelConfig, err := New(ctx, channelID, WithOrderer("localhost:7054"))
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Expecting error since orderer is not setup
	_, err = channelConfig.Query()
	if err == nil {
		t.Fatalf("Should have failed since orderer is not available")
	}

}

func setupTestChannel(name string) (*channel.Channel, error) {
	ctx := setupTestContext()
	return channel.New(ctx, mocks.NewMockChannelCfg(name))
}

func setupTestContext() fab.Context {
	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)
	return ctx
}

func getPeerWithConfigBlockPayload(t *testing.T) fab.Peer {

	// create config block builder in order to create valid payload
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				"Org1MSP",
				"Org2MSP",
			},
			OrdererAddress: "localhost:7054",
			RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}

	payload, err := proto.Marshal(builder.Build())
	if err != nil {
		t.Fatalf("Failed to marshal mock block")
	}

	// peer with valid config block payload
	peer := &mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: payload, Status: 200}

	return peer
}

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
