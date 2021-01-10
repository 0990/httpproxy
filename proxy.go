package httpproxy

import "net"

type Server interface {
	ListenAndServe() error
}

type server struct {
	cfg Config
}

func NewServer(cfg Config) Server {
	return &server{
		cfg: cfg,
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
	}

}
