/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package test

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

// nolint
// Logf writes to stdout and flushes. Applicable for when t.Logf can't be used.
func Logf(template string, args ...interface{}) {
	f := bufio.NewWriter(os.Stdout)
	defer f.Flush()

	f.Write([]byte(fmt.Sprintf(template, args...)))
	f.Write([]byte(fmt.Sprintln()))
}

// nolint
// Failf - as t.Fatalf() is not goroutine safe, this function behaves like t.Fatalf().
func Failf(t *testing.T, template string, args ...interface{}) {
	f := bufio.NewWriter(os.Stdout)
	defer f.Flush()

	f.Write([]byte(fmt.Sprintf(template, args...)))
	f.Write([]byte(fmt.Sprintln()))
	t.Fail()
}
