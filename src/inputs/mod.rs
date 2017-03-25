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

pub fn start(source: Sources, url: &str, tx: &Sender<Element>){
    let tx_copy = tx.clone();
    let str = String::from(url);
    thread::spawn(move||{
        match source{
            Sources::Email => {
                println!("Spawning email: {}", str);
                email::new(&str, tx_copy);
            },
            Sources::Files => {
                println!("Spawning file-watcher: {}", str);
                files::new(&str, tx_copy);
            }
        }
    });
}

pub fn new_element(src: Sources, url: &str, data: Vec<u8>, tags: Vec<Tag>) -> Element{
    Element{ source: src, url: String::from(url), data: data, tags: tags}
}