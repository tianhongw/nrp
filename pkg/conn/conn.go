package conn

import "net"

type Conn interface {
	net.Conn
}
