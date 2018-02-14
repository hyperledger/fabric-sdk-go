/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resource provides access to fabric network resource management, typically using system channel queries.
package resource

import (
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	fcutils "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/util"
	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	ccomm "github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Resource is a client that provides access to fabric network resource management.
type Resource struct {
	clientContext fab.Context
}

// New returns a Client instance with the SDK context.
func New(ctx fab.Context) *Resource {
	c := Resource{clientContext: ctx}
	return &c
}

// ExtractChannelConfig extracts the protobuf 'ConfigUpdate' object out of the 'ConfigEnvelope'.
func (c *Resource) ExtractChannelConfig(configEnvelope []byte) ([]byte, error) {
	logger.Debug("extractConfigUpdate - start")

	envelope := &common.Envelope{}
	err := proto.Unmarshal(configEnvelope, envelope)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config envelope failed")
	}

	payload := &common.Payload{}
	err = proto.Unmarshal(envelope.Payload, payload)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal envelope payload failed")
	}

	configUpdateEnvelope := &common.ConfigUpdateEnvelope{}
	err = proto.Unmarshal(payload.Data, configUpdateEnvelope)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config update envelope")
	}

	return configUpdateEnvelope.ConfigUpdate, nil
}

// SignChannelConfig signs a configuration.
func (c *Resource) SignChannelConfig(config []byte, signer fab.IdentityContext) (*common.ConfigSignature, error) {
	logger.Debug("SignChannelConfig - start")

	if config == nil {
		return nil, errors.New("channel configuration required")
	}

	signingUser := signer
	// If signing user is not provided default to client's user context
	if signingUser == nil {
		signingUser = c.clientContext
	}

	if signingUser == nil {
		return nil, errors.New("user context required")
	}

	creator, err := signingUser.Identity()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get user context's identity")
	}
	txnID, err := txn.NewID(signingUser)
	if err != nil {
		return nil, errors.Wrap(err, "New Transaction ID failed")
	}

	// signature is across a signature header and the config update
	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   txnID.Nonce,
	}
	signatureHeaderBytes, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal signatureHeader failed")
	}

	signingMgr := c.clientContext.SigningManager()
	if signingMgr == nil {
		return nil, errors.New("signing manager is nil")
	}

	// get all the bytes to be signed together, then sign
	signingBytes := fcutils.ConcatenateBytes(signatureHeaderBytes, config)
	signature, err := signingMgr.Sign(signingBytes, signingUser.PrivateKey())
	if err != nil {
		return nil, errors.WithMessage(err, "signing of channel config failed")
	}

	// build the return object
	configSignature := &common.ConfigSignature{
		SignatureHeader: signatureHeaderBytes,
		Signature:       signature,
	}
	return configSignature, nil
}

// CreateChannel calls the orderer to start building the new channel.
func (c *Resource) CreateChannel(request fab.CreateChannelRequest) (fab.TransactionID, error) {
	haveEnvelope := false
	if request.Envelope != nil {
		logger.Debug("createChannel - have envelope")
		haveEnvelope = true
	}

	if !haveEnvelope && request.TxnID.ID == "" {
		txnID, err := txn.NewID(c.clientContext)
		if err != nil {
			return txnID, err
		}
		request.TxnID = txnID
	}

	return request.TxnID, c.createOrUpdateChannel(request, haveEnvelope)
}

// GenesisBlockFromOrderer returns the genesis block from the defined orderer that may be
// used in a join request
func (c *Resource) GenesisBlockFromOrderer(channelName string, orderer fab.Orderer) (*common.Block, error) {

	orderers := []fab.Orderer{orderer}

	txnID, err := txn.NewID(c.clientContext)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to calculate transaction id")
	}

	// now build the seek info , will be used once the channel is created
	// to get the genesis block back
	seekStart := newSpecificSeekPosition(0)
	seekStop := newSpecificSeekPosition(0)
	seekInfo := &ab.SeekInfo{
		Start:    seekStart,
		Stop:     seekStop,
		Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
	}

	tlsCertHash := ccomm.TLSCertHash(c.clientContext.Config())
	channelHeaderOpts := txn.ChannelHeaderOpts{
		ChannelID:   channelName,
		TxnID:       txnID,
		TLSCertHash: tlsCertHash,
	}
	seekInfoHeader, err := txn.CreateChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, channelHeaderOpts)
	if err != nil {
		return nil, errors.Wrap(err, "BuildChannelHeader failed")
	}
	seekHeader, err := txn.CreateHeader(c.clientContext, seekInfoHeader, txnID)
	if err != nil {
		return nil, errors.Wrap(err, "BuildHeader failed")
	}
	seekPayload := &common.Payload{
		Header: seekHeader,
		Data:   protos_utils.MarshalOrPanic(seekInfo),
	}
	seekPayloadBytes := protos_utils.MarshalOrPanic(seekPayload)

	signedEnvelope, err := txn.SignPayload(c.clientContext, seekPayloadBytes)
	if err != nil {
		return nil, errors.WithMessage(err, "SignPayload failed")
	}

	block, err := txn.SendEnvelope(c.clientContext, signedEnvelope, orderers)
	if err != nil {
		return nil, errors.WithMessage(err, "SendEnvelope failed")
	}
	return block, nil
}

// JoinChannel sends a join channel proposal to the target peer.
//
// TODO extract targets from request into parameter.
func (c *Resource) JoinChannel(request fab.JoinChannelRequest) error {

	if request.GenesisBlock == nil {
		return errors.New("missing block input parameter with the required genesis block")
	}

	genesisBlockBytes, err := proto.Marshal(request.GenesisBlock)
	if err != nil {
		return errors.Wrap(err, "marshal genesis block failed")
	}

	// Create join channel transaction proposal for target peers
	var args [][]byte
	args = append(args, genesisBlockBytes)

	pr := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cscc",
		Fcn:         "JoinChain",
		Args:        args,
	}

	_, err = c.queryChaincode(pr, request.Targets)
	return err
}

// createOrUpdateChannel creates a new channel or updates an existing channel.
func (c *Resource) createOrUpdateChannel(request fab.CreateChannelRequest, haveEnvelope bool) error {
	// Validate request
	if request.Config == nil && !haveEnvelope {
		return errors.New("missing envelope request parameter containing the configuration of the new channel")
	}

	if request.Signatures == nil && !haveEnvelope {
		return errors.New("missing signatures request parameter for the new channel")
	}

	if request.TxnID.ID == "" && !haveEnvelope {
		return errors.New("txId required")
	}

	if request.TxnID.Nonce == nil && !haveEnvelope {
		return errors.New("nonce required")
	}

	if request.Orderer == nil {
		return errors.New("missing orderer request parameter for the initialize channel")
	}

	if request.Name == "" {
		return errors.New("missing name request parameter for the new channel")
	}

	// channel = null;
	var signature []byte
	var payloadBytes []byte

	if haveEnvelope {
		logger.Debug("createOrUpdateChannel - have envelope")
		envelope := &common.Envelope{}
		err := proto.Unmarshal(request.Envelope, envelope)
		if err != nil {
			return errors.Wrap(err, "unmarshal request envelope failed")
		}
		signature = envelope.Signature
		payloadBytes = envelope.Payload
	} else {
		logger.Debug("createOrUpdateChannel - have config_update")
		configUpdateEnvelope := &common.ConfigUpdateEnvelope{
			ConfigUpdate: request.Config,
			Signatures:   request.Signatures,
		}

		// TODO: Move
		channelHeaderOpts := txn.ChannelHeaderOpts{
			ChannelID:   request.Name,
			TxnID:       request.TxnID,
			TLSCertHash: ccomm.TLSCertHash(c.clientContext.Config()),
		}
		channelHeader, err := txn.CreateChannelHeader(common.HeaderType_CONFIG_UPDATE, channelHeaderOpts)
		if err != nil {
			return errors.WithMessage(err, "BuildChannelHeader failed")
		}

		header, err := txn.CreateHeader(c.clientContext, channelHeader, request.TxnID)
		if err != nil {
			return errors.Wrap(err, "BuildHeader failed")
		}
		configUpdateEnvelopeBytes, err := proto.Marshal(configUpdateEnvelope)
		if err != nil {
			return errors.Wrap(err, "marshal configUpdateEnvelope failed")
		}
		payload := &common.Payload{
			Header: header,
			Data:   configUpdateEnvelopeBytes,
		}
		payloadBytes, err = proto.Marshal(payload)
		if err != nil {
			return errors.Wrap(err, "marshal payload failed")
		}

		signingMgr := c.clientContext.SigningManager()
		if signingMgr == nil {
			return errors.New("signing manager is nil")
		}

		signature, err = signingMgr.Sign(payloadBytes, c.clientContext.PrivateKey())
		if err != nil {
			return errors.WithMessage(err, "signing payload failed")
		}
	}

	// Send request
	_, err := request.Orderer.SendBroadcast(&fab.SignedEnvelope{
		Signature: signature,
		Payload:   payloadBytes,
	})
	if err != nil {
		return errors.WithMessage(err, "failed broadcast to orderer")
	}

	return nil
}

// QueryChannels queries the names of all the channels that a peer has joined.
func (c *Resource) QueryChannels(peer fab.ProposalProcessor) (*pb.ChannelQueryResponse, error) {

	if peer == nil {
		return nil, errors.New("peer required")
	}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cscc",
		Fcn:         "GetChannels",
	}
	payload, err := c.queryChaincodeWithTarget(request, peer)
	if err != nil {
		return nil, errors.WithMessage(err, "cscc.GetChannels failed")
	}

	response := new(pb.ChannelQueryResponse)
	err = proto.Unmarshal(payload, response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal ChannelQueryResponse failed")
	}
	return response, nil
}

// QueryInstalledChaincodes queries the installed chaincodes on a peer.
// Returns the details of all chaincodes installed on a peer.
func (c *Resource) QueryInstalledChaincodes(peer fab.ProposalProcessor) (*pb.ChaincodeQueryResponse, error) {

	if peer == nil {
		return nil, errors.New("peer required")
	}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "lscc",
		Fcn:         "getinstalledchaincodes",
	}
	payload, err := c.queryChaincodeWithTarget(request, peer)
	if err != nil {
		return nil, errors.WithMessage(err, "lscc.getinstalledchaincodes failed")
	}

	response := new(pb.ChaincodeQueryResponse)
	err = proto.Unmarshal(payload, response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal ChaincodeQueryResponse failed")
	}

	return response, nil
}

// InstallChaincode sends an install proposal to one or more endorsing peers.
func (c *Resource) InstallChaincode(req fab.InstallChaincodeRequest) ([]*fab.TransactionProposalResponse, string, error) {

	if req.Name == "" {
		return nil, "", errors.New("chaincode name required")
	}
	if req.Path == "" {
		return nil, "", errors.New("chaincode path required")
	}
	if req.Version == "" {
		return nil, "", errors.New("chaincode version required")
	}
	if req.Package == nil {
		return nil, "", errors.New("chaincode package is required")
	}

	now := time.Now()
	cds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: req.Package.Type, ChaincodeId: &pb.ChaincodeID{Name: req.Name, Path: req.Path, Version: req.Version}},
		CodePackage: req.Package.Code, EffectiveDate: &google_protobuf.Timestamp{Seconds: int64(now.Second()), Nanos: int32(now.Nanosecond())}}

	creator, err := c.clientContext.Identity()
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to get creator identity")
	}

	// create an install from a chaincodeDeploymentSpec
	proposal, txID, err := protos_utils.CreateInstallProposalFromCDS(cds, creator)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to create chaincode deploy proposal")
	}
	proposalBytes, err := protos_utils.GetBytesProposal(proposal)
	if err != nil {
		return nil, "", err
	}
	user := c.clientContext
	if user == nil {
		return nil, "", errors.New("User context is nil")
	}

	signingMgr := c.clientContext.SigningManager()
	if signingMgr == nil {
		return nil, "", errors.Errorf("signing manager is nil")
	}

	signature, err := signingMgr.Sign(proposalBytes, user.PrivateKey())
	if err != nil {
		return nil, "", err
	}

	signedProposal := &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}

	txnID := fab.TransactionID{ID: txID} // Nonce is missing

	transactionProposalResponse, err := txn.SendProposal(&fab.TransactionProposal{
		SignedProposal: signedProposal,
		Proposal:       proposal,
		TxnID:          txnID,
	}, req.Targets)

	return transactionProposalResponse, txID, err
}

func (c *Resource) queryChaincode(request fab.ChaincodeInvokeRequest, targets []fab.ProposalProcessor) ([][]byte, error) {
	var errors multi.Errors
	responses := [][]byte{}

	for _, target := range targets {
		resp, err := c.queryChaincodeWithTarget(request, target)
		responses = append(responses, resp)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return responses, errors.ToError()
}

func (c *Resource) queryChaincodeWithTarget(request fab.ChaincodeInvokeRequest, target fab.ProposalProcessor) ([]byte, error) {
	const systemChannel = ""

	targets := []fab.ProposalProcessor{target}

	tp, err := txn.NewProposal(c.clientContext, systemChannel, request)
	if err != nil {
		return nil, errors.WithMessage(err, "NewProposal failed")
	}

	tpr, err := txn.SendProposal(tp, targets)
	if err != nil {
		return nil, errors.WithMessage(err, "SendProposal failed")
	}

	err = validateResponse(tpr[0])
	if err != nil {
		return nil, errors.WithMessage(err, "transaction proposal failed")
	}

	return tpr[0].ProposalResponse.GetResponse().Payload, nil
}

func validateResponse(response *fab.TransactionProposalResponse) error {
	if response.Status != http.StatusOK {
		return errors.Errorf("bad status from %s (%d)", response.Endorser, response.Status)
	}

	return nil
}

// newSpecificSeekPosition returns a SeekPosition that requests the block at the given index
func newSpecificSeekPosition(index uint64) *ab.SeekPosition {
	return &ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: index}}}
}
