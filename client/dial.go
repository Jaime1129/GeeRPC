package client

import (
	"github.com/Jaime1129/GeeRPC/codec"
	"log"
	"net"
)

func Dial(network, addr string) (client *Client, err error) {
	opt := codec.DefaultOption
	conn, err := net.Dial(network, addr)
	if err != nil {
		log.Println("dial err=", err)
		return nil, err
	}
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()

	client, err = NewClient(conn, opt)
	if err != nil {
		log.Println("client init err=", err)
		return nil, err
	}

	return client, nil
}

