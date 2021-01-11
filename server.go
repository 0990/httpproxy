package httpproxy

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

type Server interface {
	ListenAndServe() error
	AddCallBack(reqCB, failCB, relayCB func(r *http.Request))
}

type server struct {
	cfg        Config
	matcherMgr *matcherManger
	logger     Logger
	sess       int64

	reqCallback, failCallback, relayCallback func(r *http.Request)
}

func NewServer(cfg Config) Server {
	return &server{
		cfg:        cfg,
		logger:     log.New(os.Stderr, "", log.LstdFlags),
		matcherMgr: newMatcherManager(cfg.Hosts),
	}
}

func (s *server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.cfg.BindAddr)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		s.sess++

		proxyConn := newProxyConn(s.sess, conn, s)

		go func() {
			err := proxyConn.Handle()
			if err != nil {
				fmt.Println(err)
				proxyConn.Close()
			}
		}()
	}
}

func (p *server) AddCallBack(reqCB, failCB, relayCB func(r *http.Request)) {
	p.reqCallback = reqCB
	p.failCallback = failCB
	p.relayCallback = relayCB
}
