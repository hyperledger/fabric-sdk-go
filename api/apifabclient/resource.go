/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// Resource is a client that provides access to fabric resources such as chaincode.
type Resource interface {
	CreateChannel(request CreateChannelRequest) (TransactionID, error)
	InstallChaincode(request InstallChaincodeRequest) ([]*TransactionProposalResponse, string, error)
	QueryInstalledChaincodes(peer ProposalProcessor) (*pb.ChaincodeQueryResponse, error)
	QueryChannels(peer ProposalProcessor) (*pb.ChannelQueryResponse, error)

	GenesisBlockFromOrderer(channelName string, orderer Orderer) (*common.Block, error)
	JoinChannel(request JoinChannelRequest) error

	// TODO - the following methods are utilities
	ExtractChannelConfig(configEnvelope []byte) ([]byte, error)
	SignChannelConfig(config []byte, signer IdentityContext) (*common.ConfigSignature, error)
}

// CreateChannelRequest requests channel creation on the network
type CreateChannelRequest struct {
	// required - The name of the new channel
	Name string
	// required - The Orderer to send the update request
	Orderer Orderer
	// optional - the envelope object containing all
	// required settings and signatures to initialize this channel.
	// This envelope would have been created by the command
	// line tool "configtx"
	Envelope []byte
	// optional - ConfigUpdate object built by the
	// buildChannelConfig() method of this package
	Config []byte
	// optional - the list of collected signatures
	// required by the channel create policy when using the `config` parameter.
	// see signChannelConfig() method of this package
	Signatures []*common.ConfigSignature

	// TODO: InvokeChannelRequest allows the TransactionID to be passed in.
	// This request struct also has the field for consistency but perhaps it should be removed.
	TxnID TransactionID
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
	// required - proposal processor list
	Targets []ProposalProcessor
}

// CCPackage contains package type and bytes required to create CDS
type CCPackage struct {
	Type pb.ChaincodeSpec_Type
	Code []byte
}

// JoinChannelRequest allows a set of peers to transact on a channel on the network
type JoinChannelRequest struct {
	// The name of the channel to be joined.
	Name         string
	GenesisBlock *common.Block
	Targets      []ProposalProcessor
}
