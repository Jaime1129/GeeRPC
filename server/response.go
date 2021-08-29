package server

import (
	"github.com/Jaime1129/GeeRPC/codec"
	"log"
	"sync"
)

func (s *Server)sendResponse(cc codec.Codec, header *codec.Header, body interface{}, sending *sync.Mutex)  {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(header, body); err != nil {
		log.Println("rpc server: write response err=", err)
	}
}
