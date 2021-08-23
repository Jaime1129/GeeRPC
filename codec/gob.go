package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser // connection instance retrieved from TCP connection
	buf  *bufio.Writer      // buffer for messages
	dec  *gob.Decoder       // decoder
	enc  *gob.Encoder       // encoder
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	return nil
}

func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

func (c *GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	// after writing data into buffer, flush it so that the other side of connection will receive the data
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			// if error happens when writing data into buffer, close connection
			_ = c.Close()
		}
	}()

	// write header first
	if err = c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return
	}

	// write body then
	if err = c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return
	}

	return
}

func (c *GobCodec) Close() error {
	return c.conn.Close()
}