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

func getOrdererOrg() *genesisconfig.Organization {

	return &genesisconfig.Organization{
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
}

func mockSampleSingleMSPSoloProfile() *genesisconfig.Profile {

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
					getOrdererOrg(),
				},
			},
		},
	}
}

func getApplication() *genesisconfig.Application {
	return &genesisconfig.Application{
		ACLs: map[string]string{
			"_lifecycle/CommitChaincodeDefinition": "/Channel/Application/Writers",
			"_lifecycle/QueryChaincodeDefinition":  "/Channel/Application/Readers",
			"_lifecycle/QueryNamespaceDefinitions": "/Channel/Application/Readers",
			"lscc/ChaincodeExists":                 "/Channel/Application/Readers",
			"lscc/GetDeploymentSpec":               "/Channel/Application/Readers",
			"lscc/GetChaincodeData":                "/Channel/Application/Readers",
			"lscc/GetInstantiatedChaincodes":       "/Channel/Application/Readers",
			"qscc/GetChainInfo":                    "/Channel/Application/Readers",
			"qscc/GetBlockByNumber":                "/Channel/Application/Readers",
			"qscc/GetBlockByHash":                  "/Channel/Application/Readers",
			"qscc/GetTransactionByID":              "/Channel/Application/Readers",
			"qscc/GetBlockByTxID":                  "/Channel/Application/Readers",
			"cscc/GetConfigBlock":                  "/Channel/Application/Readers",
			"cscc/GetConfigTree":                   "/Channel/Application/Readers",
			"cscc/SimulateConfigTreeUpdate":        "/Channel/Application/Readers",
			"peer/Propose":                         "/Channel/Application/Writers",
			"peer/ChaincodeToChaincode":            "/Channel/Application/Readers",
			"event/Block":                          "/Channel/Application/Readers",
			"event/FilteredBlock":                  "/Channel/Application/Readers",
		},
		Organizations: []*genesisconfig.Organization{},
		Policies: map[string]*genesisconfig.Policy{
			"LifecycleEndorsement": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Endorsement",
			},
			"Endorsement": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Endorsement",
			},
			"Readers": {
				Type: "ImplicitMeta",
				Rule: "ANY Readers",
			},
			"Writers": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
			"Admins": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Admins",
			},
		},
		Capabilities: map[string]bool{
			"V2_0": true,
			"V1_3": false,
			"V1_2": false,
			"V1_1": false,
		},
	}
}

func mockSampleSingleMSPChannelProfile() *genesisconfig.Profile {

	return &genesisconfig.Profile{
		Policies:    getPolicies(),
		Application: getApplication(),
		Consortium:  "SampleConsortium",
	}
}

func TestCreateAndInspectGenesiBlock(t *testing.T) {

	b, err := CreateGenesisBlock(mockSampleSingleMSPSoloProfile(), "mychannel")
	require.NoError(t, err, "Failed to create genesis block")
	require.NotNil(t, b, "Failed to create genesis block")

	s, err := InspectGenesisBlock(b)
	require.NoError(t, err, "Failed to inspect genesis block")
	require.False(t, s == "", "Failed to inspect genesis block")
}

func TestCreateAndInspectConfigTx(t *testing.T) {

	e, err := CreateChannelCreateTx(mockSampleSingleMSPChannelProfile(), nil, "foo")
	require.NoError(t, err, "Failed to create channel create tx")
	require.NotNil(t, e, "Failed to create channel create tx")

	s, err := InspectChannelCreateTx(e)
	require.NoError(t, err, "Failed to inspect channel create tx")
	require.False(t, s == "", "Failed to inspect channel create tx")
}
