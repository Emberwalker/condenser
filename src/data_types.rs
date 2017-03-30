use chrono::prelude::*;

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct APIKey {
    pub key: String,
    pub name: String,
}

#[derive(Deserialize, Debug)]
pub struct ShortenRequest {
    pub url: String,
    pub code: Option<String>,
    pub meta: Option<String>
}

#[derive(Serialize, Debug)]
pub struct ShortenResponse {
    pub short_url: String,
}

#[derive(Deserialize, Debug)]
pub struct DeleteRequest {
    pub code: String,
}

#[derive(Serialize, Debug)]
pub struct DeleteResponse {
    pub code: String,
    pub status: String,
}

#[derive(Serialize, Debug)]
pub struct GenericError<'a> {
    pub error: &'a str,
    pub message: Option<&'a str>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct CodeMeta {
    pub owner: String,
    pub time: DateTime<UTC>,
    pub user_meta: Option<String>,
}

#[derive(Serialize, Debug)]
pub struct CodeMetaResponse {
    pub full_url: String,
    pub meta: CodeMeta,
}

pub static NO_APIKEY_ERROR: GenericError = GenericError {
    error: "nokey",
    message: Some("No API key in X-API-Key header.")};

pub static KEY_IN_USE_ERROR: GenericError = GenericError {
    error: "keyexists",
    message: Some("The provided shortcode is already in use. Delete it or change code.")};
