use anyhow::Result;
use sandbox::Sandbox;
use std::env;
use uuid::Uuid;

mod analyzer;
mod sandbox;
mod wrapper;

fn main() -> Result<()> {
    env::set_var("RUST_LOG", "info");
    env_logger::init();

    let sandbox = Sandbox::new(Uuid::new_v4().to_string(), "".to_string(), None);
    sandbox.run()?;

    Ok(())
}
