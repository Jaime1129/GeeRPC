package client

import (
	"errors"
	"github.com/Jaime1129/GeeRPC/codec"
	"io"
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
	client.mu.Lock()
	defer client.mu.Unlock()
	if !client.IsAvailable() {
		return 0, ErrShutdown
	}

	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq ++
	return call.Seq, nil
}

// removeCall removes Call from pending
func (client *Client) removeCall(seq uint64) *Call  {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

// terminatePendingCalls marks Client as shutdown, and terminate all pending Calls with provided error
func (client *Client) terminatePendingCalls(err error)  {
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
