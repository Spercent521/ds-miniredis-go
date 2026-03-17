package server

import (
	"bufio"
	"errors"
	"io"
	"net"

	"goredis/internal/command"
	"goredis/internal/config"
	"goredis/internal/pool"
	"goredis/internal/protocol"
)

type Server struct {
	cfg        config.Config
	listener   net.Listener
	parser     protocol.Parser
	dispatcher *command.Dispatcher
	pool       *pool.WorkerPool
}

// New creates a server with a worker pool of cfg.Workers goroutines.
// Each accepted connection is dispatched as a task to the pool.
func New(cfg config.Config, parser protocol.Parser, dispatcher *command.Dispatcher) *Server {
	return &Server{
		cfg:        cfg,
		parser:     parser,
		dispatcher: dispatcher,
		pool:       pool.New(cfg.Workers),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return err
	}
	s.listener = ln

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		// Submit connection handling to the worker pool.
		// In Go 1.22+ the loop variable 'conn' is correctly scoped per iteration.
		s.pool.Submit(func() { s.handleConn(conn) })
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		argv, err := s.parser.ParseArrayString(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			_, _ = conn.Write([]byte(protocol.EncodeError(err.Error())))
			return
		}

		resp, err := s.dispatcher.Dispatch(argv)
		if err != nil {
			if errors.Is(err, command.ErrUnknownCommand) && len(argv) > 0 {
				_, _ = conn.Write([]byte(protocol.EncodeError("unknown command '" + argv[0] + "'")))
				continue
			}
			_, _ = conn.Write([]byte(protocol.EncodeError(err.Error())))
			continue
		}

		if _, err := conn.Write([]byte(resp)); err != nil {
			return
		}
	}
}
