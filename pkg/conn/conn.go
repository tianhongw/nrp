package conn

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"net/url"
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
			log.WithLevel(cfg.Log.Level), log.WithPrefix(fmt.Sprint(id, "-")))
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

func Join(c IConn, c2 IConn) (int64, int64) {
	var wait sync.WaitGroup

	pipe := func(to IConn, from IConn, bytesCopied *int64) {
		defer to.Close()
		defer from.Close()
		defer wait.Done()

		var err error
		*bytesCopied, err = io.Copy(to, from)
		if err != nil {
			from.Errorf("Copied %d bytes failing with error %v", *bytesCopied, err)
		}
	}

	wait.Add(2)
	var fromBytes, toBytes int64
	go pipe(c, c2, &fromBytes)
	go pipe(c2, c, &toBytes)
	c.Info("Joined with connection %s", c2)
	wait.Wait()
	return fromBytes, toBytes
}

// func Join(c1, c2 IConn) (fromBytes int64, toBytes int64) {
//	var wg sync.WaitGroup
//	wg.Add(2)

//	go pipe(c1, c2, &wg, &fromBytes)
//	go pipe(c2, c1, &wg, &toBytes)

//	wg.Wait()

//	return
// }

// func pipe(to, from IConn, wg *sync.WaitGroup, bytesCopied *int64) {
//	defer func() {
//		to.Close()
//		from.Close()
//		wg.Done()
//	}()

//	var err error
//	*bytesCopied, err = io.Copy(to, from)
//	if err != nil {
//		from.Errorf("copy bytes failed: %v", err)
//	}
// }

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

func Dial(addr, typ string, tlsCfg *tls.Config) (*loggedConn, error) {
	rawConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	conn := WrapConn(rawConn, typ)
	if tlsCfg != nil {
		conn.StartTLS(tlsCfg)
	}

	return conn, nil
}

func DialHttpProxy(proxyUrl, addr, typ string, tlsCfg *tls.Config) (*loggedConn, error) {
	u, err := url.Parse(proxyUrl)
	if err != nil {
		return nil, err
	}

	var proxyAuth string
	if u.User != nil {
		proxyAuth = "Basic " + base64.StdEncoding.EncodeToString([]byte(u.User.String()))
	}

	var proxyTlsCfg *tls.Config
	switch u.Scheme {
	case "http":
		// do nothin
	case "https":
		proxyTlsCfg = &tls.Config{}
	default:
		return nil, fmt.Errorf("unsupported proxy url schema: %s", u.Scheme)
	}

	conn, err := Dial(addr, typ, proxyTlsCfg)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodConnect, "https://"+addr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; nrp)")
	if proxyAuth != "" {
		req.Header.Set("Proxy-Authorization", proxyAuth)
	}

	if err = req.Write(conn); err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non 200 status code from proxy server: %d", resp.StatusCode)
	}

	if tlsCfg != nil {
		conn.StartTLS(tlsCfg)
	}

	return conn, nil
}
