/*
Copyright IBM Corp. 2016 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package utils

import (
	"fmt"

	"encoding/hex"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/util"
	factory "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/cryptosuitebridge"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// GetHeader Get Header from bytes
func GetHeader(bytes []byte) (*common.Header, error) {
	hdr := &common.Header{}
	err := proto.Unmarshal(bytes, hdr)
	return hdr, err
}

// GetChaincodeHeaderExtension get chaincode header extension given header
func GetChaincodeHeaderExtension(hdr *common.Header) (*peer.ChaincodeHeaderExtension, error) {
	chdr, err := UnmarshalChannelHeader(hdr.ChannelHeader)
	if err != nil {
		return nil, err
	}

	chaincodeHdrExt := &peer.ChaincodeHeaderExtension{}
	err = proto.Unmarshal(chdr.Extension, chaincodeHdrExt)
	return chaincodeHdrExt, err
}

// GetChaincodeAction gets the ChaincodeAction given chaicnode action bytes
func GetChaincodeAction(caBytes []byte) (*peer.ChaincodeAction, error) {
	chaincodeAction := &peer.ChaincodeAction{}
	err := proto.Unmarshal(caBytes, chaincodeAction)
	return chaincodeAction, err
}

// GetChaincodeEvents gets the ChaincodeEvents given chaincode event bytes
func GetChaincodeEvents(eBytes []byte) (*peer.ChaincodeEvent, error) {
	chaincodeEvent := &peer.ChaincodeEvent{}
	err := proto.Unmarshal(eBytes, chaincodeEvent)
	return chaincodeEvent, err
}

// GetProposalResponsePayload gets the proposal response payload
func GetProposalResponsePayload(prpBytes []byte) (*peer.ProposalResponsePayload, error) {
	prp := &peer.ProposalResponsePayload{}
	err := proto.Unmarshal(prpBytes, prp)
	return prp, err
}

// GetPayload Get Payload from Envelope message
func GetPayload(e *common.Envelope) (*common.Payload, error) {
	payload := &common.Payload{}
	err := proto.Unmarshal(e.Payload, payload)
	return payload, err
}

// GetTransaction Get Transaction from bytes
func GetTransaction(txBytes []byte) (*peer.Transaction, error) {
	tx := &peer.Transaction{}
	err := proto.Unmarshal(txBytes, tx)
	return tx, err
}

// GetChaincodeActionPayload Get ChaincodeActionPayload from bytes
func GetChaincodeActionPayload(capBytes []byte) (*peer.ChaincodeActionPayload, error) {
	cap := &peer.ChaincodeActionPayload{}
	err := proto.Unmarshal(capBytes, cap)
	return cap, err
}

// GetChaincodeProposalPayload Get ChaincodeProposalPayload from bytes
func GetChaincodeProposalPayload(bytes []byte) (*peer.ChaincodeProposalPayload, error) {
	cpp := &peer.ChaincodeProposalPayload{}
	err := proto.Unmarshal(bytes, cpp)
	return cpp, err
}

// GetSignatureHeader Get SignatureHeader from bytes
func GetSignatureHeader(bytes []byte) (*common.SignatureHeader, error) {
	sh := &common.SignatureHeader{}
	err := proto.Unmarshal(bytes, sh)
	return sh, err
}

// CreateChaincodeProposal creates a proposal from given input.
// It returns the proposal and the transaction id associated to the proposal
func CreateChaincodeProposal(typ common.HeaderType, chainID string, cis *peer.ChaincodeInvocationSpec, creator []byte) (*peer.Proposal, string, error) {
	return CreateChaincodeProposalWithTransient(typ, chainID, cis, creator, nil)
}

// CreateChaincodeProposalWithTransient creates a proposal from given input
// It returns the proposal and the transaction id associated to the proposal
func CreateChaincodeProposalWithTransient(typ common.HeaderType, chainID string, cis *peer.ChaincodeInvocationSpec, creator []byte, transientMap map[string][]byte) (*peer.Proposal, string, error) {
	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, "", err
	}

	// compute txid
	txid, err := ComputeProposalTxID(nonce, creator)
	if err != nil {
		return nil, "", err
	}

	return CreateChaincodeProposalWithTxIDNonceAndTransient(txid, typ, chainID, cis, nonce, creator, transientMap)
}

// CreateChaincodeProposalWithTxIDNonceAndTransient creates a proposal from given input
func CreateChaincodeProposalWithTxIDNonceAndTransient(txid string, typ common.HeaderType, chainID string, cis *peer.ChaincodeInvocationSpec, nonce, creator []byte, transientMap map[string][]byte) (*peer.Proposal, string, error) {
	ccHdrExt := &peer.ChaincodeHeaderExtension{ChaincodeId: cis.ChaincodeSpec.ChaincodeId}
	ccHdrExtBytes, err := proto.Marshal(ccHdrExt)
	if err != nil {
		return nil, "", err
	}

	cisBytes, err := proto.Marshal(cis)
	if err != nil {
		return nil, "", err
	}

	ccPropPayload := &peer.ChaincodeProposalPayload{Input: cisBytes, TransientMap: transientMap}
	ccPropPayloadBytes, err := proto.Marshal(ccPropPayload)
	if err != nil {
		return nil, "", err
	}

	// TODO: epoch is now set to zero. This must be changed once we
	// get a more appropriate mechanism to handle it in.
	var epoch uint64 = 0

	timestamp := util.CreateUtcTimestamp()

	hdr := &common.Header{ChannelHeader: MarshalOrPanic(&common.ChannelHeader{
		Type:      int32(typ),
		TxId:      txid,
		Timestamp: timestamp,
		ChannelId: chainID,
		Extension: ccHdrExtBytes,
		Epoch:     epoch}),
		SignatureHeader: MarshalOrPanic(&common.SignatureHeader{Nonce: nonce, Creator: creator})}

	hdrBytes, err := proto.Marshal(hdr)
	if err != nil {
		return nil, "", err
	}

	return &peer.Proposal{Header: hdrBytes, Payload: ccPropPayloadBytes}, txid, nil
}

// GetBytesProposalResponsePayload gets proposal response payload
func GetBytesProposalResponsePayload(hash []byte, response *peer.Response, result []byte, event []byte, ccid *peer.ChaincodeID) ([]byte, error) {
	cAct := &peer.ChaincodeAction{Events: event, Results: result, Response: response, ChaincodeId: ccid}
	cActBytes, err := proto.Marshal(cAct)
	if err != nil {
		return nil, err
	}

	prp := &peer.ProposalResponsePayload{Extension: cActBytes, ProposalHash: hash}
	prpBytes, err := proto.Marshal(prp)
	return prpBytes, err
}

// GetBytesChaincodeProposalPayload gets the chaincode proposal payload
func GetBytesChaincodeProposalPayload(cpp *peer.ChaincodeProposalPayload) ([]byte, error) {
	cppBytes, err := proto.Marshal(cpp)
	return cppBytes, err
}

// GetBytesChaincodeEvent gets the bytes of ChaincodeEvent
func GetBytesChaincodeEvent(event *peer.ChaincodeEvent) ([]byte, error) {
	eventBytes, err := proto.Marshal(event)
	return eventBytes, err
}

// GetBytesChaincodeActionPayload get the bytes of ChaincodeActionPayload from the message
func GetBytesChaincodeActionPayload(cap *peer.ChaincodeActionPayload) ([]byte, error) {
	capBytes, err := proto.Marshal(cap)
	return capBytes, err
}

// GetBytesProposal returns the bytes of a proposal message
func GetBytesProposal(prop *peer.Proposal) ([]byte, error) {
	propBytes, err := proto.Marshal(prop)
	return propBytes, err
}

// GetBytesTransaction get the bytes of Transaction from the message
func GetBytesTransaction(tx *peer.Transaction) ([]byte, error) {
	bytes, err := proto.Marshal(tx)
	return bytes, err
}

// GetBytesPayload get the bytes of Payload from the message
func GetBytesPayload(payl *common.Payload) ([]byte, error) {
	bytes, err := proto.Marshal(payl)
	return bytes, err
}

// GetBytesEnvelope get the bytes of Envelope from the message
func GetBytesEnvelope(env *common.Envelope) ([]byte, error) {
	bytes, err := proto.Marshal(env)
	return bytes, err
}

// CreateProposalFromCIS returns a proposal given a serialized identity and a ChaincodeInvocationSpec
func CreateProposalFromCIS(typ common.HeaderType, chainID string, cis *peer.ChaincodeInvocationSpec, creator []byte) (*peer.Proposal, string, error) {
	return CreateChaincodeProposal(typ, chainID, cis, creator)
}

// CreateInstallProposalFromCDS returns a install proposal given a serialized identity and a ChaincodeDeploymentSpec
func CreateInstallProposalFromCDS(ccpack proto.Message, creator []byte) (*peer.Proposal, string, error) {
	return createProposalFromCDS("", ccpack, creator, nil, nil, nil, "install")
}

// CreateDeployProposalFromCDS returns a deploy proposal given a serialized identity and a ChaincodeDeploymentSpec
func CreateDeployProposalFromCDS(chainID string, cds *peer.ChaincodeDeploymentSpec, creator []byte, policy []byte, escc []byte, vscc []byte) (*peer.Proposal, string, error) {
	return createProposalFromCDS(chainID, cds, creator, policy, escc, vscc, "deploy")
}

// CreateUpgradeProposalFromCDS returns a upgrade proposal given a serialized identity and a ChaincodeDeploymentSpec
func CreateUpgradeProposalFromCDS(chainID string, cds *peer.ChaincodeDeploymentSpec, creator []byte, policy []byte, escc []byte, vscc []byte) (*peer.Proposal, string, error) {
	return createProposalFromCDS(chainID, cds, creator, policy, escc, vscc, "upgrade")
}

// createProposalFromCDS returns a deploy or upgrade proposal given a serialized identity and a ChaincodeDeploymentSpec
func createProposalFromCDS(chainID string, msg proto.Message, creator []byte, policy []byte, escc []byte, vscc []byte, propType string) (*peer.Proposal, string, error) {
	//in the new mode, cds will be nil, "deploy" and "upgrade" are instantiates.
	var ccinp *peer.ChaincodeInput
	var b []byte
	var err error
	if msg != nil {
		b, err = proto.Marshal(msg)
		if err != nil {
			return nil, "", err
		}
	}
	switch propType {
	case "deploy":
		fallthrough
	case "upgrade":
		cds, ok := msg.(*peer.ChaincodeDeploymentSpec)
		if !ok || cds == nil {
			return nil, "", fmt.Errorf("invalid message for creating lifecycle chaincode proposal from")
		}
		ccinp = &peer.ChaincodeInput{Args: [][]byte{[]byte(propType), []byte(chainID), b, policy, escc, vscc}}
	case "install":
		ccinp = &peer.ChaincodeInput{Args: [][]byte{[]byte(propType), b}}
	}

	//wrap the deployment in an invocation spec to lscc...
	lsccSpec := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			Type:        peer.ChaincodeSpec_GOLANG,
			ChaincodeId: &peer.ChaincodeID{Name: "lscc"},
			Input:       ccinp}}

	//...and get the proposal for it
	return CreateProposalFromCIS(common.HeaderType_ENDORSER_TRANSACTION, chainID, lsccSpec, creator)
}

// ComputeProposalTxID computes TxID as the Hash computed
// over the concatenation of nonce and creator.
func ComputeProposalTxID(nonce, creator []byte) (string, error) {
	// TODO: Get the Hash function to be used from
	// channel configuration
	digest, err := factory.GetDefault().Hash(
		append(nonce, creator...),
		factory.GetSHA256Opts())
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digest), nil
}
