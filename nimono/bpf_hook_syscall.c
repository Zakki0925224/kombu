// +build ignore

#define __TARGET_ARCH_x86

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

char _license[] SEC("license") = "Dual MIT/GPL";

struct syscall_event
{
    __u64 timestamp;
    __u32 syscall_nr;
    __u32 pid;
    __u32 ppid;
    char comm[TASK_COMM_LEN];
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
    struct task_struct *task;
    struct task_struct *parent_task;

    event = bpf_ringbuf_reserve(&events, sizeof(struct syscall_event), 0);

    // failed to reserve space in ringbuf
    if (!event)
    {
        return 0;
    }

    event->timestamp = bpf_ktime_get_ns();
    event->syscall_nr = nr;
    event->pid = bpf_get_current_pid_tgid() >> 32;

    task = (struct task_struct *)bpf_get_current_task();
    event->ppid = (__u32)BPF_CORE_READ(task, real_parent, tgid);
    bpf_get_current_comm(&event->comm, sizeof(event->comm));

    bpf_ringbuf_submit(event, 0);

    return 0;
}
