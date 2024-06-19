package internal

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"golang.org/x/sys/unix"
)

// reference: https://github.com/opencontainers/runc/blob/main/libcontainer/sync_unix.go
type SyncSocket struct {
	F      *os.File
	closed atomic.Bool
}

func NewSocket(f *os.File) *SyncSocket {
	return &SyncSocket{F: f}
}

func NewPairSocket(name string) (parent, child *SyncSocket, err error) {
	fds, err := unix.Socketpair(unix.AF_LOCAL, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, nil, err
	}
	pFile := os.NewFile(uintptr(fds[1]), name+"-p")
	cFile := os.NewFile(uintptr(fds[0]), name+"-c")
	return NewSocket(pFile), NewSocket(cFile), nil
}

func (s *SyncSocket) Close() error {
	s.closed.Store(true)
	return s.F.Close()
}

func (s *SyncSocket) IsClose() bool {
	return s.closed.Load()
}

func (s *SyncSocket) Write(b []byte) (int, error) {
	return s.F.Write(b)
}

func (s *SyncSocket) Read() ([]byte, error) {
	size, _, err := unix.Recvfrom(int(s.F.Fd()), nil, unix.MSG_TRUNC|unix.MSG_PEEK)
	if err != nil {
		return nil, fmt.Errorf("fetch packet length from socket: %s", err)
	}

	if size == 0 {
		return nil, io.EOF
	}
	buf := make([]byte, size)
	n, err := s.F.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, fmt.Errorf("packet read too short: expected %d byte but %d bytes read", size, n)
	}

	return buf, nil
}

func GetSocketFromChild(name string) *SyncSocket {
	f := os.NewFile(uintptr(3), name)
	return NewSocket(f)
}
