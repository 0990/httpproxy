package httpproxy

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpConn struct {
	net.Conn
	gfwList         *gfwlist.GFWList
	remoteProxyPool *ConnPool
	req             *http.Request
	reqBody         []byte
}

func NewHttpConnProxy(conn net.Conn) *HttpConn {
	return &HttpConn{
		Conn:            conn,
		gfwList:         nil,
		remoteProxyPool: nil,
		r:               nil,
		rBody:           nil,
	}
}

func (p *HttpConn) Handle() error {
	hReader := NewHistoryReader(p.Conn)
	bReader := bufio.NewReader(hReader)
	r, err := http.ReadRequest(bReader)
	if err != nil {
		return err
	}

	p.req = r
	p.reqBody = hReader.History()

	if r.Method == "CONNECT" {
		return p.handleConnect()
	} else {
		return p.handleNoConnect()
	}
}

func (p *HttpConn) handleConnect() error {

}

func (p *HttpConn) handleNoConnect() error {
	req := p.req
	if !p.req.URL.IsAbs() {
		return nil
	}

	if p.isLocalDstHost(p.req.Host) {
		resp, err := http.DefaultTransport.RoundTrip(p.req)
		if err != nil {
			return err
		}
		return nil
	}

}

func (p *HttpConn) isLocalDstHost(host string) bool {
	return true
}

func (p *HttpConn) Handle() error {

	conn := p.Conn
	defer conn.Close()
	var b [1024]byte
	n, err := conn.Read(b[:])
	if err != nil {
		return
	}
	var method, host, remoteAddr string
	c := bytes.IndexByte(b[:], '\n')
	if c == -1 {
		return
	}
	fmt.Sscanf(string(b[:c]), "%s%s", &method, &host)

	hostPortURL, err := url.Parse(host)
	if err != nil {
		//"192.168.1.1:8080"这种也解析出错
		if !strings.Contains(host, "//") {
			hostPortURL, err = url.Parse(fmt.Sprintf("http://%s", host))
			if err != nil {
				logrus.WithField("host", host).WithError(err).Error("url.Parse")
				return
			}
		} else {
			logrus.WithField("host", host).WithError(err).Error("url.Parse")
			return
		}
	}

	if hostPortURL.Opaque == "443" { //https访问
		remoteAddr = hostPortURL.Scheme + ":443"
	} else { //http访问
		if strings.Index(hostPortURL.Host, ":") == -1 { //host不带端口， 默认80
			remoteAddr = hostPortURL.Host + ":80"
		} else {
			remoteAddr = hostPortURL.Host
		}
	}

	s := strings.Split(remoteAddr, ":")
	useRemote := true
	if p.pac && !p.gfwList.IsDomainBlocked(s[0]) {
		useRemote = false
	}
	if useRemote {
		logrus.WithField("remoteAddr", remoteAddr).Info("HttpConn remote proxy")
		s, err := p.remoteProxyPool.GetConn()
		if err != nil {
			logrus.WithError(err).Error("remoteConnPool.GetConn")
			conn.Write([]byte("HTTP/1.1 404 Not found\r\n\r\n"))
			return
		}
		defer s.Close()
		frame := make([]byte, 3, 100)
		frame[0] = p.headerTag
		binary.BigEndian.PutUint16(frame[1:], uint16(len(remoteAddr)))
		frame = append(frame, []byte(remoteAddr)...)
		msg.EncodeWrite(s, frame)

		s.SetReadDeadline(time.Now().Add(time.Second * 3))
		ret, err := msg.DecodeRead(s)
		s.SetReadDeadline(time.Time{})

		if err != nil || len(ret) != 1 {
			return
		}
		if ret[0] != 0x00 {
			conn.Write([]byte("HTTP/1.1 404 Not found\r\n\r\n"))
			return
		}

		if method == "CONNECT" {
			conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		} else {
			msg.EncodeWrite(s, b[0:n])
		}

		go func() {
			err := msg.DecodeCopy(conn, s)
			if err != nil {
				conn.Close()
				s.Close()
			}
		}()
		msg.EncodeCopy(s, conn)
	} else {
		logrus.WithField("remoteAddr", remoteAddr).Info("HttpConn local proxy")
		s, err := net.DialTimeout("tcp", remoteAddr, time.Second)
		if err != nil {
			conn.Write([]byte("HTTP/1.1 404 Not found\r\n\r\n"))
			return
		}
		defer s.Close()
		if method == "CONNECT" {
			conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		} else {
			s.Write(b[0:n])
		}
		go func() {
			io.Copy(conn, s)
		}()
		io.Copy(s, conn)
	}
}
