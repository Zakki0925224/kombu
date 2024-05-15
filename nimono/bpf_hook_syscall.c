// +build ignore

#define __TARGET_ARCH_x86

#include <linux/bpf.h>
#include <linux/version.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

char _license[] SEC("license") = "Dual MIT/GPL";

struct syscall_event
{
    __u64 timestamp;
    __u32 syscall_nr;
    __u32 pid;
};

struct
{
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24);
} events SEC(".maps");

// linux/arch/x86/entry/syscall_64.c
// long x64_sys_call(const struct pt_regs *regs, unsigned int nr)
SEC("fentry/x64_sys_call")
int BPF_PROG(hook_x64_sys_call, const struct pt_regs *regs, unsigned int nr)
{
    struct syscall_event *event;
    event = bpf_ringbuf_reserve(&events, sizeof(struct syscall_event), 0);

    // failed to reserve space in ringbuf
    if (!event)
    {
        return 0;
    }

    event->timestamp = bpf_ktime_get_ns();
    event->syscall_nr = nr;
    event->pid = bpf_get_current_pid_tgid() >> 32;

    bpf_ringbuf_submit(event, 0);

    return 0;
}
