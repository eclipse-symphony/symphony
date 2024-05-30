/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

use std::collections::HashMap;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
pub struct Token {
    #[serde(rename = "accessToken")]
    pub access_token: String,
    #[serde(rename = "tokenType")]
    token_type: String,
    #[serde(rename = "username")]
    user_name: String,
    roles: Option<Vec<String>>
}
#[derive(Serialize, Deserialize)]
pub struct ComponentSpec {
    pub name: String,
    pub properties: Option<HashMap<String, String>>,
    #[serde(rename = "type")]
    pub component_type: String,
}
#[derive(Serialize, Deserialize)]
pub struct ObjectRef {
    #[serde(rename = "siteId")]
    site_id: String,
    name: String,
    group: String,
    version: String,
    kind: String,
    namespace: String,
}
#[derive(Serialize, Deserialize)]
pub struct StagedProperties {
    pub components: Option<Vec<ComponentSpec>>,
    #[serde(rename = "removed-components")]
    removed_components: Option<Vec<ComponentSpec>>,
}
#[derive(Serialize, Deserialize)]
pub struct CatalogSpec {
    #[serde(rename = "type")]
    pub catalog_type: String,
    pub properties: StagedProperties,
    #[serde(rename = "objectRef")]
    object_ref: Option<ObjectRef>,
    generation: String,
}
#[derive(Serialize, Deserialize)]
pub struct CatalogStatus {
    properties: Option<HashMap<String, String>>,
}
#[derive(Serialize, Deserialize)]
pub struct ObjectMeta {
    namespace: Option<String>,
    pub name: String,
    labels: Option<HashMap<String, String>>,
    annotations: Option<HashMap<String, String>>,
}
#[derive(Serialize, Deserialize)]
pub struct CatalogState {
    pub metadata: ObjectMeta,
    pub spec: CatalogSpec,
    status: Option<CatalogStatus>,
}