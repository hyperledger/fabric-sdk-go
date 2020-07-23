/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	reqContext "context"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
)

// LifecycleInstallCCRequest contains the parameters for installing chaincode
type LifecycleInstallCCRequest struct {
	Label   string
	Package []byte
}

// LifecycleInstallCCResponse contains the response from a chaincode installation
type LifecycleInstallCCResponse struct {
	Target    string
	Status    int32
	PackageID string
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

func (rc *Client) LifecycleQueryCommitted(ccName string, channelName string, options ...RequestOption) ([]*resource.LifecycleQueryCommittedResponse, error) {
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

	response, err := rc.lifecycleProcessor.QueryCommitted(reqCtx, ccName, channelName, opts.Targets[0])

	var errs multi.Errors
	if err != nil {
		queryErrs, ok := err.(multi.Errors)
		if ok {
			errs = append(errs, queryErrs)
		} else {
			errs = append(errs, err)
		}
	}

	return response, errs.ToError()
}
