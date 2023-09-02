use std::{process::Command, collections::HashMap};
use serde_json::json;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
struct Token {
    #[serde(rename = "accessToken")]
    access_token: String,
    #[serde(rename = "tokenType")]
    token_type: String,
    #[serde(rename = "username")]
    user_name: String,
    roles: Option<Vec<String>>
}

#[derive(Serialize, Deserialize)]
struct ComponentSpec {
    name: String,
    properties: Option<HashMap<String, String>>,
    #[serde(rename = "type")]
    component_type: String,
}

#[derive(Serialize, Deserialize)]
struct ObjectRef {
    #[serde(rename = "siteId")]
    site_id: String,
    name: String,
    group: String,
    version: String,
    kind: String,
    scope: String,
}

#[derive(Serialize, Deserialize)]
struct StagedProperties {
    components: Option<Vec<ComponentSpec>>,
    #[serde(rename = "removed-components")]
    removed_components: Option<Vec<ComponentSpec>>,
}

#[derive(Serialize, Deserialize)]
struct CatalogSpec {
    #[serde(rename = "siteId")]
    site_id: String,
    name: String,
    #[serde(rename = "type")]
    catalog_type: String,
    properties: StagedProperties,
    #[serde(rename = "objectRef")]
    object_ref: Option<ObjectRef>,
    generation: String,
}

#[derive(Serialize, Deserialize)]
struct CatalogStatus {
    properties: Option<HashMap<String, String>>,
}

#[derive(Serialize, Deserialize)]
struct CatalogState {
    id: String,
    spec: CatalogSpec,
    status: Option<CatalogStatus>,
}
fn main()  {
    let token = auth();
    let catalogs = getCatalogs(&token);
    for catalog in catalogs {
        for component in catalog.spec.properties.components.unwrap() {
            let _cmd = Command::new("docker")
            .arg("run")
            .arg("-d")
            .arg("--name")
            .arg(component.name)
            .arg(component.properties.unwrap().get("container.image").unwrap())
            .spawn()
            .expect("failed to execute command");
        }
    }
    // let _child = Command::new("ls")
    // .arg("-a")
    // .spawn()
    // .expect("failed to execute command");    
}
fn getCatalogs(token: &str) -> Vec<CatalogState> {
    let req = attohttpc::get("http://localhost:8080/v1alpha2/catalogs/registry").bearer_auth(token).send();
    if req.is_err() {
        return vec![];
    }
    let resp = req.unwrap();
    if resp.is_success() {        
        let catalogs = resp.json::<Vec<CatalogState>>();
        if catalogs.is_err() {
            println!("catalogs error: {:?}", catalogs.err().unwrap());
            return vec![];
        }
        return catalogs.unwrap();
    }
    vec![]
}
fn auth() -> String {
    let body = json!({
        "username": "admin",
        "password": ""
    });
    let req = attohttpc::post("http://localhost:8080/v1alpha2/users/auth").json(&body);
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