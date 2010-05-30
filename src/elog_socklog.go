package elog

import (
	"os"
	"fmt"
	"net"
	"json"
)

const (
	EOM = "\n"
)

// This log writer sends output to a socket
type SocketLogWriter struct {
	sock net.Conn
}

// This is the SocketLogWriter's output method
func (slw *SocketLogWriter) LogWrite(rec *LogRecord) (int, os.Error) {
	if !slw.Good() {
		return -1, os.NewError("Socket was not opened successfully")
	}

	// Marshall into JSON
	js, err := json.Marshal(rec)
	if err != nil { return 0, err }

	// Write to socket
	n, err := slw.sock.Write(js)
	if err != nil {
		return n, err
	}

	// Send End Of Message
	m, err := slw.sock.Write([]byte(EOM))

	return n+m, err
}

func (slw *SocketLogWriter) Good() bool {
	return slw.sock != nil
}

func (slw *SocketLogWriter) Close() {
	slw.sock.Close()
	slw.sock = nil
}

func NewSocketLogWriter(proto, hostport string) *SocketLogWriter {
	s, err := net.Dial(proto, "", hostport)
	slw := new(SocketLogWriter)

	if err != nil || s == nil {
		fmt.Fprintf(os.Stderr, "NewSocketLogWriter: %s\n", err)
	}

	slw.sock = s
	return slw
}
