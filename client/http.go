package client

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/tianhongw/grp/pkg/conn"
	"github.com/tianhongw/grp/pkg/util"
)

type HttpRequest struct {
	*http.Request
	BodyBytes []byte
}

type HttpResponse struct {
	*http.Response
	BodyBytes []byte
}

type HttpTxn struct {
	Req         *HttpRequest
	Resp        *HttpResponse
	Start       time.Time
	Duration    time.Duration
	UserCtx     interface{}
	ConnUserCtx interface{}
}

type HttpWrapper struct {
	Txns util.Broadcast
}

func NewHttpWrapper() *HttpWrapper {
	return &HttpWrapper{
		Txns: *util.NewBroadcast(),
	}
}

func (h *HttpWrapper) wrapConn(c conn.IConn, ctx interface{}) conn.IConn {
	tee := conn.NewTee(c)
	lastTxn := make(chan *HttpTxn)

	go h.readRequest(tee, lastTxn, ctx)
	go h.readResponses(tee, lastTxn)

	return tee
}

func (h *HttpWrapper) readRequest(tee *conn.Tee, lastTxn chan *HttpTxn, connCtx interface{}) {
	defer close(lastTxn)

	for {
		req, err := http.ReadRequest(tee.WriteBuffer())
		if err != nil {
			log.Printf("read request failed: %v", err)
			break
		}

		_, err = httputil.DumpRequest(req, true)
		if err != nil {
			log.Printf("dump request failed: %v", err)
		}

		req.URL.Scheme = "http"
		req.URL.Host = req.Host

		txn := &HttpTxn{
			Start:       time.Now(),
			ConnUserCtx: connCtx,
			Req:         &HttpRequest{Request: req},
		}

		if req.Body != nil {
			txn.Req.BodyBytes, txn.Req.Body, err = extractBody(req.Body)
			if err != nil {
				log.Printf("extract request body failed: %v", err)
			}
		}

		lastTxn <- txn

		h.Txns.In() <- txn
	}
}

func (h *HttpWrapper) readResponses(tee *conn.Tee, lastTxn chan *HttpTxn) {
	for txn := range lastTxn {
		resp, err := http.ReadResponse(tee.ReadBuffer(), txn.Req.Request)
		txn.Duration = time.Since(txn.Start)
		if err != nil {
			log.Printf("error reading response from server: %v", err)
			// no more responses to be read, we're done
			break
		}
		// make sure we read the body of the response so that
		// we don't block the reader
		_, _ = httputil.DumpResponse(resp, true)

		txn.Resp = &HttpResponse{Response: resp}
		// apparently, Body can be nil in some cases
		if resp.Body != nil {
			txn.Resp.BodyBytes, txn.Resp.Body, err = extractBody(resp.Body)
			if err != nil {
				log.Printf("failed to extract response body: %v", err)
			}
		}

		h.Txns.In() <- txn

		// XXX: remove web socket shim in favor of a real websocket protocol analyzer
		if txn.Req.Header.Get("Upgrade") == "websocket" {
			tee.Info("Upgrading to websocket")
			var wg sync.WaitGroup

			// shim for websockets
			// in order for websockets to work, we need to continue reading all of the
			// the bytes in the analyzer so that the joined connections will continue
			// sending bytes to each other
			wg.Add(2)
			go func() {
				ioutil.ReadAll(tee.WriteBuffer())
				wg.Done()
			}()

			go func() {
				ioutil.ReadAll(tee.ReadBuffer())
				wg.Done()
			}()

			wg.Wait()
			break
		}
	}
}

func extractBody(r io.Reader) ([]byte, io.ReadCloser, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r)
	return buf.Bytes(), ioutil.NopCloser(buf), err
}
