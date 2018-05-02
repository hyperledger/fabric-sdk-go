/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package multi

import (
	"fmt"
)

func Example() {
	errs := Errors{}
	errs = append(errs, fmt.Errorf("peer0 failed"))
	errs = append(errs, fmt.Errorf("peer1 failed"))

	// Multi errors implement the standard error interface and are returned as regular errors
	err := interface{}(errs).(error)

	// We can extract multi errors from a standard error
	errs, ok := err.(Errors)
	fmt.Println(ok)

	// And handle each error individually
	for _, e := range errs {
		fmt.Println(e)
	}

	// Output:
	// true
	// peer0 failed
	// peer1 failed
}
