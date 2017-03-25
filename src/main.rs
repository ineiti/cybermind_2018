extern crate cybermind;
use std::sync::mpsc::{Sender, Receiver};
use std::sync::mpsc;
use cybermind::inputs;
use cybermind::inputs::Sources;

fn main() {
    let (tx, rx): (Sender<inputs::Element>, Receiver<inputs::Element>) = mpsc::channel();

    inputs::start(Sources::Email, "test@test.com", &tx);
    inputs::start(Sources::Email, "test1@test.com", &tx);
    inputs::start(Sources::Files, "/Users", &tx);

    loop{
        let e = rx.recv().unwrap();
        println!("Received from {}: {} - {}", e.source, e.url, e.tags[0]);
    }
}
