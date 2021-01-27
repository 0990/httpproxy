package httpproxy

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type ProxyConn struct {
	net.Conn
	proxy *server

	req *http.Request

	sessId int64
}

func newProxyConn(sessId int64, conn net.Conn, proxy *server) *ProxyConn {
	return &ProxyConn{
		sessId: sessId,
		Conn:   conn,
		proxy:  proxy,
		req:    nil,
	}
}

func (p *ProxyConn) reqCallBack() {
	if p.proxy.reqCallback != nil {
		p.proxy.reqCallback(p.req)
	}
}

func (p *ProxyConn) failCallBack() {
	if p.proxy.failCallback != nil {
		p.proxy.failCallback(p.req)
	}
}

func (p *ProxyConn) relayCallBack() {
	if p.proxy.relayCallback != nil {
		p.proxy.relayCallback(p.req)
	}
}

func (p *ProxyConn) Handle() error {
	hReader := NewHistoryReader(p.Conn)
	bReader := bufio.NewReader(hReader)
	r, err := http.ReadRequest(bReader)
	if err != nil {
		return fmt.Errorf("http.ReadRequest %w", err)
	}

	p.req = r

	p.reqCallBack()

	reqBody := hReader.History()

	targetAddr := hostport(r)

	host, _, err := net.SplitHostPort(targetAddr)
	if err != nil {
		return fmt.Errorf("net.SplitHostPort %w,hostport:%s", err, targetAddr)
	}

	//本地代理
	if p.proxy.matcherMgr.Match(host) {
		if r.Method == "CONNECT" {
			p.Logf("Accepting CONNECT to %s", targetAddr)
			p.handleConnect(targetAddr)
			return nil
		}

		p.Logf("Got request %v %v %v %v", r.URL.Path, r.Host, r.Method, r.URL.String())
		p.handleNoConnect(targetAddr, reqBody)
		return nil
	}

	//其它代理处理
	p.Logf("Relay to next proxy:%s", targetAddr)
	p.relayToNextHttpProxy(reqBody)
	return nil
}

func (p *ProxyConn) relayToNextHttpProxy(reqData []byte) {
	p.relayCallBack()

	clientConn := p.Conn

	proxyConn, err := net.DialTimeout("tcp", p.proxy.cfg.NextProxyAddr, 3*time.Second)
	if err != nil {
		p.failCallBack()
		httpError(clientConn, p, fmt.Errorf("dial %s fail,error: %w", p.proxy.cfg.NextProxyAddr, err))
		return
	}

	proxyConn.Write(reqData)
	CopyEachAndClose(proxyConn, clientConn, p)
}

func (p *ProxyConn) handleConnect(targetAddr string) {
	clientConn := p.Conn

	targetConn, err := net.DialTimeout("tcp", targetAddr, 3*time.Second)
	if err != nil {
		p.failCallBack()
		httpError(clientConn, p, fmt.Errorf("dial %s fail,error: %w", targetAddr, err))
		return
	}

	clientConn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))

	CopyEachAndClose(targetConn, clientConn, p)
}

func httpError(w io.WriteCloser, ctx *ProxyConn, err error, response ...string) {
	s := "HTTP/1.1 502 Bad Gateway\r\n\r\n"
	if len(response) > 0 {
		s = response[0]
	}
	if _, err := io.WriteString(w, s); err != nil {
		ctx.Warnf("Error responding to client: %s", err)
	}
	if err := w.Close(); err != nil {
		ctx.Warnf("Error closing client connection: %s", err)
	}

	ctx.Logf("httpError,%s", err)
}

func (p *ProxyConn) handleNoConnect(targetAddr string, reqData []byte) {
	clientConn := p.Conn
	req := p.req

	if !req.URL.IsAbs() {
		p.failCallBack()
		httpError(clientConn, p, errors.New("url not abs"), "HTTP/1.1 500 This is a proxy server. Does not respond to non-proxy requests.\r\n\r\n")
		return
	}

	targetConn, err := net.DialTimeout("tcp", targetAddr, 3*time.Second)
	if err != nil {
		p.failCallBack()
		httpError(clientConn, p, fmt.Errorf("dial %s fail,error: %w", targetAddr, err))
		return
	}

	targetConn.Write(reqData)

	CopyEachAndClose(targetConn, clientConn, p)
}

func (p *ProxyConn) printf(msg string, argv ...interface{}) {
	p.proxy.logger.Printf("[%03d] "+msg+"\n", append([]interface{}{p.sessId & 0xFF}, argv...)...)
}

func (p *ProxyConn) Logf(msg string, argv ...interface{}) {
	if p.proxy.cfg.Verbose {
		p.printf("INFO: "+msg, argv...)
	}
}

func (p *ProxyConn) Warnf(msg string, argv ...interface{}) {
	p.printf("WARN: "+msg, argv...)
}

func CopyEachAndClose(a, b net.Conn, ctx *ProxyConn) {

	aTcp, aOK := a.(*net.TCPConn)
	bTcp, bOK := b.(*net.TCPConn)
	if aOK && bOK {
		go copyAndClose(aTcp, bTcp, ctx)
		go copyAndClose(bTcp, aTcp, ctx)
	} else {
		var wg sync.WaitGroup
		wg.Add(2)
		go copyOrWarn(a, b, ctx, &wg)
		go copyOrWarn(b, a, ctx, &wg)
		wg.Wait()
		a.Close()
		b.Close()
	}
}

func copyAndClose(dst, src *net.TCPConn, ctx *ProxyConn) {
	if _, err := io.Copy(dst, src); err != nil {
		ctx.Warnf("Error copying to client: %s", err)
	}

	dst.CloseWrite()
	src.CloseRead()
}

func copyOrWarn(dst io.WriteCloser, src io.ReadCloser, ctx *ProxyConn, wg *sync.WaitGroup) {
	if _, err := io.Copy(dst, src); err != nil {
		ctx.Warnf("Error copying to client: %s", err)
	}
	wg.Done()
}
