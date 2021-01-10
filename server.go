package httpproxy

import (
	"fmt"
	"log"
	"net"
	"os"
)

type Server interface {
	ListenAndServe() error
}

type server struct {
	cfg        Config
	matcherMgr *matcherManger
	logger     Logger
	sess       int64
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
