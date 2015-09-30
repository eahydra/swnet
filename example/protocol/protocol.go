package protocol

import (
	"errors"
	"io"
	"net"
)

const (
	DataPacketType    = 0xFDFD
	ControlPacketType = 0xFEFE
)

var (
	ErrUnknwonPacketHeader = errors.New("ICafeProtocol: Unknown packet header")
	ErrBodyTooBig          = errors.New("ICafeProtocol: packet body is too big")
)

var (
	MaxPacketBodySize uint32 = 512 * 1024
)

type BodyWriter interface {
	GetLength(packet interface{}) int
	Write(packet interface{}, buff []byte) error
}

type BodyReader interface {
	ReadBody(buff []byte) (interface{}, error)
}

type ProtocolImpl struct {
	Writer BodyWriter
	Reader BodyReader
}

type SWPacketHeader struct {
	Flag       uint16
	PacketFlag uint16
	BodyLength uint32
	SrcLength  uint32
}

func parseHeader(buff []byte) (*SWPacketHeader, error) {
	stream := NewBigEndianStream(buff)
	var header SWPacketHeader
	var err error
	if header.Flag, err = stream.ReadUint16(); err != nil {
		return nil, err
	}
	if header.PacketFlag, err = stream.ReadUint16(); err != nil {
		return nil, err
	}
	if header.BodyLength, err = stream.ReadUint32(); err != nil {
		return nil, err
	}
	if header.SrcLength, err = stream.ReadUint32(); err != nil {
		return nil, err
	}
	return &header, nil
}

func NewDefaultProtocol(cacher PacketCacher, bigEndian bool) *ProtocolImpl {
	bodyrw := NewDefaultBodyReadWriter(cacher, bigEndian)
	return &ProtocolImpl{
		Reader: bodyrw,
		Writer: bodyrw,
	}
}

func (d *ProtocolImpl) ReadPacket(conn net.Conn, buff []byte) (interface{}, []byte, error) {
	if cap(buff) < 12 {
		buff = make([]byte, 12)
	}
	var header *SWPacketHeader
	var err error
L:
	for {
		if _, err = io.ReadAtLeast(conn, buff[:12], 12); err != nil {
			return nil, nil, err
		}
		header, err = parseHeader(buff[:12])
		if err != nil {
			return nil, nil, err
		}

		switch header.Flag {
		case ControlPacketType:
			{
				controlPacket := []byte{0xFE, 0xFE, 0x00, 0x10, 0x7F, 0xCD,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
				if _, err = conn.Write(controlPacket); err != nil {
					return nil, nil, err
				}
			}
		case DataPacketType:
			{
				if header.BodyLength >= MaxPacketBodySize {
					return nil, nil, ErrBodyTooBig
				}
				break L
			}
		default:
			{
				return nil, nil, ErrUnknwonPacketHeader
			}
		}
	}

	if cap(buff) < int(header.BodyLength) {
		buff = make([]byte, header.BodyLength+12)
	}
	if _, err := io.ReadFull(conn, buff[:header.BodyLength]); err != nil {
		return nil, nil, err
	}

	packet, err := d.Reader.ReadBody(buff[:header.BodyLength])
	if err != nil {
		return nil, nil, err
	}
	return packet, buff[:], nil
}

func (d *ProtocolImpl) BuildPacket(packet interface{}, buff []byte) ([]byte, error) {
	length := d.Writer.GetLength(packet) + 12
	if cap(buff) < length {
		buff = make([]byte, length)
	}
	stream := NewBigEndianStream(buff[:12])
	if err := stream.WriteUint16(DataPacketType); err != nil {
		return nil, err
	}
	if err := stream.WriteUint16(0); err != nil {
		return nil, err
	}
	if err := stream.WriteUint32(uint32(length - 12)); err != nil {
		return nil, err
	}
	if err := stream.WriteUint32(uint32(length - 12)); err != nil {
		return nil, err
	}

	if err := d.Writer.Write(packet, buff[12:length]); err != nil {
		return nil, err
	}
	return buff[:length], nil
}

func (d *ProtocolImpl) WritePacket(conn net.Conn, buff []byte) error {
	_, err := conn.Write(buff)
	return err
}
