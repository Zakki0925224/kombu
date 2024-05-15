package main

import (
	"bytes"
	_ "embed"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

//go:embed hook_syscall.o
var bpfBin []byte

type bpfObject struct {
	MyMap          *ebpf.Map     `ebpf:"my_map"`
	HookX64SysCall *ebpf.Program `ebpf:"hook_x64_sys_call"`
}

func (o *bpfObject) Close() error {
	if err := o.MyMap.Close(); err != nil {
		return err
	}

	if err := o.HookX64SysCall.Close(); err != nil {
		return err
	}

	return nil
}

func main() {
	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(bpfBin))
	if err != nil {
		panic(err)
	}

	if err := rlimit.RemoveMemlock(); err != nil {
		panic(err)
	}

	var o bpfObject
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
