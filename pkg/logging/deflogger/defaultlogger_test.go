/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deflogger

import (
	"bytes"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	utils "github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

func TestDefaultLoggingWithoutCallerInfo(t *testing.T) {
	HideCallerInfo(moduleName, apilogging.WARNING)
	testDefaultLogging(t)
}

func TestDefaultLoggingWithCallerInfo(t *testing.T) {
	ShowCallerInfo(moduleName, apilogging.WARNING)
	testDefaultLogging(t)
}

func testDefaultLogging(t *testing.T) {

	logger := LoggerProvider().GetLogger(moduleName)

	//change the output to buffer
	var buf bytes.Buffer
	logger.(*Logger).ChangeOutput(&buf)

	//No level set for this module so log level should be info
	utils.VerifyTrue(t, apilogging.INFO == GetLevel(moduleName), " default log level is INFO")

	//Test logger.print outputs
	VerifyBasicLogging(t, -1, logger.Print, nil, &buf, false)
	VerifyBasicLogging(t, -1, logger.Println, nil, &buf, false)
	VerifyBasicLogging(t, -1, nil, logger.Printf, &buf, false)

	//Test logger.info outputs
	VerifyBasicLogging(t, apilogging.INFO, logger.Info, nil, &buf, false)
	VerifyBasicLogging(t, apilogging.INFO, logger.Infoln, nil, &buf, false)
	VerifyBasicLogging(t, apilogging.INFO, nil, logger.Infof, &buf, false)

	//Test logger.warn outputs
	VerifyBasicLogging(t, apilogging.WARNING, logger.Warn, nil, &buf, false)
	VerifyBasicLogging(t, apilogging.WARNING, logger.Warnln, nil, &buf, false)
	VerifyBasicLogging(t, apilogging.WARNING, nil, logger.Warnf, &buf, false)

	//Test logger.error outputs
	VerifyBasicLogging(t, apilogging.ERROR, logger.Error, nil, &buf, false)
	VerifyBasicLogging(t, apilogging.ERROR, logger.Errorln, nil, &buf, false)
	VerifyBasicLogging(t, apilogging.ERROR, nil, logger.Errorf, &buf, false)

	/*
		SINCE DEBUG LOG IS NOT YET ENABLED, LOG OUTPUT SHOULD BE EMPTY
	*/
	//Test logger.debug outputs when DEBUG level is not enabled
	logger.Debug("brown fox jumps over the lazy dog")
	logger.Debugln("brown fox jumps over the lazy dog")
	logger.Debugf("brown %s jumps over the lazy %s", "fox", "dog")

	utils.VerifyEmpty(t, buf.String(), "debug log isn't supposed to show up for info level")

	//Should be false
	utils.VerifyFalse(t, IsEnabledFor(moduleName, apilogging.DEBUG), "apiapilogging.IsEnabled for is not working as expected, expected false but got true")

	//Now change the log level to apilogging.DEBUG
	SetLevel(moduleName, apilogging.DEBUG)

	//Should be false
	utils.VerifyTrue(t, IsEnabledFor(moduleName, apilogging.DEBUG), "apiapilogging.IsEnabled for is not working as expected, expected true but got false")

	//Test logger.debug outputs
	VerifyBasicLogging(t, apilogging.DEBUG, logger.Debug, nil, &buf, false)
	VerifyBasicLogging(t, apilogging.DEBUG, logger.Debugln, nil, &buf, false)
	VerifyBasicLogging(t, apilogging.DEBUG, nil, logger.Debugf, &buf, false)

	//Reset module levels for next test
	moduleLevels = &moduleLeveled{}
}

func TestDefaultLoggingPanic(t *testing.T) {

	//Reset custom logger, need default one
	logger := LoggerProvider().GetLogger(moduleName)
	//change the output to buffer
	var buf bytes.Buffer
	logger.(*Logger).ChangeOutput(&buf)

	VerifyCriticalLoggings(t, apilogging.CRITICAL, logger.Panic, nil, &buf)
	VerifyCriticalLoggings(t, apilogging.CRITICAL, logger.Panicln, nil, &buf)
	VerifyCriticalLoggings(t, apilogging.CRITICAL, nil, logger.Panicf, &buf)

}
