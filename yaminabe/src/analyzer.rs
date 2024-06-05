use anyhow::Result;
use serde::Deserialize;

use crate::{syscall::X86_64Syscalls, uptime::Uptime};

pub type SyscallEvents = Vec<SyscallEvent>;

#[derive(Deserialize, Debug)]
pub struct SyscallEvent {
    timestamp: u64,
    nr: u32,
    pid: u32,
    ppid: u32,
    comm: [u8; 16],
}

impl SyscallEvent {
    fn comm_to_string(&self) -> String {
        String::from_utf8_lossy(&self.comm)
            .to_string()
            .replace('\0', "")
    }

    fn nr_to_syscall(&self) -> Option<X86_64Syscalls> {
        if let Ok(s) = X86_64Syscalls::try_from(self.nr) {
            Some(s)
        } else {
            None
        }
    }

    fn timestamp_to_uptime(&self) -> Uptime {
        Uptime::from(self.timestamp)
    }
}

pub struct Analyzer {
    syscall_events: SyscallEvents,
    target_pid: u32,
}

impl Analyzer {
    pub fn new(syscall_events: SyscallEvents, target_pid: u32) -> Self {
        Self {
            syscall_events,
            target_pid,
        }
    }

    pub fn analyze(&self) -> Result<()> {
        let filtered = self
            .syscall_events
            .iter()
            .filter(|e| e.comm_to_string() == "target")
            .collect::<Vec<_>>();

        for e in &filtered {
            println!(
                "uptime: {}, syscall: {:?} (nr: {}), pid: {}, ppid: {}, comm: {:?}",
                e.timestamp_to_uptime(),
                e.nr_to_syscall(),
                e.nr,
                e.pid,
                e.ppid,
                e.comm_to_string()
            );
        }
        println!("events count: {}", self.syscall_events.len());

        Ok(())
    }
}
