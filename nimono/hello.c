// +build ignore

#define __TARGET_ARCH_x86

#include <linux/bpf.h>
#include <linux/version.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

char _license[] SEC("license") = "GPL";
int _version SEC("version") = LINUX_VERSION_CODE;

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32);
    __type(value, __u64);
    __uint(max_entries, 1);
} my_map SEC(".maps");

// linux/arch/x86/entry/syscall_64.c
// long x64_sys_call(const struct pt_regs *regs, unsigned int nr)
SEC("fentry/x64_sys_call")
int BPF_PROG(hook_x64_sys_call, const struct pt_regs *regs, unsigned int nr)
{
    char msg[] = "x64_sys_call called, nr: %d\n";
    bpf_trace_printk(msg, sizeof(msg), nr);
    return 0;
}

SEC("socket")
int hello()
{
    char msg[] = "Hello, world!\n";
    bpf_trace_printk(msg, sizeof(msg));
    return 0;
}
