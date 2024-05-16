use crate::wrapper;
use anyhow::Result;
use log::info;
use std::{
    fs::{self, File},
    io::Write,
    path::{Path, PathBuf},
    time::Duration,
};

const SETUP_SH: &str = include_str!("./setup.sh");
const NIMONO_BIN: &[u8] = include_bytes!("../../build/nimono");

pub struct Sandbox {
    container_id: String,
    target_program_path: String,
    timeout_dur: Option<Duration>,
}

impl Drop for Sandbox {
    fn drop(&mut self) {
        let _ = wrapper::delete_container(&self.container_id);
        //let _ = self.remove_mount_dir();
    }
}

impl Sandbox {
    pub fn new(
        container_id: String,
        target_program_path: String,
        timeout_sec: Option<u64>,
    ) -> Self {
        Self {
            container_id,
            target_program_path,
            timeout_dur: timeout_sec.map(Duration::from_secs),
        }
    }

    pub fn run(&self) -> Result<()> {
        self.create_mount_dir()?;
        wrapper::download_oci_container_bundle("ubuntu", "latest")?;
        wrapper::create_container(&self.container_id, "./bundles/ubuntu-latest")?;

        let mount_dir_path = self.mount_dir_path();
        let mount_source_path = mount_dir_path.to_str().unwrap();
        let mount_dest_path = "/mnt";

        // setup container
        info!("Execute setup script...");
        wrapper::start_container(
            &self.container_id,
            mount_source_path,
            mount_dest_path,
            Some(&["sh", "/mnt/setup.sh"]),
            self.timeout_dur,
        )?;

        // restart and run target program
        info!("Execute target program...");
        wrapper::start_container(
            &self.container_id,
            mount_source_path,
            mount_dest_path,
            Some(&["/mnt/target"]),
            self.timeout_dur,
        )?;

        // remove container when dropped this sandbox instance

        Ok(())
    }

    fn mount_dir_path(&self) -> PathBuf {
        PathBuf::from(format!("mount-{}", self.container_id))
    }

    fn create_mount_dir(&self) -> Result<()> {
        // check target
        let target_program_path = Path::new(&self.target_program_path);
        if !target_program_path.exists() {
            return Err(anyhow::anyhow!("Target program does not exist"));
        }

        if !target_program_path.is_file() {
            return Err(anyhow::anyhow!("Target program is not a file"));
        }

        fs::create_dir(self.mount_dir_path())?;
        let mut setup_sh = File::create(self.mount_dir_path().join("setup.sh"))?;
        setup_sh.write_all(SETUP_SH.as_bytes())?;

        let mut nimono_bin = File::create(self.mount_dir_path().join("nimono"))?;
        nimono_bin.write_all(NIMONO_BIN)?;

        // copy target to mount directory
        fs::copy(target_program_path, self.mount_dir_path().join("target"))?;

        info!("Created mount directory");
        Ok(())
    }

    fn remove_mount_dir(&self) -> Result<()> {
        fs::remove_dir_all(self.mount_dir_path())?;

        info!("Removed mount directory");
        Ok(())
    }
}
