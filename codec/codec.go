package codec

import "io"

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

type Header struct {
	ServiceMethod string // cmd name
	Seq           uint64 // sequence number chosen by client
	Error         string
}

// Codec is for encoding and decoding RPC messages
type Codec interface {
	io.Closer

	// ReadHeader decodes the header of message
	ReadHeader(header *Header) error

	// ReadBody decodes the body of message
	ReadBody(body interface{}) error

	// Write encodes both the header and body of message
	Write(header *Header, body interface{}) error
}

// NewCodecFunc is the construct function of Codec instance. Similarly to factory method, the client and server can acquire
// the actual constructor struct by Type.
type NewCodecFunc func(closer io.ReadWriteCloser) Codec

type Type string

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
