package server

import (
	"encoding/json"
	"github.com/Jaime1129/GeeRPC/codec"
	"io"
	"log"
	"net"
	"sync"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()
var InvalidRequest = struct {}{}

// Accept waits on listener for new connections.
// For each handle each newly created connections in new go routine.
func (s *Server) Accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("rpc server: accept err=", err)
		}
		go s.HandleConn(conn)
	}
}

func Accept(listener net.Listener) {
	DefaultServer.Accept(listener)
}

// HandleConn handles each connection established by communication.
func (s *Server) HandleConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()

	var opt codec.Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: decode options err=", err)
		return
	}

	if opt.MagicNumber != codec.MagicNumber {
		log.Println("rpc server: invalid magic number=", opt.MagicNumber)
		return
	}

	// acquire Codec constructor
	f, ok := codec.NewCodecFuncMap[opt.CodecType]
	if !ok {
		log.Println("rpc server: invalid codec type=", opt.CodecType)
		return
	}

	// construct Codec and apply to connection instance
	s.ApplyCodec(f(conn))
}

// ApplyCodec applies Codec methods to decode request sent via connection, and then process them by invoking request
// handler in dedicated go routines.
func (s *Server) ApplyCodec(cc codec.Codec) {
	sending := &sync.Mutex{} // protect write access for response
	wg := &sync.WaitGroup{}
	for {
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break // cannot recover if request is nil, so close the connection
			}
			req.h.Error = err.Error()
			// send error response
			s.sendResponse(cc, req.h, InvalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(cc, req, sending, wg)
	}

	// wait for all ongoing requests have been processed
	wg.Wait()
	_ = cc.Close()
}
