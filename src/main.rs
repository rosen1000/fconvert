use chalk::{colors::Colors::*, Chalk};
use std::process::exit;
use std::{env::args, process::Command, time::Instant};

const USAGE_STRING: &str = "Usage: fconvert [FORMAT] FILES...
Try 'fconvert --help' for more information.";
const HELP_STRING: &str = "Usage: fconvert [FORMAT] FILES...
Convert batch of files into target format
Example: fconvert mp3 video1.mp4 video2.mov";

fn main() {
    let args = args().collect::<Vec<String>>();

    if args.contains(&"--help".to_string()) || args.contains(&"-h".to_string()) {
        println!("{}", HELP_STRING);
        exit(0);
    }

    if args.len() < 3 {
        println!("{}", USAGE_STRING);
        exit(0);
    }

    let mut args = args.iter();
    args.next();
    let wanted_format = args.next().unwrap();
    let file_list = args;

    for file in file_list {
        let now = Instant::now();
        let wanted_file_name = format!("{}.{}", format_file_name(file), wanted_format);
        let output = Command::new("ffmpeg")
            .arg("-hide_banner")
            .arg("-i")
            .arg(file)
            .arg(&wanted_file_name)
            .output();

        if let Ok(v) = output {
            if v.status.success() {
                println!(
                    "Formated {} ({}ms)",
                    wanted_file_name,
                    now.elapsed().as_millis()
                );
            } else {
                let stderr = String::from_utf8(v.clone().stderr).unwrap();
                if stderr.contains("already exists") {
                    println!("Failed {} (already exists)", wanted_file_name);
                } else if stderr.contains("suitable output") {
                    println!(
                        "Failed {} (unable to find a suitable output format)",
                        wanted_file_name
                    );
                } else {
                    println!("Failed {} (unknown)", wanted_file_name);
                    println!("{:?}", stderr);
                }
            }
        } else if let Err(e) = output {
            if (e.raw_os_error()) == Some(2) {
                println!(
                    "{}: ffmpeg not found!\nTry adding ffmpeg to PATH",
                    Chalk::new(Red, "ERROR").color()
                );
                exit(1);
            }
        } else {
            panic!("Something went horribly wrong");
        }
    }

    println!("Done!");
}

fn format_file_name(file_name: &String) -> String {
    let mut parts: Vec<_> = file_name.split(".").collect();
    parts.pop();
    return parts.join(".");
}
