clang -O2 -g -c -target bpf bpf_hook_syscall.c
go build
./nimono &
pid=$!
sleep 1
kill -SIGINT "$pid"
