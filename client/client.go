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

	sending sync.Mutex // protect header
	header  codec.Header

	mu       sync.Mutex // protect following
	seq      uint64
	pending  map[uint64]*Call
	closing  bool // if user has called Close
	shutdown bool //  if server has notify for shutting down
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

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
