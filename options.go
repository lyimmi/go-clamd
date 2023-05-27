package clamd

import "time"

// Option sets an option to Clamd.
type Option func(*Clamd)

// WithDefaultTCP sets up the client to use the default TCP connection on 127.0.0.1:3310
func WithDefaultTCP() Option {
	return func(c *Clamd) {
		c.connType = socketTypeTcp
	}
}

// WithTCP sets up the client with a custom TCP connection.
func WithTCP(host string, port int) Option {
	return func(c *Clamd) {
		c.connType = socketTypeTcp
		c.tcpHost = host
		c.tcpPort = port
	}
}

// WithUnix sets up the client to use a custom unix socket.
func WithUnix(socket string) Option {
	return func(c *Clamd) {
		c.connType = socketTypeUnix
		c.unixSocketName = socket
	}
}

// WithTimeout sets a timeout on the connection.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Clamd) {
		c.connType = socketTypeUnix
		c.timeout = timeout
	}
}
