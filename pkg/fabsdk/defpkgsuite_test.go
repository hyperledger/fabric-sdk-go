/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import "testing"

func TestNewPkgSuite(t *testing.T) {
	pkgsuite := defPkgSuite{}

	core, err := pkgsuite.Core()
	if err != nil {
		t.Fatalf("Unexpected error getting default core factory")
	}
	if core == nil {
		t.Fatalf("Core is nil")
	}

	service, err := pkgsuite.Service()
	if err != nil {
		t.Fatalf("Unexpected error getting default service factory")
	}
	if service == nil {
		t.Fatalf("service is nil")
	}

	logger, err := pkgsuite.Logger()
	if err != nil {
		t.Fatalf("Unexpected error getting default logger factory")
	}
	if logger == nil {
		t.Fatalf("logger is nil")
	}
}
