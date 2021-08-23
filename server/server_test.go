package server

import (
	"encoding/json"
	"fmt"
	"github.com/Jaime1129/GeeRPC/codec"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

func startServer(addr chan<- string) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("listen err=", err)
	}
	log.Println("start listing addr=", l.Addr().String())
	addr <- l.Addr().String()
	server := NewServer()
	server.Accept(l)
}

func TestServer_Accept(t *testing.T) {
	addr := make(chan string)
	go startServer(addr)

	conn, err := net.Dial("tcp", <-addr)
	if err != nil {
		log.Println("dial tcp err=", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// wait until server start accepting connections
	time.Sleep(time.Second)

	// send options first
	_ = json.NewEncoder(conn).Encode(codec.DefaultOption)
	cc := codec.NewGobCodec(conn)

	wg := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			h := &codec.Header{
				ServiceMethod: "foo.sum",
				Seq:           uint64(seq),
			}
			_ = cc.Write(h, fmt.Sprintf("req %d", h.Seq))

			replyHeader := &codec.Header{}
			_ = cc.ReadHeader(replyHeader)

			var reply string
			_ = cc.ReadBody(&reply)
			log.Println("reply: ", reply)
		}(i)
	}
	wg.Wait()
}
