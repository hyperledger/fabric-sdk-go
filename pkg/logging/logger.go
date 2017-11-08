/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
)

var mutex = &sync.Mutex{}

//Logger basic implementation of api.Logger interface
type Logger struct {
	logger apilogging.Logger
	module string
}

var moduleLevels apilogging.Leveled = &moduleLeveled{}
var callerInfos = callerInfo{}

var customLogger apilogging.Logger

// GetLogger creates and returns a Logger object based on the module name.
func GetLogger(module string) (*Logger, error) {
	return &Logger{logger: getDefaultLogger(module), module: module}, nil
}

// NewLogger is like GetLogger but panics if the logger can't be created.
func NewLogger(module string) *Logger {
	logger, err := GetLogger(module)
	if err != nil {
		panic("logger: " + module + ": " + err.Error())
	}
	return logger
}

//SetCustomLogger sets new custom logger which takes over logging operations already created and
//new logger which are going to be created. Care should be taken while using this method.
//It is recommended to add Custom loggers before making any loggings.
func SetCustomLogger(newCustomLogger apilogging.Logger) {
	mutex.Lock()
	customLogger = newCustomLogger
	mutex.Unlock()
}

//SetModuleLevels replaces existing levelled logging modules
func SetModuleLevels(customModuleLevels apilogging.Leveled) {
	moduleLevels = customModuleLevels
}

//SetLevel - setting log level for given module
func SetLevel(level apilogging.Level, module string) {
	moduleLevels.SetLevel(level, module)
}

//GetLevel - getting log level for given module
func GetLevel(module string) apilogging.Level {
	return moduleLevels.GetLevel(module)
}

//IsEnabledFor - Check if given log level is enabled for given module
func IsEnabledFor(level apilogging.Level, module string) bool {
	return moduleLevels.IsEnabledFor(level, module)
}

// IsEnabledForLogger will return true if given logging level is enabled for the given logger.
func IsEnabledForLogger(level apilogging.Level, logger *Logger) bool {
	return moduleLevels.IsEnabledFor(level, logger.module)
}

//ShowCallerInfo - Show caller info in log lines
func ShowCallerInfo(module string) {
	callerInfos.ShowCallerInfo(module)
}

//HideCallerInfo - Do not show caller info in log lines
func HideCallerInfo(module string) {
	callerInfos.HideCallerInfo(module)
}

//IsCallerInfoEnabled - Check if caller info is enabled for given module
func IsCallerInfoEnabled(module string) bool {
	return callerInfos.IsCallerInfoEnabled(module)
}

//Fatal calls Fatal function of underlying logger
func (l *Logger) Fatal(args ...interface{}) {
	l.getCurrentLogger().Fatal(args...)
}

//Fatalf calls Fatalf function of underlying logger
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.getCurrentLogger().Fatalf(format, args...)
}

//Fatalln calls Fatalln function of underlying logger
func (l *Logger) Fatalln(args ...interface{}) {
	l.getCurrentLogger().Fatalln(args...)
}

//Panic calls Panic function of underlying logger
func (l *Logger) Panic(args ...interface{}) {
	l.getCurrentLogger().Panic(args...)
}

//Panicf calls Panicf function of underlying logger
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.getCurrentLogger().Panicf(format, args...)
}

//Panicln calls Panicln function of underlying logger
func (l *Logger) Panicln(args ...interface{}) {
	l.getCurrentLogger().Panicln(args...)
}

//Print calls Print function of underlying logger
func (l *Logger) Print(args ...interface{}) {
	l.getCurrentLogger().Print(args...)
}

//Printf calls Printf function of underlying logger
func (l *Logger) Printf(format string, args ...interface{}) {
	l.getCurrentLogger().Printf(format, args...)
}

//Println calls Println function of underlying logger
func (l *Logger) Println(args ...interface{}) {
	l.getCurrentLogger().Println(args...)
}

//Debug calls Debug function of underlying logger
func (l *Logger) Debug(args ...interface{}) {
	if IsEnabledFor(DEBUG, l.module) {
		l.getCurrentLogger().Debug(args...)
	}
}

//Debugf calls Debugf function of underlying logger
func (l *Logger) Debugf(format string, args ...interface{}) {
	if IsEnabledFor(DEBUG, l.module) {
		l.getCurrentLogger().Debugf(format, args...)
	}
}

//Debugln calls Debugln function of underlying logger
func (l *Logger) Debugln(args ...interface{}) {
	if IsEnabledFor(DEBUG, l.module) {
		l.getCurrentLogger().Debugln(args...)
	}
}

//Info calls Info function of underlying logger
func (l *Logger) Info(args ...interface{}) {
	if IsEnabledFor(INFO, l.module) {
		l.getCurrentLogger().Info(args...)
	}
}

//Infof calls Infof function of underlying logger
func (l *Logger) Infof(format string, args ...interface{}) {
	if IsEnabledFor(INFO, l.module) {
		l.getCurrentLogger().Infof(format, args...)
	}
}

//Infoln calls Infoln function of underlying logger
func (l *Logger) Infoln(args ...interface{}) {
	if IsEnabledFor(INFO, l.module) {
		l.getCurrentLogger().Infoln(args...)
	}
}

//Warn calls Warn function of underlying logger
func (l *Logger) Warn(args ...interface{}) {
	if IsEnabledFor(WARNING, l.module) {
		l.getCurrentLogger().Warn(args...)
	}
}

//Warnf calls Warnf function of underlying logger
func (l *Logger) Warnf(format string, args ...interface{}) {
	if IsEnabledFor(WARNING, l.module) {
		l.getCurrentLogger().Warnf(format, args...)
	}
}

//Warnln calls Warnln function of underlying logger
func (l *Logger) Warnln(args ...interface{}) {
	if IsEnabledFor(WARNING, l.module) {
		l.getCurrentLogger().Warnln(args...)
	}
}

//Error calls Error function of underlying logger
func (l *Logger) Error(args ...interface{}) {
	if IsEnabledFor(ERROR, l.module) {
		l.getCurrentLogger().Error(args...)
	}
}

//Errorf calls Errorf function of underlying logger
func (l *Logger) Errorf(format string, args ...interface{}) {
	if IsEnabledFor(ERROR, l.module) {
		l.getCurrentLogger().Errorf(format, args...)
	}
}

//Errorln calls Errorln function of underlying logger
func (l *Logger) Errorln(args ...interface{}) {
	if IsEnabledFor(ERROR, l.module) {
		l.getCurrentLogger().Errorln(args...)
	}
}

//getCurrentLogger - returns customlogger is set, or default logger
func (l *Logger) getCurrentLogger() apilogging.Logger {
	if customLogger != nil {
		return customLogger
	}
	return l.logger
}
