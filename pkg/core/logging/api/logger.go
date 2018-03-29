/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

// Level defines all available log levels for log messages.
type Level int

// Log levels.
const (
	CRITICAL Level = iota
	ERROR
	WARNING
	INFO
	DEBUG
)

//Logger - Standard logger interface
type Logger interface {
	Fatal(v ...interface{})

	Fatalf(format string, v ...interface{})

	Fatalln(v ...interface{})

	Panic(v ...interface{})

	Panicf(format string, v ...interface{})

	Panicln(v ...interface{})

	Print(v ...interface{})

	Printf(format string, v ...interface{})

	Println(v ...interface{})

	Debug(args ...interface{})

	Debugf(format string, args ...interface{})

	Debugln(args ...interface{})

	Info(args ...interface{})

	Infof(format string, args ...interface{})

	Infoln(args ...interface{})

	Warn(args ...interface{})

	Warnf(format string, args ...interface{})

	Warnln(args ...interface{})

	Error(args ...interface{})

	Errorf(format string, args ...interface{})

	Errorln(args ...interface{})
}

// LoggerProvider is a factory for module loggers
// TODO: should this be renamed to LoggerFactory?
type LoggerProvider interface {
	GetLogger(module string) Logger
}

// LoggingType defines the level of logging in config
type LoggingType struct {
	Level string
}
