use redis;
use redis::Commands;
use rand;
use rand::Rng;
use chrono::prelude::*;
use serde_json;

use conf;
use data_types::*;

#[derive(Debug)]
pub enum LookupError {
    NoSuchShortcode,
    DBError(redis::RedisError),
}

#[derive(Debug)]
pub enum InsertError {
    ExistingCodeClash,
    DBError(redis::RedisError),
}

#[derive(Debug)]
pub enum DeleteError {
    DoesNotExist,
    DBError(redis::RedisError),
}

fn get_client() -> Result<redis::Client, redis::RedisError> {
    let conn_str = conf::get_config().redis_conn;
    redis::Client::open(conn_str.as_str())
}

pub fn get_full_url(code: &str) -> Result<String, LookupError> {
    let client = try!(get_client().map_err(|e| LookupError::DBError(e)));
    let conn = try!(client.get_connection().map_err(|e| LookupError::DBError(e)));
    match conn.get(code.to_uppercase()) {
        Ok(url) => Ok(url),
        Err(err) => {
            match err.kind() {
                // TypeError = got a nil (no such key)
                redis::ErrorKind::TypeError => Err(LookupError::NoSuchShortcode),
                _ => Err(LookupError::DBError(err)),
            }
        },
    }
}

pub fn get_code_meta(code: &str) -> Result<CodeMeta, LookupError> {
    let key = format!("meta/{}", code.clone().to_uppercase());
    let client = try!(get_client().map_err(|e| LookupError::DBError(e)));
    let conn = try!(client.get_connection().map_err(|e| LookupError::DBError(e)));
    let resp: redis::RedisResult<String> = conn.get(key);
    match resp {
        Ok(meta) => Ok(serde_json::from_str(meta.as_str()).unwrap()),
        Err(err) => {
            match err.kind() {
                // TypeError = got a nil (no such key)
                redis::ErrorKind::TypeError => Err(LookupError::NoSuchShortcode),
                _ => Err(LookupError::DBError(err)),
            }
        },
    }
}

pub fn add_url(url: &str, owner: APIKey, user_meta: Option<String>) -> Result<String, InsertError> {
    let code_len = conf::get_config().codelen;
    let mut rng = rand::thread_rng();
    loop {
        let shortcode = rng.gen_ascii_chars().take(code_len).collect::<String>();
        match add_url_with_code(url, shortcode.as_str(), owner.clone(), user_meta.clone()) {
            Ok(new_url) => return Ok(new_url),
            Err(err) => {
                match err {
                    InsertError::ExistingCodeClash => {
                        println!("!! Random code clashed with entry in database. Retrying with new code. ({})", shortcode);
                        continue;
                    },
                    InsertError::DBError(_) => return Err(err),
                }
            }
        }
    }
}

pub fn add_url_with_code(url: &str, code: &str, owner: APIKey, user_meta: Option<String>) -> Result<String, InsertError> {
    let shortcode = code.to_uppercase();
    let client = try!(get_client().map_err(|e| InsertError::DBError(e)));
    let conn = try!(client.get_connection().map_err(|e| InsertError::DBError(e)));
    let exists: redis::RedisResult<bool> = conn.exists(shortcode.clone());
    match exists {
        Ok(true) => return Err(InsertError::ExistingCodeClash),
        Ok(false) => (),
        Err(e) => return Err(InsertError::DBError(e)),
    }
    try!(conn.set(shortcode.clone(), url).map_err(|e| InsertError::DBError(e)));
    let meta = CodeMeta {
        owner: owner.name,
        time: UTC::now(),
        user_meta: user_meta,
    };
    let res: redis::RedisResult<bool> = conn.set(format!("meta/{}", shortcode), serde_json::to_string(&meta).unwrap());
    match res {
        Ok(_) => Ok(shortcode),
        Err(e) => Err(InsertError::DBError(e)),
    }
}

pub fn delete_code(code: String) -> Result<String, DeleteError> {
    let shortcode = code.to_uppercase();
    let client = try!(get_client().map_err(|e| DeleteError::DBError(e)));
    let conn = try!(client.get_connection().map_err(|e| DeleteError::DBError(e)));
    let exists: redis::RedisResult<bool> = conn.exists(shortcode.clone());
    match exists {
        Ok(true) => (),
        Ok(false) => return Err(DeleteError::DoesNotExist),
        Err(e) => return Err(DeleteError::DBError(e)),
    }
    let del_code: redis::RedisResult<i8> = conn.del(shortcode.clone());
    let del_meta: redis::RedisResult<i8> = conn.del(format!("meta/{}", shortcode.clone()));
    match del_code {
        Ok(_) => match del_meta {
            Ok(_) => Ok(shortcode),
            Err(e) => Err(DeleteError::DBError(e)),
        },
        Err(e) => Err(DeleteError::DBError(e)),
    }
}
