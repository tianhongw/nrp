package server

import (
	"github.com/tianhongw/grp/pkg/conn"
	"github.com/tianhongw/grp/pkg/message"
)

func newProxy(conn conn.IConn, req *message.ProxyReg) {
	conn.Infof("new proxy for client: %s", req.ClientId)
	ctl := gControlRegistry.Get(req.ClientId)
	if ctl == nil {
		panic("no control find for client: " + req.ClientId)
	}

	ctl.registerProxy(conn)
}
