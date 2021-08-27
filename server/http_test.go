package server

import (
	"context"
	"github.com/Jaime1129/GeeRPC/client"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func startHTTPServer(svc interface{}, addr chan<- string) {
	err := Register(svc)
	if err != nil {
		log.Fatalf("register service err=%s", err.Error())
	}

	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatal("listen err=", err)
	}
	HandleHTTP()
	addr <- l.Addr().String()
	log.Println("start listing addr=", l.Addr().String())
	_ = http.Serve(l, nil)
}

func call(addrCh chan string) {
	// client dial to server
	cli, _ := client.DialHTTP("tcp", <-addrCh)
	defer func() {
		_ = cli.Close()
	}()

	time.Sleep(time.Second)

	wg := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			args := Args{
				Num1: seq,
				Num2: seq * seq,
			}
			var reply int
			err := cli.Call(context.Background(), "Foo.Sum", args, &reply)
			if err != nil {
				log.Fatalf("call Foo.Sum seq=%d, err=%s", seq, err.Error())
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}

	wg.Wait()
}

func main() {
	log.SetFlags(0)
	ch := make(chan string)
	go call(ch)

	var foo Foo
	startHTTPServer(&foo, ch)
}
