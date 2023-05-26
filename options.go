package clamav

type Option func(*ClamAV)

func WithTCP() Option {
	return func(c *ClamAV) {
		c.connType = socketTypeTCP
	}
}

func WithCustomTCP(host string, port int) Option {
	return func(c *ClamAV) {
		c.connType = socketTypeTCP
		c.TCPHost = host
		c.TCPPort = port
	}
}

func WithUnix() Option {
	return func(c *ClamAV) {
		c.connType = socketTypeUnix
	}
}

func WithCustomUnix(name string) Option {
	return func(c *ClamAV) {
		c.connType = socketTypeUnix
		c.unixSocketName = name
	}
}
