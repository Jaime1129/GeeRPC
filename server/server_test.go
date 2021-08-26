package server

import (
	"github.com/Jaime1129/GeeRPC/client"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

func startServer(addr chan<- string) {
	var foo Foo
	err := Register(&foo)
	if err != nil {
		log.Fatalf("register service err=%s", err.Error())
	}

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("listen err=", err)
	}

	addr <- l.Addr().String()
	log.Println("start listing addr=", l.Addr().String())
	Accept(l)
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
			args := Args{
				Num1: seq,
				Num2: seq*seq,
			}
			var reply int
			if err := cli.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum err=", err)
			}
			log.Println("reply: ", reply)
		}(i)
	}
	wg.Wait()
}
