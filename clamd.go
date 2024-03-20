// Package clamd is a client for ClamAV daemon over TCP or UNIX socket.
package clamd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"
)

// dataChunkSize is the chunk size for stream scan.
const dataChunkSize = 1024

// Available socket types.
const (
	socketTypeTcp  = "tcp"
	socketTypeUnix = "unix"
)

// Commands and responses.
const (
	cmdPing             = "PING"
	cmdVersion          = "VERSION"
	cmdReload           = "RELOAD"
	cmdShutdown         = "SHUTDOWN"
	cmdInstream         = "INSTREAM"
	cmdScan             = "SCAN"
	cmdContscan         = "CONTSCAN"
	resOk               = "OK"
	resFound            = "FOUND"
	resPong             = "PONG"
	resReloading        = "RELOADING"
	resNoSuchFile       = "No such file or directory. ERROR"
	resPermissionDenied = "Permission denied. ERROR"
	resCantOpenFile     = "Can't open file or directory ERROR"
)

// NewClamd returns a Clamd client with default options.
//
// Default connection is a UNIX socket on /var/run/clamav/clamd.ctl with 30 second timeout. Defaults can be changed by
// passing in Option functions.
func NewClamd(opts ...Option) *Clamd {
	const (
		defaultSocketType     = socketTypeUnix
		defaultUnixSocketName = "/var/run/clamav/clamd.ctl"
		defaultTCPHost        = "127.0.0.1"
		defaultTCPPort        = 3310
		defaultTimeout        = 60 * time.Second
	)

	c := &Clamd{
		mu:             sync.Mutex{},
		connType:       defaultSocketType,
		unixSocketName: defaultUnixSocketName,
		tcpHost:        defaultTCPHost,
		tcpPort:        defaultTCPPort,
		timeout:        defaultTimeout,
		conn:           nil,
	}

	for _, opt := range opts {
		opt(c)
	}

	switch c.connType {
	case socketTypeTcp:
		c.connStr = fmt.Sprintf("%s:%d", c.tcpHost, c.tcpPort)
	case socketTypeUnix:
		c.connStr = c.unixSocketName
	}

	c.dialer = net.Dialer{
		Timeout: c.timeout,
	}

	return c
}

// Clamd is a client for ClamAV's daemon clamd.
type Clamd struct {
	mu             sync.Mutex
	connType       string
	connStr        string
	unixSocketName string
	tcpHost        string
	tcpPort        int
	timeout        time.Duration
	dialer         net.Dialer
	conn           net.Conn
}

func (c *Clamd) l() {
	c.mu.Lock()
}

func (c *Clamd) ul() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.mu.Unlock()
}

func (c *Clamd) writeCmd(ctx context.Context, command string) error {
	var err error
	c.conn, err = c.dialer.DialContext(ctx, c.connType, c.connStr)
	if err != nil {
		return errors.Join(ErrDial, err)
	}

	_, err = c.conn.Write([]byte(fmt.Sprintf("n%s\n", command)))
	if err != nil {
		return errors.Join(ErrCommandCall, err)
	}

	return err
}

func (c *Clamd) readData() (string, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, c.conn)
	if err != nil {
		return "", errors.Join(ErrCommandRead, err)
	}

	return strings.TrimSuffix(buf.String(), "\n"), err
}

func (c *Clamd) writeCmdReadData(ctx context.Context, command string) (res string, err error) {
	err = c.writeCmd(ctx, command)
	if err != nil {
		return "", errors.Join(ErrCommandCall, err)
	}

	res, err = c.readData()
	if err != nil {
		return "", errors.Join(ErrCommandCall, err)
	}

	return res, err
}

// sendData sends a stream of data to clamd
//
// [dutchcoders/go-clamd]: https://github.com/dutchcoders/go-clamd
func (c *Clamd) sendData(data []byte) error {
	var buf [4]byte
	lenData := len(data)
	buf[0] = byte(lenData >> 24)
	buf[1] = byte(lenData >> 16)
	buf[2] = byte(lenData >> 8)
	buf[3] = byte(lenData >> 0)

	a := buf

	b := make([]byte, len(a))
	for i := range a {
		b[i] = a[i]
	}

	_, err := c.conn.Write(b)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(data)
	return err
}

// Ping checks if Clamd is up and responsive.
func (c *Clamd) Ping(ctx context.Context) (bool, error) {
	c.l()
	defer c.ul()

	res, err := c.writeCmdReadData(ctx, cmdPing)
	if err != nil {
		return false, err
	}

	if res != resPong {
		return false, errors.Join(ErrInvalidResponse, fmt.Errorf("%s", res))
	}

	return true, nil
}

// Version returns Clamd's version.
func (c *Clamd) Version(ctx context.Context) (string, error) {
	c.l()
	defer c.ul()

	res, err := c.writeCmdReadData(ctx, cmdVersion)
	if err != nil {
		return "", err
	}

	return res, nil
}

// Reload Clamd virus databases.
func (c *Clamd) Reload(ctx context.Context) (bool, error) {
	c.l()
	defer c.ul()

	res, err := c.writeCmdReadData(ctx, cmdReload)
	if err != nil {
		return false, err
	}

	if res != resReloading {
		return false, errors.Join(ErrInvalidResponse, fmt.Errorf("%s", res))
	}

	return true, nil
}

// Shutdown stops Clamd cleanly.
func (c *Clamd) Shutdown(ctx context.Context) (bool, error) {
	c.l()
	defer c.ul()

	_, err := c.writeCmdReadData(ctx, cmdShutdown)
	if err != nil {
		return false, err
	}
	return true, nil
}

func parseErr(res string, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	if strings.HasSuffix(res, resOk) {
		return true, nil
	}
	if strings.HasSuffix(res, resFound) {
		return false, nil
	}
	if strings.HasSuffix(res, resNoSuchFile) {
		return false, errors.Join(ErrNoSuchFileOrDir, fmt.Errorf("%s", res))
	}
	if strings.HasSuffix(res, resPermissionDenied) {
		return false, errors.Join(ErrPermissionDenied, fmt.Errorf("%s", res))
	}
	if strings.HasSuffix(res, resCantOpenFile) {
		return false, errors.Join(ErrCantOpenFile, fmt.Errorf("%s", res))
	}

	return false, errors.Join(ErrUnknown, fmt.Errorf("%s", res))
}

// Scan a file or a directory (recursively) with archive support enabled (if not disabled in clamd.conf). A full path is required.
func (c *Clamd) Scan(ctx context.Context, src string) (bool, error) {
	c.l()
	defer c.ul()

	if src == "" {
		return false, ErrEmptySrc
	}

	res, err := c.writeCmdReadData(ctx, fmt.Sprintf("%s %s", cmdScan, src))

	return parseErr(res, err)
}

// ScanStream Scan a stream of data.This avoids the overhead of establishing new TCP connections and problems with NAT.
//
// Note: do not exceed StreamMaxLength as defined in clamd.conf, otherwise clamd will reply with INSTREAM size limit
// exceeded and close the connection. (default is 25M)
func (c *Clamd) ScanStream(ctx context.Context, r io.Reader) (bool, error) {
	c.l()
	defer c.ul()

	err := c.writeCmd(ctx, cmdInstream)
	if err != nil {
		return false, err
	}

	for {
		buf := make([]byte, dataChunkSize)
		n, err := r.Read(buf)
		if n > 0 {
			err = c.sendData(buf[0:n])
			if err != nil {
				if errors.Is(err, io.ErrClosedPipe) {
					return false, errors.Join(ErrSreamLimitExceeded, err)
				}
				if errors.Is(err, net.ErrClosed) {
					return false, errors.Join(ErrSreamLimitExceeded, err)
				}
				if errors.Is(err, syscall.EPIPE) {
					return false, errors.Join(ErrSreamLimitExceeded, err)
				}
				return false, err
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return false, err
		}
	}

	_, err = c.conn.Write([]byte{0, 0, 0, 0})
	if err != nil {
		return false, errors.Join(ErrUnknown, err)
	}

	res, err := c.readData()

	return parseErr(res, err)
}

// ScanAll Scan file or directory (recursively) with archive support enabled and don't stop the scanning when a virus is found.
func (c *Clamd) ScanAll(ctx context.Context, src string) (bool, error) {
	c.l()
	defer c.ul()

	if src == "" {
		return false, ErrEmptySrc
	}

	res, err := c.writeCmdReadData(ctx, fmt.Sprintf("%s %s", cmdContscan, src))

	return parseErr(res, err)
}

// Stats Replies with statistics about the scan queue, contents of scan queue, and memory usage.
func (c *Clamd) Stats(ctx context.Context) (*Stats, error) {
	c.l()
	defer c.ul()

	res, err := c.writeCmdReadData(ctx, "STATS")
	if err != nil {
		return nil, err
	}

	stats, err := parseStats(res)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
