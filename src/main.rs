extern crate cybermind;
use std::sync::mpsc::{Sender, Receiver};
use std::sync::mpsc;
use cybermind::inputs;

fn main() {
    let (tx, rx): (Sender<inputs::Element>, Receiver<inputs::Element>) = mpsc::channel();

    inputs::start(inputs::Sources::Email, String::from("test@test.com"), tx.clone());
    inputs::start(inputs::Sources::Email, String::from("test1@test.com"), tx.clone());
    inputs::start(inputs::Sources::Files, String::from("/Users"), tx.clone());

    loop{
        let e = rx.recv().unwrap();
        println!("Received from {}: {}", e.source, e.name);
    }
}
