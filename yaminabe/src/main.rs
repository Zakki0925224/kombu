use analyzer::Analyzer;
use anyhow::Result;
use clap::Parser;
use log::info;
use sandbox::Sandbox;
use std::env;
use uuid::Uuid;

mod analyzer;
mod detection_rule;
mod sandbox;
mod syscall;
mod uptime;
mod wrapper;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long)]
    target_program_path: String,
    #[arg(short, long)]
    detection_rules_dir_path: String,
}

fn main() -> Result<()> {
    env::set_var("RUST_LOG", "info");
    env_logger::init();

    let args = Args::parse();
    let sandbox = Sandbox::new(Uuid::new_v4().to_string(), args.target_program_path);
    let sandbox_result = sandbox.run()?;
    let analyzer = Analyzer::new(sandbox_result, args.detection_rules_dir_path)?;
    let analyze_result = analyzer.analyze()?;

    if let Some(violated_detection_rule) = analyze_result.violated_detection_rule {
        info!("\x1b[31mviolation detected!\x1b[0m");
        info!("violated detection rule: {:?}", violated_detection_rule);
        info!("message: {:?}", analyze_result.message)
    } else {
        info!("\x1b[32mno violation detected!\x1b[0m");
    }

    Ok(())
}
