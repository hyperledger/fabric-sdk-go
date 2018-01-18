/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package testdata

import (
	"fmt"
	"log"

	"bytes"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
)

var logPrefixFormatter = " [%s] "

func GetSampleLoggingProvider(output *bytes.Buffer) apilogging.LoggerProvider {
	return &sampleLoggingProvider{output}
}

/*
	Sample logging provider
*/
type sampleLoggingProvider struct {
	buf *bytes.Buffer
}

//GetLogger returns default logger implementation
func (p *sampleLoggingProvider) GetLogger(module string) apilogging.Logger {
	sampleLogger := log.New(p.buf, fmt.Sprintf(logPrefixFormatter, module), log.Ldate|log.Ltime|log.LUTC)
	return &SampleLogger{customLogger: sampleLogger, module: module}
}

/*
	Sample logger
*/

type SampleLogger struct {
	customLogger *log.Logger
	module       string
}

//
func (l *SampleLogger) Fatal(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Fatalf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Fatalln(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Panic(v ...interface{})   { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Panicf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Panicln(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Print(v ...interface{})   { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Printf(format string, v ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Println(v ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Debug(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Debugf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Debugln(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Info(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Infof(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Infoln(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Warn(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Warnf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Warnln(args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Error(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
func (l *SampleLogger) Errorf(format string, args ...interface{}) {
	l.customLogger.Print("CUSTOM LOG OUTPUT")
}
func (l *SampleLogger) Errorln(args ...interface{}) { l.customLogger.Print("CUSTOM LOG OUTPUT") }
