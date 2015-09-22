package swnet

import (
	"net"
)

// Server is tcp server wrapper
type Server struct {
	listener       net.Listener
	sendChanSize   int
	packetHandler  PacketHandler
	packetProtocol PacketProtocol
}

// NewServer creates a Server, you can set PacketProtocol, PacketHandler and
// the send channel size, so these parameters will be pass to Session.
func NewServer(listener net.Listener, packetProtocol PacketProtocol,
	packetHandler PacketHandler, sendChanSize int) *Server {
	return &Server{
		listener:       listener,
		sendChanSize:   sendChanSize,
		packetHandler:  packetHandler,
		packetProtocol: packetProtocol,
	}
}

func Listen(network, address string,
	protocol PacketProtocol, handler PacketHandler, sendChanSize int) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, protocol, handler, sendChanSize), nil
}

// Close destory the listener
func (s *Server) Close() error {
	return s.listener.Close()
}

// AcceptLoop begin accept connection.
// When incoming new session, you can get a chance to close the session,
// for example, you get a new session, and then find the remote address is in
// black list, so you can close the session.
// And, you can also set some parameter with the new session, such as
// send buffer size, recv buffer size and so on.
// If got a new session, but you don't want to start, must call Session.Close to close it.
// If everything is ok, you must call Session.Start to begin work.
func (s *Server) AcceptLoop(newSessionCallback func(*Session)) error {
	defer s.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			} else {
				return err
			}
		}
		session := NewSession(conn, s.packetProtocol, s.packetHandler, s.sendChanSize)
		newSessionCallback(session)
	}
}
