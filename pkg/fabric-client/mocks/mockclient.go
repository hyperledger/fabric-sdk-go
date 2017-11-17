/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"bytes"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

// MockClient ...
type MockClient struct {
	channels       map[string]fab.Channel
	cryptoSuite    apicryptosuite.CryptoSuite
	stateStore     fab.KeyValueStore
	userContext    fab.User
	config         config.Config
	errorScenario  bool
	signingManager fab.SigningManager
}

// NewMockClient ...
/*
 * Returns a FabricClient instance
 */
func NewMockClient() *MockClient {
	channels := make(map[string]fab.Channel)
	c := &MockClient{channels: channels, cryptoSuite: nil, stateStore: nil, userContext: nil, config: NewMockConfig(), signingManager: NewMockSigningManager()}
	return c
}

//NewMockInvalidClient : Returns new Mock FabricClient with error flag on used to test invalid scenarios
func NewMockInvalidClient() *MockClient {
	channels := make(map[string]fab.Channel)
	c := &MockClient{channels: channels, cryptoSuite: nil, stateStore: nil, userContext: nil, config: NewMockConfig(), errorScenario: true}
	return c
}

// NewChannel ...
func (c *MockClient) NewChannel(name string) (fab.Channel, error) {
	if name == "error" {
		return nil, errors.New("Genererate error in new channel")
	}
	return nil, nil
}

// SetChannel convenience method to set channel
func (c *MockClient) SetChannel(id string, channel fab.Channel) {
	c.channels[id] = channel
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
	return nil, errors.New("Not implemented yet")
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
func (c *MockClient) SetCryptoSuite(cryptoSuite apicryptosuite.CryptoSuite) {
	c.cryptoSuite = cryptoSuite
}

// CryptoSuite ...
func (c *MockClient) CryptoSuite() apicryptosuite.CryptoSuite {
	return c.cryptoSuite
}

// SigningManager returns the signing manager
func (c *MockClient) SigningManager() fab.SigningManager {
	return c.signingManager
}

// SetSigningManager mocks setting signing manager
func (c *MockClient) SetSigningManager(signingMgr fab.SigningManager) {
	c.signingManager = signingMgr
}

// SaveUserToStateStore ...
func (c *MockClient) SaveUserToStateStore(user fab.User, skipPersistence bool) error {
	return errors.New("Not implemented yet")

}

// LoadUserFromStateStore ...
func (c *MockClient) LoadUserFromStateStore(name string) (fab.User, error) {
	if c.errorScenario {
		return nil, errors.New("just to test error scenario")
	}
	return NewMockUser("test"), nil
}

// ExtractChannelConfig ...
func (c *MockClient) ExtractChannelConfig(configEnvelope []byte) ([]byte, error) {
	if bytes.Compare(configEnvelope, []byte("ExtractChannelConfigError")) == 0 {
		return nil, errors.New("Mock extract channel config error")
	}

	return configEnvelope, nil
}

// SignChannelConfig ...
func (c *MockClient) SignChannelConfig(config []byte, signer fab.User) (*common.ConfigSignature, error) {
	if bytes.Compare(config, []byte("SignChannelConfigError")) == 0 {
		return nil, errors.New("Mock sign channel config error")
	}
	return nil, nil
}

// CreateChannel ...
func (c *MockClient) CreateChannel(request fab.CreateChannelRequest) (apitxn.TransactionID, error) {
	if c.errorScenario {
		return apitxn.TransactionID{}, errors.New("Create Channel Error")
	}

	return apitxn.TransactionID{}, nil
}

//QueryChannels ...
func (c *MockClient) QueryChannels(peer fab.Peer) (*pb.ChannelQueryResponse, error) {
	return nil, errors.New("Not implemented yet")
}

//QueryInstalledChaincodes ...
func (c *MockClient) QueryInstalledChaincodes(peer fab.Peer) (*pb.ChaincodeQueryResponse, error) {
	return nil, errors.New("Not implemented yet")
}

// InstallChaincode ...
func (c *MockClient) InstallChaincode(chaincodeName string, chaincodePath string, chaincodeVersion string,
	chaincodePackage []byte, targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, string, error) {
	return nil, "", errors.New("Not implemented yet")

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
