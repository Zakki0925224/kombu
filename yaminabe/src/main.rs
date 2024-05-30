use anyhow::Result;
use clap::Parser;
use sandbox::Sandbox;
use std::env;
use uuid::Uuid;

mod analyzer;
mod sandbox;
mod wrapper;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long)]
    target_program_path: String,
}

fn main() -> Result<()> {
    env::set_var("RUST_LOG", "info");
    env_logger::init();

    let args = Args::parse();
    let sandbox = Sandbox::new(Uuid::new_v4().to_string(), args.target_program_path);
    let analyzer = sandbox.run()?;
    analyzer.analyze()?;

    Ok(())
}
