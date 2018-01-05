/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabapi provides a default implementation of the fabric API for fabsdk
package fabapi

import (
	"github.com/hyperledger/fabric-sdk-go/def/factory/defclient"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

// NewSDKOpts returns SDK options populated with the default implementation referenced by the fabapi package
// TODO: Refactor option style
func NewSDKOpts() fabsdk.Options {
	opts := fabsdk.Options{}

	PopulateSDKOpts(&opts)

	return opts
}

// PopulateSDKOpts populates an SDK options with the default implementation referenced by the fabapi package
// TODO: Refactor option style
func PopulateSDKOpts(opts *fabsdk.Options) {
	if opts.LoggerFactory == nil {
		opts.LoggerFactory = deflogger.LoggerProvider()
	}
	if opts.CoreFactory == nil {
		opts.CoreFactory = defcore.NewProviderFactory()
	}
	if opts.ServiceFactory == nil {
		opts.ServiceFactory = defsvc.NewProviderFactory()
	}
	if opts.ContextFactory == nil {
		opts.ContextFactory = defclient.NewOrgClientFactory()
	}
	if opts.SessionFactory == nil {
		opts.SessionFactory = defclient.NewSessionClientFactory()
	}
}
