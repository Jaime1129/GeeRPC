package server

import (
	"context"
	"github.com/Jaime1129/GeeRPC/client"
	"github.com/Jaime1129/GeeRPC/codec"
	"github.com/Jaime1129/GeeRPC/util"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func startServer(svc interface{}, addr chan<- string) {
	err := Register(svc)
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
	var foo Foo
	go startServer(&foo, addr)

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
				Num2: seq * seq,
			}
			var reply int
			if err := cli.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum err=", err)
			}
			log.Println("reply: ", reply)
		}(i)
	}
	wg.Wait()
}

type Bar int

func (b Bar) Timeout(argv int, reply *int) error {
	time.Sleep(2 * time.Second)
	*reply = argv + 1
	return nil
}

func TestServer_Call(t *testing.T) {
	t.Parallel()
	addrCh := make(chan string)
	var bar Bar
	go startServer(&bar, addrCh)

	addr := <-addrCh
	time.Sleep(time.Second)

	t.Run("client timeout", func(t *testing.T) {
		client, _ := client.Dial("tcp", addr)
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		var reply int
		err := client.Call(ctx, "Bar.Timeout", 1, &reply)
		util.Assert(err != nil && strings.Contains(err.Error(), ctx.Err().Error()), "expect client call timeout")
	})

	t.Run("server handle timeout", func(t *testing.T) {
		client, _ := client.Dial("tcp", addr, &codec.Option{HandleTimeout: time.Second})
		var reply int
		// client set not timeout in context
		err := client.Call(context.Background(), "Bar.Timeout", 1, &reply)
		util.Assert(err != nil && strings.Contains(err.Error(), "handle timeout"), "expect handling timeout")
	})
}

func Test_XDial(t *testing.T) {
	if runtime.GOOS != "linux" {
		return
	}

	ch := make(chan struct{})
	addr := "/tmp/geerpc.sock"
	go func() {
		_ = os.Remove(addr)
		l, err := net.Listen("unix", addr)
		if err != nil {
			t.Fatal("failed to listen unix socket")
		}
		ch <- struct{}{}
		Accept(l)
	}()

	<-ch
	_, err := client.XDial("unix@" + addr)
	util.Assert(err == nil, "failed to connect unix socket")
}
