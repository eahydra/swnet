package protocol

import (
	"fmt"
	"sync"

	"github.com/eahydra/swnet"
)

type PacketHandler func(session *swnet.Session, packet Packet)

type Dispatcher struct {
	rwlock     sync.RWMutex
	handlerMap map[uint32]PacketHandler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlerMap: make(map[uint32]PacketHandler),
	}
}

func (p *Dispatcher) AddHandler(id uint32, handler PacketHandler) {
	p.rwlock.Lock()
	defer p.rwlock.Unlock()
	p.handlerMap[id] = handler
}

func (p *Dispatcher) DelHandler(id uint32, handler PacketHandler) {
	p.rwlock.Lock()
	defer p.rwlock.Unlock()
	delete(p.handlerMap, id)
}

func (p *Dispatcher) Handle(session *swnet.Session, packet interface{}) {
	if t, ok := packet.(Packet); ok {
		p.rwlock.RLock()
		defer p.rwlock.RUnlock()
		h, ok := p.handlerMap[t.GetPacketType()]
		if ok {
			h(session, t)
		} else {
			fmt.Println("NOT FOUND")
		}
	}
}
