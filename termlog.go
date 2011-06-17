// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"os"
	"fmt"
)

// This is the standard writer that prints to standard output
type ConsoleLogWriter struct{}

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter() *ConsoleLogWriter { return new(ConsoleLogWriter) }

// This is the ConsoleLogWriter's output method
func (slw *ConsoleLogWriter) LogWrite(rec *LogRecord) (n int, err os.Error) {
	return fmt.Fprint(os.Stdout, "[", rec.Created.Format("01/02/06 15:04:05"), "] [", levelStrings[rec.Level], "] ", rec.Message, "\n")
}

// The standard output logger should always be writable
func (slw *ConsoleLogWriter) Good() bool { return true }

// The standard output logger never really closes
func (slw *ConsoleLogWriter) Close() {}
