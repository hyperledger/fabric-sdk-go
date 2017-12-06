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

// JoinChannelOpts contains options for peers joining channel
type JoinChannelOpts struct {
	Targets      []fab.Peer   // target peers
	TargetFilter TargetFilter // peer filter
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
	Err    error
}

// InstallCCOpts contains options for installing chaincode
type InstallCCOpts struct {
	Targets      []fab.Peer   // target peers
	TargetFilter TargetFilter // target filter
}

// InstantiateCCRequest contains instantiate chaincode request parameters
type InstantiateCCRequest struct {
	Name    string
	Path    string
	Version string
	Args    [][]byte
	Policy  *common.SignaturePolicyEnvelope
}

// InstantiateCCOpts contains options for instantiating chaincode
type InstantiateCCOpts struct {
	Targets      []fab.Peer   // target peers
	TargetFilter TargetFilter // target filter
	Timeout      time.Duration
}

// UpgradeCCRequest contains upgrade chaincode request parameters
type UpgradeCCRequest struct {
	Name    string
	Path    string
	Version string
	Args    [][]byte
	Policy  *common.SignaturePolicyEnvelope
}

// UpgradeCCOpts contains options for upgrading chaincode
type UpgradeCCOpts struct {
	Targets      []fab.Peer   // target peers
	TargetFilter TargetFilter // target filter
	Timeout      time.Duration
}

// ResourceMgmtClient is responsible for managing resources: peers joining channels, and installing and instantiating chaincodes(TODO).
type ResourceMgmtClient interface {

	// InstallCC installs chaincode
	InstallCC(req InstallCCRequest) ([]InstallCCResponse, error)

	// InstallCCWithOpts installs chaincode with custom options (specific peers, filtered peers)
	InstallCCWithOpts(req InstallCCRequest, opts InstallCCOpts) ([]InstallCCResponse, error)

	// InstantiateCC instantiates chaincode using default settings
	InstantiateCC(channelID string, req InstantiateCCRequest) error

	// InstantiateCCWithOpts instantiates chaincode with custom options (target peers, filtered peers)
	InstantiateCCWithOpts(channelID string, req InstantiateCCRequest, opts InstantiateCCOpts) error

	// UpgradeCC upgrades chaincode using default settings
	UpgradeCC(channelID string, req UpgradeCCRequest) error

	// UpgradeCCWithOpts upgrades chaincode with custom options (target peers, filtered peers)
	UpgradeCCWithOpts(channelID string, req UpgradeCCRequest, opts UpgradeCCOpts) error

	// JoinChannel allows for peers to join existing channel
	JoinChannel(channelID string) error

	//JoinChannelWithOpts allows for customizing set of peers about to join the channel (specific peers/filtered peers)
	JoinChannelWithOpts(channelID string, opts JoinChannelOpts) error
}
