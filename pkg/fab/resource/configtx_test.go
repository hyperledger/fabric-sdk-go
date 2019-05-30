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

// Mock profiles are based on https://github.com/hyperledger/fabric/blob/v2.0.0-alpha/sampleconfig/configtx.yaml

func channelCapabilities() map[string]bool {
	return map[string]bool{
		"V1_3": true,
	}
}

func channelDefaults() (map[string]*genesisconfig.Policy, map[string]bool) {

	policies := map[string]*genesisconfig.Policy{
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
	return policies, channelCapabilities()
}

func ordererCapabilities() map[string]bool {
	return map[string]bool{
		"V1_1": true,
	}
}

func ordererDefauls() *genesisconfig.Orderer {
	return &genesisconfig.Orderer{
		OrdererType:  "solo",
		Addresses:    []string{"orderer.example.org:7050"},
		BatchTimeout: time.Duration(2 * time.Second),
		BatchSize: genesisconfig.BatchSize{
			MaxMessageCount:   500,
			AbsoluteMaxBytes:  10 * 1024 * 1024,
			PreferredMaxBytes: 2 * 1024 * 1024,
		},
		MaxChannels: 0,
		Policies: map[string]*genesisconfig.Policy{
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
				Rule: "ANY Admins",
			},
			"BlockValidation": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
		},
		Capabilities: ordererCapabilities(),
	}
}

func sampleOrgPolicies() map[string]*genesisconfig.Policy {
	return map[string]*genesisconfig.Policy{
		"Readers": {
			Type: "Signature",
			Rule: "OR('SampleOrg.member')",
		},
		"Writers": {
			Type: "Signature",
			Rule: "OR('SampleOrg.member')",
		},
		"Admins": {
			Type: "Signature",
			Rule: "OR('SampleOrg.admin')",
		},
		"Endorsement": {
			Type: "Signature",
			Rule: "OR('SampleOrg.member')",
		},
	}
}

func sampleOrg() *genesisconfig.Organization {
	return &genesisconfig.Organization{
		Name:          "SampleOrg",
		SkipAsForeign: false,
		ID:            "SampleOrg",
		MSPDir:        filepath.Join(metadata.GetProjectPath(), "test/fixtures/fabric/v1/crypto-config/ordererOrganizations/example.com/msp"),
		MSPType:       "bccsp",
		Policies:      sampleOrgPolicies(),
		AnchorPeers: []*genesisconfig.AnchorPeer{
			{
				Host: "127.0.0.1",
				Port: 7051,
			},
		},
	}
}

func sampleSingleMSPSolo() *genesisconfig.Profile {

	policies, _ := channelDefaults()
	orderer := ordererDefauls()
	orderer.Organizations = []*genesisconfig.Organization{
		sampleOrg(),
	}

	return &genesisconfig.Profile{
		Policies: policies,
		Orderer:  orderer,
		Consortiums: map[string]*genesisconfig.Consortium{
			"SampleConsortium": {
				Organizations: []*genesisconfig.Organization{
					sampleOrg(),
				},
			},
		},
	}
}

func applicationDefaults() *genesisconfig.Application {

	_, capabilities := channelDefaults()

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
		Capabilities: capabilities,
	}
}

func sampleSingleMSPChannel() *genesisconfig.Profile {

	policies, _ := channelDefaults()
	appDefaults := applicationDefaults()
	appDefaults.Organizations = []*genesisconfig.Organization{
		sampleOrg(),
	}

	return &genesisconfig.Profile{
		Policies:    policies,
		Consortium:  "SampleConsortium",
		Application: appDefaults,
	}
}

func TestInspectMissing(t *testing.T) {
	_, err := InspectBlock(nil)
	require.Error(t, err, "Missing block")
}

func TestMissingOrdererSection(t *testing.T) {
	config := sampleSingleMSPSolo()
	config.Orderer = nil

	_, err := CreateGenesisBlock(config, "mychannel")
	require.Error(t, err, "Missing orderer section")
}

func TestMissingConsortiumSection(t *testing.T) {
	config := sampleSingleMSPSolo()
	config.Consortiums = nil

	_, err := CreateGenesisBlock(config, "mychannel")
	require.NoError(t, err, "Missing consortiums section")
}

func TestForOrdererMissingConsortiumSection(t *testing.T) {
	config := sampleSingleMSPSolo()
	config.Consortiums = nil

	_, err := CreateGenesisBlockForOrderer(config, "mychannel")
	require.Error(t, err, "Missing consortiums section")
}

func TestCreateAndInspectGenesiBlock(t *testing.T) {

	b, err := CreateGenesisBlock(sampleSingleMSPSolo(), "mychannel")
	require.NoError(t, err, "Failed to create genesis block")
	require.NotNil(t, b, "Failed to create genesis block")

	s, err := InspectBlock(b)
	require.NoError(t, err, "Failed to inspect genesis block")
	require.False(t, s == "", "Failed to inspect genesis block")
}

func TestCreateAndInspectGenesiBlockForOrderer(t *testing.T) {

	b, err := CreateGenesisBlockForOrderer(sampleSingleMSPSolo(), "mychannel")
	require.NoError(t, err, "Failed to create genesis block")
	require.NotNil(t, b, "Failed to create genesis block")

	s, err := InspectBlock(b)
	require.NoError(t, err, "Failed to inspect genesis block")
	require.False(t, s == "", "Failed to inspect genesis block")
}

func TestMissingConsortiumValue(t *testing.T) {
	config := sampleSingleMSPChannel()
	config.Consortium = ""

	_, err := CreateChannelCreateTx(config, nil, "configtx")
	require.Error(t, err, "Missing Consortium value in Application Profile definition")
}

func TestMissingApplicationValue(t *testing.T) {
	config := sampleSingleMSPChannel()
	config.Application = nil

	_, err := CreateChannelCreateTx(config, nil, "configtx")
	require.Error(t, err, "Missing Application value in Application Profile definition")
}

func TestCreateAndInspectConfigTx(t *testing.T) {

	e, err := CreateChannelCreateTx(sampleSingleMSPChannel(), nil, "foo")
	require.NoError(t, err, "Failed to create channel create tx")
	require.NotNil(t, e, "Failed to create channel create tx")

	s, err := InspectChannelCreateTx(e)
	require.NoError(t, err, "Failed to inspect channel create tx")
	require.False(t, s == "", "Failed to inspect channel create tx")
}

func TestGenerateAnchorPeersUpdate(t *testing.T) {

	e, err := CreateAnchorPeersUpdate(sampleSingleMSPChannel(), "foo", "SampleOrg")
	require.NoError(t, err, "Failed to create anchor peers update")
	require.NotNil(t, e, "Failed to create anchor peers update")
}

func TestBadAnchorPeersUpdates(t *testing.T) {

	config := sampleSingleMSPChannel()

	_, err := CreateAnchorPeersUpdate(config, "foo", "")
	require.Error(t, err, "Bad anchorPeerUpdate request - asOrg empty")

	backupApplication := config.Application
	config.Application = nil
	_, err = CreateAnchorPeersUpdate(config, "foo", "SampleOrg")
	require.Error(t, err, "Bad anchorPeerUpdate request")

	config.Application = backupApplication

	config.Application.Organizations[0] = &genesisconfig.Organization{Name: "FakeOrg", ID: "FakeOrg"}
	_, err = CreateAnchorPeersUpdate(config, "foo", "SampleOrg")
	require.Error(t, err, "Bad anchorPeerUpdate request - fake org")
}
