package server

import "net"

type ConnHandler interface {
	Handle(conn net.Conn)
}
