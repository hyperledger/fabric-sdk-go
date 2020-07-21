/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	mspcfg "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/stretchr/testify/require"
)

func randomMspDir() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return "/tmp/msp/" + fmt.Sprintf("%d", rnd.Uint64())
}

func TestGenerateMspDir(t *testing.T) {

	ordererMspDir := filepath.Join(metadata.GetProjectPath(), "test/fixtures/fabric/v1/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp")
	cfg, err := mspcfg.GetVerifyingMspConfig(ordererMspDir, "mymspid", "bccsp")
	require.NoError(t, err, "Error generating msp config from dir")
	mspConfig := &msp.FabricMSPConfig{}
	err = proto.Unmarshal(cfg.Config, mspConfig)
	require.NoError(t, err, "Error unmarshaling msp config")

	dir := randomMspDir()

	err = GenerateMspDir(dir, cfg)
	require.NoError(t, err, "Error generating msp dir")

	cfg1, err := mspcfg.GetVerifyingMspConfig(dir, "mymspid", "bccsp")
	require.NoError(t, err, "Error generating msp config from dir1")
	mspConfig1 := &msp.FabricMSPConfig{}
	err = proto.Unmarshal(cfg1.Config, mspConfig1)
	require.NoError(t, err, "Error unmarshaling msp config1")

	require.Equal(t, mspConfig.RootCerts, mspConfig1.RootCerts, "RootCerts are different")
	require.Equal(t, mspConfig.RevocationList, mspConfig1.RevocationList, "RevocationList are different")
	require.Equal(t, mspConfig.TlsIntermediateCerts, mspConfig1.TlsIntermediateCerts, "TlsIntermediateCerts are different")
	require.Equal(t, mspConfig.TlsRootCerts, mspConfig1.TlsRootCerts, "TlsRootCerts are different")
	require.Equal(t, mspConfig.IntermediateCerts, mspConfig1.IntermediateCerts, "IntermediateCerts are different")
	require.Equal(t, mspConfig.Admins, mspConfig1.Admins, "Admins are different")
}
