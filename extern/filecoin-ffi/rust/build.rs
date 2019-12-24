use std::env;
use std::path::Path;

fn main() {
    println!("cargo:rerun-if-changed=src/");

    let out_dir = env::var("OUT_DIR").unwrap();
    let hdr_out = Path::new(&out_dir).join("include/filecoin.h");

    cbindgen::generate(std::env::var("CARGO_MANIFEST_DIR").unwrap())
        .expect("Could not generate header")
        .write_to_file(hdr_out);
}
