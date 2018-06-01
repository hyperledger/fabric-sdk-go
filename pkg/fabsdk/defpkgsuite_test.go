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
		t.Fatal("Unexpected error getting default core factory")
	}
	if core == nil {
		t.Fatal("Core is nil")
	}

	msp, err := pkgsuite.MSP()
	if err != nil {
		t.Fatal("Unexpected error getting default MSP factory")
	}
	if msp == nil {
		t.Fatal("MSP is nil")
	}

	service, err := pkgsuite.Service()
	if err != nil {
		t.Fatal("Unexpected error getting default service factory")
	}
	if service == nil {
		t.Fatal("service is nil")
	}

	logger, err := pkgsuite.Logger()
	if err != nil {
		t.Fatal("Unexpected error getting default logger factory")
	}
	if logger == nil {
		t.Fatal("logger is nil")
	}
}
