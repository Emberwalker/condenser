use rocket::Outcome;
use rocket::http::Status;
use rocket::request::{self, Request, FromRequest};

use data_types::*;
use conf;

// This file is heavily based on the sample,
// https://api.rocket.rs/rocket/request/trait.FromRequest.html

impl<'a, 'r> FromRequest<'a, 'r> for APIKey {
    type Error = ();

    fn from_request(request: &'a Request<'r>) -> request::Outcome<APIKey, ()> {
        let keys: Vec<_> = request.headers().get("X-API-Key").collect();
        if keys.len() != 1 {
            return Outcome::Failure((Status::BadRequest, ()));
        }

        let key = keys[0];
        let key_entry = get_key_entry(key);

        match key_entry {
            None => Outcome::Failure((Status::Unauthorized, ())),
            Some(found_key) => Outcome::Success(found_key),
        }
    }
}

fn get_key_entry(key: &str) -> Option<APIKey> {
    let all_keys = conf::get_config().api_keys;
    for k in &all_keys {
        if key == k.key {
            return Some(k.clone());
        }
    }
    None
}
