package codec

const MagicNumber = 0x3bef5c

/*
| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
| <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|

| Option | Header1 | Body1 | Header2 | Body2 | ...
*/

// Option serves as the first segmentation of GeeRPC messages, which is of fixed size encoding type Json
type Option struct {
	MagicNumber int  // MagicNumber indicates this is a GeeRPC message
	CodecType   Type // the client can specify different CodecType
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   GobType,
}


