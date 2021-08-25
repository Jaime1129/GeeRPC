package server

import (
	"fmt"
	"github.com/Jaime1129/GeeRPC/client"
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

	cli, _ := client.Dial("tcp", <-addr)
	defer func() {
		_ = cli.Close()
	}()

	// wait until server start accepting connections
	time.Sleep(time.Second)

	wg := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			args := fmt.Sprintf("req %d", seq)
			var reply string
			if err := cli.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum err=", err)
			}
			log.Println("reply: ", reply)
		}(i)
	}
	wg.Wait()
}
