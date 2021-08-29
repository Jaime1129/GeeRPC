package client

import (
	"errors"
	"fmt"
	"github.com/Jaime1129/GeeRPC/codec"
	"net"
	"time"
)

type clientResult struct {
	client *Client
	err    error
}

type newClientFunc func(conn net.Conn, opt *codec.Option) (client *Client, err error)

func Dial(network, addr string, opts ...*codec.Option) (client *Client, err error) {
	return dialTimeout(NewClient, network, addr, opts...)
}

func dialTimeout(f newClientFunc, network, address string, opts ...*codec.Option) (cli *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout(network, address, opt.ConnTimeout)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = conn.Close()
		}
	}()

	ch := make(chan clientResult)
	go func() {
		// need to communicate with server to create new client
		cli, err := f(conn, opt)
		ch <- clientResult{client: cli, err: err}
	}()

	if opt.ConnTimeout == 0 {
		result := <-ch
		return result.client, result.err
	}

	select {
	case <-time.After(opt.ConnTimeout):
		return nil, fmt.Errorf("rpc client: connect timeout: expect within %s", opt.ConnTimeout.String())
	case result := <-ch:
		return result.client, result.err
	}
}

func parseOptions(opts ...*codec.Option) (*codec.Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return codec.DefaultOption, nil
	}

	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}

	opt := opts[0]
	opt.MagicNumber = codec.DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = codec.DefaultOption.CodecType
	}

	return opt, nil
}
