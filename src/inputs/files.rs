use inputs;
use inputs::{Sources, Value, Tag};
use std::thread;
use std::time;
use std::sync::mpsc::Sender;
use std::fs;

pub fn new(path: &str, inform: Sender<inputs::Element>){
    loop {
        thread::sleep(time::Duration::new(1,0));

        let paths = match fs::read_dir(&path){
            Err(e) => {
                println!("Error while reading directory {}: {}", path, e);
                continue;
            },
            Ok(p) => p,
        };

        for p in paths {
            let pa = p.unwrap().path();
            let name = String::from(pa.to_str().unwrap());
            let e = inputs::new_element(Sources::Files, path.clone(),
                                        vec![0; 0],
                                        vec![Tag{key: String::from("name"),
                                           value: Value::Str(name)}]);
            inform.send(e).unwrap();
        }
    }
}