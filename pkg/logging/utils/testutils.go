/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"testing"
)

//VerifyTrue verifies if boolean input is true, if false then fails test
func VerifyTrue(t *testing.T, input bool, msgAndArgs ...interface{}) {
	if !input {
		failTest(t, msgAndArgs)
	}
}

//VerifyFalse verifies if boolean input is false, if true then fails test
func VerifyFalse(t *testing.T, input bool, msgAndArgs ...interface{}) {
	if input {
		failTest(t, msgAndArgs)
	}
}

//VerifyEmpty Verifies if input is empty, fails test if not empty
func VerifyEmpty(t *testing.T, in interface{}, msgAndArgs ...interface{}) {
	if in == nil {
		return
	} else if in == "" {
		return
	}
	failTest(t, msgAndArgs...)
}

//VerifyNotEmpty Verifies if input is not empty, fails test if empty
func VerifyNotEmpty(t *testing.T, in interface{}, msgAndArgs ...interface{}) {
	if in != nil {
		return
	} else if in != "" {
		return
	}
	failTest(t, msgAndArgs...)
}

func failTest(t *testing.T, msgAndArgs ...interface{}) {
	if len(msgAndArgs) == 1 {
		t.Fatal(msgAndArgs[0])
	}
	if len(msgAndArgs) > 1 {
		t.Fatalf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
}
