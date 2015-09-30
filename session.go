package swnet

import (
	"errors"
	"net"
	"sync/atomic"
)

var (
	// ErrStoped means session had been closed
	ErrStoped = errors.New("swnet: session had stoped")
	// ErrSendChanBlocking means the chan of send is full
	ErrSendChanBlocking = errors.New("swnet: Send Channel blocking")
)

// PacketHandler is used to process packet that recved from remote session
// When got a valid packet from PacketReader, you can dispatch it.

type PacketHandler func(s *Session, packet interface{})

// PacketReader is used to unmarshal a complete packet from buff
type PacketReader interface {
	// Read data from conn and build a complete packet.
	// How to read from conn is up to you. You can set read timeout or other option.
	// If buff's capacity is small, you can make a new buff, then return it,
	// so can reuse to reduce memory overhead.
	ReadPacket(conn net.Conn, buff []byte) (interface{}, []byte, error)
}

// PacketWriter is used to marshal packet into buff
type PacketWriter interface {
	// Build a complete packet. If buff's capacity is too small,  you can make a new one
	// and return it to reuse.
	BuildPacket(packet interface{}, buff []byte) ([]byte, error)

	// How to write data to conn is up to you. So you can set write timeout or other option.
	WritePacket(conn net.Conn, buff []byte) error
}

// PacketProtocol just a composite interface
type PacketProtocol interface {
	PacketReader
	PacketWriter
}

// Session is a tcp connection wrapper. It recved data in silence, and
// queue data to send.
type Session struct {
	closed         int32
	conn           net.Conn
	sendChan       chan interface{}
	stopedChan     chan struct{}
	closeCallback  func(*Session)
	sendCallback   func(*Session, interface{})
	packetHandler  PacketHandler
	packetProtocol PacketProtocol
}

// NewSession new a session. You can set PacketProtocol, PacketHandler. and you can set
// the chan size of send to ensure fairness.
func NewSession(conn net.Conn, protocol PacketProtocol, handler PacketHandler, sendChanSize int) *Session {
	return &Session{
		closed:         -1,
		conn:           conn,
		stopedChan:     make(chan struct{}),
		sendChan:       make(chan interface{}, sendChanSize),
		packetHandler:  handler,
		packetProtocol: protocol,
	}
}

func Dial(network, address string, protocol PacketProtocol, handler PacketHandler, sendChanSize int) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewSession(conn, protocol, handler, sendChanSize), nil
}

// RawConn return net.Conn, so you can set/get parameter with it
func (s *Session) RawConn() net.Conn {
	return s.conn
}

// Close the session, destory other resource.
func (s *Session) Close() error {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		s.conn.Close()
		close(s.stopedChan)
		if s.closeCallback != nil {
			s.closeCallback(s)
		}
	}
	return nil
}

// SetCloseCallback can set a callback that be invoked when session closed.
func (s *Session) SetCloseCallback(callback func(*Session)) {
	s.closeCallback = callback
}

// SetSendCallback can set a callback when send complete, so you can reuse the packet
func (s *Session) SetSendCallback(callback func(*Session, interface{})) {
	s.sendCallback = callback
}

// SetPacketHandler can set a new packet handler. For example when server create a new session, and
// at this you can change the packet handler to process different operation.
func (s *Session) SetPacketHandler(handler PacketHandler) {
	s.packetHandler = handler
}

// SetProtocol can set a new PacketProtocol.
func (s *Session) SetProtocol(protocol PacketProtocol) {
	s.packetProtocol = protocol
}

// SetSendChanSize can change the chan size of send
func (s *Session) SetSendChanSize(chanSize int) {
	s.sendChan = make(chan interface{}, chanSize)
}

// GetSendChanSize return the chan size of send
func (s *Session) GetSendChanSize() int {
	return cap(s.sendChan)
}

// Start can call when new session created by server or a client session to start
func (s *Session) Start() {
	if atomic.CompareAndSwapInt32(&s.closed, -1, 0) {
		go s.sendLoop()
		go s.recvLoop()
	}
}

func (s *Session) recvLoop() {
	defer s.Close()

	var recvBuff []byte
	var packet interface{}
	var err error
	for {
		packet, recvBuff, err = s.packetProtocol.ReadPacket(s.conn, recvBuff)
		if err != nil {
			break
		}
		s.packetHandler(s, packet)
	}
}

func (s *Session) sendLoop() {
	defer s.Close()

	var sendBuff []byte
	var err error

	for {
		select {
		case packet, ok := <-s.sendChan:
			{
				if !ok {
					return
				}

				if sendBuff, err = s.packetProtocol.BuildPacket(packet, sendBuff); err == nil {
					err = s.packetProtocol.WritePacket(s.conn, sendBuff)
				}
				if err != nil {
					return
				}
				if s.sendCallback != nil {
					s.sendCallback(s, packet)
				}
			}
		case <-s.stopedChan:
			{
				return
			}
		}
	}
}

// AsyncSend queue the packet to the chan of send,
// if the send channel is full, return ErrSendChanBlocking.
// if the session had been closed, return ErrStoped
func (s *Session) AsyncSend(packet interface{}) error {
	select {
	case s.sendChan <- packet:
	case <-s.stopedChan:
		return ErrStoped
	default:
		return ErrSendChanBlocking
	}
	return nil
}
