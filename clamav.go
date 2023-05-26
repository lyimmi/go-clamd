package clamd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

const (
	socketTypeTCP  string = "tcp"
	socketTypeUnix string = "unix"
)

var (
	cmdPING    = []byte("PING")
	cmdVERSION = []byte("VERSION")
	cmdRELOAD  = []byte("RELOAD")
	//cmdSHOUTDOWN = []byte("SHUTDOWN")
	cmdSTREAM    = []byte("zINSTREAM\\0")
	cmdSCAN      = []byte("SCAN ")
	cmdCONTSCAN  = []byte("CONTSCAN ")
	resOK        = []byte("OK\n")
	resPONG      = []byte("PONG\n")
	resRELOADING = []byte("RELOADING\n")
)

var (
	ErrFilePathIsEmpty = errors.New("file path is empty")
)

func NewClamAV(opts ...Option) *ClamAV {
	const (
		defaultSocketType     = socketTypeUnix
		defaultUnixSocketName = "/var/run/clamav/clamd.ctl"
		defaultTCPHost        = "127.0.0.1"
		defaultTCPPort        = 3310
		defaultTimeout        = 30 * time.Second
	)

	c := &ClamAV{
		connType:       defaultSocketType,
		unixSocketName: defaultUnixSocketName,
		TCPHost:        defaultTCPHost,
		TCPPort:        defaultTCPPort,
		timeout:        defaultTimeout,
	}

	for _, opt := range opts {
		opt(c)
	}

	switch c.connType {
	case socketTypeTCP:
		c.connStr = fmt.Sprintf("%s:%d", c.TCPHost, c.TCPPort)
	case socketTypeUnix:
		c.connStr = defaultUnixSocketName
	}

	c.dialer = net.Dialer{
		Timeout: c.timeout,
	}
	return c
}

type ClamAV struct {
	connType       string
	connStr        string
	unixSocketName string
	TCPHost        string
	TCPPort        int
	timeout        time.Duration
	dialer         net.Dialer
}

func (c ClamAV) call(ctx context.Context, command []byte) ([]byte, error) {
	conn, err := c.dialer.DialContext(ctx, c.connType, c.connStr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write(command)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, conn)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

// Ping checks if ClamAV is up and responsive.
func (c ClamAV) Ping(ctx context.Context) (bool, error) {
	res, err := c.call(ctx, cmdPING)
	if err != nil {
		return false, err
	}

	return bytes.Equal(res, resPONG), nil
}

// Version returns ClamAV's version.
func (c ClamAV) Version(ctx context.Context) (string, error) {
	res, err := c.call(ctx, cmdVERSION)
	if err != nil {
		return "", err
	}
	return string(bytes.Trim(res, "\n")), nil
}

// Reload ClamAV virus databases.
func (c ClamAV) Reload(ctx context.Context) (bool, error) {
	res, err := c.call(ctx, cmdRELOAD)
	if err != nil {
		return false, err
	}
	return bytes.Equal(res, resRELOADING), nil
}

//	func (c ClamAV) ShoutDown(ctx context.Context) (string, error) {
//		return "", nil
//	}

// Scan a file or a directory (recursively) with archive support enabled (if not disabled in clamd.conf). A full path is required.
func (c ClamAV) Scan(ctx context.Context, src string) (bool, error) {
	if src == "" {
		return false, ErrFilePathIsEmpty
	}
	var (
		err error
		res []byte
	)
	cmd := strings.Builder{}
	_, err = cmd.Write(cmdSCAN)
	if err != nil {
		return false, err
	}
	_, err = cmd.WriteString(src)
	if err != nil {
		return false, err
	}
	res, err = c.call(ctx, []byte(cmd.String()))
	if err != nil {
		return false, err
	}

	return bytes.HasSuffix(res, resOK), nil
}

// ScanStream todo: implement reader and stream
func (c ClamAV) ScanStream(ctx context.Context) (bool, error) {

	return false, nil
}

// ScanAll Scan file or directory (recursively) with archive support enabled and don't stop the scanning when a virus is found.
func (c ClamAV) ScanAll(ctx context.Context, src string) (bool, error) {
	if src == "" {
		return false, ErrFilePathIsEmpty
	}

	var (
		err error
		res []byte
	)

	cmd := strings.Builder{}
	_, err = cmd.Write(cmdCONTSCAN)
	if err != nil {
		return false, err
	}
	_, err = cmd.WriteString(src)
	if err != nil {
		return false, err
	}
	res, err = c.call(ctx, []byte(cmd.String()))
	if err != nil {
		return false, err
	}

	return bytes.HasSuffix(res, resOK), nil
}
