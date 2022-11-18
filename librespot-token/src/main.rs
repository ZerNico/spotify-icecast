use std::env;
use std::process::ExitCode;

use librespot::core::{authentication::Credentials, config::SessionConfig, session::Session};

const SCOPES: &str =
    "streaming,user-read-playback-state,user-modify-playback-state,user-read-currently-playing";

#[tokio::main]
async fn main() -> ExitCode {
    let session_config = SessionConfig::default();

    let args: Vec<_> = env::args().collect();
    if args.len() != 3 {
        eprintln!("Usage: {} USERNAME PASSWORD", args[0]);
        std::process::exit(1);
    }

    let credentials = Credentials::with_password(&args[1], &args[2]);
    let session = Session::new(session_config, None);

    match session.connect(credentials, false).await {
        Ok(()) => {
            let response =  session.token_provider().get_token(SCOPES).await.unwrap();
            println!("{}", response.access_token);

            std::process::exit(0);
        }
        Err(e) => {
            println!("Error connecting: {}", e);
            std::process::exit(1);
        }
    }
}