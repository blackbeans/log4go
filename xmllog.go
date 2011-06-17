// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"os"
	"fmt"
	"time"
	"sync"
)

// This log writer sends output to a file
type XMLLogWriter struct {
	// thread safety (TODO: goroutine+chans instead?)
	lock *sync.Mutex

	// The opened file
	filename string
	file     *os.File

	// Rotate at linecount
	maxrecords            int
	maxrecords_currecords int

	// Rotate at size
	maxsize         int
	maxsize_cursize int

	// Rotate daily
	daily          bool
	daily_opendate int

	// Keep old logfiles (.001, .002, etc)
	rotate bool
}

// This is the XMLLogWriter's output method
func (xlw *XMLLogWriter) LogWrite(rec *LogRecord) (n int, err os.Error) {
	xlw.lock.Lock()
	defer xlw.lock.Unlock()

	// First, check if we've gone over any of our rotate triggers
	if xlw.Good() {
		if (xlw.maxrecords > 0 && xlw.maxrecords_currecords >= xlw.maxrecords) ||
			(xlw.maxsize > 0 && xlw.maxsize_cursize >= xlw.maxsize) ||
			(xlw.daily && time.LocalTime().Day != xlw.daily_opendate) {
			xlw.intRotate()
		}
	}

	// Make sure the writer is (still) good
	if !xlw.Good() {
		return -1, os.NewError("File was not opened successfully")
	}

	// Perform the write
	n, err = xlw.file.Write([]byte(FormatLogRecord("\t<record level=\"%L\">\n\t\t<timestamp>%D %T</timestamp>\n\t\t<source>%S</source>\n\t\t<message>%M</message>\n\t</record>\n", rec)))

	// Update the counts
	if err == nil {
		xlw.maxrecords_currecords++
		xlw.maxsize_cursize += n
	}

	return n, err
}

func (xlw *XMLLogWriter) Good() bool {
	return xlw.file != nil
}

func (xlw *XMLLogWriter) Close() {
	// Write the closing tag
	if xlw.file != nil {
		xlw.file.Write([]byte("</log>\n"))
	}

	xlw.file.Close()
	xlw.file = nil
}

func NewXMLLogWriter(fname string, rotate bool) *XMLLogWriter {
	xlw := new(XMLLogWriter)
	xlw.lock = new(sync.Mutex)
	xlw.filename = fname
	xlw.file = nil
	xlw.rotate = rotate

	xlw.intRotate() // open the file for the first time

	return xlw
}

// Request that the logs rotate
func (xlw *XMLLogWriter) Rotate() {
	xlw.lock.Lock()
	defer xlw.lock.Unlock()
	xlw.intRotate()
}

// If this is called in a threaded context, it MUST be synchronized
func (xlw *XMLLogWriter) intRotate() {
	// Close any log file that may be open
	xlw.Close()

	// If we are keeping log files, move it to the next available number
	if xlw.rotate {
		_, err := os.Lstat(xlw.filename)
		if err == nil { // file exists
			// Find the next available number
			num := 1
			fname := ""
			for ; err == nil && num < 999; num++ {
				fname = xlw.filename + fmt.Sprintf(".%03d", num)
				_, err = os.Lstat(fname)
			}
			if err != nil {
				// Rename the file to its newfound home
				//fmt.Fprintf(os.Stderr, "XMLLogWriter.intRotate: Moving %s to %s\n", xlw.filename, fname)
				os.Rename(xlw.filename, fname)
			} else {
				fmt.Fprintf(os.Stderr, "XMLLogWriter.intRotate: Cannot find free log number to rename %s\n", xlw.filename)
			}
		}
	}

	// Open the log file
	fd, err := os.OpenFile(xlw.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return
	}
	xlw.file = fd

	// Write the closing tag
	xlw.file.Write([]byte("<log created=\"" + time.LocalTime().Format("2006/01/02 15:04:05 MST") + "\">\n"))

	// Set the daily open date to the current date
	xlw.daily_opendate = time.LocalTime().Day

	// initialize rotation values
	xlw.maxrecords_currecords = 0
	xlw.maxsize_cursize = 0
}

// Set rotate at linecount
func (xlw *XMLLogWriter) SetRotateRecords(maxrecords int) {
	xlw.maxrecords = maxrecords
}

// Set rotate at size
func (xlw *XMLLogWriter) SetRotateSize(maxsize int) {
	xlw.maxsize = maxsize
}

// Set rotate daily
func (xlw *XMLLogWriter) SetRotateDaily(daily bool) {
	xlw.daily = daily
}

// Set keep old
func (xlw *XMLLogWriter) SetRotate(rotate bool) {
	xlw.rotate = rotate
}
