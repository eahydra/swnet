package protocol

import "fmt"

type Packet interface {
	GetID() uint32
	GetPacketType() uint32
	Length() int
	AdjustLength()
	Read(stream ReadStream) error
	Write(stream WriteStream) error
}

var (
	ErrUnknownPacket = fmt.Errorf("unknown packet")
)

const (
	PKTTYPE_KEEPALIVE    uint32 = 0x00000001
	PKTTYPE_KEEPALIVEACK uint32 = 0x80000001
)

type PacketHeader struct {
	ID         uint32
	PacketType uint32
	Len        uint32
	Version    uint32
	Ack        uint32
	Token      uint32
}

func (p *PacketHeader) GetID() uint32 { return p.ID }

func (p *PacketHeader) GetPacketType() uint32 { return p.PacketType }

func (p *PacketHeader) Length() int { return 36 }

func (p *PacketHeader) AdjustLength() { p.Len = uint32(p.Length()) }

func (p *PacketHeader) Read(stream ReadStream) error {
	var err error
	if p.ID, err = stream.ReadUint32(); err != nil {
		return err
	}
	if p.PacketType, err = stream.ReadUint32(); err != nil {
		return err
	}
	if p.Len, err = stream.ReadUint32(); err != nil {
		return err
	}
	if p.Version, err = stream.ReadUint32(); err != nil {
		return err
	}
	if p.Ack, err = stream.ReadUint32(); err != nil {
		return err
	}
	if p.Token, err = stream.ReadUint32(); err != nil {
		return err
	}
	return nil
}

func (w *PacketHeader) Write(stream WriteStream) error {
	var err error
	if err = stream.WriteUint32(w.ID); err != nil {
		return err
	}
	if err = stream.WriteUint32(w.PacketType); err != nil {
		return err
	}
	if err = stream.WriteUint32(w.Len); err != nil {
		return err
	}
	if err = stream.WriteUint32(w.Version); err != nil {
		return err
	}
	if err = stream.WriteUint32(w.Ack); err != nil {
		return err
	}
	if err = stream.WriteUint32(w.Token); err != nil {
		return err
	}
	return nil
}

type Keepalive struct {
	PacketHeader
}

func (s *Keepalive) Length() int                    { return s.PacketHeader.Length() }
func (s *Keepalive) AdjustLength()                  { s.Len = uint32(s.Length()) }
func (s *Keepalive) Read(stream ReadStream) error   { return nil }
func (s *Keepalive) Write(stream WriteStream) error { return s.PacketHeader.Write(stream) }
func NewKeepalive() *Keepalive {
	return &Keepalive{
		PacketHeader: PacketHeader{
			PacketType: PKTTYPE_KEEPALIVE,
		},
	}
}

type KeepaliveAck struct {
	PacketHeader
}

func (s *KeepaliveAck) Length() int                    { return s.PacketHeader.Length() }
func (s *KeepaliveAck) AdjustLength()                  { s.Len = uint32(s.Length()) }
func (s *KeepaliveAck) Read(stream ReadStream) error   { return nil }
func (s *KeepaliveAck) Write(stream WriteStream) error { return s.PacketHeader.Write(stream) }
func NewKeepaliveAck() *KeepaliveAck {
	return &KeepaliveAck{
		PacketHeader: PacketHeader{
			PacketType: PKTTYPE_KEEPALIVEACK,
		},
	}
}

type PacketCacher interface {
	Get(id uint32, header *PacketHeader) Packet
	Put(id uint32, packet Packet)
}

type PacketFactory struct {
	Cacher PacketCacher
}

func NewPacketFactory(cacher PacketCacher) *PacketFactory {
	return &PacketFactory{
		Cacher: cacher,
	}
}

func (p *PacketFactory) CreatePacket(stream ReadStream) (newPacket Packet, err error) {
	var header PacketHeader
	if err = header.Read(stream); err != nil {
		return nil, err
	}
	if p.Cacher != nil {
		newPacket = p.Cacher.Get(header.PacketType, &header)
	}

	if newPacket == nil {
		switch header.PacketType {
		case PKTTYPE_KEEPALIVE:
			{
				newPacket = &Keepalive{PacketHeader: header}
			}
		case PKTTYPE_KEEPALIVEACK:
			{
				newPacket = &KeepaliveAck{PacketHeader: header}
			}
		default:
			{
				return nil, ErrUnknownPacket
			}
		}
	}

	if err = newPacket.Read(stream); err != nil {
		return nil, err
	}
	return newPacket, nil
}
