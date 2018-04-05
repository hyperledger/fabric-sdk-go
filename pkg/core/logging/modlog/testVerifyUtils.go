/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package modlog

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/metadata"
	"github.com/stretchr/testify/assert"
)

const (
	basicLevelOutputWithCallerInfoExpectedRegex = "\\[%s\\] .* UTC - modlog.* -> %4.4s brown fox jumps over the lazy dog"
	basicLevelOutputExpectedRegex               = "\\[%s\\] .* UTC .*-> %4.4s brown fox jumps over the lazy dog"
	printLevelOutputExpectedRegex               = "\\[%s\\] .* brown fox jumps over the lazy dog"
	customLevelOutputExpectedRegex              = "\\[%s\\] .* CUSTOM LOG OUTPUT"
	moduleName                                  = "module-xyz"
)

type fn func(...interface{})
type fnf func(string, ...interface{})

//VerifyCriticalLoggings utility func which does job calling and verifying CRITICAL log level functions - PANIC
func VerifyCriticalLoggings(t *testing.T, level api.Level, loggerFunc fn, loggerFuncf fnf, buf fmt.Stringer) {
	//Handling panic as well as checking log output
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%v was supposed to panic", loggerFunc)
		}
		var regex string
		opts := getLoggerOpts(moduleName, level)
		if opts.callerInfoEnabled {
			//with caller info
			regex = fmt.Sprintf(basicLevelOutputWithCallerInfoExpectedRegex, moduleName, metadata.ParseString(level))
		} else {
			//without caller info
			regex = fmt.Sprintf(basicLevelOutputExpectedRegex, moduleName, metadata.ParseString(level))
		}
		match, err := regexp.MatchString(regex, buf.String())
		assert.Empty(t, err, "error while matching regex with logoutput wasnt expected")
		assert.True(t, match, "CRITICAL logger isn't producing output as expected, \n logoutput:%s\n regex: %s", buf.String(), regex)

	}()

	//Call logger func
	if loggerFunc != nil {
		loggerFunc("brown fox jumps over the lazy dog")
	} else if loggerFuncf != nil {
		loggerFuncf("brown %s jumps over the lazy %s", "fox", "dog")
	}
}

//VerifyBasicLogging utility func which does job calling and verifying basic log level functions - DEBUG, INFO, ERROR, WARNING
func VerifyBasicLogging(t *testing.T, level api.Level, loggerFunc fn, loggerFuncf fnf, buf *bytes.Buffer, verifyCustom bool, moduleName string) {

	//Call logger func
	if loggerFunc != nil {
		loggerFunc("brown fox jumps over the lazy dog")
	} else if loggerFuncf != nil {
		loggerFuncf("brown %s jumps over the lazy %s", "fox", "dog")
	}

	//check output
	var regex string
	levelName := "print"

	if verifyCustom {
		levelName = metadata.ParseString(level)
		regex = fmt.Sprintf(customLevelOutputExpectedRegex, moduleName)
	} else if level > 0 && !verifyCustom {
		levelName = metadata.ParseString(level)
		opts := getLoggerOpts(moduleName, level)
		if opts.callerInfoEnabled {
			//with caller info
			regex = fmt.Sprintf(basicLevelOutputWithCallerInfoExpectedRegex, moduleName, levelName)
		} else {
			//without caller info
			regex = fmt.Sprintf(basicLevelOutputExpectedRegex, moduleName, levelName)
		}
	} else {
		regex = fmt.Sprintf(printLevelOutputExpectedRegex, moduleName)
	}
	match, err := regexp.MatchString(regex, buf.String())

	assert.Empty(t, err, "error while matching regex with logoutput wasnt expected")
	assert.True(t, match, "%s logger isn't producing output as expected, \n logoutput:%s\n regex: %s", levelName, buf.String(), regex)

	//Reset output buffer, for next use
	buf.Reset()
}
