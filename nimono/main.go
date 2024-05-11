package main

import (
	"C"

	"github.com/cilium/ebpf"
)
import "fmt"

type myObjs struct {
	MyMap *ebpf.Map     `ebpf:"my_map"`
	Hello *ebpf.Program `ebpf:"hello"`
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

	var o myObjs
	if err := spec.LoadAndAssign(&o, nil); err != nil {
		panic(err)
	}
	defer o.Close()

	opt := ebpf.RunOptions{
		Data: make([]byte, 64),
	}
	ret, err := o.Hello.Run(&opt)
	if err != nil {
		panic(err)
	}

	fmt.Printf("ret: %d\n", ret)
}
