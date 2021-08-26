package server

import (
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
	mtype  *methodType   // RPC method to be invoked
	svc    *service      // service instance which provides the RPC method
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
	// specify the service instance and method to be invoked
	req.svc, req.mtype, err = s.findService(header.ServiceMethod)
	if err != nil {
		log.Println("rpc server: find service err=", err)
		return nil, err
	}
	// create the argument and reply instance for reading from connection
	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReplyv()

	// make sure argvi is a pointer, as ReadyBody need pointer argument
	var argvi interface{}
	if req.argv.Type().Kind() == reflect.Ptr {
		// if argv is Ptr type, Interface() will return it as interface{} type
		argvi = req.argv.Interface()
	} else {
		// else, acquire the Address of argv first, then convert to interface{} type
		argvi = req.argv.Addr().Interface()
	}

	// deserialize the body as argument
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err=", err)
		return nil, err
	}

	// construct request
	return req, nil
}

func (s *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	// call registered rpc methods to get the right replyv
	log.Println(req.h, req.argv.Elem())
	err := req.svc.call(req.mtype, req.argv, req.replyv)
	if err != nil {
		req.h.Error = err.Error()
		s.sendResponse(cc, req.h, InvalidRequest, sending)
		return
	}

	s.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}
