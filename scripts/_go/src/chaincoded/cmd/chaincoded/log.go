/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"
	"os"
)

func isVerbose() bool {
	_, ok := os.LookupEnv("CHAINCODED_VERBOSE")
	return ok
}

func logDebugf(format string, v ...interface{}) {
	if isVerbose() {
		log.Printf(format, v...)
	}
}

func logInfof(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func logWarningf(format string, v ...interface{}) {
	log.Printf("Warning: "+format, v...)
}

func logFatalf(format string, v ...interface{}) {
	log.Printf("Fatal: "+format, v...)
}
