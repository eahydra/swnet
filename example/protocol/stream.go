package protocol

import (
	"encoding/binary"
	"fmt"
)

var ErrBuffOverflow = fmt.Errorf("DSProtocol: buff is too small to io")

type ReadStream interface {
	Size() int
	Left() int
	Reset([]byte)
	Data() []byte
	ReadByte() (b byte, err error)
	ReadUint16() (b uint16, err error)
	ReadUint32() (b uint32, err error)
	ReadUint64() (b uint64, err error)
	ReadBuff(size int) (b []byte, err error)
	CopyBuff(b []byte) error
}

type WriteStream interface {
	Size() int
	Left() int
	Reset([]byte)
	Data() []byte
	WriteByte(b byte) error
	WriteUint16(b uint16) error
	WriteUint32(b uint32) error
	WriteUint64(b uint64) error
	WriteBuff(b []byte) error
}

type BigEndianStreamImpl struct {
	pos  int
	buff []byte
}

type LittleEndianStreamImpl struct {
	pos  int
	buff []byte
}

func NewBigEndianStream(buff []byte) *BigEndianStreamImpl {
	return &BigEndianStreamImpl{
		buff: buff,
	}
}

func NewLittleEndianStream(buff []byte) *LittleEndianStreamImpl {
	return &LittleEndianStreamImpl{
		buff: buff,
	}
}

func (impl *BigEndianStreamImpl) Size() int { return len(impl.buff) }

func (impl *BigEndianStreamImpl) Data() []byte { return impl.buff }

func (impl *BigEndianStreamImpl) Left() int { return len(impl.buff) - impl.pos }

func (impl *BigEndianStreamImpl) Reset(buff []byte) { impl.pos = 0; impl.buff = buff }

func (impl *BigEndianStreamImpl) ReadByte() (b byte, err error) {
	if impl.Left() < 1 {
		return 0, ErrBuffOverflow
	}
	b = impl.buff[impl.pos]
	impl.pos += 1
	return b, nil
}

func (impl *BigEndianStreamImpl) ReadUint16() (b uint16, err error) {
	if impl.Left() < 2 {
		return 0, ErrBuffOverflow
	}
	b = binary.BigEndian.Uint16(impl.buff[impl.pos:])
	impl.pos += 2
	return b, nil
}

func (impl *BigEndianStreamImpl) ReadUint32() (b uint32, err error) {
	if impl.Left() < 4 {
		return 0, ErrBuffOverflow
	}
	b = binary.BigEndian.Uint32(impl.buff[impl.pos:])
	impl.pos += 4
	return b, nil
}

func (impl *BigEndianStreamImpl) ReadUint64() (b uint64, err error) {
	if impl.Left() < 8 {
		return 0, ErrBuffOverflow
	}
	b = binary.BigEndian.Uint64(impl.buff[impl.pos:])
	impl.pos += 8
	return b, nil
}

func (impl *BigEndianStreamImpl) ReadBuff(size int) (buff []byte, err error) {
	if impl.Left() < size {
		return nil, ErrBuffOverflow
	}
	buff = make([]byte, size, size)
	copy(buff, impl.buff[impl.pos:impl.pos+size])
	impl.pos += size
	return buff, nil
}

func (impl *BigEndianStreamImpl) CopyBuff(b []byte) error {
	if impl.Left() < len(b) {
		return ErrBuffOverflow
	}
	copy(b, impl.buff[impl.pos:impl.pos+len(b)])
	return nil
}

func (impl *BigEndianStreamImpl) WriteByte(b byte) error {
	if impl.Left() < 1 {
		return ErrBuffOverflow
	}
	impl.buff[impl.pos] = b
	impl.pos += 1
	return nil
}

func (impl *BigEndianStreamImpl) WriteUint16(b uint16) error {
	if impl.Left() < 2 {
		return ErrBuffOverflow
	}
	binary.BigEndian.PutUint16(impl.buff[impl.pos:], b)
	impl.pos += 2
	return nil
}

func (impl *BigEndianStreamImpl) WriteUint32(b uint32) error {
	if impl.Left() < 4 {
		return ErrBuffOverflow
	}
	binary.BigEndian.PutUint32(impl.buff[impl.pos:], b)
	impl.pos += 4
	return nil
}

func (impl *BigEndianStreamImpl) WriteUint64(b uint64) error {
	if impl.Left() < 8 {
		return ErrBuffOverflow
	}
	binary.BigEndian.PutUint64(impl.buff[impl.pos:], b)
	impl.pos += 8
	return nil
}

func (impl *BigEndianStreamImpl) WriteBuff(buff []byte) error {
	if impl.Left() < len(buff) {
		return ErrBuffOverflow
	}
	copy(impl.buff[impl.pos:], buff)
	impl.pos += len(buff)
	return nil
}

func (impl *LittleEndianStreamImpl) Size() int { return len(impl.buff) }

func (impl *LittleEndianStreamImpl) Data() []byte { return impl.buff }

func (impl *LittleEndianStreamImpl) Left() int { return len(impl.buff) - impl.pos }

func (impl *LittleEndianStreamImpl) Reset(buff []byte) { impl.pos = 0; impl.buff = buff }

func (impl *LittleEndianStreamImpl) ReadByte() (b byte, err error) {
	if impl.Left() < 1 {
		return 0, ErrBuffOverflow
	}
	b = impl.buff[impl.pos]
	impl.pos += 1
	return b, nil
}

func (impl *LittleEndianStreamImpl) ReadUint16() (b uint16, err error) {
	if impl.Left() < 2 {
		return 0, ErrBuffOverflow
	}
	b = binary.LittleEndian.Uint16(impl.buff[impl.pos:])
	impl.pos += 2
	return b, nil
}

func (impl *LittleEndianStreamImpl) ReadUint32() (b uint32, err error) {
	if impl.Left() < 4 {
		return 0, ErrBuffOverflow
	}
	b = binary.LittleEndian.Uint32(impl.buff[impl.pos:])
	impl.pos += 4
	return b, nil
}

func (impl *LittleEndianStreamImpl) ReadUint64() (b uint64, err error) {
	if impl.Left() < 8 {
		return 0, ErrBuffOverflow
	}
	b = binary.LittleEndian.Uint64(impl.buff[impl.pos:])
	impl.pos += 8
	return b, nil
}

func (impl *LittleEndianStreamImpl) ReadBuff(size int) (buff []byte, err error) {
	if impl.Left() < size {
		return nil, ErrBuffOverflow
	}
	buff = make([]byte, size, size)
	copy(buff, impl.buff[impl.pos:impl.pos+size])
	impl.pos += size
	return buff, nil
}

func (impl *LittleEndianStreamImpl) CopyBuff(b []byte) error {
	if impl.Left() < len(b) {
		return ErrBuffOverflow
	}
	copy(b, impl.buff[impl.pos:impl.pos+len(b)])
	return nil
}

func (impl *LittleEndianStreamImpl) WriteByte(b byte) error {
	if impl.Left() < 1 {
		return ErrBuffOverflow
	}
	impl.buff[impl.pos] = b
	impl.pos += 1
	return nil
}

func (impl *LittleEndianStreamImpl) WriteUint16(b uint16) error {
	if impl.Left() < 2 {
		return ErrBuffOverflow
	}
	binary.LittleEndian.PutUint16(impl.buff[impl.pos:], b)
	impl.pos += 2
	return nil
}

func (impl *LittleEndianStreamImpl) WriteUint32(b uint32) error {
	if impl.Left() < 4 {
		return ErrBuffOverflow
	}
	binary.LittleEndian.PutUint32(impl.buff[impl.pos:], b)
	impl.pos += 4
	return nil
}

func (impl *LittleEndianStreamImpl) WriteUint64(b uint64) error {
	if impl.Left() < 8 {
		return ErrBuffOverflow
	}
	binary.LittleEndian.PutUint64(impl.buff[impl.pos:], b)
	impl.pos += 8
	return nil
}

func (impl *LittleEndianStreamImpl) WriteBuff(buff []byte) error {
	if impl.Left() < len(buff) {
		return ErrBuffOverflow
	}
	copy(impl.buff[impl.pos:], buff)
	impl.pos += len(buff)
	return nil
}
