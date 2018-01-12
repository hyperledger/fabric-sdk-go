/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defpkgsuite

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestSDKOpt(t *testing.T) {
	opt := SDKOpt()

	_, err := fabsdk.New(opt, fabsdk.ConfigFile("../../../test/fixtures/config/config_test.yaml"))
	if err != nil {
		t.Fatalf("Unexpected error constructing SDK: %v", err)
	}
}

func TestNewPkgSuite(t *testing.T) {
	pkgsuite := newPkgSuite()

	if pkgsuite.Context == nil {
		t.Fatalf("Context is nil")
	}
	if pkgsuite.Core == nil {
		t.Fatalf("Core is nil")
	}
	if pkgsuite.Logger == nil {
		t.Fatalf("Logger is nil")
	}
	if pkgsuite.Service == nil {
		t.Fatalf("Service is nil")
	}
	if pkgsuite.Session == nil {
		t.Fatalf("Session is nil")
	}
}
