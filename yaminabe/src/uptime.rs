use std::fmt;

use anyhow::Result;

#[derive(Debug)]
pub struct Uptime {
    pub day: u64,
    pub hour: u64,
    pub min: u64,
    pub sec: u64,
    pub ms: u64,
    pub micro: u64,
    pub nano: u64,
}

impl From<u64> for Uptime {
    fn from(timestamp_ns: u64) -> Self {
        let sec = timestamp_ns / 1_000_000_000;
        let micro = (timestamp_ns % 1_000_000_000) / 1000;
        let nano = timestamp_ns % 1000;
        let day = sec / 86400;
        let hour = (sec % 86400) / 3600;
        let min = (sec % 3600) / 60;
        let sec = sec % 60;
        Uptime {
            day,
            hour,
            min,
            sec,
            ms: micro / 1000,
            micro: micro % 1000,
            nano,
        }
    }
}

impl fmt::Display for Uptime {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "{}d {}:{}:{}:{}.{}.{}",
            self.day, self.hour, self.min, self.sec, self.ms, self.micro, self.nano
        )
    }
}
