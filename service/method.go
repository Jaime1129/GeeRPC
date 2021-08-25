package service

import (
	"reflect"
	"sync/atomic"
)

// methodType stores the necessary method information for invoking a method in reflection manner
type methodType struct {
	method    reflect.Method // method represents the actual method to be invoked
	ArgType   reflect.Type   // argv is the argument type, which is the first argument of local method
	ReplyType reflect.Type   // replyv is the reply type, which is the second argument of local method
	numCalls  uint64         // record the number of calls made to local method
}

// NumCalls returns the number of calls made so far to this method
func (m *methodType) NumCalls() uint64 {
	// read in atomic manner
	return atomic.LoadUint64(&m.numCalls)
}

// newArgv constructs a new zero value of ArgType
func (m *methodType) newArgv() reflect.Value {
	var val reflect.Value
	// ArgType can be pointer or value type
	switch m.ArgType.Kind() {
	case reflect.Ptr:
		// Type.Elem() returns the underlying Type of pointer
		val = reflect.New(m.ArgType.Elem())
	default:
		// Value.Elem() returns the underlying Value of pointer
		val = reflect.New(m.ArgType).Elem()
	}

	return val
}

// newReplyv constructs a pointer pointing to zero value of underlying type
func (m *methodType) newReplyv() reflect.Value {
	// ReplyType must be pointer type
	// replyv is a PtrTo value pointing to the element type, possibly array/map/slice/channel or pointed type
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}
