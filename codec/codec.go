package codec

type Header struct {
	ServiceMethod string // cmd name
	Seq           uint64 // sequence number chosen by client
	Error         string
}


