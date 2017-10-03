/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package errors

import (
	e "errors"
	"fmt"
	"testing"

	perrors "github.com/pkg/errors"
)

func TestNewBridge(t *testing.T) {
	msg1 := "error msg"
	err := New(msg1)

	if err.Error() != msg1 {
		t.Fatal("Unexpected error text")
	}
	testStack(t, err)
}

func TestErrorfBridge(t *testing.T) {
	msg1 := "formatted error msg"
	err := Errorf("%s", msg1)

	if err.Error() != msg1 {
		t.Fatal("Unexpected error text")
	}
	testStack(t, err)
}

func TestWrapBridge(t *testing.T) {
	msg1 := "error msg"
	err1 := e.New(msg1)

	msg2 := "wrap msg"
	err2 := Wrap(err1, msg2)

	testWrapMsg(t, err1, err2, msg1, msg2)
	testCause(t, err1, err2)
	testStack(t, err2)
}

func TestWrapfBridge(t *testing.T) {
	msg1 := "error msg"
	err1 := e.New(msg1)

	msg2 := "formatted wrap msg"
	err2 := Wrapf(err1, "%s", msg2)

	testWrapMsg(t, err1, err2, msg1, msg2)
	testCause(t, err1, err2)
	testStack(t, err2)
}

func TestWithMessageBridge(t *testing.T) {
	msg1 := "error msg"
	err1 := e.New(msg1)

	msg2 := "formatted wrap msg"
	err2 := WithMessage(err1, msg2)

	testWrapMsg(t, err1, err2, msg1, msg2)
	testCause(t, err1, err2)
	testNoStack(t, err2)
}

func TestWithStackBridge(t *testing.T) {
	msg1 := "error msg"
	err1 := e.New(msg1)

	err2 := WithStack(err1)

	testCause(t, err1, err2)
	testStack(t, err2)
}

func testWrapMsg(t *testing.T, err1 error, err2 error, msg1 string, msg2 string) {
	txt := err2.Error()
	expected := fmt.Sprintf("%s: %s", msg2, msg1)
	if txt != expected {
		t.Fatalf("Unexpected error text [txt:%v: expected:%v]", txt, expected)
	}
}

func testCause(t *testing.T, err1 error, err2 error) {
	cause := Cause(err2)
	if cause != err1 {
		t.Fatal("Unexpected cause")
	}
}

type stackTracer interface {
	StackTrace() perrors.StackTrace
}

func testStack(t *testing.T, err error) {
	if _, ok := err.(stackTracer); !ok {
		t.Fatal("stackTracer interface expected")
	}
}

func testNoStack(t *testing.T, err error) {
	if _, ok := err.(stackTracer); ok {
		t.Fatal("stackTracer interface not expected")
	}
}
