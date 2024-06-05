use std::fs;

fn main() {
    if let Err(err) = fs::remove_dir_all("/home/root") {
        println!("{:?}", err);
        return;
    }

    println!("removed successfully!");
}
