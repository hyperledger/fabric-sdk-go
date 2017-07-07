/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"io/ioutil"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

func TestChannelConfigs(t *testing.T) {

	client := mocks.NewMockClient()
	user := mocks.NewMockUser("test")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetCryptoSuite(cryptoSuite)

	channel, _ := NewChannel("testChannel", client)

	if channel.IsReadonly() {
		//TODO: Rightnow it is returning false always, need to revisit test once actual implementation is provided
		t.Fatal("Is Readonly test failed")
	}

	if channel.UpdateChannel() {
		//TODO: Rightnow it is returning false always, need to revisit test once actual implementation is provided
		t.Fatal("UpdateChannel test failed")
	}

	channel.SetMSPManager(nil)

}

func TestLoadConfigUpdateEnvelope(t *testing.T) {
	//Get Channel
	channel, _ := setupTestChannel()

	//Read config file from test directory
	fileLoc := "../../../test/fixtures/channel/mychanneltx.tx"
	res, err := ioutil.ReadFile(fileLoc)

	//Pass config to LoadConfigUpdateEnvelope and test
	err = channel.LoadConfigUpdateEnvelope(res)

	if err != nil {
		t.Fatalf("LoadConfigUpdateEnvelope Test Failed with, Cause '%s'", err.Error())
	}

	err = channel.Initialize(res)

	if err == nil {
		t.Fatalf("Initialize Negative Test Failed with, Cause '%s'", err.Error())
	}

	org1MSPID := "ORG1MSP"
	org2MSPID := "ORG2MSP"

	builder := &mocks.MockConfigUpdateEnvelopeBuilder{}

	err = channel.LoadConfigUpdateEnvelope(builder.BuildBytes())

	if err == nil {
		t.Fatal("Expected error was : channel initialization error: unable to load MSPs from config")
	}

	builder = &mocks.MockConfigUpdateEnvelopeBuilder{
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

	//Create mock orderer
	configBuilder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
			//RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}
	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(configBuilder.Build())
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(configBuilder.Build())
	channel.AddOrderer(orderer)

	//Add a second orderer
	configBuilder = &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
			//RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}
	orderer = mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(configBuilder.Build())
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(configBuilder.Build())
	channel.AddOrderer(orderer)
	err = channel.Initialize(nil)

	if err == nil {
		t.Fatal("Initialize on orderers config supposed to fail with 'could not decode pem bytes'")
	}

}

func TestChannelInitialize(t *testing.T) {
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

	mspManager := channel.MSPManager()
	if mspManager == nil {
		t.Fatalf("nil MSPManager on new channel")
	}

}

//func TestChannelInitializeFromUpdate(t *testing.T) {
//	org1MSPID := "ORG1MSP"
//	org2MSPID := "ORG2MSP"
//
//	client := mocks.NewMockClient()
//	user := mocks.NewMockUser("test", )
//	cryptoSuite := &mocks.MockCryptoSuite{}
//	client.SaveUserToStateStore(user, true)
//	client.SetCryptoSuite(cryptoSuite)
//	channel, _ := NewChannel("testChannel", client)
//
//	builder := &mocks.MockConfigUpdateEnvelopeBuilder{
//		ChannelID: "mychannel",
//		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
//			ModPolicy: "Admins",
//			MSPNames: []string{
//				org1MSPID,
//				org2MSPID,
//			},
//			OrdererAddress: "localhost:7054",
//			RootCA:         validRootCA,
//		},
//	}
//
//	err := channel.Initialize(builder.BuildBytes())
//	if err != nil {
//		t.Fatalf("channel Initialize failed : %v", err)
//	}
//
//	mspManager := channel.MSPManager()
//	if mspManager == nil {
//		t.Fatalf("nil MSPManager on new channel")
//	}
//
//}
