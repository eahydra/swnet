### Description  

swnet is a simple network framework write by Golang. It just support TCP/IP. It can recv and send in background.
You can implement PacketHandler to dispatch or route packet. And can implement PacketProtocol to marshal/unmarshal packet.

### Example  

server:
```go
package main

import (
	"example/protocol"
	"fmt"
	"net"
	"swnet"
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
```

client:
```go
package main

import (
	"example/protocol"
	"fmt"
	"os"
	"os/signal"
	"swnet"
)

func onKeepaliveAck(session *swnet.Session, packet protocol.Packet) {
	fmt.Println("keepalive ack")
	ack := packet.(*protocol.KeepaliveAck)
	req := protocol.NewKeepalive()
	req.Token = ack.Token + 1
	req.Version = ack.Version
	session.AsyncSend(req)
}

func main() {
	swProtocol := protocol.NewDefaultProtocol(nil, false)
	dispatcher := protocol.NewDispatcher()
	dispatcher.AddHandler(protocol.PKTTYPE_KEEPALIVEACK, onKeepaliveAck)
	session, err := swnet.Dial("tcp4", "127.0.0.1:19905", swProtocol, dispatcher.Handle, 1024)
	if err != nil {
		fmt.Println("swnet.Dial failed, err:", err)
		return
	}
	defer session.Close()

	session.SetCloseCallback(func(*swnet.Session) {
		fmt.Println("exit")
		os.Exit(0)
	})
	session.Start()

	req := protocol.NewKeepalive()
	if err := session.AsyncSend(req); err != nil {
		fmt.Println("session.AsyncSend, err:", err)
		return
	}
	osSignal := make(chan os.Signal, 2)
	signal.Notify(osSignal, os.Kill, os.Interrupt)
	<-osSignal
}
```
