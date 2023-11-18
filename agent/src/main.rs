// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
// SPDX-License-Identifier: MIT

mod models;
mod symphony_api;

use std::{process::{Command, ExitStatus}, collections::HashMap, os::unix::process::ExitStatusExt};
use std::sync::Mutex;
use models::ComponentSpec;
use crate::symphony_api::{auth, get_catalogs};

lazy_static::lazy_static! {
    static ref PROCESS_HASHMAP: Mutex<HashMap<String, String>> = Mutex::new(HashMap::new());
}

fn main()  {
    println!("+============================+");
    println!("| SYMPHONY PICCOLO ver 0.0.1 |");
    println!("+============================+\n");
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
                                    let status = deploy_docker(&catalog.spec.name, &component);
                                    if status.success() {
                                        println!("Docker deployment is done.");
                                    } else {
                                        println!("failed");
                                    }
                                },
                                "wasm" => {
                                    let status = deploy_wasmedge(&catalog.spec.name, &component);
                                    if status.success() {
                                        println!("WASM deployment is done.");
                                    } else {
                                        println!("failed");
                                    }
                                },
                                "ebpf" => {
                                    let status = deploy_ebpf(&catalog.spec.name, &component);
                                    if status.success() {
                                        println!("eBPF deployment is done.");
                                    } else {
                                        println!("failed");
                                    }
                                }
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
fn deploy_ebpf(_name: &str, component: &ComponentSpec) -> ExitStatus {
    //let key = format!("{}-{}", name, component.name);
    let output = Command::new("bpftool")
    .arg("prog").arg("show").arg("name").arg(component.name.clone())
    .output();
    if output.is_ok() && output.unwrap().stdout.len() > 0 {
        println!("skipped");
        return ExitStatus::from_raw(0);
    }
    
    let file_name = match download_file(component.properties.as_ref().unwrap().get("ebpf.url").unwrap()) {
        Ok(value) => value,
        Err(value) => return value,
    };

    println!("loading {}...", file_name);

    let output = Command::new("bpftool")
    .arg("prog").arg("load").arg(file_name).arg(format!("/sys/fs/bpf/{}", component.name.clone()))
    .output();
    if !output.is_ok() {        
        return ExitStatus::from_raw(1);
    }

    if component.properties.as_ref().unwrap().contains_key("ebpf.event") {
        let event = component.properties.as_ref().unwrap().get("ebpf.event").unwrap();
        let output = Command::new("bpftool")
        .arg("net").arg("attach").arg(event).arg("name").arg(component.name.clone()).arg("dev").arg("eth0")
        .output();
        if !output.is_ok() {        
            return ExitStatus::from_raw(1);
        }
    }
    
    return ExitStatus::from_raw(0);
}
fn deploy_wasmedge(name :&str, component: &ComponentSpec) -> ExitStatus {
    let key = format!("{}-{}", name, component.name);

    if check_process(&key) {
        println!("skipped");
        return ExitStatus::from_raw(0);
    }
    
    let file_name = match download_file(component.properties.as_ref().unwrap().get("wasm.url").unwrap()) {
        Ok(value) => value,
        Err(value) => return value,
    };

    let mut cmd = Command::new("wasmedge");
    if component.properties.as_ref().unwrap().contains_key("wasm.dir") {
        let dir = component.properties.as_ref().unwrap().get("wasm.dir").unwrap();
        cmd.arg("--dir").arg(dir);  
    }
    cmd.arg(file_name);

    launch_process(cmd, key)
}

fn launch_process(mut cmd: Command, key: String) -> ExitStatus {
    let mut process_hashmap = PROCESS_HASHMAP.lock().unwrap();
    let _ = match cmd.spawn() {
        Ok(child) => {
            let pid = child.id().to_string();
            process_hashmap.insert(key, pid);
            ExitStatus::from_raw(0)
        },
        Err(e) => {
            eprintln!("Failed to launch command: {}", e);
            ExitStatus::from_raw(1)
        }
    };
    ExitStatus::from_raw(0)    
}

fn check_process(key: &String) -> bool {
    let process_hashmap = PROCESS_HASHMAP.lock().unwrap();
    if process_hashmap.contains_key(key.as_str()) {
        let pid = process_hashmap.get(key.as_str()).unwrap();
        let output = Command::new("ps")
        .arg("-p")
        .arg(pid)
        .output();

        if output.is_ok() && output.unwrap().stdout.len() > 0 {
           return true;
        }
    }
    return false;
}

fn download_file(address :&str) -> Result<&str, ExitStatus> {
    let file_name = address.split("/").last().unwrap();
    let output = Command::new("wget")
    .arg("-O")
    .arg(file_name)
    .arg(address)
    .output();
    if output.is_err() {
        return Err(ExitStatus::from_raw(1));
    }
    Ok(file_name)
}

fn deploy_docker(_name: &str, component: &ComponentSpec) -> ExitStatus {
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