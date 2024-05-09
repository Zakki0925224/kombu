use anyhow::Result;
use log::info;
use std::{
    process::{Command, Output},
    time::Duration,
};
use wait_timeout::ChildExt;

const RUNTIME_NAME: &str = "./build/dashi";

fn runtime_cmd(args: &[&str], as_root: bool) -> Command {
    if as_root {
        let mut cmd = Command::new("sudo");
        cmd.args(&[&[RUNTIME_NAME], args].concat());
        cmd
    } else {
        let mut cmd = Command::new(RUNTIME_NAME);
        cmd.args(args);
        cmd
    }
}

fn output_to_result(output: Output) -> Result<()> {
    if output.status.success() {
        return Ok(());
    }

    Err(anyhow::anyhow!(
        "Failed to execute command: {:?}",
        &[
            String::from_utf8_lossy(&output.stdout),
            String::from_utf8_lossy(&output.stderr)
        ]
    ))
}

pub fn download_oci_container_bundle(docker_image_name: &str, tag: &str) -> Result<()> {
    let mut cmd = runtime_cmd(&["download", docker_image_name, tag], true);
    let output = cmd.output()?;
    output_to_result(output)?;
    Ok(())
}

pub fn create_container(container_id: &str, oci_runtime_bundle_path: &str) -> Result<()> {
    let mut cmd = runtime_cmd(&["create", container_id, oci_runtime_bundle_path], false);
    let output = cmd.output()?;
    output_to_result(output)?;
    Ok(())
}

pub fn start_container(
    container_id: &str,
    mount_source_path: &str,
    mount_dest_path: &str,
    command: Option<&[&str]>,
    timeout_dur: Option<Duration>,
) -> Result<()> {
    let mut cmd = runtime_cmd(
        &[
            "start",
            &format!("-mount-source={}", mount_source_path),
            &format!("-mount-dest={}", mount_dest_path),
            "-user",
            container_id,
        ],
        false,
    );
    if let Some(command) = command {
        info!("Running command in the container: {:?}", command);
        cmd.args(command);
    }

    if let Some(timeout_dur) = timeout_dur {
        let mut child = cmd.spawn()?;
        let status = match child.wait_timeout(timeout_dur) {
            Ok(s) => s.unwrap(),
            Err(_) => {
                child.kill()?;
                child.wait()?;
                info!("Timed out");
                return Ok(());
            }
        };

        if !status.success() {
            return Err(anyhow::anyhow!("Failed to execute command"));
        }

        return Ok(());
    }

    let output = cmd.output()?;
    output_to_result(output)?;
    Ok(())
}

pub fn delete_container(container_id: &str) -> Result<()> {
    let mut cmd = runtime_cmd(&["delete", container_id], true);
    let output = cmd.output()?;
    output_to_result(output)?;
    Ok(())
}
