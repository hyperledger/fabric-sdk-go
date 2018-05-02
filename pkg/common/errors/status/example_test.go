/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package status

import "fmt"

func Example() {
	// Status errors are returned for certain transient errors by clients in the SDK
	statusError := New(ClientStatus, EndorsementMismatch.ToInt32(), "proposal responses do not match", nil)

	// Status errors implement the standard error interface and are returned as regular errors
	err := interface{}(statusError).(error)

	// A user can extract status information from a status
	status, ok := FromError(err)
	fmt.Println(ok)
	fmt.Println(status.Group)
	fmt.Println(Code(status.Code))
	fmt.Println(status.Message)

	// Output:
	// true
	// Client Status
	// ENDORSEMENT_MISMATCH
	// proposal responses do not match
}
