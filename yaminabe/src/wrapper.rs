use anyhow::Result;
use log::info;
use std::process::{Command, Output};

const RUNTIME_NAME: &str = "./build/dashi";

fn runtime_cmd(args: &[&str]) -> Command {
    let mut cmd = Command::new("sudo");
    cmd.args(&[&[RUNTIME_NAME], args].concat());
    cmd
}

fn output_to_result(output: Output) -> Result<()> {
    info!("{:?}", output);

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
    let mut cmd = runtime_cmd(&["download", docker_image_name, tag]);
    let output = cmd.output()?;
    output_to_result(output)?;
    Ok(())
}

pub fn create_container(container_id: &str, oci_runtime_bundle_path: &str) -> Result<()> {
    let mut cmd = runtime_cmd(&["create", container_id, oci_runtime_bundle_path]);
    let output = cmd.output()?;
    output_to_result(output)?;
    Ok(())
}

pub fn start_container(container_id: &str, command: Option<&[&str]>) -> Result<()> {
    let mut cmd = runtime_cmd(&["start", container_id]);
    if let Some(command) = command {
        cmd.args(command);
    }
    let output = cmd.output()?;
    output_to_result(output)?;
    Ok(())
}

pub fn delete_container(container_id: &str) -> Result<()> {
    let mut cmd = runtime_cmd(&["delete", container_id]);
    let output = cmd.output()?;
    output_to_result(output)?;
    Ok(())
}
