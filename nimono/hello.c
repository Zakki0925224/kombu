#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

char _license[] SEC("license") = "GPL";

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32);
    __type(value, __u64);
    __uint(max_entries, 1);
} my_map SEC(".maps");

SEC("socket")
int hello()
{
    char msg[] = "Hello, world!\n";
    bpf_trace_printk(msg, sizeof(msg));
    return 0;
}
