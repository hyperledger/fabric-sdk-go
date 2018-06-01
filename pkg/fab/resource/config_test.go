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
		t.Fatal(err)
	}

	_, err = ExtractChannelConfig(configTx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateConfigSignature(t *testing.T) {
	ctx := setupContext()

	configTx, err := ioutil.ReadFile(path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = CreateConfigSignature(ctx, configTx)
	if err != nil {
		t.Fatalf("Expected 'channel configuration required %s", err)
	}
}
