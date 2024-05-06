use anyhow::Result;
use std::time::Duration;

pub struct Sandbox {
    container_name: String,
    specimen_program_path: String,
    timeout_dur: Option<Duration>,
}

impl Sandbox {
    pub fn new(
        container_name: String,
        specimen_program_path: String,
        timeout_sec: Option<u64>,
    ) -> Self {
        Self {
            container_name,
            specimen_program_path,
            timeout_dur: timeout_sec.map(Duration::from_secs),
        }
    }

    pub fn run(&self) -> Result<()> {
        Ok(())
    }
}
