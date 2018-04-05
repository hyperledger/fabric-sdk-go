/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package testdata

import (
	"fmt"
	"log"

	"bytes"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
)

var logPrefixFormatter = " [%s] "

//GetSampleLoggingProvider provide sample logging
func GetSampleLoggingProvider(output *bytes.Buffer) api.LoggerProvider {
	return &sampleLoggingProvider{output}
}

/*
	Sample logging provider
*/
type sampleLoggingProvider struct {
	buf *bytes.Buffer
}

//GetLogger returns default logger implementation
func (p *sampleLoggingProvider) GetLogger(module string) api.Logger {
	sampleLogger := log.New(p.buf, fmt.Sprintf(logPrefixFormatter, module), log.Ldate|log.Ltime|log.LUTC)
	return &SampleLogger{customLogger: sampleLogger, module: module}
}

//SampleLogger ...
type SampleLogger struct {
	customLogger *log.Logger
	module       string
}

//Fatal logging
func (l *SampleLogger) Fatal(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }

//Fatalf logging
func (l *SampleLogger) Fatalf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Fatalln logging
func (l *SampleLogger) Fatalln(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }

//Panic logging
func (l *SampleLogger) Panic(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }

//Panicf logging
func (l *SampleLogger) Panicf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Panicln logging
func (l *SampleLogger) Panicln(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }

//Print logging
func (l *SampleLogger) Print(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }

//Printf logging
func (l *SampleLogger) Printf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Println logging
func (l *SampleLogger) Println(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }

//Debug logging
func (l *SampleLogger) Debug(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Debugf logging
func (l *SampleLogger) Debugf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Debugln logging
func (l *SampleLogger) Debugln(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Info logging
func (l *SampleLogger) Info(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Infof logging
func (l *SampleLogger) Infof(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Infoln logging
func (l *SampleLogger) Infoln(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Warn logging
func (l *SampleLogger) Warn(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Warnf logging
func (l *SampleLogger) Warnf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Warnln logging
func (l *SampleLogger) Warnln(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Error logging
func (l *SampleLogger) Error(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }

//Errorf logging
func (l *SampleLogger) Errorf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}

//Errorln logging
func (l *SampleLogger) Errorln(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
