package main

import (
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

type myObjs struct {
	MyMap          *ebpf.Map     `ebpf:"my_map"`
	HookX64SysCall *ebpf.Program `ebpf:"hook_x64_sys_call"`
	Hello          *ebpf.Program `ebpf:"hello"`
}

func (o *myObjs) Close() error {
	if err := o.MyMap.Close(); err != nil {
		return err
	}

	if err := o.Hello.Close(); err != nil {
		return err
	}

	return nil
}

func main() {
	spec, err := ebpf.LoadCollectionSpec("hello.o")
	if err != nil {
		panic(err)
	}

	if err := rlimit.RemoveMemlock(); err != nil {
		panic(err)
	}

	var o myObjs
	if err := spec.LoadAndAssign(&o, nil); err != nil {
		panic(err)
	}
	defer o.Close()

	link, err := link.AttachTracing(link.TracingOptions{
		Program: o.HookX64SysCall,
	})
	if err != nil {
		panic(err)
	}
	defer link.Close()

	for {
	}
}
