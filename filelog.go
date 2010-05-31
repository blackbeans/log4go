// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"os"
	"fmt"
	"time"
	"sync"
	"strings"
	"container/vector"
)

// This log writer sends output to a file
type FileLogWriter struct {
	// thread safety (TODO: goroutine+chans instead?)
	lock *sync.Mutex

	// The opened file
	filename string
	file  *os.File

	// The logging format
	format string

	// Rotate at linecount
	maxlines int
	maxlines_curlines int

	// Rotate at size
	maxsize int
	maxsize_cursize int

	// Rotate daily
	daily bool
	daily_opendate int

	// Keep old logfiles (.001, .002, etc)
	rotate bool
}

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
	pieces := strings.Split(format, "%", 0)
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

	return strings.Join(ovec, "")+"\n"
}

// This is the FileLogWriter's output method
func (flw *FileLogWriter) LogWrite(rec *LogRecord) (n int, err os.Error) {
	flw.lock.Lock()
	defer flw.lock.Unlock()

	// First, check if we've gone over any of our rotate triggers
	if flw.Good() {
		if (flw.maxlines > 0 && flw.maxlines_curlines >= flw.maxlines) ||
			(flw.maxsize > 0 && flw.maxsize_cursize >= flw.maxsize) ||
			(flw.daily && time.LocalTime().Day != flw.daily_opendate) {
			flw.intRotate()
		}
	}

	// Make sure the writer is (still) good
	if !flw.Good() {
		return -1, os.NewError("File was not opened successfully")
	}

	// Perform the write
	n,err = flw.file.Write([]byte(FormatLogRecord(flw.format, rec)))

	// Update the counts
	if err == nil {
		flw.maxlines_curlines++
		flw.maxsize_cursize += n
	}

	return n, err
}

func (flw *FileLogWriter) Good() bool {
	return flw.file != nil
}

func (flw *FileLogWriter) Close() {
	flw.file.Close()
	flw.file = nil
}

func NewFileLogWriter(fname string, rotate bool) *FileLogWriter {
	flw := new(FileLogWriter)
	flw.lock = new(sync.Mutex)
	flw.filename = fname
	flw.format = "[%D %T] [%L] (%S) %M"
	flw.file = nil
	flw.rotate = rotate

	flw.intRotate() // open the file for the first time

	return flw
}

// Request that the logs rotate
func (flw *FileLogWriter) Rotate() {
	flw.lock.Lock()
	defer flw.lock.Unlock()
	flw.intRotate()
}

// If this is called in a threaded context, it MUST be synchronized
func (flw *FileLogWriter) intRotate() {
	// Close any log file that may be open
	if flw.file != nil {
		flw.file.Close()
	}

	// If we are keeping log files, move it to the next available number
	if flw.rotate {
		_, err := os.Lstat(flw.filename)
		if err == nil { // file exists
			//fmt.Fprintf(os.Stderr, "FileLogWriter.intRotate: file %s exists, searching for extension\n", flw.filename)
			// Find the next available number
			num := 1
			fname := ""
			for ; err == nil && num < 999; num++ {
				fname = flw.filename + fmt.Sprintf(".%03d", num)
				_, err = os.Lstat(fname)
			}
			if err != nil {
				// Rename the file to its newfound home
				//fmt.Fprintf(os.Stderr, "FileLogWriter.intRotate: Moving %s to %s\n", flw.filename, fname)
				os.Rename(flw.filename, fname)
			} else {
				//fmt.Fprintf(os.Stderr, "FileLogWriter.intRotate: Cannot find free log number to rename %s\n", flw.filename)
			}
		}
	}

	// Open the log file
	fd, err := os.Open(flw.filename, os.O_WRONLY|os.O_APPEND|os.O_CREAT, 0660)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "FileLogWriter.intRotate: %s\n", err)
		return
	}
	flw.file = fd

	// Set the daily open date to the current date
	flw.daily_opendate = time.LocalTime().Day

	// initialize rotation values
	flw.maxlines_curlines = 0
	flw.maxsize_cursize = 0
}

// Set the logging format
func (flw *FileLogWriter) SetFormat(format string) {
	//fmt.Fprintf(os.Stderr, "FileLogWriter.SetFormat: %v\n", format)
	flw.format = format
}

// Set rotate at linecount
func (flw *FileLogWriter) SetRotateLines(maxlines int) {
	//fmt.Fprintf(os.Stderr, "FileLogWriter.SetRotateLines: %v\n", maxlines)
	flw.maxlines = maxlines
}

// Set rotate at size
func (flw *FileLogWriter) SetRotateSize(maxsize int) {
	//fmt.Fprintf(os.Stderr, "FileLogWriter.SetRotateSize: %v\n", maxsize)
	flw.maxsize = maxsize
}

// Set rotate daily
func (flw *FileLogWriter) SetRotateDaily(daily bool) {
	//fmt.Fprintf(os.Stderr, "FileLogWriter.SetRotateDaily: %v\n", daily)
	flw.daily = daily
}

// Set keep old
func (flw *FileLogWriter) SetRotate(rotate bool) {
	//fmt.Fprintf(os.Stderr, "FileLogWriter.SetRotate: %v\n", rotate)
	flw.rotate = rotate
}
