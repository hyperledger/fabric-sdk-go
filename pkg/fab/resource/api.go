/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	common "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// CreateChannelRequest requests channel creation on the network
type CreateChannelRequest struct {
	// required - The name of the new channel
	Name string
	// required - The Orderer to send the update request
	Orderer fab.Orderer
	// optional - the envelope object containing all
	// required settings and signatures to initialize this channel.
	// This envelope would have been created by the command
	// line tool "configtx"
	Envelope []byte
	// optional - ConfigUpdate object built by the
	// buildChannelConfig() method of this package
	Config []byte
	// optional - the list of collected signatures
	// required by the channel create policy when using the `apiconfig` parameter.
	// see signChannelConfig() method of this package
	Signatures []*common.ConfigSignature
}

// InstallChaincodeRequest requests chaincode installation on the network
type InstallChaincodeRequest struct {
	// required - name of the chaincode
	Name string
	// required - path to the location of chaincode sources (path from GOPATH/src folder)
	Path string
	// chaincodeVersion: required - version of the chaincode
	Version string
	// required - package (chaincode package type and bytes)
	Package *CCPackage
}

// JoinChannelRequest allows a set of peers to transact on a channel on the network
type JoinChannelRequest struct {
	// The name of the channel to be joined.
	Name         string
	GenesisBlock *common.Block
}

// CCPackage contains package type and bytes required to create CDS
type CCPackage struct {
	Type pb.ChaincodeSpec_Type
	Code []byte
}

// LifecycleInstallProposalResponse is the response from an install proposal request
type LifecycleInstallProposalResponse struct {
	*fab.TransactionProposalResponse
	*lb.InstallChaincodeResult
}

// CCReference contains the name and version of an instantiated chaincode that
// references the installed chaincode package.
type CCReference struct {
	Name    string
	Version string
}

// LifecycleInstalledCC contains the package ID and label of the installed chaincode,
// including a map of channel name to chaincode name and version
// pairs of chaincode definitions that reference this chaincode package.
type LifecycleInstalledCC struct {
	PackageID  string
	Label      string
	References map[string][]CCReference
}

// LifecycleQueryInstalledCCResponse contains the response for a LifecycleQueryInstalledCC request.
type LifecycleQueryInstalledCCResponse struct {
	*fab.TransactionProposalResponse
	InstalledChaincodes []LifecycleInstalledCC
}

// LifecycleQueryApprovedCCResponse contains the response for a LifecycleQueryApprovedCC request
type LifecycleQueryApprovedCCResponse struct {
	*fab.TransactionProposalResponse
	ApprovedChaincode *LifecycleApprovedCC
}

// LifecycleApprovedCC contains information about an approved chaincode
type LifecycleApprovedCC struct {
	Name                string
	Version             string
	Sequence            int64
	EndorsementPlugin   string
	ValidationPlugin    string
	SignaturePolicy     *common.SignaturePolicyEnvelope
	ChannelConfigPolicy string
	CollectionConfig    []*pb.CollectionConfig
	InitRequired        bool
	PackageID           string
}
