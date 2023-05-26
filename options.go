package clamd

type Option func(*Clamd)

func WithDefaultTCP() Option {
	return func(c *Clamd) {
		c.connType = SOCKET_TYPE_TCP
	}
}

func WithTCP(host string, port int) Option {
	return func(c *Clamd) {
		c.connType = SOCKET_TYPE_TCP
		c.TCPHost = host
		c.TCPPort = port
	}
}

func WithUnix(name string) Option {
	return func(c *Clamd) {
		c.connType = SOCKET_TYPE_UNIX
		c.unixSocketName = name
	}
}
