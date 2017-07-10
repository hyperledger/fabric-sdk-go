/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// MockClient ...
type MockClient struct {
	channels      map[string]fab.Channel
	cryptoSuite   bccsp.BCCSP
	stateStore    fab.KeyValueStore
	userContext   fab.User
	config        config.Config
	errorScenario bool
}

// NewMockClient ...
/*
 * Returns a FabricClient instance
 */
func NewMockClient() fab.FabricClient {
	channels := make(map[string]fab.Channel)
	c := &MockClient{channels: channels, cryptoSuite: nil, stateStore: nil, userContext: nil, config: NewMockConfig()}
	return c
}

//NewMockInvalidClient : Returns new Mock FabricClient with error flag on used to test invalid scenarios
func NewMockInvalidClient() fab.FabricClient {
	channels := make(map[string]fab.Channel)
	c := &MockClient{channels: channels, cryptoSuite: nil, stateStore: nil, userContext: nil, config: NewMockConfig(), errorScenario: true}
	return c
}

// NewChannel ...
func (c *MockClient) NewChannel(name string) (fab.Channel, error) {
	return nil, nil
}

// Channel ...
func (c *MockClient) Channel(id string) fab.Channel {
	return c.channels[id]
}

// Config ...
func (c *MockClient) Config() config.Config {
	return c.config
}

// QueryChannelInfo ...
func (c *MockClient) QueryChannelInfo(name string, peers []fab.Peer) (fab.Channel, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

// SetStateStore ...
func (c *MockClient) SetStateStore(stateStore fab.KeyValueStore) {
	c.stateStore = stateStore
}

// StateStore ...
func (c *MockClient) StateStore() fab.KeyValueStore {
	return c.stateStore
}

// SetCryptoSuite ...
func (c *MockClient) SetCryptoSuite(cryptoSuite bccsp.BCCSP) {
	c.cryptoSuite = cryptoSuite
}

// CryptoSuite ...
func (c *MockClient) CryptoSuite() bccsp.BCCSP {
	return c.cryptoSuite
}

// SaveUserToStateStore ...
func (c *MockClient) SaveUserToStateStore(user fab.User, skipPersistence bool) error {
	return fmt.Errorf("Not implemented yet")

}

// LoadUserFromStateStore ...
func (c *MockClient) LoadUserFromStateStore(name string) (fab.User, error) {
	if c.errorScenario {
		return nil, fmt.Errorf("just to test error scenario")
	}
	return NewMockUser("test"), nil
}

// ExtractChannelConfig ...
func (c *MockClient) ExtractChannelConfig(configEnvelope []byte) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented yet")

}

// SignChannelConfig ...
func (c *MockClient) SignChannelConfig(config []byte) (*common.ConfigSignature, error) {
	return nil, fmt.Errorf("Not implemented yet")

}

// CreateChannel ...
func (c *MockClient) CreateChannel(request fab.CreateChannelRequest) (apitxn.TransactionID, error) {
	return apitxn.TransactionID{}, fmt.Errorf("Not implemented yet")

}

//QueryChannels ...
func (c *MockClient) QueryChannels(peer fab.Peer) (*pb.ChannelQueryResponse, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

//QueryInstalledChaincodes ...
func (c *MockClient) QueryInstalledChaincodes(peer fab.Peer) (*pb.ChaincodeQueryResponse, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

// InstallChaincode ...
func (c *MockClient) InstallChaincode(chaincodeName string, chaincodePath string, chaincodeVersion string,
	chaincodePackage []byte, targets []fab.Peer) ([]*apitxn.TransactionProposalResponse, string, error) {
	return nil, "", fmt.Errorf("Not implemented yet")

}

// UserContext ...
func (c *MockClient) UserContext() fab.User {
	return c.userContext
}

// SetUserContext ...
func (c *MockClient) SetUserContext(user fab.User) {
	c.userContext = user
}

// NewTxnID computes a TransactionID for the current user context
func (c *MockClient) NewTxnID() (apitxn.TransactionID, error) {
	return apitxn.TransactionID{
		ID:    "1234",
		Nonce: []byte{1, 2, 3, 4, 5},
	}, nil
}
