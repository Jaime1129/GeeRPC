package client

import (
	"github.com/Jaime1129/GeeRPC/codec"
	"github.com/Jaime1129/GeeRPC/util"
	"net"
	"strings"
	"testing"
	"time"
)

func Test_client_dialTimeout(t *testing.T) {
	t.Parallel()
	l, _ := net.Listen("tcp", ":0")

	f:= func(conn net.Conn, opt *codec.Option) (client *Client, err error){
		_ = conn.Close()
		time.Sleep(time.Second * 2)
		return nil, nil
	}

	t.Run("timeout", func(t *testing.T) {
		_, err := dialTimeout(f, "tcp", l.Addr().String(), &codec.Option{ConnTimeout: time.Second})
		util.Assert(err != nil && strings.Contains(err.Error(), "connect timeout"), "expect timeout error")
	})

	t.Run("unlimited timeout", func(t *testing.T) {
		_, err := dialTimeout(f, "tcp", l.Addr().String(), &codec.Option{ConnTimeout: 0})
		util.Assert(err == nil, "0 timeout")
	})
}
