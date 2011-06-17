// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"os"
	"strings"
	"container/vector"
)

const (
	// Only allow these three outputs, so make an enum
	STDIN = iota
	STDOUT
	STDERR
)

const (
	FORMAT_DEFAULT = "[%D %T] [%L] (%S) %M"
	FORMAT_SHORT   = "[%t %d] [%L] %M"
	FORMAT_ABBREV  = "[%L] %M"
)

// Known format codes:
// %T - Time (15:04:05 MST)
// %t - Time (15:04)
// %D - Date (2006/01/02)
// %d - Date (01/02/06)
// %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
// %S - Source
// %M - Message
// Ignores unknown formats
// Recommended: "[%D %T] [%L] (%S) %M"
func FormatLogRecord(format string, rec *LogRecord) string {
	var ovec vector.StringVector

	// Split the string into pieces by % signs
	pieces := strings.Split(format, "%", -1)
	ovec.Resize(0, 2*len(pieces)+2) // allocate enough pieces for each piece and its previous plus an extra for the first and last piece for good measure

	// Iterate over the pieces, replacing known formats
	for i, piece := range pieces {
		if i > 0 && len(piece) > 0 {
			switch piece[0] {
			case 'T':
				ovec.Push(rec.Created.Format("15:04:05 MST"))
			case 't':
				ovec.Push(rec.Created.Format("15:04"))
			case 'D':
				ovec.Push(rec.Created.Format("2006/01/02"))
			case 'd':
				ovec.Push(rec.Created.Format("01/02/06"))
			case 'L':
				ovec.Push(levelStrings[rec.Level])
			case 'S':
				ovec.Push(rec.Source)
			case 'M':
				ovec.Push(rec.Message)
			}
			if len(piece) > 1 {
				ovec.Push(piece[1:])
			}
		} else if len(piece) > 0 {
			ovec.Push(piece)
		}
	}

	return strings.Join(ovec, "") + "\n"
}

// This log writer sends output to a file
type FormatLogWriter struct {
	// The logging format
	format string

	// The output file (stdin, stdout, stderr)
	out *os.File
}

// This is the FormatLogWriter's output method
func (flw *FormatLogWriter) LogWrite(rec *LogRecord) (n int, err os.Error) {
	// Perform the write
	return flw.out.Write([]byte(FormatLogRecord(flw.format, rec)))
}

func (flw *FormatLogWriter) Good() bool {
	return true
}

func (flw *FormatLogWriter) Close() {}

func NewFormatLogWriter(file int, format string) *FormatLogWriter {
	flw := new(FormatLogWriter)
	flw.format = format

	switch file {
	case STDIN:
		flw.out = os.Stdin
	case STDOUT:
		flw.out = os.Stdout
	case STDERR:
		flw.out = os.Stderr
	default:
		flw.out = os.Stdin
	}

	return flw
}

// Set the logging format
func (flw *FormatLogWriter) SetFormat(format string) {
	//fmt.Fprintf(os.Stderr, "FormatLogWriter.SetFormat: %v\n", format)
	flw.format = format
}
