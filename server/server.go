package server

import (
	"crypto/tls"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tianhongw/grp/conf"
	"github.com/tianhongw/grp/pkg/conn"
	"github.com/tianhongw/grp/pkg/log"
	"github.com/tianhongw/grp/pkg/message"
)

var (
	gListeners       map[string]*conn.Listener
	gTunnelRegistry  *TunnelRegistry
	gControlRegistry *ControlRegistry
)

const (
	defaultConnReadTimeoutSec  = 10
	defaultConnWriteTimeoutSec = 10
	defaultProxyConnTimeout    = 3 * time.Minute
)

type Server struct {
	cfg *conf.Config

	log.Logger

	wg sync.WaitGroup

	exitChan chan struct{}

	isExiting int32
}

func NewServer(cfg *conf.Config) *Server {
	srv := &Server{
		cfg:      cfg,
		exitChan: make(chan struct{}),
	}

	lg, err := log.NewLogger(cfg.Log.Type,
		log.WithLevel(cfg.Log.Level),
		log.WithPrefix("server"))
	if err != nil {
		panic(err)
	}

	srv.Logger = lg

	return srv
}

func (s *Server) Run() error {
	gTunnelRegistry = newTunnelRegistry(s.cfg)

	gControlRegistry = newControlRegistry(s.cfg)

	gListeners = make(map[string]*conn.Listener)

	if s.cfg.Server.HTTPAddr != "" {
		httpListener, err := startHttpListener(s.cfg.Server.HTTPAddr, nil)
		if err != nil {
			return err
		}
		s.Infof("http listening on: %s", httpListener.Addr)
		gListeners["http"] = httpListener
	}

	if err := s.tunnelListener(s.cfg.Server.TunnelPort, nil); err != nil {
		s.Errorf("start tunnel listener failed: %v", err)
		return err
	}

	return nil
}

func (s *Server) Exit() error {
	if !atomic.CompareAndSwapInt32(&s.isExiting, 0, 1) {
		return nil
	}

	s.Info("exiting server")

	gControlRegistry.exit()

	close(s.exitChan)

	s.wg.Wait()

	s.Info("server exiting down")

	return nil
}

func (s *Server) tunnelListener(addr string, tlsConfig *tls.Config) error {
	s.wg.Add(1)
	defer s.wg.Done()

	listener, err := conn.Listen(addr, "tunnel", tlsConfig)
	if err != nil {
		return err
	}

	s.Infof("tunnel listening on %s", listener.Addr)

	for {
		select {
		case conn := <-listener.Conns:
			go s.tunnelHandler(conn)
		case <-s.exitChan:
			return nil
		}
	}
}

func (s *Server) tunnelHandler(conn conn.IConn) {
	readTimeout := s.cfg.Server.ConnReadTimeoutSec
	if readTimeout == 0 {
		readTimeout = defaultConnReadTimeoutSec
	}

	conn.SetReadDeadline(time.Now().Add(time.Duration(readTimeout) * time.Second))
	rawMsg, err := message.ReadMsg(conn)
	if err != nil {
		conn.Close()
		return
	}

	conn.SetReadDeadline(time.Time{})

	switch m := rawMsg.(type) {
	case *message.AuthRequest:
		newControl(s.cfg, conn, m)
	case *message.ProxyReg:
		newProxy(conn, m)
	default:
		conn.Close()
	}
}
