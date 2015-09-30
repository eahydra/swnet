package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/eahydra/swnet"
	"github.com/eahydra/swnet/example/protocol"
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
