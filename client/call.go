package client

// Call represents an active rpc call
type Call struct {
	Seq           uint64      // Seq is a sequence number chosen by client
	ServiceMethod string      // format <service>.<method>
	Args          interface{} // arguments of rpc method
	Reply         interface{} // reply of rpc method
	Error         error
	Done          chan *Call // strobes when call is complete, for supporting async call
}

func (call *Call) done() {
	call.Done <- call
}
