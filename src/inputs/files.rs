use inputs;
use std::thread;
use std::time;
use std::sync::mpsc::Sender;
use std::fs;

pub fn new(path: String, inform: Sender<inputs::Element>){
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
            let e = inputs::Element{source: String::from("file:") + &path,
                date: String::from("Constant"), name: name};
            inform.send(e).unwrap();
        }
    }
}