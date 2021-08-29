package client

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/Jaime1129/GeeRPC/codec"
	"github.com/Jaime1129/GeeRPC/common"
	"io"
	"net"
	"net/http"
	"strings"
)

// NewHTTPClient establishes HTTP connection from client to server via HTTP CONNECT method
func NewHTTPClient(conn net.Conn, opt *codec.Option) (*Client, error) {
	// send CONNECT HTTP request
	_, _ = io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", common.DefaultRPCPath))

	// require success response
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == common.Connected {
		return NewClient(conn, opt)
	}

	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}

	return nil, err
}

// DialHTTP is a convenient function for establishing HTTP connection from client to server
func DialHTTP(network, addr string, opts ...*codec.Option) (*Client, error) {
	return dialTimeout(NewHTTPClient, network, addr, opts...)
}

func XDial(rpcAddr string, opts ...*codec.Option) (*Client, error) {
	parts := strings.Split(rpcAddr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("rpc client err: wrong format of rpc address %s", rpcAddr)
	}

	protocol, addr := parts[0], parts[1]
	switch protocol {
	case "http":
		return DialHTTP("tcp", addr, opts...)
	default:
		return Dial(protocol, addr, opts...)
	}
}