/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	reqContext "context"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
)

// LifecycleInstallCCRequest contains the parameters for installing chaincode
type LifecycleInstallCCRequest struct {
	Label   string `json:"label,omitempty"`
	Package []byte `json:"package,omitempty"`
}

// LifecycleInstallCCResponse contains the response from a chaincode installation
type LifecycleInstallCCResponse struct {
	Target    string `json:"target,omitempty"`
	Status    int32  `json:"status,omitempty"`
	PackageID string `json:"packageID,omitempty"`
}

// CCReference contains the name and version of an instantiated chaincode that
// references the installed chaincode package.
type CCReference struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// LifecycleInstalledCC contains the package ID and label of the installed chaincode,
// including a map of channel name to chaincode name and version
// pairs of chaincode definitions that reference this chaincode package.
type LifecycleInstalledCC struct {
	PackageID  string                   `json:"packageID,omitempty"`
	Label      string                   `json:"label,omitempty"`
	References map[string][]CCReference `json:"references,omitempty"`
}

// LifecycleApproveCCRequest contains the parameters for approving a chaincode for an org.
type LifecycleApproveCCRequest struct {
	Name                string                          `json:"name,omitempty"`
	Version             string                          `json:"version,omitempty"`
	PackageID           string                          `json:"packageID,omitempty"`
	Sequence            int64                           `json:"sequence,omitempty"`
	EndorsementPlugin   string                          `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string                          `json:"validationPlugin,omitempty"`
	SignaturePolicy     *common.SignaturePolicyEnvelope `json:"signaturePolicy,omitempty"`
	ChannelConfigPolicy string                          `json:"channelConfigPolicy,omitempty"`
	CollectionConfig    []*pb.CollectionConfig          `json:"collectionConfig,omitempty"`
	InitRequired        bool                            `json:"initRequired,omitempty"`
}

// LifecycleQueryApprovedCCRequest contains the parameters for querying approved chaincodes
type LifecycleQueryApprovedCCRequest struct {
	Name     string `json:"name,omitempty"`
	Sequence int64  `json:"sequence,omitempty"`
}

// LifecycleApprovedChaincodeDefinition contains information about the approved chaincode
type LifecycleApprovedChaincodeDefinition struct {
	Name                string                          `json:"name,omitempty"`
	Version             string                          `json:"version,omitempty"`
	Sequence            int64                           `json:"sequence,omitempty"`
	EndorsementPlugin   string                          `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string                          `json:"validationPlugin,omitempty"`
	SignaturePolicy     *common.SignaturePolicyEnvelope `json:"signaturePolicy,omitempty"`
	ChannelConfigPolicy string                          `json:"channelConfigPolicy,omitempty"`
	CollectionConfig    []*pb.CollectionConfig          `json:"collectionConfig,omitempty"`
	InitRequired        bool                            `json:"initRequired,omitempty"`
	PackageID           string                          `json:"packageID,omitempty"`
}

// LifecycleCheckCCCommitReadinessRequest contains the parameters for checking the 'commit readiness' of a chaincode
type LifecycleCheckCCCommitReadinessRequest struct {
	Name                string                          `json:"name,omitempty"`
	Version             string                          `json:"version,omitempty"`
	Sequence            int64                           `json:"sequence,omitempty"`
	EndorsementPlugin   string                          `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string                          `json:"validationPlugin,omitempty"`
	SignaturePolicy     *common.SignaturePolicyEnvelope `json:"signaturePolicy,omitempty"`
	ChannelConfigPolicy string                          `json:"channelConfigPolicy,omitempty"`
	CollectionConfig    []*pb.CollectionConfig          `json:"collectionConfig,omitempty"`
	InitRequired        bool                            `json:"initRequired,omitempty"`
}

// LifecycleCheckCCCommitReadinessResponse contains the org approvals for the chaincode
type LifecycleCheckCCCommitReadinessResponse struct {
	Approvals map[string]bool `json:"approvals,omitempty"`
}

// LifecycleCommitCCRequest contains the parameters for committing a chaincode
type LifecycleCommitCCRequest struct {
	Name                string                          `json:"name,omitempty"`
	Version             string                          `json:"version,omitempty"`
	Sequence            int64                           `json:"sequence,omitempty"`
	EndorsementPlugin   string                          `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string                          `json:"validationPlugin,omitempty"`
	SignaturePolicy     *common.SignaturePolicyEnvelope `json:"signaturePolicy,omitempty"`
	ChannelConfigPolicy string                          `json:"channelConfigPolicy,omitempty"`
	CollectionConfig    []*pb.CollectionConfig          `json:"collectionConfig,omitempty"`
	InitRequired        bool                            `json:"initRequired,omitempty"`
}

// LifecycleQueryCommittedCCRequest contains the parameters to query committed chaincodes.
// If name is not provided then all committed chaincodes on the given channel are returned,
// otherwise only the chaincode with the given name is returned.
type LifecycleQueryCommittedCCRequest struct {
	Name string `json:"name,omitempty"`
}

// LifecycleChaincodeDefinition contains information about a committed chaincode.
// Note that approvals are only returned if a chaincode name is provided in the request.
type LifecycleChaincodeDefinition struct {
	Name                string                          `json:"name,omitempty"`
	Version             string                          `json:"version,omitempty"`
	Sequence            int64                           `json:"sequence,omitempty"`
	EndorsementPlugin   string                          `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string                          `json:"validationPlugin,omitempty"`
	SignaturePolicy     *common.SignaturePolicyEnvelope `json:"signaturePolicy,omitempty"`
	ChannelConfigPolicy string                          `json:"channelConfigPolicy,omitempty"`
	CollectionConfig    []*pb.CollectionConfig          `json:"collectionConfig,omitempty"`
	InitRequired        bool                            `json:"initRequired,omitempty"`
	Approvals           map[string]bool                 `json:"approvals,omitempty"`
}

// LifecycleInstallCC installs a chaincode package using Fabric 2.0 chaincode lifecycle.
func (rc *Client) LifecycleInstallCC(req LifecycleInstallCCRequest, options ...RequestOption) ([]LifecycleInstallCCResponse, error) {
	err := rc.lifecycleProcessor.verifyInstallParams(req)
	if err != nil {
		return nil, err
	}

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for LifecycleInstallCC")
	}

	rc.resolveTimeouts(&opts)

	parentReqCtx, parentReqCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(opts.Timeouts[fab.ResMgmt]), contextImpl.WithParent(opts.ParentContext))
	parentReqCtx = reqContext.WithValue(parentReqCtx, contextImpl.ReqContextTimeoutOverrides, opts.Timeouts)
	defer parentReqCancel()

	defaultTargets, err := rc.resolveDefaultTargets(&opts)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get default targets for InstallCC")
	}

	targets, err := rc.calculateTargets(defaultTargets, opts.TargetFilter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for install cc")
	}

	if len(targets) == 0 {
		return nil, errors.WithStack(status.New(status.ClientStatus, status.NoPeersFound.ToInt32(), "no targets available", nil))
	}

	newTargets, errs := rc.lifecycleProcessor.adjustTargetsForInstall(targets, req, opts.Retry, parentReqCtx)

	if len(newTargets) == 0 {
		// CC is already installed on all targets and/or
		// we are unable to verify if cc is installed on target(s)
		logger.Debugf("Chaincode [%s] has already been installed on all peers", req.Label)

		return nil, errs.ToError()
	}

	reqCtx, cancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(fab.ResMgmt), contextImpl.WithParent(parentReqCtx))
	defer cancel()

	responses, err := rc.lifecycleProcessor.install(reqCtx, req.Package, newTargets)

	if err != nil {
		installErrs, ok := err.(multi.Errors)
		if ok {
			errs = append(errs, installErrs)
		} else {
			errs = append(errs, err)
		}
	}

	return responses, errs.ToError()
}

// LifecycleQueryInstalledCC returns the chaincodes that were installed on a given peer with Fabric 2.0 chaincode lifecycle.
func (rc *Client) LifecycleQueryInstalledCC(options ...RequestOption) ([]LifecycleInstalledCC, error) {
	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for QueryInstalledCC")
	}

	if len(opts.Targets) != 1 {
		return nil, errors.New("only one target is supported")
	}

	rc.resolveTimeouts(&opts)

	parentReqCtx, parentReqCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(opts.Timeouts[fab.ResMgmt]), contextImpl.WithParent(opts.ParentContext))
	parentReqCtx = reqContext.WithValue(parentReqCtx, contextImpl.ReqContextTimeoutOverrides, opts.Timeouts)
	defer parentReqCancel()

	reqCtx, cancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(fab.ResMgmt), contextImpl.WithParent(parentReqCtx))
	defer cancel()

	responses, err := rc.lifecycleProcessor.queryInstalled(reqCtx, opts.Targets[0])

	var errs multi.Errors
	if err != nil {
		installErrs, ok := err.(multi.Errors)
		if ok {
			errs = append(errs, installErrs)
		} else {
			errs = append(errs, err)
		}
	}

	return responses, errs.ToError()
}

// LifecycleGetInstalledCCPackage retrieves the installed chaincode package for the given package ID.
// NOTE: The package ID may be computed with fab/ccpackager/lifecycle.ComputePackageID.
func (rc *Client) LifecycleGetInstalledCCPackage(packageID string, options ...RequestOption) ([]byte, error) {
	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for GetInstalledCCPackage")
	}

	if len(opts.Targets) != 1 {
		return nil, errors.New("only one target is supported")
	}

	rc.resolveTimeouts(&opts)

	parentReqCtx, parentReqCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(opts.Timeouts[fab.ResMgmt]), contextImpl.WithParent(opts.ParentContext))
	parentReqCtx = reqContext.WithValue(parentReqCtx, contextImpl.ReqContextTimeoutOverrides, opts.Timeouts)
	defer parentReqCancel()

	reqCtx, cancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(fab.ResMgmt), contextImpl.WithParent(parentReqCtx))
	defer cancel()

	response, err := rc.lifecycleProcessor.GetInstalledPackage(reqCtx, packageID, opts.Targets[0])

	var errs multi.Errors
	if err != nil {
		installErrs, ok := err.(multi.Errors)
		if ok {
			errs = append(errs, installErrs)
		} else {
			errs = append(errs, err)
		}
	}

	return response, errs.ToError()
}

// LifecycleApproveCC approves a chaincode for an organization.
func (rc *Client) LifecycleApproveCC(channelID string, req LifecycleApproveCCRequest, options ...RequestOption) (fab.TransactionID, error) {
	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return "", errors.WithMessage(err, "failed to get opts for ApproveCC")
	}

	reqCtx, cancel := rc.createRequestContext(opts, fab.ResMgmt)
	defer cancel()

	return rc.lifecycleProcessor.approve(reqCtx, channelID, req, opts)
}

// LifecycleQueryApprovedCC returns information about the approved chaincode definition
func (rc *Client) LifecycleQueryApprovedCC(channelID string, req LifecycleQueryApprovedCCRequest, options ...RequestOption) (LifecycleApprovedChaincodeDefinition, error) {
	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return LifecycleApprovedChaincodeDefinition{}, errors.WithMessage(err, "failed to get opts for QueryApprovedCCDefinition")
	}

	if len(opts.Targets) != 1 {
		return LifecycleApprovedChaincodeDefinition{}, errors.New("only one target is supported")
	}

	rc.resolveTimeouts(&opts)

	parentReqCtx, parentReqCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(opts.Timeouts[fab.ResMgmt]), contextImpl.WithParent(opts.ParentContext))
	parentReqCtx = reqContext.WithValue(parentReqCtx, contextImpl.ReqContextTimeoutOverrides, opts.Timeouts)
	defer parentReqCancel()

	reqCtx, cancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(fab.ResMgmt), contextImpl.WithParent(parentReqCtx))
	defer cancel()

	return rc.lifecycleProcessor.queryApproved(reqCtx, channelID, req, opts.Targets[0])
}

// LifecycleCheckCCCommitReadiness checks the 'commit readiness' of a chaincode. Returned are the org approvals.
func (rc *Client) LifecycleCheckCCCommitReadiness(channelID string, req LifecycleCheckCCCommitReadinessRequest, options ...RequestOption) (LifecycleCheckCCCommitReadinessResponse, error) {
	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return LifecycleCheckCCCommitReadinessResponse{}, errors.WithMessage(err, "failed to get opts for CheckCCCommitReadiness")
	}

	reqCtx, cancel := rc.createRequestContext(opts, fab.ResMgmt)
	defer cancel()

	return rc.lifecycleProcessor.checkCommitReadiness(reqCtx, channelID, req, opts)
}

// LifecycleCommitCC commits the chaincode to the given channel
func (rc *Client) LifecycleCommitCC(channelID string, req LifecycleCommitCCRequest, options ...RequestOption) (fab.TransactionID, error) {
	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return fab.EmptyTransactionID, errors.WithMessage(err, "failed to get opts for CommitCC")
	}

	reqCtx, cancel := rc.createRequestContext(opts, fab.ResMgmt)
	defer cancel()

	return rc.lifecycleProcessor.commit(reqCtx, channelID, req, opts)
}

// LifecycleQueryCommittedCC queries for committed chaincodes on a given channel
func (rc *Client) LifecycleQueryCommittedCC(channelID string, req LifecycleQueryCommittedCCRequest, options ...RequestOption) ([]LifecycleChaincodeDefinition, error) {
	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for QueryCommittedCC")
	}

	reqCtx, cancel := rc.createRequestContext(opts, fab.ResMgmt)
	defer cancel()

	return rc.lifecycleProcessor.queryCommitted(reqCtx, channelID, req, opts)
}
