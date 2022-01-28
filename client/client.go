package client

import (
	"crypto/tls"
	"errors"
	"strings"
	"sync/atomic"
	"time"

	"github.com/tianhongw/grp/conf"
	"github.com/tianhongw/grp/pkg/conn"
	"github.com/tianhongw/grp/pkg/log"
	"github.com/tianhongw/grp/pkg/message"
	"github.com/tianhongw/grp/pkg/proto"
	"github.com/tianhongw/grp/pkg/util"
)

type tunnel struct {
	PublicUrl string
	LocalAddr string
	Protocol  string
}

type Client struct {
	// client id
	id string

	ctlConn conn.IConn

	log.Logger

	tunnels map[string]*tunnel

	protocols []proto.Protocol

	cfg *conf.Config

	tlsCfg *tls.Config

	waitGroup util.WaitGroupWrapper

	isExiting int32

	exitChan chan struct{}

	lastPing time.Time
	lastPong atomic.Value
}

func NewClient(cfg *conf.Config) *Client {
	c := &Client{
		cfg:      cfg,
		exitChan: make(chan struct{}),
		lastPing: time.Now(),
		tunnels:  map[string]*tunnel{},
	}

	lg, _ := log.NewLogger(cfg.Log.Type,
		log.WithLevel(cfg.Log.Level), log.WithPrefix("client "))

	c.Logger = lg

	return c
}

const (
	maxFailCount = 30
	maxWaitTime  = 1 * time.Minute
)

func (c *Client) Run() error {
	wait := 1 * time.Second
	failCount := 0

	for {
		err := c.loop()

		failCount++
		if failCount > maxFailCount {
			return err
		}

		wait = 2 * wait
		if wait > maxWaitTime {
			wait = maxWaitTime
		}

		time.Sleep(wait)
	}

}

func (c *Client) loop() error {
	var (
		ctlConn conn.IConn
		err     error
	)

	if c.cfg.Client.HTTPProxy != "" {
		ctlConn, err = conn.DialHttpProxy(c.cfg.Client.HTTPProxy, c.cfg.Client.ServerAddr, "control", c.tlsCfg)
	} else {
		ctlConn, err = conn.Dial(c.cfg.Client.ServerAddr, "control", c.tlsCfg)
	}

	if err != nil {
		return err
	}

	c.ctlConn = ctlConn

	defer ctlConn.Close()

	authReq := &message.AuthRequest{
		ClientId: c.id,
		User:     c.cfg.Client.AuthToken,
	}

	if err := message.WriteMsg(ctlConn, authReq); err != nil {
		return err
	}

	msg, err := message.ReadMsg(ctlConn)
	if err != nil {
		return err
	}

	authResp, ok := msg.(*message.AuthResponse)
	if !ok {
		return errors.New("not auth response")
	}

	if authResp.ErrorMsg != "" {
		c.Error(authResp.ErrorMsg)
		return errors.New(authResp.ErrorMsg)
	}

	c.id = authResp.ClientId

	c.Infof("client: %s successfully connect to server, control conn established at: %v",
		c.id, ctlConn.LocalAddr())

	// request tunnels
	reqIdToTunnelCfg := make(map[string]*conf.TunnelOption)
	for _, cfg := range c.cfg.Client.Tunnels {
		var protocols []string
		for proto := range cfg.Protocols {
			protocols = append(protocols, proto)
		}

		tunnelRequest := &message.TunnelRequest{
			RequestId:  util.NewStringID(),
			Protocol:   strings.Join(protocols, ","),
			HostName:   cfg.HostName,
			SubDomain:  cfg.SubDomain,
			HttpAuth:   cfg.HttpAuth,
			RemotePort: cfg.RemotePort,
		}

		if err := message.WriteMsg(ctlConn, tunnelRequest); err != nil {
			return err
		}

		reqIdToTunnelCfg[tunnelRequest.RequestId] = cfg
	}

	c.waitGroup.Wrap(c.heartbeat)

	for {
		select {
		case <-c.exitChan:
			return errors.New("client exited")
		default:
		}
		rawMsg, err := message.ReadMsg(ctlConn)
		if err != nil {
			return err
		}

		switch m := rawMsg.(type) {
		case *message.Pong:
			c.lastPong.Store(time.Now())
		case *message.TunnelResponse:
			if m.ErrorMsg != "" {
				c.Errorf("new tunnel failed: %v", err)
				continue
			}
			t := &tunnel{
				PublicUrl: m.URL,
				LocalAddr: reqIdToTunnelCfg[m.RequestId].Protocols[m.Protocol],
				Protocol:  m.Protocol,
			}
			c.tunnels[t.PublicUrl] = t
			c.Infof("tunnel established, public url: %s, local addr: %s",
				t.PublicUrl, t.LocalAddr)
		case *message.ProxyRequest:
			c.waitGroup.Wrap(c.proxy)
		}
	}
}

const (
	defaultPingInterval      = 3 * time.Second
	defaultPongCheckInterval = 10 * time.Second
)

func (c *Client) heartbeat() {
	conn := c.ctlConn
	ping := time.NewTicker(defaultPingInterval)
	pongCheck := time.NewTicker(defaultPongCheckInterval)

	defer func() {
		ping.Stop()
		pongCheck.Stop()
	}()

	for {
		select {
		case <-ping.C:
			if err := message.WriteMsg(conn, &message.Ping{}); err != nil {
				c.Errorf("client write ping message failed: %v", err)
				go c.Exit()
				return
			}
			c.lastPing = time.Now()
		case <-pongCheck.C:
			lastPong := c.lastPong.Load().(time.Time)
			if c.lastPing.Sub(lastPong) > 2*defaultPingInterval {
				c.Errorf("client have not recived ping message from server side, last ping at: %v", c.lastPing)
				go c.Exit()
				return
			}
		case <-c.exitChan:
			return
		}
	}
}

func (c *Client) Exit() error {
	if !atomic.CompareAndSwapInt32(&c.isExiting, 0, 1) {
		return nil
	}

	close(c.exitChan)

	c.waitGroup.Wait()

	if err := c.ctlConn.Close(); err != nil {
		c.Errorf("close control conn failed: %v", err)
	}

	c.Infof("client: %s exit success", c.id)

	return nil
}
