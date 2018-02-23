/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resource provides access to fabric network resource management, typically using system channel queries.
package resource

import (
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	ccomm "github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/resource/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Resource is a client that provides access to fabric network resource management.
type Resource struct {
	clientContext context.Context
}

// New returns a Client instance with the SDK context.
func New(ctx context.Context) *Resource {
	c := Resource{clientContext: ctx}
	return &c
}

type fabCtx struct {
	context.ProviderContext
	context.IdentityContext
}

// SignChannelConfig signs a configuration.
func (c *Resource) SignChannelConfig(config []byte, signer context.IdentityContext) (*common.ConfigSignature, error) {
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

	ctx := fabCtx{
		ProviderContext: c.clientContext,
		IdentityContext: signingUser,
	}

	return CreateConfigSignature(ctx, config)
}

// CreateChannel calls the orderer to start building the new channel.
func (c *Resource) CreateChannel(request api.CreateChannelRequest) (fab.TransactionID, error) {
	if request.Orderer == nil {
		return fab.TransactionID{}, errors.New("missing orderer request parameter for the initialize channel")
	}

	if request.Name == "" {
		return fab.TransactionID{}, errors.New("missing name request parameter for the new channel")
	}

	if request.Envelope != nil {
		return c.createChannelFromEnvelope(request)
	}

	if request.Config == nil {
		return fab.TransactionID{}, errors.New("missing envelope request parameter containing the configuration of the new channel")
	}

	if request.Signatures == nil {
		return fab.TransactionID{}, errors.New("missing signatures request parameter for the new channel")
	}

	txnID, err := txn.NewID(c.clientContext)
	if err != nil {
		return txnID, err
	}

	return txnID, c.createOrUpdateChannel(txnID, request)
}

// TODO: this function was extracted from createOrUpdateChannel, but needs a closer examination.
func (c *Resource) createChannelFromEnvelope(request api.CreateChannelRequest) (fab.TransactionID, error) {
	env, err := c.extractSignedEnvelope(request.Envelope)
	if err != nil {
		return fab.TransactionID{}, errors.WithMessage(err, "signed envelope not valid")
	}

	// Send request
	_, err = request.Orderer.SendBroadcast(env)
	if err != nil {
		return fab.TransactionID{}, errors.WithMessage(err, "failed broadcast to orderer")
	}
	return fab.TransactionID{}, nil
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
	seekInfoBytes, err := proto.Marshal(seekInfo)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling of seek info header failed")
	}

	tlsCertHash := ccomm.TLSCertHash(c.clientContext.Config())
	channelHeaderOpts := txn.ChannelHeaderOpts{
		ChannelID:   channelName,
		TxnID:       txnID,
		TLSCertHash: tlsCertHash,
	}
	seekInfoHeader, err := txn.CreateChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, channelHeaderOpts)
	if err != nil {
		return nil, errors.Wrap(err, "CreateChannelHeader failed")
	}

	payload, err := txn.CreatePayload(txnID, seekInfoHeader, seekInfoBytes)
	if err != nil {
		return nil, errors.Wrap(err, "CreatePayload failed")
	}

	block, err := txn.SendPayload(c.clientContext, payload, orderers)
	if err != nil {
		return nil, errors.WithMessage(err, "SendEnvelope failed")
	}
	return block, nil
}

// JoinChannel sends a join channel proposal to the target peer.
//
// TODO extract targets from request into parameter.
func (c *Resource) JoinChannel(request api.JoinChannelRequest) error {

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

func (c *Resource) extractSignedEnvelope(reqEnvelope []byte) (*fab.SignedEnvelope, error) {
	envelope := &common.Envelope{}
	err := proto.Unmarshal(reqEnvelope, envelope)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal request envelope failed")
	}
	se := fab.SignedEnvelope{
		Signature: envelope.Signature,
		Payload:   envelope.Payload,
	}
	return &se, nil
}

// createOrUpdateChannel creates a new channel or updates an existing channel.
func (c *Resource) createOrUpdateChannel(txnID fab.TransactionID, request api.CreateChannelRequest) error {

	configUpdateEnvelope := &common.ConfigUpdateEnvelope{
		ConfigUpdate: request.Config,
		Signatures:   request.Signatures,
	}
	configUpdateEnvelopeBytes, err := proto.Marshal(configUpdateEnvelope)
	if err != nil {
		return errors.Wrap(err, "marshal configUpdateEnvelope failed")
	}

	channelHeaderOpts := txn.ChannelHeaderOpts{
		ChannelID:   request.Name,
		TxnID:       txnID,
		TLSCertHash: ccomm.TLSCertHash(c.clientContext.Config()),
	}
	channelHeader, err := txn.CreateChannelHeader(common.HeaderType_CONFIG_UPDATE, channelHeaderOpts)
	if err != nil {
		return errors.WithMessage(err, "CreateChannelHeader failed")
	}

	payload, err := txn.CreatePayload(txnID, channelHeader, configUpdateEnvelopeBytes)
	if err != nil {
		return errors.WithMessage(err, "CreatePayload failed")
	}

	_, err = txn.BroadcastPayload(c.clientContext, payload, []fab.Orderer{request.Orderer})
	if err != nil {
		return errors.WithMessage(err, "SendEnvelope failed")
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
func (c *Resource) InstallChaincode(req api.InstallChaincodeRequest) ([]*fab.TransactionProposalResponse, string, error) {

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

	propReq := ChaincodeInstallRequest{
		Name:    req.Name,
		Path:    req.Path,
		Version: req.Version,
		Package: &ChaincodePackage{
			Type: req.Package.Type,
			Code: req.Package.Code,
		},
	}

	txid, err := txn.NewID(c.clientContext)
	if err != nil {
		return nil, "", errors.WithMessage(err, "create transaction ID failed")
	}

	prop, err := CreateChaincodeInstallProposal(txid, propReq)
	if err != nil {
		return nil, "", errors.WithMessage(err, "creation of install chaincode proposal failed")
	}

	transactionProposalResponse, err := txn.SendProposal(c.clientContext, prop, req.Targets)

	return transactionProposalResponse, prop.TxnID.ID, err
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

	txid, err := txn.NewID(c.clientContext)
	if err != nil {
		return nil, errors.WithMessage(err, "create transaction ID failed")
	}

	tp, err := txn.CreateChaincodeInvokeProposal(txid, systemChannel, request)
	if err != nil {
		return nil, errors.WithMessage(err, "NewProposal failed")
	}

	tpr, err := txn.SendProposal(c.clientContext, tp, targets)
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
