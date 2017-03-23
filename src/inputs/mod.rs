use std::thread;
use std::sync::mpsc::Sender;
use std::fmt;

pub mod email;
pub mod files;

pub struct Element {
    pub source: Sources,
    pub url: String,
    pub data: Vec<u8>,
    pub tags: Vec<Tag>,
}

pub struct Tag{
    pub key: String,
    pub value: Value,
}

pub enum Value{
    Str(String),
    Elm(Element),
    Data(Vec<u8>),
}

pub enum Sources {
    Email,
    Files,
}

impl fmt::Display for Sources {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}",
        match *self{
            Sources::Email => "email",
            Sources::Files => "files",
        } )
    }
}

impl fmt::Display for Tag {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}:{}", self.key, self.value)
    }
}

impl fmt::Display for Value{
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result{
        let s = String::from("Data");
        write!(f, "{}", match *self{
            Value::Str(ref s) => s,
            Value::Elm(ref e) => &e.url,
            Value::Data(_) => &s,
        })
    }
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

pub fn new_element(src: Sources, url: String, data: Vec<u8>, tags: Vec<Tag>) -> Element{
    Element{ source: src, url: url, data: data, tags: tags}
}