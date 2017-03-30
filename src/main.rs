#![feature(plugin)]
#![plugin(rocket_codegen)]

extern crate rand;
#[macro_use]
extern crate lazy_static;
extern crate rocket;
extern crate rocket_contrib;
#[macro_use]
extern crate serde_derive;
extern crate serde_json;
extern crate redis;
extern crate chrono;

use rocket::response::{Failure, Redirect, content};
use rocket::http::Status;
use rocket_contrib::JSON;

mod data_types;
mod conf;
mod db;
mod security;

use data_types::*;

#[error(401)]
fn unauthenticated() -> content::JSON<String> {
    content::JSON(serde_json::to_string_pretty(&NO_APIKEY_ERROR).unwrap())
}

#[error(409)]
fn conflict() -> content::JSON<String> {
    content::JSON(serde_json::to_string_pretty(&KEY_IN_USE_ERROR).unwrap())
}

#[post("/api/shorten", format = "application/json", data = "<json>")]
fn shorten(api_key: APIKey, json: JSON<ShortenRequest>) -> Result<content::JSON<String>, Failure> {
    let resp = match json.code {
        None => db::add_url(json.url.as_str(), api_key, json.meta.clone()),
        Some(ref code) => db::add_url_with_code(json.url.as_str(), code.as_str(), api_key, json.meta.clone()),
    };
    match resp {
        Ok(ans) => {
            let final_url = format!("{}/{}", conf::get_config().server_url, ans);
            Ok(content::JSON(serde_json::to_string_pretty(&ShortenResponse { short_url: final_url }).unwrap()))
        },
        Err(db::InsertError::ExistingCodeClash) => Err(Failure(Status::Conflict)),
        Err(db::InsertError::DBError(_)) => Err(Failure(Status::InternalServerError))
    }
}

#[post("/api/delete", format = "application/json", data = "<json>")]
#[allow(unused_variables)]
fn delete(api_key: APIKey, json: JSON<DeleteRequest>) -> Result<content::JSON<String>, Failure> {
    let resp = db::delete_code(json.code.clone());
    match resp {
        Ok(ans) => Ok(content::JSON(serde_json::to_string_pretty(&DeleteResponse { code: ans, status: "deleted".to_string() }).unwrap())),
        Err(db::DeleteError::DoesNotExist) => Ok(content::JSON(serde_json::to_string_pretty(&DeleteResponse { code: json.code.to_uppercase(), status: "noexist".to_string() }).unwrap())),
        Err(db::DeleteError::DBError(e)) => Err(Failure(Status::InternalServerError)),
    }
}

#[get("/api/meta/<code>")]
fn shortmeta<'a>(code: &str) -> Result<content::JSON<String>, Failure> {
    match db::get_code_meta(code) {
        Ok(meta) => {
            let json = serde_json::to_string_pretty(&meta).unwrap();
            Ok(content::JSON(json))
        },
        Err(db::LookupError::NoSuchShortcode) => Err(Failure(Status::NotFound)),
        Err(db::LookupError::DBError(inner)) => {
            println!("!! DB error in meta handler: {:?}", inner);
            Err(Failure(Status::InternalServerError))
        },
    }
}

#[get("/<code>")]
fn shortcode(code: &str) -> Result<Redirect, Failure> {
    let target = match db::get_full_url(code) {
        Ok(url) => url,
        Err(err) => match err {
            db::LookupError::NoSuchShortcode => return Err(Failure(Status::NotFound)),
            db::LookupError::DBError(inner) => {
                println!("!! DB error in shortcode handler: {:?}", inner);
                return Err(Failure(Status::InternalServerError));
            },
        }
    };
    Ok(Redirect::to(target.as_str()))
}

fn main() {
    rocket::ignite().mount("/", routes![
        shorten,
        delete,
        shortmeta,
        shortcode
    ]).catch(errors![
        unauthenticated,
        conflict
    ]).launch();
}
