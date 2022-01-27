package server

import (
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/tianhongw/grp/conf"
	"github.com/tianhongw/grp/pkg/conn"
	"github.com/tianhongw/grp/pkg/log"
	"github.com/tianhongw/grp/pkg/message"
	"github.com/tianhongw/grp/pkg/util"
)

type Tunnel struct {
	req *message.TunnelRequest

	start time.Time

	// public url
	url string

	listener *net.TCPListener

	lg log.Logger

	ctl *Control

	isExiting int32
}

func (t *Tunnel) exit() {
	if !atomic.CompareAndSwapInt32(&t.isExiting, 0, 1) {
		return
	}

	if t.listener != nil {
		if err := t.listener.Close(); err != nil {
			t.lg.Errorf("close tunnel listener failed: %v", err)
		}
	}
}

func (t *Tunnel) handlePublicConn(pubConn conn.IConn) {
	defer pubConn.Close()

	proxyConn, err := t.ctl.getProxy()
	if err != nil {
		t.lg.Errorf("get proxy failed: %v", err)
		return
	}

	startProxyReq := &message.ProxyStart{
		URL:        t.url,
		ClientAddr: pubConn.RemoteAddr().String(),
	}

	if err := message.WriteMsg(proxyConn, startProxyReq); err != nil {
		t.lg.Errorf("write start proxy request failed: %v", err)
		return
	}

	proxyConn.SetDeadline(time.Time{})

	conn.Join(pubConn, proxyConn)
}

func NewTunnel(req *message.TunnelRequest, ctl *Control, cfg conf.Config) (*Tunnel, error) {
	tunnel := &Tunnel{
		req:   req,
		start: time.Now(),
		ctl:   ctl,
		lg:    ctl.lg,
	}

	proto := tunnel.req.Protocal

	switch proto {
	case "http", "https":
		l, ok := gListeners[proto]
		if !ok {
			return nil, fmt.Errorf("not listening for %s connections", proto)
		}
		if err := registerVHost(tunnel, cfg.Server.Domain, proto, l.Addr.(*net.TCPAddr).Port); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("protocol: %s not supported yet", proto)
	}

	return tunnel, nil
}

func registerVHost(t *Tunnel, domain, proto string, port int) error {
	vhost := strings.ToLower(fmt.Sprintf("%s:%d", domain, port))

	hostName := strings.ToLower(strings.TrimSpace(t.req.HostName))
	if hostName != "" {
		t.url = fmt.Sprintf("%s://%s", proto, hostName)
		return gTunnelRegistry.Register(t, t.url)
	}

	subDomain := strings.ToLower(strings.TrimSpace(t.req.SubDomain))
	if subDomain != "" {
		t.url = fmt.Sprintf("%s://%s.%s", proto, subDomain, vhost)
		return gTunnelRegistry.Register(t, t.url)
	}

	t.url = fmt.Sprintf("%s://%d.%s", proto, util.NewIntID(), vhost)
	return gTunnelRegistry.Register(t, t.url)
}
