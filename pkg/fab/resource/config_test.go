/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

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
