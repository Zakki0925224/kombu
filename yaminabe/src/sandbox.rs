use anyhow::Result;
use std::time::Duration;

use crate::wrapper;

pub struct Sandbox {
    container_id: String,
    specimen_program_path: String,
    timeout_dur: Option<Duration>,
}

impl Sandbox {
    pub fn new(
        container_id: String,
        specimen_program_path: String,
        timeout_sec: Option<u64>,
    ) -> Self {
        Self {
            container_id,
            specimen_program_path,
            timeout_dur: timeout_sec.map(Duration::from_secs),
        }
    }

    pub fn run(&self) -> Result<()> {
        wrapper::download_oci_container_bundle("ubuntu", "latest")?;
        wrapper::create_container(&self.container_id, "./bundles/ubuntu-latest")?;
        wrapper::start_container(&self.container_id, Some(&["uname", "-a"]))?;
        wrapper::delete_container(&self.container_id)?;

        Ok(())
    }
}
