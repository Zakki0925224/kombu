package main

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

//go:embed bpf_hook_syscall.o
var bpfBin []byte

type BpfObject struct {
	Events         *ebpf.Map     `ebpf:"events"`
	HookX64SysCall *ebpf.Program `ebpf:"hook_x64_sys_call"`
}

type SyscallEvent struct {
	Timestamp uint64
	SyscallNr uint32
	Pid       uint32
}

func (o *BpfObject) Close() error {
	if err := o.Events.Close(); err != nil {
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

	var o BpfObject
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

	rd, err := ringbuf.NewReader(o.Events)
	if err != nil {
		panic(err)
	}

	var event SyscallEvent
	for {
		record, err := rd.Read()
		if err != nil {
			if err == ringbuf.ErrClosed {
				panic(err)
			}
			continue
		}

		// parse record
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.NativeEndian, &event); err != nil {
			fmt.Printf("Failed to parse syscall event: %s\n", err)
			continue
		}

		fmt.Printf("%#v\n", event)
	}
}
