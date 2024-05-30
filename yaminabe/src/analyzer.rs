use anyhow::Result;
use serde::Deserialize;

pub type SyscallEvents = Vec<SyscallEvent>;

#[derive(Deserialize, Debug)]
pub struct SyscallEvent {
    timestamp: i64,
    nr: i64,
    pid: i64,
}

pub struct Analyzer {
    syscall_events: SyscallEvents,
    target_pid: i64,
}

impl Analyzer {
    pub fn new(syscall_events: SyscallEvents, target_pid: i64) -> Self {
        Self {
            syscall_events,
            target_pid,
        }
    }

    pub fn analyze(&self) -> Result<()> {
        //println!("{:?}", self.syscall_events);
        let filtered_events = self
            .syscall_events
            .iter()
            .filter(|e| e.pid == self.target_pid)
            .collect::<Vec<_>>();
        println!("{:?}", filtered_events);

        // pid list
        let mut pid_list: Vec<i64> = Vec::new();
        for e in self.syscall_events.iter() {
            let pid = e.pid;

            if pid_list.iter().find(|p| **p == pid).is_none() {
                pid_list.push(pid);
            }
        }
        pid_list.sort();
        println!("{:?}", pid_list);

        Ok(())
    }
}
