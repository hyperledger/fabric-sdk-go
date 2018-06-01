/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package logging

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/testdata"
)

var modName = "module-xyz"

func Example() {

	Initialize(testdata.GetSampleLoggingProvider(&buf))
	//Create new logger
	logger := NewLogger(modName)

	logger.Info("log test data")

	fmt.Println("log info is completed")

	// Output: log info is completed

}

func ExampleNewLogger() {

	Initialize(testdata.GetSampleLoggingProvider(&buf))
	//Create new logger
	NewLogger(modName)

	fmt.Println("log is completed")

	// Output: log is completed

}

func ExampleInitialize() {

	Initialize(testdata.GetSampleLoggingProvider(&buf))

	fmt.Println("log is completed")

	// Output: log is completed

}

func ExampleSetLevel() {

	Initialize(testdata.GetSampleLoggingProvider(&buf))

	SetLevel(modName, INFO)

	fmt.Println("log is completed")

	// Output: log is completed

}

func ExampleGetLevel() {

	Initialize(testdata.GetSampleLoggingProvider(&buf))

	SetLevel(modName, DEBUG)

	l := GetLevel(modName)
	if l != DEBUG {
		fmt.Println("log level is not debug")
		return
	}

	fmt.Println("log is completed")

	// Output: log is completed

}

func ExampleIsEnabledFor() {

	Initialize(testdata.GetSampleLoggingProvider(&buf))

	isEnabled := IsEnabledFor(modName, DEBUG)

	if !isEnabled {
		fmt.Println("log level debug is enabled")
		return
	}

	fmt.Println("log is completed")

	// Output: log is completed

}

func ExampleLogLevel() {

	Initialize(testdata.GetSampleLoggingProvider(&buf))

	level, err := LogLevel("debug")
	if err != nil {
		fmt.Printf("failed LogLevel: %s\n", err)
		return
	}

	if level != DEBUG {
		fmt.Println("log level is not debug")
		return
	}
	fmt.Println("log is completed")

	// Output: log is completed

}
