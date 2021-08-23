package server

import (
	"fmt"
	"github.com/Jaime1129/GeeRPC/codec"
	"io"
	"log"
	"reflect"
	"sync"
)

type request struct {
	h      *codec.Header
	argv   reflect.Value // argument value of request
	replyv reflect.Value // reply value of request
}

func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	// read header
	header := &codec.Header{}
	err := cc.ReadHeader(header)
	if err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header err=", err)
		}
		return nil, err
	}

	req := &request{h: header}
	// read body
	// TODO Day 1: assuming the type of response is string
	// check the type of request argument
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read body err=", err)
		return nil, err
	}

	// construct request
	return req, nil
}

func (s *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	// TODO should call registered rpc methods to get the right replyv
	defer wg.Done()
	log.Println(req.h, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("resp %d", req.h.Seq))
	s.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}