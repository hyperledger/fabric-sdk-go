/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCreateChaincodeInstallProposal(t *testing.T) {
	c := setupTestClient()
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "peer1.example.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Status: 200}

	request := ChaincodeInstallRequest{
		Name:    "examplecc",
		Path:    "github.com/examplecc",
		Version: "1",
		Package: &ChaincodePackage{},
	}

	txid, err := txn.NewID(c.clientContext)
	assert.Nil(t, err, "create transaction ID failed")

	prop, err := CreateChaincodeInstallProposal(txid, request)
	assert.Nil(t, err, "CreateChaincodeInstallProposal failed")

	_, err = txn.SendProposal(c.clientContext, prop, []fab.ProposalProcessor{&peer})
	assert.Nil(t, err, "sending mock proposal failed")
}

func TestExtractChannelConfig(t *testing.T) {
	configTx, err := ioutil.ReadFile(path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"))
	if err != nil {
		t.Fatalf(err.Error())
	}

	_, err = ExtractChannelConfig(configTx)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestCreateConfigSignature(t *testing.T) {
	client := setupTestClient()

	configTx, err := ioutil.ReadFile(path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"))
	if err != nil {
		t.Fatalf(err.Error())
	}

	_, err = CreateConfigSignature(client.clientContext, configTx)
	if err != nil {
		t.Fatalf("Expected 'channel configuration required %v", err)
	}
}
