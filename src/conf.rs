use std::fs::File;
use std::sync::{Arc, RwLock};
use std::borrow::Borrow;
use std::ops::Deref;

use serde_json;

use data_types::*;

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct Config {
    pub api_keys: Vec<APIKey>,
    pub redis_conn: String,
    pub codelen: usize,
    pub server_url: String,
}

impl Default for Config {
    fn default() -> Config {
        Config {
            api_keys: Vec::new(),
            redis_conn: "redis://127.0.0.1:6379".to_string(),
            codelen: 6,
            server_url: "http://localhost:8000".to_string(),
        }
    }
}

lazy_static! {
    pub static ref CONF: Arc<RwLock<Config>> = {
        let mut config_raw: Option<Config> = None;
        let fs_option = File::open("conf.json");
        if let Ok(file) = fs_option {
            let config_raw_opt = serde_json::from_reader(file);
            if let Ok(config) = config_raw_opt {
                config_raw = Some(config);
            }
        }

        let out_conf: Config = match config_raw {
            Some(conf) => conf,
            None => Config::default(),
        };
        Arc::new(RwLock::new(out_conf))
    };
}

pub fn get_config() -> Config {
    let lock = CONF.borrow().read().unwrap();
    let cnf = lock.deref();
    cnf.clone()
}
