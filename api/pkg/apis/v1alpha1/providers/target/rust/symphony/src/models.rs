/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

 use serde::{Deserialize, Serialize};
 use std::collections::HashMap;
 use std::time::SystemTime;
 
 use serde_json::Value;
 
 pub type ProviderConfig = Value;
 
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 pub struct ValidationRule {
     #[serde(rename = "requiredType")]
     pub required_component_type: String,
     #[serde(rename = "componentValidationRule")]
     pub component_validation_rule: ComponentValidationRule,
     #[serde(rename = "sidecarValidationRule")]
     pub sidecar_validation_rule: ComponentValidationRule,
     #[serde(rename = "allowSidecar")]
     pub allow_sidecar: bool,
     #[serde(rename = "supportScopes")]
     pub scope_isolation: bool,
     #[serde(rename = "instanceIsolation")]
     pub instance_isolation: bool,
 }
 
 impl ValidationRule {
    pub fn new() -> Self {
        ValidationRule {
            required_component_type: "".to_string(),
            component_validation_rule: ComponentValidationRule {
                required_component_type: "".to_string(),
                change_detection_properties: vec![
                    PropertyDesc {
                        ignore_case: true,
                        is_component_name: false,
                        name: "*".to_string(),
                        skip_if_missing: true,
                        prefix_match: false,
                    },
                ],
                change_detection_metadata: vec![],
                required_properties: vec![],
                optional_properties: vec![],
                required_metadata: vec![],
                optional_metadata: vec![],
            },
            sidecar_validation_rule: ComponentValidationRule {
                required_component_type: "".to_string(),
                change_detection_properties: vec![],
                change_detection_metadata: vec![],
                required_properties: vec![],
                optional_properties: vec![],
                required_metadata: vec![],
                optional_metadata: vec![],
            },
            allow_sidecar: true,
            scope_isolation: true,
            instance_isolation: true,
        }
    }
}

 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct DeploymentSpec {
     pub solution_name: String,
     pub solution: SolutionState,
     pub instance: InstanceState,
     pub targets: HashMap<String, TargetState>,
     pub devices: Option<Vec<DeviceSpec>>,
     pub assignments: Option<HashMap<String, String>>,
     pub component_start_index: Option<i32>,
     pub component_end_index: Option<i32>,
     pub active_target: Option<String>,
     pub generation: Option<String>,
     pub object_namespace: Option<String>,
     pub hash: Option<String>,
 }
 
 impl DeploymentSpec {
    /// Creates an empty `DeploymentSpec` with default values.
    pub fn empty() -> Self {
        DeploymentSpec {
            solution_name: String::new(),
            solution: SolutionState {
                metadata: ObjectMeta {
                    namespace: None,
                    name: None,
                    generation: None,
                    labels: None,
                    annotations: None,
                },
                spec: None,
            },
            instance: InstanceState {
                metadata: ObjectMeta {
                    namespace: None,
                    name: None,
                    generation: None,
                    labels: None,
                    annotations: None,
                },
                spec: None,
                status: InstanceStatus {}, // Assuming `InstanceStatus` has a default implementation
            },
            targets: HashMap::new(),
            devices: None,
            assignments: None,
            component_start_index: None,
            component_end_index: None,
            active_target: None,
            generation: None,
            object_namespace: None,
            hash: None,
        }
    }
}

 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct ComponentSpec {
     pub name: String,
     #[serde(rename = "type")]
     pub component_type: Option<String>,
     pub metadata: Option<HashMap<String, String>>,
     pub properties: Option<HashMap<String, serde_json::Value>>,
     pub parameters: Option<HashMap<String, String>>,
     pub routes: Option<Vec<RouteSpec>>,
     pub constraints: Option<String>,
     pub dependencies: Option<Vec<String>>,
     pub skills: Option<Vec<String>>,
     pub sidecars: Option<Vec<SidecarSpec>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 pub struct PropertyDesc {
     pub name: String,
     #[serde(rename = "ignoreCase")]
     pub ignore_case: bool,
     #[serde(rename = "skipIfMissing")]
     pub skip_if_missing: bool,
     #[serde(rename = "prefixMatch")]
     pub prefix_match: bool,
     #[serde(rename = "isComponentName")]
     pub is_component_name: bool,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 pub struct ComponentValidationRule {
     #[serde(rename = "requiredType")]
     pub required_component_type: String,
     #[serde(rename = "changeDetection")]
     pub change_detection_properties: Vec<PropertyDesc>,
     #[serde(rename = "changeDetectionMetadata")]
     pub change_detection_metadata: Vec<PropertyDesc>,
     #[serde(rename = "requiredProperties")]
     pub required_properties: Vec<String>,
     #[serde(rename = "optionalProperties")]
     pub optional_properties: Vec<String>,
     #[serde(rename = "requiredMetadata")]
     pub required_metadata: Vec<String>,
     #[serde(rename = "optionalMetadata")]
     pub optional_metadata: Vec<String>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct RouteSpec {
     pub route: String,
     #[serde(rename = "type")]
     pub route_type: String,
     pub properties: Option<HashMap<String, String>>,
     pub filters: Option<Vec<FilterSpec>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct SidecarSpec {
     pub name: Option<String>,
     #[serde(rename = "type")]
     pub sidecar_type: Option<String>,
     pub properties: Option<HashMap<String, serde_json::Value>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct ComponentStep {
     pub action: ComponentAction,
     pub component: ComponentSpec,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct DeploymentStep {
     pub target: Option<String>,
     #[serde(default)]
     pub components: Vec<ComponentStep>,
     pub role: String,
     pub is_first: bool,
 }

//  impl DeploymentStep {
//     // The get_components method that returns a Vec<ComponentSpec>
//     pub fn get_components(&self) -> Vec<ComponentSpec> {
//         let mut ret = Vec::new();
//         for component_step in &self.components {
//             ret.push(component_step.component.clone());
//         }
//         ret
//     }
// }
 
 #[derive(Serialize, Deserialize, Debug, Clone, PartialEq)]
 #[serde(rename_all = "camelCase")]
 pub enum ComponentAction {
     Update,
     Delete
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct SolutionState {
     pub metadata: ObjectMeta,
     pub spec: Option<SolutionSpec>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct SolutionSpec {
     pub display_name: Option<String>,
     pub metadata: Option<HashMap<String, String>>,
     pub components: Option<Vec<ComponentSpec>>,
     pub version: Option<String>,
     pub root_resource: Option<String>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct InstanceState {
     pub metadata: ObjectMeta,
     pub spec: Option<InstanceSpec>,
     pub status: InstanceStatus,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct InstanceSpec {
     pub display_name: Option<String>,
     pub scope: Option<String>,
     pub parameters: Option<HashMap<String, String>>,
     pub metadata: Option<HashMap<String, String>>,
     pub solution: String,
     pub target: Option<TargetSelector>, // Include TargetSelector
     pub topologies: Option<Vec<TopologySpec>>,
     pub pipelines: Option<Vec<PipelineSpec>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct TargetSelector {
     pub name: Option<String>,
     pub selector: Option<HashMap<String, String>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct TargetState {
     pub metadata: ObjectMeta,
     pub status: TargetStatus,
     pub spec: Option<TargetSpec>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct TargetSpec {
     pub display_name: Option<String>,
     pub scope: Option<String>,
     pub metadata: Option<HashMap<String, String>>,
     pub properties: Option<HashMap<String, String>>,
     pub components: Option<Vec<ComponentSpec>>,
     pub constraints: Option<String>,
     pub topologies: Option<Vec<TopologySpec>>,
     pub force_redeploy: Option<bool>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct DeviceSpec {
     pub display_name: Option<String>,
     pub properties: Option<HashMap<String, String>>,
     pub bindings: Option<Vec<BindingSpec>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct BindingSpec {
     pub role: String,
     pub provider: String,
     pub config: Option<HashMap<String, String>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct FilterSpec {
     pub direction: String,
     pub filter_type: String,
     pub parameters: Option<HashMap<String, String>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct TopologySpec {
     pub device: Option<String>,
     pub selector: Option<HashMap<String, String>>,
     pub bindings: Option<Vec<BindingSpec>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct ObjectMeta {
     pub namespace: Option<String>,
     pub name: Option<String>,
     pub generation: Option<String>,
     pub labels: Option<HashMap<String, String>>,
     pub annotations: Option<HashMap<String, String>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct InstanceStatus {
     // Fields for instance status
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct TargetStatus {
     // Fields for target status
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct PipelineSpec {
     pub name: String,
     pub skill: String,
     pub parameters: Option<HashMap<String, String>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct DeployableStatus {
     pub properties: Option<HashMap<String, String>>,
     pub provisioning_status: ProvisioningStatus,
     pub last_modified: Option<SystemTime>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct ProvisioningStatus {
     pub operation_id: String,
     pub status: String,
     pub failure_cause: Option<String>,
     pub log_errors: Option<bool>,
     pub error: Option<ErrorType>,
     pub output: Option<HashMap<String, String>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct ErrorType {
     pub code: Option<String>,
     pub message: Option<String>,
     pub target: Option<String>,
     pub details: Option<Vec<TargetError>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct TargetError {
     pub code: Option<String>,
     pub message: Option<String>,
     pub target: Option<String>,
     pub details: Option<Vec<ComponentError>>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct ComponentError {
     pub code: Option<String>,
     pub message: Option<String>,
     pub target: Option<String>,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone)]
 #[serde(rename_all = "camelCase")]
 pub struct ComponentResultSpec {
     pub status: State,
     pub message: String,
 }
 
 #[derive(Serialize, Deserialize, Debug, Clone, Copy, PartialEq, Eq, Hash)]
 #[serde(into = "u16", from = "u16")]
 pub enum State {
     // HTTP Status codes
     OK = 200,
     Accepted = 202,
     BadRequest = 400,
     Unauthorized = 403,
     NotFound = 404,
     MethodNotAllowed = 405,
     Conflict = 409,
     InternalError = 500,
 
     // Config errors
     BadConfig = 1000,
     MissingConfig = 1001,
 
     // API invocation errors
     InvalidArgument = 2000,
     APIRedirect = 3030,
 
     // IO errors
     FileAccessError = 4000,
 
     // Serialization errors
     SerializationError = 5000,
     DeserializeError = 5001,
 
     // Async requests
     DeleteRequested = 6000,
 
     // Operation results
     UpdateFailed = 8001,
     DeleteFailed = 8002,
     ValidateFailed = 8003,
     Updated = 8004,
     Deleted = 8005,
 
     // Workflow status
     Running = 9994,
     Paused = 9995,
     Done = 9996,
     Delayed = 9997,
     Untouched = 9998,
     NotImplemented = 9999,
 
     // Detailed error codes
     InitFailed = 10000,
     CreateActionConfigFailed = 10001,
     HelmActionFailed = 10002,
     GetComponentSpecFailed = 10003,
     CreateProjectorFailed = 10004,
     K8sRemoveServiceFailed = 10005,
     K8sRemoveDeploymentFailed = 10006,
     K8sDeploymentFailed = 10007,
     ReadYamlFailed = 10008,
     ApplyYamlFailed = 10009,
     ReadResourcePropertyFailed = 10010,
     ApplyResourceFailed = 10011,
     DeleteYamlFailed = 10012,
     DeleteResourceFailed = 10013,
     CheckResourceStatusFailed = 10014,
     ApplyScriptFailed = 10015,
     RemoveScriptFailed = 10016,
     YamlResourcePropertyNotFound = 10017,
     GetHelmPropertyFailed = 10018,
     HelmChartPullFailed = 10019,
     HelmChartLoadFailed = 10020,
     HelmChartApplyFailed = 10021,
     HelmChartUninstallFailed = 10022,
     IngressApplyFailed = 10023,
     HttpNewRequestFailed = 10024,
     HttpSendRequestFailed = 10025,
     HttpErrorResponse = 10026,
     MqttPublishFailed = 10027,
     MqttApplyFailed = 10028,
     MqttApplyTimeout = 10029,
     ConfigMapApplyFailed = 10030,
     HttpBadWaitStatusCode = 10031,
     HttpNewWaitRequestFailed = 10032,
     HttpSendWaitRequestFailed = 10033,
     HttpErrorWaitResponse = 10034,
     HttpBadWaitExpression = 10035,
     ScriptExecutionFailed = 10036,
     ScriptResultParsingFailed = 10037,
     WaitToGetInstancesFailed = 10038,
     WaitToGetSitesFailed = 10039,
     WaitToGetCatalogsFailed = 10040,
     InvalidWaitObjectType = 10041,
     CatalogsGetFailed = 10042,
     InvalidInstanceCatalog = 10043,
     CreateInstanceFromCatalogFailed = 10044,
     InvalidSolutionCatalog = 10045,
     CreateSolutionFromCatalogFailed = 10046,
     InvalidTargetCatalog = 10047,
     CreateTargetFromCatalogFailed = 10048,
     InvalidCatalogCatalog = 10049,
     CreateCatalogFromCatalogFailed = 10050,
     ParentObjectMissing = 10051,
     ParentObjectCreateFailed = 10052,
     MaterializeBatchFailed = 10053,
     DeleteInstanceFailed = 10054,
     CreateInstanceFailed = 10055,
     DeploymentNotReached = 10056,
     InvalidObjectType = 10057,
     UnsupportedAction = 10058,
 
     // Instance controller errors
     SolutionGetFailed = 11000,
     TargetCandidatesNotFound = 11001,
     TargetListGetFailed = 11002,
     ObjectInstanceConversionFailed = 11003,
     TimedOut = 11004,
 
     // Target controller errors
     TargetPropertyNotFound = 12000,
 }
 
 impl Into<u16> for State {
     fn into(self) -> u16 {
         self as u16
     }
 }
 
 impl From<u16> for State {
     fn from(value: u16) -> Self {
         match value {
             // HTTP Status codes
             200 => State::OK,
             202 => State::Accepted,
             400 => State::BadRequest,
             403 => State::Unauthorized,
             404 => State::NotFound,
             405 => State::MethodNotAllowed,
             409 => State::Conflict,
             500 => State::InternalError,
 
             // Config errors
             1000 => State::BadConfig,
             1001 => State::MissingConfig,
 
             // API invocation errors
             2000 => State::InvalidArgument,
             3030 => State::APIRedirect,
 
             // IO errors
             4000 => State::FileAccessError,
 
             // Serialization errors
             5000 => State::SerializationError,
             5001 => State::DeserializeError,
 
             // Async requests
             6000 => State::DeleteRequested,
 
             // Operation results
             8001 => State::UpdateFailed,
             8002 => State::DeleteFailed,
             8003 => State::ValidateFailed,
             8004 => State::Updated,
             8005 => State::Deleted,
 
             // Workflow status
             9994 => State::Running,
             9995 => State::Paused,
             9996 => State::Done,
             9997 => State::Delayed,
             9998 => State::Untouched,
             9999 => State::NotImplemented,
 
             // Detailed error codes
             10000 => State::InitFailed,
             10001 => State::CreateActionConfigFailed,
             10002 => State::HelmActionFailed,
             10003 => State::GetComponentSpecFailed,
             10004 => State::CreateProjectorFailed,
             10005 => State::K8sRemoveServiceFailed,
             10006 => State::K8sRemoveDeploymentFailed,
             10007 => State::K8sDeploymentFailed,
             10008 => State::ReadYamlFailed,
             10009 => State::ApplyYamlFailed,
             10010 => State::ReadResourcePropertyFailed,
             10011 => State::ApplyResourceFailed,
             10012 => State::DeleteYamlFailed,
             10013 => State::DeleteResourceFailed,
             10014 => State::CheckResourceStatusFailed,
             10015 => State::ApplyScriptFailed,
             10016 => State::RemoveScriptFailed,
             10017 => State::YamlResourcePropertyNotFound,
             10018 => State::GetHelmPropertyFailed,
             10019 => State::HelmChartPullFailed,
             10020 => State::HelmChartLoadFailed,
             10021 => State::HelmChartApplyFailed,
             10022 => State::HelmChartUninstallFailed,
             10023 => State::IngressApplyFailed,
             10024 => State::HttpNewRequestFailed,
             10025 => State::HttpSendRequestFailed,
             10026 => State::HttpErrorResponse,
             10027 => State::MqttPublishFailed,
             10028 => State::MqttApplyFailed,
             10029 => State::MqttApplyTimeout,
             10030 => State::ConfigMapApplyFailed,
             10031 => State::HttpBadWaitStatusCode,
             10032 => State::HttpNewWaitRequestFailed,
             10033 => State::HttpSendWaitRequestFailed,
             10034 => State::HttpErrorWaitResponse,
             10035 => State::HttpBadWaitExpression,
             10036 => State::ScriptExecutionFailed,
             10037 => State::ScriptResultParsingFailed,
             10038 => State::WaitToGetInstancesFailed,
             10039 => State::WaitToGetSitesFailed,
             10040 => State::WaitToGetCatalogsFailed,
             10041 => State::InvalidWaitObjectType,
             10042 => State::CatalogsGetFailed,
             10043 => State::InvalidInstanceCatalog,
             10044 => State::CreateInstanceFromCatalogFailed,
             10045 => State::InvalidSolutionCatalog,
             10046 => State::CreateSolutionFromCatalogFailed,
             10047 => State::InvalidTargetCatalog,
             10048 => State::CreateTargetFromCatalogFailed,
             10049 => State::InvalidCatalogCatalog,
             10050 => State::CreateCatalogFromCatalogFailed,
             10051 => State::ParentObjectMissing,
             10052 => State::ParentObjectCreateFailed,
             10053 => State::MaterializeBatchFailed,
             10054 => State::DeleteInstanceFailed,
             10055 => State::CreateInstanceFailed,
             10056 => State::DeploymentNotReached,
             10057 => State::InvalidObjectType,
             10058 => State::UnsupportedAction,
 
             // Instance controller errors
             11000 => State::SolutionGetFailed,
             11001 => State::TargetCandidatesNotFound,
             11002 => State::TargetListGetFailed,
             11003 => State::ObjectInstanceConversionFailed,
             11004 => State::TimedOut,
 
             // Target controller errors
             12000 => State::TargetPropertyNotFound,
 
             _ => State::InternalError, // Default case for unknown values
         }
     }
 }