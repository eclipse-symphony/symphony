/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

use serde_json::json;
use crate::models::{Token, CatalogVersionState};

static BASE_URL: &str = "http://localhost:8080/v1alpha2";

pub fn get_catalogversions(token: &str) -> Vec<CatalogVersionState> {
    let req = attohttpc::get(format!("{}/catalogversions/registry", BASE_URL)).bearer_auth(token).send();
    if req.is_err() {
        return vec![];
    }
    let resp = req.unwrap();
    if resp.is_success() {        
        let catalogversions = resp.json::<Vec<CatalogVersionState>>();
        if catalogversions.is_err() {
            println!("catalogversions error: {:?}", catalogversions.err().unwrap());
            return vec![];
        }
        return catalogversions.unwrap();
    }
    vec![]
}
pub fn auth() -> String {
    let body = json!({
        "username": "admin",
        "password": ""
    });
    let req = attohttpc::post(format!("{}/users/auth", BASE_URL)).json(&body);
    if req.is_err() {
        return "".to_string();
    }
    let resp = req.unwrap().send();

    if resp.is_err() {
        return "".to_string();
    }
    let resp = resp.unwrap();
    if resp.is_success() {
        let token = resp.json::<Token>();
        if token.is_err() {
            println!("token error: {:?}", token.err().unwrap());
            return "".to_string();
        }
        return token.unwrap().access_token;
    }
    "".to_string()
}