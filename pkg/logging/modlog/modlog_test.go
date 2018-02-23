/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package modlog

import (
	"bytes"
	"testing"

	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/logging/loglevel"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/testdata"
	utils "github.com/hyperledger/fabric-sdk-go/pkg/logging/testutils"
)

//change the output to buffer for tests
var buf bytes.Buffer

func TestDefaultLoggingWithoutCallerInfo(t *testing.T) {
	HideCallerInfo(moduleName, loglevel.WARNING)
	testDefaultLogging(t)
}

func TestDefaultLoggingWithCallerInfo(t *testing.T) {
	ShowCallerInfo(moduleName, loglevel.WARNING)
	testDefaultLogging(t)
}

func testDefaultLogging(t *testing.T) {

	logger := LoggerProvider().GetLogger(moduleName)

	//logger.(*Logger).ChangeOutput(&buf)
	logger.(*Log).ChangeOutput(&buf)

	//No level set for this module so log level should be info
	utils.VerifyTrue(t, loglevel.INFO == GetLevel(moduleName), " default log level is INFO")

	//Test logger.print outputs
	VerifyBasicLogging(t, -1, logger.Print, nil, &buf, false, moduleName)
	VerifyBasicLogging(t, -1, logger.Println, nil, &buf, false, moduleName)
	VerifyBasicLogging(t, -1, nil, logger.Printf, &buf, false, moduleName)

	//Test logger.info outputs
	VerifyBasicLogging(t, loglevel.INFO, logger.Info, nil, &buf, false, moduleName)
	VerifyBasicLogging(t, loglevel.INFO, logger.Infoln, nil, &buf, false, moduleName)
	VerifyBasicLogging(t, loglevel.INFO, nil, logger.Infof, &buf, false, moduleName)

	//Test logger.warn outputs
	VerifyBasicLogging(t, loglevel.WARNING, logger.Warn, nil, &buf, false, moduleName)
	VerifyBasicLogging(t, loglevel.WARNING, logger.Warnln, nil, &buf, false, moduleName)
	VerifyBasicLogging(t, loglevel.WARNING, nil, logger.Warnf, &buf, false, moduleName)

	//Test logger.error outputs
	VerifyBasicLogging(t, loglevel.ERROR, logger.Error, nil, &buf, false, moduleName)
	VerifyBasicLogging(t, loglevel.ERROR, logger.Errorln, nil, &buf, false, moduleName)
	VerifyBasicLogging(t, loglevel.ERROR, nil, logger.Errorf, &buf, false, moduleName)

	/*
		SINCE DEBUG LOG IS NOT YET ENABLED, LOG OUTPUT SHOULD BE EMPTY
	*/
	//Test logger.debug outputs when DEBUG level is not enabled
	logger.Debug("brown fox jumps over the lazy dog")
	logger.Debugln("brown fox jumps over the lazy dog")
	logger.Debugf("brown %s jumps over the lazy %s", "fox", "dog")

	utils.VerifyEmpty(t, buf.String(), "debug log isn't supposed to show up for info level")

	//Should be false
	utils.VerifyFalse(t, IsEnabledFor(moduleName, loglevel.DEBUG), "apiapilogging.IsEnabled for is not working as expected, expected false but got true")

	//Now change the log level to apilogging.DEBUG
	SetLevel(moduleName, loglevel.DEBUG)

	//Should be false
	utils.VerifyTrue(t, IsEnabledFor(moduleName, loglevel.DEBUG), "apiapilogging.IsEnabled for is not working as expected, expected true but got false")

	//Test logger.debug outputs
	VerifyBasicLogging(t, loglevel.DEBUG, logger.Debug, nil, &buf, false, moduleName)
	VerifyBasicLogging(t, loglevel.DEBUG, logger.Debugln, nil, &buf, false, moduleName)
	VerifyBasicLogging(t, loglevel.DEBUG, nil, logger.Debugf, &buf, false, moduleName)

	//Reset module levels for next test
	moduleLevels = &loglevel.ModuleLevels{}
}

func TestDefaultLoggingPanic(t *testing.T) {

	//Reset custom logger, need default one
	logger := LoggerProvider().GetLogger(moduleName)
	//change the output to buffer
	var buf bytes.Buffer
	logger.(*Log).ChangeOutput(&buf)

	VerifyCriticalLoggings(t, loglevel.CRITICAL, logger.Panic, nil, &buf)
	VerifyCriticalLoggings(t, loglevel.CRITICAL, logger.Panicln, nil, &buf)
	VerifyCriticalLoggings(t, loglevel.CRITICAL, nil, logger.Panicf, &buf)

}

func resetLoggerInstance() {
	loggerProviderInstance = nil
	loggerProviderOnce = sync.Once{}
}

func TestDefaultCustomModulledLogging(t *testing.T) {

	//Init logger with default logger
	resetLoggerInstance()
	//Set custom logger in place of default logger
	InitLogger(testdata.GetSampleLoggingProvider(&buf))
	//Get new logger
	dlogger := LoggerProvider().GetLogger(moduleName2)

	//No level set for this module so log level should be info
	utils.VerifyTrue(t, loglevel.INFO == GetLevel(moduleName2), " default log level is INFO")

	//Test logger.print outputs
	VerifyBasicLogging(t, loglevel.INFO, dlogger.Print, nil, &buf, true, moduleName2)
	VerifyBasicLogging(t, loglevel.INFO, dlogger.Println, nil, &buf, true, moduleName2)
	VerifyBasicLogging(t, loglevel.INFO, nil, dlogger.Printf, &buf, true, moduleName2)

	//Test logger.info outputs
	VerifyBasicLogging(t, loglevel.INFO, dlogger.Info, nil, &buf, true, moduleName2)
	VerifyBasicLogging(t, loglevel.INFO, dlogger.Infoln, nil, &buf, true, moduleName2)
	VerifyBasicLogging(t, loglevel.INFO, nil, dlogger.Infof, &buf, true, moduleName2)

	//Test logger.warn outputs
	VerifyBasicLogging(t, loglevel.WARNING, dlogger.Warn, nil, &buf, true, moduleName2)
	VerifyBasicLogging(t, loglevel.WARNING, dlogger.Warnln, nil, &buf, true, moduleName2)
	VerifyBasicLogging(t, loglevel.WARNING, nil, dlogger.Warnf, &buf, true, moduleName2)

	//Test logger.error outputs
	VerifyBasicLogging(t, loglevel.ERROR, dlogger.Error, nil, &buf, true, moduleName2)
	VerifyBasicLogging(t, loglevel.ERROR, dlogger.Errorln, nil, &buf, true, moduleName2)
	VerifyBasicLogging(t, loglevel.ERROR, nil, dlogger.Errorf, &buf, true, moduleName2)

	//Should be false
	utils.VerifyFalse(t, IsEnabledFor(moduleName2, loglevel.DEBUG), "logging.IsEnabled for is not working as expected, expected false but got true")

	//Now change the log level to DEBUG
	SetLevel(moduleName2, loglevel.DEBUG)

	//Should be false
	utils.VerifyTrue(t, IsEnabledFor(moduleName2, loglevel.DEBUG), "logging.IsEnabled for is not working as expected, expected true but got false")

	//Test logger.debug outputs
	VerifyBasicLogging(t, loglevel.DEBUG, dlogger.Debug, nil, &buf, true, moduleName2)
	VerifyBasicLogging(t, loglevel.DEBUG, dlogger.Debugln, nil, &buf, true, moduleName2)
	VerifyBasicLogging(t, loglevel.DEBUG, nil, dlogger.Debugf, &buf, true, moduleName2)

}

func TestCustomDefaultLoggingPanic(t *testing.T) {

	//Init logger with default logger
	resetLoggerInstance()
	//Set custom logger in place of default logger
	InitLogger(testdata.GetSampleLoggingProvider(&buf))
	//Get new logger
	logger := LoggerProvider().GetLogger(moduleName2)

	VerifyBasicLogging(t, loglevel.CRITICAL, logger.Fatal, nil, &buf, true, moduleName2)
	VerifyBasicLogging(t, loglevel.CRITICAL, logger.Fatalln, nil, &buf, true, moduleName2)
	VerifyBasicLogging(t, loglevel.CRITICAL, nil, logger.Fatalf, &buf, true, moduleName2)

}
