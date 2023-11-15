/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

use std::{process::{Command, ExitStatus}, collections::HashMap, os::{unix::process::ExitStatusExt, linux::raw::stat}};
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
    println!("SYMPHONY PICCOLO ver 0.0.1");
    loop {
        println!("reconciling...");
        let token = auth();
        if token != "" {
            print!("get desired state >>> ");
            let catalogs = get_catalogs(&token);
            for catalog in catalogs {
                if catalog.spec.catalog_type == "staged" {
                    if let Some(components) = catalog.spec.properties.components {                
                        for component in components {
                            print!("reconcil {} >>> ", component.name);
                            let status: ExitStatus = ExitStatus::from_raw(0);
                            match component.component_type.as_str() {
                                "docker" => {
                                    let status = deploy_docker(&component);
                                    if status.success() {
                                        println!("done");
                                    } else {
                                        println!("failed");
                                    }
                                },
                                _ => {
                                    println!("skipped");
                                }
                            }
                            if status.success() {
                                println!("done");
                            } else {
                                println!("failed");
                            }
                        }
                    } else {
                        println!("No components found in catalog {}", catalog.spec.name);
                    }
                }
            }
        }
        std::thread::sleep(std::time::Duration::from_secs(15));
    }   
}
fn deploy_docker(component: &ComponentSpec) -> ExitStatus {
     //check if container is running
     let output = Command::new("docker")
     .arg("ps")
     .arg(format!("--filter=name={}", component.name))   
     .arg("--format")
     .arg("{{.Names}}")
     .output();

     if output.is_ok() && output.unwrap().stdout.len() > 0 {
         println!("skipped");
         return ExitStatus::from_raw(0);
     }
     
     let mut cmd = Command::new("docker");

     cmd.arg("run");
     if component.properties.as_ref().unwrap().contains_key("container.runtime") {
        let runtime = component.properties.as_ref().unwrap().get("container.runtime").unwrap();
        cmd.arg("--runtime").arg(runtime);       
     }
     cmd.arg("-d")
     .arg("--name")
     .arg(component.name.clone())
     .arg(component.properties.as_ref().unwrap().get("container.image").unwrap())
     .spawn()
     .expect("failed to execute command");

     cmd.spawn().expect("failed to wait on child").wait().expect("failed to wait on child")
}

fn get_catalogs(token: &str) -> Vec<CatalogState> {
    let req = attohttpc::get("http://52.188.128.127:8080/v1alpha2/catalogs/registry").bearer_auth(token).send();
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
    let req = attohttpc::post("http://52.188.128.127:8080/v1alpha2/users/auth").json(&body);
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