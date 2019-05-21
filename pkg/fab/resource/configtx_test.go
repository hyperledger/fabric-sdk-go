/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/genesisconfig"
	"github.com/stretchr/testify/require"
)

func getPolicies() map[string]*genesisconfig.Policy {
	return map[string]*genesisconfig.Policy{
		"Admins": {
			Type: "ImplicitMeta",
			Rule: "ANY Admins",
		},
		"Readers": {
			Type: "ImplicitMeta",
			Rule: "ANY Readers",
		},
		"Writers": {
			Type: "ImplicitMeta",
			Rule: "ANY Writers",
		},
	}
}

var ordererOrg = &genesisconfig.Organization{
	Name:          "OrdererOrg",
	SkipAsForeign: false,
	ID:            "OrdererOrg",
	MSPDir:        filepath.Join(metadata.GetProjectPath(), "test/fixtures/fabric/v1/crypto-config/ordererOrganizations/example.com/msp"),
	MSPType:       "bccsp",
	Policies: map[string]*genesisconfig.Policy{
		"Readers": {
			Type: "Signature",
			Rule: "OR('OrdererOrg.admin')",
		},
		"Writers": {
			Type: "Signature",
			Rule: "OR('OrdererOrg.admin')",
		},
		"Admins": {
			Type: "Signature",
			Rule: "OR('OrdererOrg.admin')",
		},
		"Endorsement": {
			Type: "Signature",
			Rule: "OR('OrdererOrg.admin')",
		},
	},
}

func mockProfile() *genesisconfig.Profile {

	return &genesisconfig.Profile{
		Policies: getPolicies(),
		Orderer: &genesisconfig.Orderer{
			OrdererType:  "solo",
			Addresses:    []string{"orderer.example.org:7050"},
			BatchTimeout: time.Duration(2 * time.Second),
			BatchSize: genesisconfig.BatchSize{
				MaxMessageCount:   500,
				AbsoluteMaxBytes:  10 * 1024 * 1024,
				PreferredMaxBytes: 2 * 1024 * 1024,
			},
			MaxChannels: 0,
			Policies:    getPolicies(),
		},
		Consortiums: map[string]*genesisconfig.Consortium{
			"SampleConsortium": {
				Organizations: []*genesisconfig.Organization{
					ordererOrg,
				},
			},
		},
	}
}

func TestCreateAndInspectGenesiBlock(t *testing.T) {

	b, err := CreateGenesisBlock(mockProfile(), "mychannel")
	require.NoError(t, err, "Failed to create genesis block")
	require.NotNil(t, b, "Failed to create genesis block")

	s, err := InspectGenesisBlock(b)
	require.NoError(t, err, "Failed to inspect genesis block")
	require.False(t, s == "", "Failed to inspect genesis block")
}
