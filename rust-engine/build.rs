use std::path::PathBuf;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let proto_root = PathBuf::from("../proto/definitions");

    let common_proto = proto_root.join("premierpro/common/v1/common.proto");
    let media_proto = proto_root.join("premierpro/media/v1/media.proto");

    // Verify proto files exist
    if !common_proto.exists() {
        panic!(
            "Proto file not found: {}. Run from the rust-engine directory.",
            common_proto.display()
        );
    }
    if !media_proto.exists() {
        panic!(
            "Proto file not found: {}. Run from the rust-engine directory.",
            media_proto.display()
        );
    }

    tonic_build::configure()
        .build_server(true)
        .build_client(true)
        .compile_protos(
            &[
                common_proto.to_str().unwrap(),
                media_proto.to_str().unwrap(),
            ],
            &[proto_root.to_str().unwrap()],
        )?;

    println!("cargo:rerun-if-changed=../proto/definitions/");
    Ok(())
}
