/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabapi

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestNewSDKOpts(t *testing.T) {
	opts := NewSDKOpts()
	verifySDKOpts(t, &opts)
}

func TestPopulateSDKOpts(t *testing.T) {
	opts := fabsdk.Options{}
	PopulateSDKOpts(&opts)
	verifySDKOpts(t, &opts)
}

func verifySDKOpts(t *testing.T, opts *fabsdk.Options) {
	if opts.CoreFactory == nil {
		t.Fatal("Expected CoreFactory to be populated")
	}

	if opts.ServiceFactory == nil {
		t.Fatal("Expected ServiceFactory to be populated")
	}

	if opts.ContextFactory == nil {
		t.Fatal("Expected ContextFactory to be populated")
	}

	if opts.SessionFactory == nil {
		t.Fatal("Expected SessionFactory to be populated")
	}
}
