package client

import (
	"github.com/tianhongw/grp/pkg/conn"
	"github.com/tianhongw/grp/pkg/message"
)

func (c *Client) proxy() {
	var (
		remoteConn conn.IConn
		err        error
	)

	if c.cfg.Client.HTTPProxy != "" {
		remoteConn, err = conn.DialHttpProxy(c.cfg.Client.HTTPProxy, c.cfg.Client.ServerAddr, "proxy", c.tlsCfg)
	} else {
		remoteConn, err = conn.Dial(c.cfg.Client.ServerAddr, "proxy", c.tlsCfg)
	}

	if err != nil {
		c.Errorf("failed to establish proxy connection: %v", err)
		return
	}

	defer remoteConn.Close()

	if err := message.WriteMsg(remoteConn, &message.ProxyReg{
		ClientId: c.id,
	}); err != nil {
		c.Errorf("write message failed: %v", err)
		return
	}

	rawMsg, err := message.ReadMsg(remoteConn)
	if err != nil {
		c.Errorf("read message failed: %v", err)
		return
	}

	startProxy, ok := rawMsg.(*message.ProxyStart)
	if !ok {
		c.Error("not start proxy message type")
		return
	}

	tunnel, ok := c.tunnels[startProxy.URL]
	if !ok {
		c.Errorf("could not find tunnel for proxy: %s", startProxy.URL)
		return
	}

	locConn, err := conn.Dial(tunnel.LocalAddr, "private", nil)
	if err != nil {
		c.Errorf("dial local address: %s failed: %v", tunnel.LocalAddr, err)
		return
	}
	defer locConn.Close()

	return
}
