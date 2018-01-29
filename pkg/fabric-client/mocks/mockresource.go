/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"bytes"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// MockResource ...
type MockResource struct {
	errorScenario bool
}

// NewMockInvalidResource ...
func NewMockInvalidResource() *MockResource {
	c := &MockResource{errorScenario: true}
	return c
}

// NewMockResource ...
func NewMockResource() *MockResource {
	return &MockResource{}
}

// ExtractChannelConfig ...
func (c *MockResource) ExtractChannelConfig(configEnvelope []byte) ([]byte, error) {
	if bytes.Compare(configEnvelope, []byte("ExtractChannelConfigError")) == 0 {
		return nil, errors.New("Mock extract channel config error")
	}

	return configEnvelope, nil
}

// SignChannelConfig ...
func (c *MockResource) SignChannelConfig(config []byte, signer fab.IdentityContext) (*common.ConfigSignature, error) {
	if bytes.Compare(config, []byte("SignChannelConfigError")) == 0 {
		return nil, errors.New("Mock sign channel config error")
	}
	return nil, nil
}

// CreateChannel ...
func (c *MockResource) CreateChannel(request fab.CreateChannelRequest) (apitxn.TransactionID, error) {
	if c.errorScenario {
		return apitxn.TransactionID{}, errors.New("Create Channel Error")
	}

	return apitxn.TransactionID{}, nil
}

//QueryChannels ...
func (c *MockResource) QueryChannels(peer fab.Peer) (*pb.ChannelQueryResponse, error) {
	return nil, errors.New("Not implemented yet")
}

//QueryInstalledChaincodes mocks query installed chaincodes
func (c *MockResource) QueryInstalledChaincodes(peer fab.Peer) (*pb.ChaincodeQueryResponse, error) {
	if peer == nil {
		return nil, errors.New("Generate Error")
	}
	ci := &pb.ChaincodeInfo{Name: "name", Version: "version", Path: "path"}
	response := &pb.ChaincodeQueryResponse{Chaincodes: []*pb.ChaincodeInfo{ci}}
	return response, nil
}

// InstallChaincode mocks install chaincode
func (c *MockResource) InstallChaincode(req fab.InstallChaincodeRequest) ([]*apitxn.TransactionProposalResponse, string, error) {
	if req.Name == "error" {
		return nil, "", errors.New("Generate Error")
	}

	if req.Name == "errorInResponse" {
		result := apitxn.TransactionProposalResult{Endorser: "http://peer1.com", Status: 10}
		response := &apitxn.TransactionProposalResponse{TransactionProposalResult: result, Err: errors.New("Generate Response Error")}
		return []*apitxn.TransactionProposalResponse{response}, "1234", nil
	}

	result := apitxn.TransactionProposalResult{Endorser: "http://peer1.com", Status: 0}
	response := &apitxn.TransactionProposalResponse{TransactionProposalResult: result}
	return []*apitxn.TransactionProposalResponse{response}, "1234", nil
}
