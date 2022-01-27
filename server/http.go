package server

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/inconshreveable/go-vhost"
	"github.com/tianhongw/grp/pkg/conn"
)

func startHttpListener(addr string, tlsCfg *tls.Config) (*conn.Listener, error) {
	listener, err := conn.Listen(addr, "public", tlsCfg)
	if err != nil {
		return nil, err
	}

	proto := "http"
	if tlsCfg != nil {
		proto = "https"
	}

	go func() {
		for conn := range listener.Conns {
			go httpHandle(conn, proto)
		}
	}()

	return listener, nil
}

func httpHandle(c conn.IConn, proto string) {
	defer c.Close()

	c.SetDeadline(time.Now().Add(defaultConnReadTimeoutSec * time.Second))

	vhostConn, err := vhost.HTTP(c)
	if err != nil {
		c.Errorf("bad request: %v", err)
		c.Write([]byte(conn.BadRequest))
		return
	}

	host := strings.ToLower(vhostConn.Host())

	auth := vhostConn.Request.Header.Get("Authorization")

	vhostConn.Free()

	c = conn.WrapConn(c, "public")

	tunnel := gTunnelRegistry.Get(fmt.Sprintf("%s://%s", proto, host))
	if tunnel == nil {
		c.Errorf("can not find tunnel for host: %s", host)
		c.Write([]byte(fmt.Sprintf(conn.NotFound, len(host)+8, host)))
		return
	}

	if tunnel.req.HttpAuth != "" &&
		tunnel.req.HttpAuth != auth {
		c.Error("authentication failed")
		c.Write([]byte(conn.NotAuthorized))
		return
	}

	c.SetDeadline(time.Time{})

	tunnel.handlePublicConn(c)
}
