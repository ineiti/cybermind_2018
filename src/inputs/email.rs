use inputs;
use inputs::{Sources, Tag, Value};
use std::thread;
use std::time;
use std::sync::mpsc::Sender;

pub fn new(addr: &str, inform: Sender<inputs::Element>){
    loop {
        let e = inputs::new_element(Sources::Email,
                                   addr, vec![0; 0],
            vec![Tag{key: String::from("from"), value: Value::Str(String::from("someone"))}]);
        inform.send(e).unwrap();
        thread::sleep(time::Duration::new(1,0));
    }
}