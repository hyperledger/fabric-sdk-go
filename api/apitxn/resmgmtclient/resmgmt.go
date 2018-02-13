/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmtclient

import (
	"time"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	common "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

// TargetFilter allows for filtering target peers
type TargetFilter interface {
	// Accept returns true if peer should be included in the list of target peers
	Accept(peer fab.Peer) bool
}

// InstallCCRequest contains install chaincode request parameters
type InstallCCRequest struct {
	Name    string
	Path    string
	Version string
	Package *fab.CCPackage
}

// InstallCCResponse contains install chaincode response status
type InstallCCResponse struct {
	Target string
	Status int32
	Info   string
}

// InstantiateCCRequest contains instantiate chaincode request parameters
type InstantiateCCRequest struct {
	Name       string
	Path       string
	Version    string
	Args       [][]byte
	Policy     *common.SignaturePolicyEnvelope
	CollConfig []*common.CollectionConfig
}

// UpgradeCCRequest contains upgrade chaincode request parameters
type UpgradeCCRequest struct {
	Name       string
	Path       string
	Version    string
	Args       [][]byte
	Policy     *common.SignaturePolicyEnvelope
	CollConfig []*common.CollectionConfig
}

//Opts contains options for operations performed by ResourceMgmtClient
type Opts struct {
	Targets      []fab.Peer    // target peers
	TargetFilter TargetFilter  // target filter
	Timeout      time.Duration //timeout options for instantiate and upgrade CC
}

//Option func for each Opts argument
type Option func(opts *Opts) error

// ResourceMgmtClient is responsible for managing resources: peers joining channels, and installing and instantiating chaincodes(TODO).
type ResourceMgmtClient interface {

	// InstallCC installs chaincode with optional custom options (specific peers, filtered peers)
	InstallCC(req InstallCCRequest, options ...Option) ([]InstallCCResponse, error)

	// InstantiateCC instantiates chaincode with optional custom options (specific peers, filtered peers, timeout)
	InstantiateCC(channelID string, req InstantiateCCRequest, options ...Option) error

	// UpgradeCC upgrades chaincode  with optional custom options (specific peers, filtered peers, timeout)
	UpgradeCC(channelID string, req UpgradeCCRequest, options ...Option) error

	// JoinChannel allows for peers to join existing channel with optional custom options (specific peers, filtered peers)
	JoinChannel(channelID string, options ...Option) error
}
