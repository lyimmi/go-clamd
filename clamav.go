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
	"time"
)

const (
	DATA_CHUNK_SIZE       = 1024
	SOCKET_TYPE_TCP       = "tcp"
	SOCKET_TYPE_UNIX      = "unix"
	CMD_PING              = "PING"
	CMD_VERSION           = "VERSION"
	CMD_RELOAD            = "RELOAD"
	CMD_SHUTDOWN          = "SHUTDOWN"
	CMD_INSTREAM          = "INSTREAM"
	CMD_SCAN              = "SCAN"
	CMD_CONTSCAN          = "CONTSCAN"
	RES_OK                = "OK"
	RES_PONG              = "PONG"
	RES_SHUTDOWN          = "SHUTDOWN"
	RES_RELOADING         = "RELOADING"
	RES_NO_SUCH_FILE      = "No such file or directory. ERROR"
	RES_PERMISSION_DENIED = "Permission denied. ERROR"
)

func NewClamd(opts ...Option) *Clamd {
	const (
		defaultSocketType     = SOCKET_TYPE_UNIX
		defaultUnixSocketName = "/var/run/clamav/clamd.ctl"
		defaultTCPHost        = "127.0.0.1"
		defaultTCPPort        = 3310
		defaultTimeout        = 180 * time.Second
	)

	c := &Clamd{
		mu:             sync.Mutex{},
		connType:       defaultSocketType,
		unixSocketName: defaultUnixSocketName,
		TCPHost:        defaultTCPHost,
		TCPPort:        defaultTCPPort,
		timeout:        defaultTimeout,
		conn:           nil,
	}

	for _, opt := range opts {
		opt(c)
	}

	switch c.connType {
	case SOCKET_TYPE_TCP:
		c.connStr = fmt.Sprintf("%s:%d", c.TCPHost, c.TCPPort)
	case SOCKET_TYPE_UNIX:
		c.connStr = defaultUnixSocketName
	}

	c.dialer = net.Dialer{
		Timeout: c.timeout,
	}

	return c
}

type Clamd struct {
	mu             sync.Mutex
	connType       string
	connStr        string
	unixSocketName string
	TCPHost        string
	TCPPort        int
	timeout        time.Duration
	dialer         net.Dialer
	conn           net.Conn
}

func (c *Clamd) l() {
	c.mu.Lock()
}

func (c *Clamd) ul() {
	if c.conn != nil {
		c.conn.Close()
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

	res, err := c.writeCmdReadData(ctx, CMD_PING)
	if err != nil {
		return false, err
	}

	if res != RES_PONG {
		return false, errors.Join(ErrInvalidResponse, fmt.Errorf("%s", res))
	}

	return true, nil
}

// Version returns Clamd's version.
func (c *Clamd) Version(ctx context.Context) (string, error) {
	c.l()
	defer c.ul()

	res, err := c.writeCmdReadData(ctx, CMD_VERSION)
	if err != nil {
		return "", err
	}

	return res, nil
}

// Reload Clamd virus databases.
func (c *Clamd) Reload(ctx context.Context) (bool, error) {
	c.l()
	defer c.ul()

	res, err := c.writeCmdReadData(ctx, CMD_RELOAD)
	if err != nil {
		return false, err
	}

	if res != RES_RELOADING {
		return false, errors.Join(ErrInvalidResponse, fmt.Errorf("%s", res))
	}

	return true, nil
}

func (c *Clamd) Shutdown(ctx context.Context) (bool, error) {
	c.l()
	defer c.ul()

	_, err := c.writeCmdReadData(ctx, CMD_SHUTDOWN)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Scan a file or a directory (recursively) with archive support enabled (if not disabled in clamd.conf). A full path is required.
func (c *Clamd) Scan(ctx context.Context, src string) (bool, error) {
	c.l()
	defer c.ul()

	if src == "" {
		return false, ErrEmptySrc
	}

	res, err := c.writeCmdReadData(ctx, fmt.Sprintf("%s %s", CMD_SCAN, src))
	if err != nil {
		return false, err
	}

	if strings.HasSuffix(res, RES_OK) {
		return true, nil
	}

	if strings.HasSuffix(res, RES_NO_SUCH_FILE) {
		return false, errors.Join(ErrNoSuchFileOrDir, fmt.Errorf("%s", res))
	}
	if strings.HasSuffix(res, RES_PERMISSION_DENIED) {
		return false, errors.Join(ErrPermissionDenied, fmt.Errorf("%s", res))
	}

	return false, errors.Join(ErrUnknown, fmt.Errorf("%s", res))
}

// ScanStream todo: implement reader and stream
func (c *Clamd) ScanStream(ctx context.Context, r io.Reader) (bool, error) {
	c.l()
	defer c.ul()

	err := c.writeCmd(ctx, CMD_INSTREAM)
	if err != nil {
		return false, err
	}

	for {
		buf := make([]byte, DATA_CHUNK_SIZE)
		n, err := r.Read(buf)
		if n > 0 {
			err = c.sendData(buf[0:n])
			if err != nil {
				return false, err
			}
		}
		if err != nil {
			break
		}
	}

	_, err = c.conn.Write([]byte{0, 0, 0, 0})
	if err != nil {
		return false, err
	}

	res, err := c.readData()
	if strings.HasSuffix(res, RES_OK) {
		return true, nil
	}
	return false, nil
}

// ScanAll Scan file or directory (recursively) with archive support enabled and don't stop the scanning when a virus is found.
func (c *Clamd) ScanAll(ctx context.Context, src string) (bool, error) {
	c.l()
	defer c.ul()

	if src == "" {
		return false, ErrEmptySrc
	}

	res, err := c.writeCmdReadData(ctx, fmt.Sprintf("%s %s", CMD_CONTSCAN, src))
	if err != nil {
		return false, err
	}

	if !strings.HasSuffix(res, RES_OK) {
		return false, errors.Join(ErrInvalidResponse, fmt.Errorf("%s", res))
	}

	return true, nil
}

var (
	ErrDial             = errors.New("error while connecting to clamd")
	ErrCommandCall      = errors.New("error while calling clamd")
	ErrCommandRead      = errors.New("error while reading response from clamd")
	ErrEmptySrc         = errors.New("scan source is empty")
	ErrInvalidResponse  = errors.New("invalid response from clamd")
	ErrNoSuchFileOrDir  = errors.New("clamd can't find file or directory")
	ErrPermissionDenied = errors.New("clamd can't open file or dir, permission denied")
	ErrUnknown          = errors.New("unknown error")
)
