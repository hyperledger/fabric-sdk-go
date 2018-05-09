/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package logging enables setting custom logger implementation.
//
//  Basic Flow:
//  1) Initialize logger
//  2) Create new logger for specific module
//  3) Call log info
package logging

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/metadata"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/modlog"
)

//Logger basic implementation of api.Logger interface
type Logger struct {
	instance api.Logger // access only via Logger.logger()
	module   string
	once     sync.Once
}

// logger factory singleton - access only via loggerProvider()
var loggerProviderInstance api.LoggerProvider
var loggerProviderOnce sync.Once

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

// TODO: enable leveler to redirect to loggerProvider
//var levelerProvider apilogging.Leveler

const (
	//loggerNotInitializedMsg is used when a logger is not initialized before logging
	loggerNotInitializedMsg = "Default logger initialized (please call logging.InitLogger if you wish to use a custom logger)"
	loggerModule            = "fabsdk/common"
)

// NewLogger creates and returns a Logger object based on the module name.
func NewLogger(module string) *Logger {
	// note: the underlying logger instance is lazy initialized on first use
	return &Logger{module: module}
}

func loggerProvider() api.LoggerProvider {
	loggerProviderOnce.Do(func() {
		// A custom logger must be initialized prior to the first log output
		// Otherwise the built-in logger is used
		loggerProviderInstance = modlog.LoggerProvider()
		logger := loggerProviderInstance.GetLogger(loggerModule)
		logger.Debug(loggerNotInitializedMsg)
	})
	return loggerProviderInstance
}

//Initialize sets new logger which takes over logging operations.
//It is required to call this function before making any loggings.
func Initialize(l api.LoggerProvider) {
	loggerProviderOnce.Do(func() {
		loggerProviderInstance = l
		logger := loggerProviderInstance.GetLogger(loggerModule)
		logger.Debug("Logger provider initialized")

		// TODO
		// use custom leveler implementation (otherwise fallback to default)
		//		levelerProvider, ok := loggingProvider.(apilogging.Leveler)
		//		if !ok {
		//		}
	})
}

//SetLevel - setting log level for given module
//  Parameters:
//  module is module name
//  level is logging level
func SetLevel(module string, level Level) {
	modlog.SetLevel(module, api.Level(level))
}

//GetLevel - getting log level for given module
//  Parameters:
//  module is module name
//
//  Returns:
//  logging level
func GetLevel(module string) Level {
	return Level(modlog.GetLevel(module))
}

//IsEnabledFor - Check if given log level is enabled for given module
//  Parameters:
//  module is module name
//  level is logging level
//
//  Returns:
//  is logging enabled for this module and level
func IsEnabledFor(module string, level Level) bool {
	return modlog.IsEnabledFor(module, api.Level(level))
}

// LogLevel returns the log level from a string representation.
//  Parameters:
//  level is logging level in string representation
//
//  Returns:
//  logging level
func LogLevel(level string) (Level, error) {
	l, err := metadata.ParseLevel(level)
	return Level(l), err
}

//Fatal calls Fatal function of underlying logger
func (l *Logger) Fatal(args ...interface{}) {
	l.logger().Fatal(args...)
}

//Fatalf calls Fatalf function of underlying logger
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger().Fatalf(format, args...)
}

//Fatalln calls Fatalln function of underlying logger
func (l *Logger) Fatalln(args ...interface{}) {
	l.logger().Fatalln(args...)
}

//Panic calls Panic function of underlying logger
func (l *Logger) Panic(args ...interface{}) {
	l.logger().Panic(args...)
}

//Panicf calls Panicf function of underlying logger
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.logger().Panicf(format, args...)
}

//Panicln calls Panicln function of underlying logger
func (l *Logger) Panicln(args ...interface{}) {
	l.logger().Panicln(args...)
}

//Print calls Print function of underlying logger
func (l *Logger) Print(args ...interface{}) {
	l.logger().Print(args...)
}

//Printf calls Printf function of underlying logger
func (l *Logger) Printf(format string, args ...interface{}) {
	l.logger().Printf(format, args...)
}

//Println calls Println function of underlying logger
func (l *Logger) Println(args ...interface{}) {
	l.logger().Println(args...)
}

//Debug calls Debug function of underlying logger
func (l *Logger) Debug(args ...interface{}) {
	l.logger().Debug(args...)
}

//Debugf calls Debugf function of underlying logger
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger().Debugf(format, args...)
}

//Debugln calls Debugln function of underlying logger
func (l *Logger) Debugln(args ...interface{}) {
	l.logger().Debugln(args...)
}

//Info calls Info function of underlying logger
func (l *Logger) Info(args ...interface{}) {
	l.logger().Info(args...)
}

//Infof calls Infof function of underlying logger
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger().Infof(format, args...)
}

//Infoln calls Infoln function of underlying logger
func (l *Logger) Infoln(args ...interface{}) {
	l.logger().Infoln(args...)
}

//Warn calls Warn function of underlying logger
func (l *Logger) Warn(args ...interface{}) {
	l.logger().Warn(args...)
}

//Warnf calls Warnf function of underlying logger
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger().Warnf(format, args...)
}

//Warnln calls Warnln function of underlying logger
func (l *Logger) Warnln(args ...interface{}) {
	l.logger().Warnln(args...)
}

//Error calls Error function of underlying logger
func (l *Logger) Error(args ...interface{}) {
	l.logger().Error(args...)
}

//Errorf calls Errorf function of underlying logger
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger().Errorf(format, args...)
}

//Errorln calls Errorln function of underlying logger
func (l *Logger) Errorln(args ...interface{}) {
	l.logger().Errorln(args...)
}

func (l *Logger) logger() api.Logger {
	l.once.Do(func() {
		l.instance = loggerProvider().GetLogger(l.module)
	})
	return l.instance
}
