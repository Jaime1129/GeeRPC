package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Jaime1129/GeeRPC/codec"
	"io"
	"log"
	"net"
	"sync"
)

// Client represents a rpc client.
// Multiple outstanding calls may be associated with a single Client.
// A Client can by invoked in different goroutines.
type Client struct {
	cc  codec.Codec // for encoding and decoding messages
	opt *codec.Option

	sending sync.Mutex   // protect exclusive write access to connection
	header  codec.Header // as request can only be sent exclusively, the header can be reused

	mu       sync.Mutex       // protect following
	seq      uint64           // record the sequence of request. Each request has its unique seq number
	pending  map[uint64]*Call // seq of outstanding request -> Call struct
	closing  bool             // if user has called Close
	shutdown bool             //  if server has notify for shutting down
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

func NewClient(conn net.Conn, opt *codec.Option) (*Client, error) {
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: get codec func err=", err)
		return nil, err
	}

	// send options via connection
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: send options err=", err)
		// close connection
		_ = conn.Close()
		return nil, err
	}

	client := &Client{
		cc:      f(conn),
		opt:     opt,
		seq:     1, // normal calls start from 1, while 0 means invalid call.
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client, nil
}

// receive implements receiving reply from server
func (client *Client) receive() {
	var err error
	// wait for reply in an infinite loop
	// only exit the loop when err is not nil
	for err == nil {
		// read header
		var h codec.Header
		err = client.cc.ReadHeader(&h)
		if err != nil {
			break
		}
		// remove call from pending
		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			// 1. it usually means call is already removed, implying that server write partially failed
			// discard the body
			err = client.cc.ReadBody(nil)
		case h.Error != "":
			// 2. header indicates error
			call.Error = fmt.Errorf(h.Error)
			// discard the body
			err = client.cc.ReadBody(nil)
			call.done()
		default:
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("read body err=" + err.Error())
			}
			call.done()
		}
	}

	// terminate all ongoing Calls if error happens
	client.terminatePendingCalls(err)
}

// Close closes the Client
func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		// client has been closed
		return ErrShutdown
	}
	client.closing = true
	// client close -> codec close -> connection close
	return client.cc.Close()
}

func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.closing && !client.shutdown
}

// registerCall add new Call into pending Calls
func (client *Client) registerCall(call *Call) (uint64, error) {
	if !client.IsAvailable() {
		return 0, ErrShutdown
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil
}

// removeCall removes Call from pending
func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

// terminatePendingCalls marks Client as shutdown, and terminate all pending Calls with provided error
func (client *Client) terminatePendingCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()

	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}

func (client *Client) send(call *Call) {
	// acquire exclusive write access to connection to ensure sending a complete request
	client.sending.Lock()
	defer client.sending.Unlock()

	// register this call
	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// prepare request header
	// as header is already reused, it's necessary to init every field
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""

	// encode and send request
	err = client.cc.Write(&client.header, call.Args)
	if err != nil {
		call := client.removeCall(seq)
		// if call is nil, Write may partially failed
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

// Go creates a Call struct and sends which to server asynchronously.
// The Call struct carrying reply can be received from Call.Done
func (client *Client) Go(serviceMethod string, args, reply interface{}) *Call {
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          make(chan *Call, 10),
	}

	client.send(call)
	return call
}

// Call internally invokes Go to send request to server synchronously
func (client *Client) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	call := client.Go(serviceMethod, args, reply)

	select {
	case <-ctx.Done():
		client.removeCall(call.Seq)
		return fmt.Errorf("rpc client: call failed: %s", ctx.Err().Error())
	case call = <-call.Done:
		return call.Error
	}
	return call.Error
}
