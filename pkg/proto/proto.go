package proto

import "github.com/tianhongw/grp/pkg/conn"

type Protocol interface {
	GetName() string
	WrapConn(conn.IConn, interface{}) conn.IConn
}

type Tcp struct{}

func NewTcp() *Tcp {
	return new(Tcp)
}

func (t *Tcp) GetName() string {
	return "tcp"
}

func (t *Tcp) WrapConn(c conn.IConn, ctx interface{}) conn.IConn {
	return c
}
