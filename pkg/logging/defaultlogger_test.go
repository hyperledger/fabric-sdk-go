/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"bytes"
	"testing"
)

func TestDefaultLoggingWithoutCallerInfo(t *testing.T) {
	HideCallerInfo(moduleName)
	testDefaultLogging(t)
}

func TestDefaultLoggingWithCallerInfo(t *testing.T) {
	ShowCallerInfo(moduleName)
	testDefaultLogging(t)
}

func testDefaultLogging(t *testing.T) {

	SetCustomLogger(nil)
	logger := NewLogger(moduleName)

	//change the output to buffer
	var buf bytes.Buffer
	logger.logger.(*DefaultLogger).defaultLogger.SetOutput(&buf)

	//No level set for this module so log level should be info
	verifyTrue(t, INFO == GetLevel(moduleName), " default log level is INFO")

	//Test logger.print outputs
	verifyBasicLogging(t, -1, logger.Print, nil, &buf, false)
	verifyBasicLogging(t, -1, logger.Println, nil, &buf, false)
	verifyBasicLogging(t, -1, nil, logger.Printf, &buf, false)

	//Test logger.info outputs
	verifyBasicLogging(t, INFO, logger.Info, nil, &buf, false)
	verifyBasicLogging(t, INFO, logger.Infoln, nil, &buf, false)
	verifyBasicLogging(t, INFO, nil, logger.Infof, &buf, false)

	//Test logger.warn outputs
	verifyBasicLogging(t, WARNING, logger.Warn, nil, &buf, false)
	verifyBasicLogging(t, WARNING, logger.Warnln, nil, &buf, false)
	verifyBasicLogging(t, WARNING, nil, logger.Warnf, &buf, false)

	//Test logger.error outputs
	verifyBasicLogging(t, ERROR, logger.Error, nil, &buf, false)
	verifyBasicLogging(t, ERROR, logger.Errorln, nil, &buf, false)
	verifyBasicLogging(t, ERROR, nil, logger.Errorf, &buf, false)

	/*
		SINCE DEBUG LOG IS NOT YET ENABLED, LOG OUTPUT SHOULD BE EMPTY
	*/
	//Test logger.debug outputs when DEBUG level is not enabled
	logger.Debug("brown fox jumps over the lazy dog")
	logger.Debugln("brown fox jumps over the lazy dog")
	logger.Debugf("brown %s jumps over the lazy %s", "fox", "dog")

	verifyEmpty(t, buf.String(), "debug log isn't supposed to show up for info level")

	//Should be false
	verifyFalse(t, IsEnabledForLogger(DEBUG, logger), "logging.IsEnabled for is not working as expected, expected false but got true")
	verifyFalse(t, IsEnabledFor(DEBUG, moduleName), "logging.IsEnabled for is not working as expected, expected false but got true")

	//Now change the log level to DEBUG
	SetLevel(DEBUG, moduleName)

	//Should be false
	verifyTrue(t, IsEnabledForLogger(DEBUG, logger), "logging.IsEnabled for is not working as expected, expected true but got false")
	verifyTrue(t, IsEnabledFor(DEBUG, moduleName), "logging.IsEnabled for is not working as expected, expected true but got false")

	//Test logger.debug outputs
	verifyBasicLogging(t, DEBUG, logger.Debug, nil, &buf, false)
	verifyBasicLogging(t, DEBUG, logger.Debugln, nil, &buf, false)
	verifyBasicLogging(t, DEBUG, nil, logger.Debugf, &buf, false)

	//Reset module levels for next test
	SetModuleLevels(&moduleLeveled{})
}

func TestDefaultLoggingPanic(t *testing.T) {

	//Reset custom logger, need default one
	SetCustomLogger(nil)
	logger := NewLogger(moduleName)

	verifyCriticalLoggings(t, CRITICAL, logger.Panic, nil, logger)
	verifyCriticalLoggings(t, CRITICAL, logger.Panicln, nil, logger)
	verifyCriticalLoggings(t, CRITICAL, nil, logger.Panicf, logger)

}
