package xclient

import (
	"context"
	"github.com/Jaime1129/GeeRPC/client"
	"github.com/Jaime1129/GeeRPC/codec"
	"io"
	"sync"
)

type XClient struct {
	d       Discovery // d will help determine the server which request should be send to
	mode    SelectMode
	opt     *codec.Option
	mu      sync.Mutex // protect following
	clients map[string]*client.Client
}

// ensure that XClient implements io.Closer interface
var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *codec.Option) *XClient {
	return &XClient{
		d:       d,
		mode:    mode,
		opt:     opt,
		clients: make(map[string]*client.Client),
	}
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()

	for key, client := range xc.clients {
		_ = client.Close()
		delete(xc.clients, key)
	}

	return nil
}

func (xc *XClient) dial(rpcAddr string) (*client.Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	cli, ok := xc.clients[rpcAddr]
	if ok && !cli.IsAvailable() {
		_ = cli.Close()
		delete(xc.clients, rpcAddr)
		cli = nil
	}

	if cli == nil {
		// init new client
		var err error
		cli, err = client.XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}

		xc.clients[rpcAddr] = cli
	}

	return cli, nil
}

func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {
	cli, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}

	return cli.Call(ctx, serviceMethod, args, reply)
}

func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}

	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}