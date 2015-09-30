package protocol

type DefaultBodyReadWriter struct {
	factory   *PacketFactory
	BigEndian bool
}

func NewDefaultBodyReadWriter(cacher PacketCacher, bigEndian bool) *DefaultBodyReadWriter {
	return &DefaultBodyReadWriter{
		factory:   NewPacketFactory(cacher),
		BigEndian: bigEndian,
	}
}

func (d *DefaultBodyReadWriter) ReadBody(buff []byte) (interface{}, error) {
	var readStream ReadStream
	if d.BigEndian {
		readStream = NewBigEndianStream(buff)
	} else {
		readStream = NewLittleEndianStream(buff)
	}
	return d.factory.CreatePacket(readStream)
}

func (d *DefaultBodyReadWriter) GetLength(packet interface{}) int {
	p := packet.(Packet)
	return p.Length()
}

func (d *DefaultBodyReadWriter) Write(packet interface{}, buff []byte) error {
	p := packet.(Packet)
	p.AdjustLength()
	var writeStream WriteStream
	if d.BigEndian {
		writeStream = NewBigEndianStream(buff)
	} else {
		writeStream = NewLittleEndianStream(buff)
	}
	return p.Write(writeStream)
}
