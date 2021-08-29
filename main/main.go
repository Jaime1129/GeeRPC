package main

import (
	"context"
	"github.com/Jaime1129/GeeRPC/registry"
	"github.com/Jaime1129/GeeRPC/server"
	"github.com/Jaime1129/GeeRPC/xclient"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addr string, wg *sync.WaitGroup) {
	var foo Foo
	l, _ := net.Listen("tcp", ":0")
	s := server.NewServer()
	_ = s.Register(&foo)
	registry.Heartbeat(addr, "tcp@" + l.Addr().String(), 0)
	wg.Done()
	s.Accept(l)
}

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	http.Serve(l, nil)
}

func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}

	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {

	}
}

// call single service
func call(registry string) {
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.Random, nil)
	defer xc.Close()

	// send request and receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i ++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

// broadcast to all services
func broadcast(registry string) {
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.Random, nil)
	defer xc.Close()

	// send request and receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i ++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			// expect 2-5 timeout secs
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i * i})
		}(i)
	}

	wg.Wait()
}

func main() {
	log.SetFlags(0)
	registryAddr := "http://localhost:9999/_geerpc_/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	time.Sleep(time.Second)
	wg.Add(2)
	go startServer(registryAddr, &wg)
	go startServer(registryAddr, &wg)
	wg.Wait()

	call(registryAddr)
	broadcast(registryAddr)
}
