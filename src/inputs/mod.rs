use std::thread;
use std::sync::mpsc::Sender;

pub mod email;
pub mod files;

pub struct Element {
    pub source: String,
    pub date: String,
    pub name: String,
}

pub enum Sources {
    Email,
    Files,
}

pub fn start(source: Sources, url: String, tx: Sender<Element>){
    thread::spawn(move||{
        match source{
            Sources::Email => {
                println!("Spawning email: {}", url);
                email::new(url, tx);
            },
            Sources::Files => {
                println!("Spawning file-watcher: {}", url);
                files::new(url, tx);
            }
        }
    });
}

/*
trait Input {
    // Update searches for new data.
    fn Update(&self);

    //
}
*/