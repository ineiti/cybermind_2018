use inputs;
use std::thread;
use std::time;
use std::sync::mpsc::Sender;

pub fn new(addr: String, inform: Sender<inputs::Element>){
    loop {
        let e = inputs::Element{source: String::from("email:") + &addr,
        date: String::from("Constant"), name: String::from("something")};
        inform.send(e).unwrap();
        thread::sleep(time::Duration::new(1,0));
    }
}