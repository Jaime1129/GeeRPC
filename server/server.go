package server

import (
	"encoding/json"
	"errors"
	"github.com/Jaime1129/GeeRPC/codec"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()
var InvalidRequest = struct{}{}

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
	s.ApplyCodec(f(conn), opt.HandleTimeout)
}

// ApplyCodec applies Codec methods to decode request sent via connection, and then process them by invoking request
// handler in dedicated go routines.
func (s *Server) ApplyCodec(cc codec.Codec, timeout time.Duration) {
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
		go s.handleRequest(cc, req, sending, wg, timeout)
	}

	// wait for all ongoing requests have been processed
	wg.Wait()
	_ = cc.Close()
}

// Register registers a service which provides rpc method implementation
func (s *Server) Register(rcvr interface{}) error {
	svc := NewService(rcvr)
	if _, dup := s.serviceMap.LoadOrStore(svc.Name(), svc); dup {
		return errors.New("rpc server: service is already registered" + svc.Name())
	}

	return nil
}

func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
}

// findService returns the service and method stored in Server's map
func (s *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	split := strings.Split(serviceMethod, ".")
	if len(split) != 2 {
		err = errors.New("rpc server: invalid serviceMethod " + serviceMethod)
		return
	}

	svcName, methodName := split[0], split[1]
	svcVal, ok := s.serviceMap.Load(svcName)
	if !ok {
		err = errors.New("rpc server: service not found " + svcName)
		return
	}

	svc = svcVal.(*service)
	mtype, ok = svc.method[methodName]
	if !ok {
		err = errors.New("rpc server: method not found" + methodName)
		return
	}

	return
}

