package server

import (
	"github.com/Jaime1129/GeeRPC/codec"
	"io"
	"log"
	"reflect"
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
