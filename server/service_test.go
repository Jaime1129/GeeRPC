package server

import (
	"github.com/Jaime1129/GeeRPC/util"
	"reflect"
	"testing"
)

type Foo int

type Args struct {
	Num1 int
	Num2 int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func Test_service(t *testing.T) {
	var foo Foo

	s := NewService(&foo)
	util.Assert(s.method["Sum"] != nil, "fail to register Foo.Sum")
	util.Assert(s.method["sum"] == nil, "unexported functions shouldn't be registered")
}

func Test_service_call(t *testing.T) {
	var foo Foo
	s := NewService(&foo)
	mType := s.method["Sum"]

	argv := mType.newArgv()
	replyv := mType.newReplyv()
	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 2}))
	err := s.call(mType, argv, replyv)
	util.Assert(err==nil && *replyv.Interface().(*int) == 3 && mType.NumCalls() == 1, "failed to call Foo.Sum")
}
