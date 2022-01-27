package conn

import (
	"crypto/tls"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"sync"

	"github.com/tianhongw/grp/conf"
	"github.com/tianhongw/grp/pkg/log"
	"github.com/tianhongw/grp/pkg/util"

	vhost "github.com/inconshreveable/go-vhost"
)

var _ IConn = (*loggedConn)(nil)

type IConn interface {
	net.Conn
	log.Logger
	SetType(string)
}

type loggedConn struct {
	net.Conn
	log.Logger

	id  int64
	typ string
}

func (c *loggedConn) StartTLS(tlsCfg *tls.Config) {
	c.Conn = tls.Client(c.Conn, tlsCfg)
}

func (c *loggedConn) SetType(typ string) {
	c.typ = typ
}

func (c *loggedConn) Close() error {
	return c.Conn.Close()
}

type Listener struct {
	net.Addr
	Conns chan *loggedConn
}

func WrapConn(conn net.Conn, typ string) *loggedConn {
	cfg := conf.GetConfig()
	switch c := conn.(type) {
	case *vhost.HTTPConn:
		wrapped := c.Conn.(*loggedConn)
		return &loggedConn{
			Conn:   conn,
			Logger: wrapped.Logger,
			id:     wrapped.id,
			typ:    wrapped.typ,
		}
	case *loggedConn:
		return c
	case *net.TCPConn:
		id := util.NewIntID()
		lg, _ := log.NewLogger(cfg.Log.Type,
			log.WithLevel(cfg.Log.Level), log.WithPrefix(fmt.Sprint(id)))
		wrapped := &loggedConn{
			Conn:   c,
			Logger: lg,
			id:     id,
			typ:    typ,
		}
		return wrapped
	}

	panic("unsupported connection type")
}

func Listen(addr, typ string, tlsCfg *tls.Config) (*Listener, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	l := &Listener{
		Addr:  listener.Addr(),
		Conns: make(chan *loggedConn),
	}

	go func() {
		for {
			rawConn, err := listener.Accept()
			if err != nil {
				stdlog.Printf("accept new tcp connection failed: %v", err)
				continue
			}
			c := WrapConn(rawConn, typ)
			if tlsCfg != nil {
				c.Conn = tls.Server(c.Conn, tlsCfg)
			}
			c.Logger.Infof("new connection from: %s for type: %s", c.RemoteAddr(), typ)
			l.Conns <- c
		}
	}()

	return l, nil
}

func Join(c1, c2 IConn) (fromBytes int64, toBytes int64) {
	var wg sync.WaitGroup
	wg.Add(2)

	go pipe(c1, c2, &wg, &fromBytes)
	go pipe(c2, c1, &wg, &toBytes)

	wg.Wait()

	return
}

func pipe(to, from IConn, wg *sync.WaitGroup, bytesCopied *int64) {
	defer func() {
		to.Close()
		from.Close()
		wg.Done()
	}()

	var err error
	*bytesCopied, err = io.Copy(to, from)
	if err != nil {
		from.Errorf("copy bytes failed: %v", err)
	}
}

const (
	NotAuthorized = `HTTP/1.0 401 Not Authorized
WWW-Authenticate: Basic realm="ngrok"
Content-Length: 23

Authorization required
`

	NotFound = `HTTP/1.0 404 Not Found
Content-Length: %d

Tunnel %s not found
`

	BadRequest = `HTTP/1.0 400 Bad Request
Content-Length: 12

Bad Request
`
)
