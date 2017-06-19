/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	api "github.com/hyperledger/fabric-sdk-go/api"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// MockClient ...
type MockClient struct {
	channels    map[string]api.Channel
	cryptoSuite bccsp.BCCSP
	stateStore  api.KeyValueStore
	userContext api.User
	config      api.Config
}

// NewMockClient ...
/*
 * Returns a FabricClient instance
 */
func NewMockClient() api.FabricClient {
	channels := make(map[string]api.Channel)
	c := &MockClient{channels: channels, cryptoSuite: nil, stateStore: nil, userContext: nil, config: NewMockConfig()}
	return c
}

// NewChannel ...
func (c *MockClient) NewChannel(name string) (api.Channel, error) {
	return nil, nil
}

// GetChannel ...
func (c *MockClient) GetChannel(name string) api.Channel {
	return c.channels[name]
}

// GetConfig ...
func (c *MockClient) GetConfig() api.Config {
	return c.config
}

// QueryChannelInfo ...
func (c *MockClient) QueryChannelInfo(name string, peers []api.Peer) (api.Channel, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

// SetStateStore ...
func (c *MockClient) SetStateStore(stateStore api.KeyValueStore) {
	c.stateStore = stateStore
}

// GetStateStore ...
func (c *MockClient) GetStateStore() api.KeyValueStore {
	return c.stateStore
}

// SetCryptoSuite ...
func (c *MockClient) SetCryptoSuite(cryptoSuite bccsp.BCCSP) {
	c.cryptoSuite = cryptoSuite
}

// GetCryptoSuite ...
func (c *MockClient) GetCryptoSuite() bccsp.BCCSP {
	return c.cryptoSuite
}

// SaveUserToStateStore ...
func (c *MockClient) SaveUserToStateStore(user api.User, skipPersistence bool) error {
	return fmt.Errorf("Not implemented yet")

}

// LoadUserFromStateStore ...
func (c *MockClient) LoadUserFromStateStore(name string) (api.User, error) {
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
func (c *MockClient) CreateChannel(request *api.CreateChannelRequest) error {
	return fmt.Errorf("Not implemented yet")

}

// CreateOrUpdateChannel ...
func (c *MockClient) CreateOrUpdateChannel(request *api.CreateChannelRequest, haveEnvelope bool) error {
	return fmt.Errorf("Not implemented yet")

}

//QueryChannels ...
func (c *MockClient) QueryChannels(peer api.Peer) (*pb.ChannelQueryResponse, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

//QueryInstalledChaincodes ...
func (c *MockClient) QueryInstalledChaincodes(peer api.Peer) (*pb.ChaincodeQueryResponse, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

// InstallChaincode ...
func (c *MockClient) InstallChaincode(chaincodeName string, chaincodePath string, chaincodeVersion string,
	chaincodePackage []byte, targets []api.Peer) ([]*api.TransactionProposalResponse, string, error) {
	return nil, "", fmt.Errorf("Not implemented yet")

}

// GetIdentity returns MockClient's serialized identity
func (c *MockClient) GetIdentity() ([]byte, error) {
	return []byte("test"), nil

}

// GetUserContext ...
func (c *MockClient) GetUserContext() api.User {
	return c.userContext
}

// SetUserContext ...
func (c *MockClient) SetUserContext(user api.User) {
	c.userContext = user
}
