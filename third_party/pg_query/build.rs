#![cfg_attr(feature = "clippy", feature(plugin))]
#![cfg_attr(feature = "clippy", plugin(clippy))]

use fs_extra::dir::CopyOptions;
use glob::glob;
use std::collections::hash_map::DefaultHasher;
use std::env;
use std::hash::Hasher;
use std::path::{Path, PathBuf};
use std::process::Command;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let out_dir = PathBuf::from(env::var("OUT_DIR")?);
    let build_path = Path::new(".").join("libpg_query");
    let out_header_path = out_dir.join("pg_query").with_extension("h");
    let out_protobuf_path = out_dir.join("protobuf");
    let target = env::var("TARGET").unwrap();

    println!("cargo:rerun-if-changed=libpg_query");
    println!("cargo:rustc-link-search=native={}", out_dir.display());
    println!("cargo:rustc-link-lib=static=pg_query");

    // Copy the relevant source files to the OUT_DIR, but only if they changed.
    // Without this check, fs_extra::copy_items overwrites files unconditionally,
    // updating their mtimes and causing Cargo to re-run the build script every time.
    let source_paths = vec![
        build_path.join("pg_query").with_extension("h"),
        build_path.join("postgres_deparse").with_extension("h"),
        build_path.join("Makefile"),
        build_path.join("src"),
        build_path.join("protobuf"),
        build_path.join("vendor"),
    ];

    let hash_file = out_dir.join(".source_hash");
    let current_hash = hash_source_paths(&source_paths);
    let cached_hash = std::fs::read_to_string(&hash_file).unwrap_or_default();

    if cached_hash.trim() != current_hash {
        let copy_options = CopyOptions { overwrite: true, ..CopyOptions::default() };
        fs_extra::copy_items(&source_paths, &out_dir, &copy_options)?;
        std::fs::write(&hash_file, &current_hash)?;
    }

    // Compile the C library.
    let mut build = cc::Build::new();
    build
        .files(glob(out_dir.join("src/*.c").to_str().unwrap()).unwrap().map(|p| p.unwrap()))
        .files(glob(out_dir.join("src/postgres/*.c").to_str().unwrap()).unwrap().map(|p| p.unwrap()))
        .file(out_dir.join("vendor/protobuf-c/protobuf-c.c"))
        .file(out_dir.join("vendor/xxhash/xxhash.c"))
        .file(out_dir.join("protobuf/pg_query.pb-c.c"))
        .include(out_dir.join("."))
        .include(out_dir.join("./vendor"))
        .include(out_dir.join("./src/postgres/include"))
        .include(out_dir.join("./src/include"))
        .warnings(false); // Avoid unnecessary warnings, as they are already considered as part of libpg_query development
    if env::var("PROFILE").unwrap() == "debug" || env::var("DEBUG").unwrap() == "1" {
        build.define("USE_ASSERT_CHECKING", None);
    }
    if target.contains("windows") {
        build.include(out_dir.join("./src/postgres/include/port/win32"));
        if target.contains("msvc") {
            build.include(out_dir.join("./src/postgres/include/port/win32_msvc"));
        }
    }
    build.compile("pg_query");

    // Generate bindings for Rust
    bindgen::Builder::default()
        .header(out_header_path.to_str().ok_or("Invalid header path")?)
        .generate()
        .map_err(|e| format!("Unable to generate bindings: {e}"))?
        .write_to_file(out_dir.join("bindings.rs"))?;

    // Only generate protobuf bindings if protoc is available
    let protoc_exists = Command::new("protoc").arg("--version").status().is_ok();
    // If the package is being built by docs.rs, we don't want to regenerate the protobuf bindings
    let is_built_by_docs_rs = env::var("DOCS_RS").is_ok();

    if !is_built_by_docs_rs && (env::var("REGENERATE_PROTOBUF").is_ok() || protoc_exists) {
        println!("generating protobuf bindings");
        let src_dir = PathBuf::from(env::var("CARGO_MANIFEST_DIR")?).join("src");
        let proto_tmp_dir = out_dir.join("proto_gen");
        std::fs::create_dir_all(&proto_tmp_dir)?;

        // Generate into a temp dir (inside OUT_DIR) instead of directly into src/
        env::set_var("OUT_DIR", &proto_tmp_dir);

        let mut prost_build = prost_build::Config::new();
        prost_build.type_attribute(".", "#[derive(serde::Serialize)]");
        prost_build.compile_protos(&[&out_protobuf_path.join("pg_query").with_extension("proto")], &[&out_protobuf_path])?;

        // Only update src/protobuf.rs if the generated output actually changed,
        // to avoid invalidating Cargo's fingerprint on every build
        let new_content = std::fs::read(proto_tmp_dir.join("pg_query.rs"))?;
        let existing = std::fs::read(src_dir.join("protobuf.rs")).unwrap_or_default();
        if new_content != existing {
            std::fs::copy(proto_tmp_dir.join("pg_query.rs"), src_dir.join("protobuf.rs"))?;
        }

        // Reset OUT_DIR to the original value
        env::set_var("OUT_DIR", &out_dir);
    } else {
        println!("skipping protobuf generation");
    }

    Ok(())
}

/// Hash the contents of all files under the given paths to detect changes.
fn hash_source_paths(paths: &[PathBuf]) -> String {
    let mut hasher = DefaultHasher::new();
    for path in paths {
        hash_path(&mut hasher, path);
    }
    format!("{:x}", hasher.finish())
}

fn hash_path(hasher: &mut DefaultHasher, path: &Path) {
    if path.is_file() {
        if let Ok(contents) = std::fs::read(path) {
            hasher.write(path.to_string_lossy().as_bytes());
            hasher.write(&contents);
        }
    } else if path.is_dir() {
        if let Ok(entries) = std::fs::read_dir(path) {
            let mut entries: Vec<_> = entries.filter_map(|e| e.ok()).collect();
            entries.sort_by_key(|e| e.file_name());
            for entry in entries {
                hash_path(hasher, &entry.path());
            }
        }
    }
}
