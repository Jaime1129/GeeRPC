package server

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

/* TODO: It might be better for NewService to only receive an instance which implements a specific interface as argument.
e.g.
type Processor interface {
	Process(argv, replyv interface) error
}
*/

// service holds the service that exports its methods as RPC method
type service struct {
	name   string                 // simply the name of functional instance, normally as the prefix of RPC method
	typ    reflect.Type           // a Go type of the functional instance
	rcvr   reflect.Value          // a Go Value of the functional instance
	method map[string]*methodType // where exported RPC methods of the functional instance are stored
}

// NewService receives a pointer of the actual functional instance, and create service instance accordingly.
func NewService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)

	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s service is not exported", s.name)
	}

	s.registerMethods()
	return s
}

func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		// iterate each method of service
		method := s.typ.Method(i)
		log.Printf("method name: %s.%s\n", s.typ.Name(), method.Name)
		mType := method.Type

		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			// RPC method must have two arguments, except the first one as self's pointer
			// RPC method also must have exactly one return value, which is error type
			continue
		}

		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}

		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
			numCalls:  0,
		}

		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

// call invokes registered methods in reflection manner
func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnVal := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnVal[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}

func (s *service) Name() string {
	return s.name
}