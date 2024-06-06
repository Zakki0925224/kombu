use serde::Deserialize;

#[derive(Debug, Deserialize, Clone)]
pub struct Meta {
    pub name: String,
    pub desc: String,
}

#[derive(Debug, Deserialize, Clone)]
pub struct SyscallFrequent {
    pub threshold_count: usize,
    pub number: usize,
}

#[derive(Debug, Deserialize, Clone)]
pub struct SyscallConsecutive {
    pub threshold_count: usize,
    pub number: usize,
}

#[derive(Debug, Deserialize, Clone)]
pub struct Syscall {
    pub blacklist_numbers: Vec<usize>,
    pub frequent: Vec<SyscallFrequent>,
    pub consecutive: Vec<SyscallConsecutive>,
}

#[derive(Debug, Deserialize, Clone)]
pub struct Timestamp {
    pub check_timetravel: bool,
}

#[derive(Debug, Deserialize, Clone)]
pub struct DetectionRule {
    pub meta: Meta,
    pub syscall: Option<Syscall>,
    pub timestamp: Option<Timestamp>,
}
