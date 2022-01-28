package server

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"

	"github.com/tianhongw/grp/conf"
	"github.com/tianhongw/grp/pkg/conn"
	"github.com/tianhongw/grp/pkg/log"
	"github.com/tianhongw/grp/pkg/message"
	"github.com/tianhongw/grp/pkg/util"
)

type Control struct {
	clientId string

	auth *message.AuthRequest

	// actual connection
	conn conn.IConn

	// put msg in this channel to sent it to the client
	out chan (message.Message)

	// read from the channel to get msg from client
	in chan (message.Message)

	lastPing time.Time

	proxies chan conn.IConn

	tunnels []*Tunnel

	exitChan  chan struct{}
	waitGroup util.WaitGroupWrapper

	lg log.Logger

	isExiting int32

	cfg *conf.Config
}

const (
	defaultProxyMaxSize = 10
)

func newControl(cfg *conf.Config, ctlConn conn.IConn,
	authReq *message.AuthRequest) *Control {

	c := &Control{
		auth:     authReq,
		clientId: authReq.ClientId,
		out:      make(chan message.Message),
		in:       make(chan message.Message),
		proxies:  make(chan conn.IConn, defaultProxyMaxSize),
		tunnels:  make([]*Tunnel, 0),
		exitChan: make(chan struct{}),
		lastPing: time.Now(),
		cfg:      cfg,
	}

	if c.clientId == "" {
		c.clientId = util.NewStringID()
	}

	lg, err := log.NewLogger(cfg.Log.Type,
		log.WithLevel(cfg.Log.Level),
		log.WithPrefix(fmt.Sprintf("control:%s", c.clientId)))
	if err != nil {
		panic(err)
	}

	c.lg = lg

	if replacedCtl := gControlRegistry.Add(c.clientId, c); replacedCtl != nil {
		replacedCtl.waitGroup.Wait()
	}

	go func() {
		c.out <- &message.AuthResponse{
			ClientId: c.clientId,
		}

		// ask for a proxy connection
		c.out <- &message.ProxyRequest{}
	}()

	c.waitGroup.Wrap(c.manager)
	c.waitGroup.Wrap(c.reader)
	c.waitGroup.Wrap(c.writer)

	return c
}

func (c *Control) registerTunnel(req *message.TunnelRequest) {
	for _, proto := range strings.Split(req.Protocol, ",") {
		newReq := *req
		newReq.Protocol = proto

		c.lg.Debugf("register tunnel: %v", newReq)

		t, err := NewTunnel(&newReq, c, *c.cfg)
		if err != nil {
			c.lg.Errorf("register tunnel failed: %v", err)
			c.out <- &message.TunnelResponse{
				ErrorMsg: err.Error(),
			}
			if len(c.tunnels) == 0 {
				c.exit()
			}
			return
		}

		c.tunnels = append(c.tunnels, t)
		c.out <- &message.TunnelResponse{
			RequestId: req.RequestId,
			URL:       t.url,
			Protocol:  proto,
		}
	}
}

func (c *Control) requestProxy() {
	go func() {
		c.out <- &message.ProxyRequest{}
	}()
}

func (c *Control) registerProxy(conn conn.IConn) {
	conn.SetDeadline(time.Now().Add(defaultProxyConnTimeout))
	select {
	case c.proxies <- conn:
		c.lg.Info("proxy registerd")
	default:
		c.lg.Warning("proxy buffer is full, discarding")
		conn.Close()
	}
}

func (c *Control) getProxy() (conn.IConn, error) {
	select {
	case proxyConn, ok := <-c.proxies:
		if !ok {
			return nil, errors.New("control is exiting")
		}
		return proxyConn, nil

	default:
		go func() {
			c.out <- message.ProxyRequest{}
		}()

		select {
		case proxyConn, ok := <-c.proxies:
			if !ok {
				return nil, errors.New("control is exiting")
			}
			return proxyConn, nil
		case <-time.After(defaultPingCheckInterval):
			return nil, errors.New("get proxy connection timeout")
		}
	}
}

const (
	defaultPingCheckInterval = 30 * time.Second
)

func (c *Control) manager() {
	reap := time.NewTicker(defaultPingCheckInterval)
	defer reap.Stop()

	for {
		select {
		case rawMsg := <-c.in:
			switch mt := rawMsg.(type) {
			case *message.TunnelRequest:
				c.registerTunnel(mt)
			case *message.Ping:
				c.lastPing = time.Now()
				c.out <- &message.Pong{}
			}
		case <-reap.C:
			if time.Since(c.lastPing) > defaultPingCheckInterval {
				c.lg.Errorf("lost heartbeat, last time is : %v", c.lastPing)
				go func() { c.exit() }()
			}
		case <-c.exitChan:
			return
		}
	}
}

func (c *Control) reader() {
	timeout := c.cfg.Server.ConnReadTimeoutSec
	if timeout == 0 {
		timeout = defaultConnReadTimeoutSec
	}

	for {
		select {
		case <-c.exitChan:
			return
		default:
			c.conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
			msg, err := message.ReadMsg(c.conn)
			if err != nil {
				if err == io.EOF {
					return
				} else {
					c.lg.Errorf("read message failed: %v", err)
					go func() { c.exit() }()
				}
			}
			c.in <- msg
		}
	}
}

func (c *Control) writer() {
	timeout := c.cfg.Server.ConnWriteTimeoutSec
	if timeout == 0 {
		timeout = defaultConnWriteTimeoutSec
	}

	for {
		select {
		case msg := <-c.out:
			c.conn.SetWriteDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
			if err := message.WriteMsg(c.conn, msg); err != nil {
				c.lg.Errorf("write message failed: %v", err)
				go func() { c.exit() }()
			}
		case <-c.exitChan:
			return
		}
	}
}

func (c *Control) exit() {
	if !atomic.CompareAndSwapInt32(&c.isExiting, 0, 1) {
		c.lg.Warning("control is already start exit")
		return
	}

	gControlRegistry.Remove(c.clientId)

	close(c.exitChan)

	c.waitGroup.Wait()

	close(c.in)

	close(c.out)

	if err := c.conn.Close(); err != nil {
		c.lg.Errorf("close control connection failed: %v", err)
	}

	for _, t := range c.tunnels {
		t.exit()
	}

	close(c.proxies)
	for p := range c.proxies {
		if err := p.Close(); err != nil {
			c.lg.Errorf("close proxy connection failed: %v", err)
		}
	}

	c.lg.Info("shutdown success")
}

func (c *Control) Replace(replacement *Control) {
	c.lg.Info("control is replaced")
	c.exit()
}
