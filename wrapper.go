// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log4go

import (
	"os"
	"fmt"
	"strings"
)

var (
	Global *Logger
)

func init() {
	Global = NewConsoleLogger(DEBUG)
}

// Wrapper for (*Logger).LoadConfiguration
func LoadConfiguration(filename string) {
	Global.LoadConfiguration(filename)
}

// Wrapper for (*Logger).AddFilter
func AddFilter(name string, level int, writer LogWriter) {
	Global.AddFilter(name, level, writer)
}

// Wrapper for (*Logger).Log
func Log(level int, source, message string) {
	Global.Log(level, source, message)
}

// Wrapper for (*Logger).Logf
func Logf(level int, format string, args ...interface{}) {
	Global.intLogf(level, format, args)
}

// Wrapper for (*Logger).Finest
func Finest(format string, args ...interface{}) {
	Global.intLogf(FINEST, format, args)
}

// Wrapper for (*Logger).Fine
func Fine(format string, args ...interface{}) {
	Global.intLogf(FINE, format, args)
}

// Wrapper for (*Logger).Debug
func Debug(format string, args ...interface{}) {
	Global.intLogf(DEBUG, format, args)
}

// Wrapper for (*Logger).Trace
func Trace(format string, args ...interface{}) {
	Global.intLogf(TRACE, format, args)
}

// Wrapper for (*Logger).Info
func Info(format string, args ...interface{}) {
	Global.intLogf(INFO, format, args)
}

// Wrapper for (*Logger).Warn
func Warn(format string, args ...interface{}) os.Error {
	Global.intLogf(WARNING, format, args)
	return os.NewError(fmt.Sprintf(format, args))
}

// Wrapper for (*Logger).Error
func Error(format string, args ...interface{}) os.Error {
	Global.intLogf(ERROR, format, args)
	return os.NewError(fmt.Sprintf(format, args))
}

// Wrapper for (*Logger).Critical
func Critical(format string, args ...interface{}) os.Error {
	Global.intLogf(CRITICAL, format, args)
	return os.NewError(fmt.Sprintf(format, args))
}

func Crash(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(CRITICAL, strings.Repeat(" %v", len(args))[1:], args)
	}
	panic(args)
}

// Logs the given message and crashes the program
func Crashf(format string, args ...interface{}) {
	Global.intLogf(CRITICAL, format, args)
	Global.Close() // so that hopefully the messages get logged
	panic(fmt.Sprintf(format, args))
}

func Exit(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(ERROR, strings.Repeat(" %v", len(args))[1:], args)
	}
	Global.Close() // so that hopefully the messages get logged
	os.Exit(0)
}

func Exitf(format string, args ...interface{}) {
	Global.intLogf(ERROR, format, args)
	Global.Close() // so that hopefully the messages get logged
	os.Exit(0)
}

func Stderr(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(ERROR, strings.Repeat(" %v", len(args))[1:], args)
	}
}

func Stderrf(format string, args ...interface{}) {
	Global.intLogf(ERROR, format, args)
}

func Stdout(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(INFO, strings.Repeat(" %v", len(args))[1:], args)
	}
}

func Stdoutf(format string, args ...interface{}) {
	Global.intLogf(INFO, format, args)
}
