package main

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	Timestamp uint64   `json:"timestamp"`
	SyscallNr uint32   `json:"nr"`
	Pid       uint32   `json:"pid"`
	Ppid      uint32   `json:"ppid"`
	Comm      [16]byte `json:"comm"`
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
		log.Fatalf("Failed to load BPF binary: %s\n", err)
		os.Exit(1)
	}
	log.Printf("Loaded BPF binary\n")

	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatalf("Failed to remove memlock: %s\n", err)
		os.Exit(1)
	}

	var o BpfObject
	if err := spec.LoadAndAssign(&o, nil); err != nil {
		log.Fatalf("Failed to load BPF object: %s\n", err)
		os.Exit(1)
	}
	log.Printf("Loaded BPF object\n")
	defer o.Close()

	link, err := link.AttachTracing(link.TracingOptions{
		Program: o.HookX64SysCall,
	})
	if err != nil {
		log.Fatalf("Failed to attach hook function: %s\n", err)
		os.Exit(1)
	}
	log.Printf("Attached hook function\n")
	defer link.Close()

	rd, err := ringbuf.NewReader(o.Events)
	if err != nil {
		log.Fatalf("Failed to create ringbuf reader: %s\n", err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var event SyscallEvent
	var events []SyscallEvent

l:
	for {
		select {
		case <-sigCh:
			// received signal
			break l
		default:
		}

		// read record
		record, err := rd.Read()
		if err != nil {
			if err == ringbuf.ErrClosed {
				log.Fatalf("ringbuf closed\n")
				os.Exit(1)
			}
			continue
		}

		// parse record
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.NativeEndian, &event); err != nil {
			log.Fatalf("Failed to parse syscall event: %s\n", err)
			continue
		}

		events = append(events, event)
	}

	// export json log
	file, err := os.Create("syscall_events.json")
	if err != nil {
		log.Fatalf("Failed to create file: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	if err := encoder.Encode(events); err != nil {
		log.Fatalf("Failed to encode: %s\n", err)
		os.Exit(1)
	}

	log.Printf("Successed to export log\n")
}
