use std::fs::File;
use std::io::{Read, Write};

fn main() {
    // create file
    let mut file = File::create("sample_file.txt").unwrap();
    file.write_all(b"Hello, world!").unwrap();

    // read file
    let mut file = File::open("sample_file.txt").unwrap();
    let mut contents = String::new();
    file.read_to_string(&mut contents).unwrap();
    println!("{}", contents);

    // remove file
    std::fs::remove_file("sample_file.txt").unwrap();
}
