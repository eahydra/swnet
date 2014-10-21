### Description  

swnet is a simple network framework write by Golang. It just support TCP/IP. It can recv and send in background.
You can implement PacketHandler to dispatch or route packet. And can implement PacketProtocol to marshal/unmarshal packet.

### Example  

server:
```go
package main

import (
	"fmt"
	"net"
	"protocol/dsprotocol"
	"swnet"
)

func onKeepalive(session *swnet.Session, packet dsprotocol.Packet) {
	fmt.Println("keepalive")
	req := packet.(*dsprotocol.Keepalive)
	ack := dsprotocol.NewKeepaliveAck()
	ack.Token = req.Token
	ack.Version = req.Version
	session.AsyncSend(ack)
}

func main() {
	listener, err := net.Listen("tcp", ":19905")
	if err != nil {
		fmt.Println("net.Listen:", err)
		return
	}

	dsPacket := &dsprotocol.DSPacketReadWriter{BigEndian: false}
	icafeProtocol := &dsprotocol.ICafeProtocol{
		Reader: dsPacket,
		Writer: dsPacket,
	}

	packetDispatcher := dsprotocol.NewPacketDispatcher()
	packetDispatcher.AddHandler(dsprotocol.PKTTYPE_KEEPALIVE, onKeepalive)

	server := swnet.NewServer(listener, icafeProtocol, packetDispatcher, 1024)
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
```

client:
```go
package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"protocol/dsprotocol"
	"swnet"
)

func onKeepaliveAck(session *swnet.Session, packet dsprotocol.Packet) {
	fmt.Println("Incoming keepalive ack")
	ack := packet.(*dsprotocol.KeepaliveAck)
	req := dsprotocol.NewKeepalive()
	req.Token = ack.Token + 1
	req.Version = ack.Version
	session.AsyncSend(req)
}

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:19905")
	if err != nil {
		fmt.Println("net.Dial:", err)
		return
	}

	dsPacket := &dsprotocol.DSPacketReadWriter{BigEndian: false}
	icafeProtocol := &dsprotocol.ICafeProtocol{
		Reader: dsPacket,
		Writer: dsPacket,
	}

	packetDispatcher := dsprotocol.NewPacketDispatcher()
	packetDispatcher.AddHandler(dsprotocol.PKTTYPE_KEEPALIVEACK, onKeepaliveAck)

	session := swnet.NewSession(conn, icafeProtocol, packetDispatcher, 1024)
	defer session.Close()
	session.SetCloseCallback(func(*swnet.Session) {
		fmt.Println("exit")
		os.Exit(0)
	})
	session.Start()

	req := dsprotocol.NewKeepalive()
	if err := session.AsyncSend(req); err != nil {
		fmt.Println("session.AsyncSend, err:", err)
		return
	}
	osSignal := make(chan os.Signal, 2)
	signal.Notify(osSignal, os.Kill, os.Interrupt)
	<-osSignal
}
```
