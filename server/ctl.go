package server

import (
	"time"

	"github.com/tianhongw/grp/pkg/conn"
	"github.com/tianhongw/grp/pkg/message"
	"github.com/tianhongw/grp/pkg/util"
)

type Control struct {
	id string

	auth *message.AuthRequest

	// actual connection
	conn conn.Conn

	// put msg in this channel to sent it to the client
	out chan (message.Message)

	// read from the channel to get msg from client
	in chan (message.Message)

	lastPing time.Time

	proxies chan conn.Conn

	exitChan         chan struct{}
	waitGroupWrapper util.WaitGroupWrapper

	isExiting int32
}
