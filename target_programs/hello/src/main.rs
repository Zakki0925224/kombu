use std::os::unix::fs::MetadataExt;

fn main() {
    println!("Hello, world!");
    println!(
        "metadata for {:?}",
        std::fs::metadata("/proc/self").map(|m| m.uid())
    );
    println!(
        "metadata for {:?}",
        std::fs::metadata("/proc/self").map(|m| m.gid())
    );
}
