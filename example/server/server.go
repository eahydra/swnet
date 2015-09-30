package main

import (
	"fmt"
	"net"

	"github.com/eahydra/swnet"
	"github.com/eahydra/swnet/example/protocol"
)

func onKeepalive(session *swnet.Session, packet protocol.Packet) {
	fmt.Println("keepalive")
	req := packet.(*protocol.Keepalive)
	ack := protocol.NewKeepaliveAck()
	ack.Token = req.Token
	ack.Version = req.Version
	session.AsyncSend(ack)
}

func main() {
	swProtocol := protocol.NewDefaultProtocol(nil, false)
	dispatcher := protocol.NewDispatcher()
	dispatcher.AddHandler(protocol.PKTTYPE_KEEPALIVE, onKeepalive)
	server, err := swnet.Listen("tcp4", "127.0.0.1:19905", swProtocol, dispatcher.Handle, 1024)
	if err != nil {
		fmt.Println("swnet.Listen failed. err:", err)
		return
	}
	defer server.Close()

	server.AcceptLoop(func(session *swnet.Session) {
		tcpConn := session.RawConn().(*net.TCPConn)
		fmt.Println("Incoming new session. Remote:", tcpConn.RemoteAddr().String())
		tcpConn.SetNoDelay(true)
		tcpConn.SetReadBuffer(64 * 1024)
		tcpConn.SetWriteBuffer(64 * 1024)

		session.SetCloseCallback(func(s *swnet.Session) {
			fmt.Println("session closed!")
		})
		session.Start()
	})
}
