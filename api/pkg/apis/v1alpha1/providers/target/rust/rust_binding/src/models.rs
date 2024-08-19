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

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(rename_all = "camelCase")]
pub struct ComponentSpec {
    pub name: String,
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
    pub route_type: String,
    pub properties: Option<HashMap<String, String>>,
    pub filters: Option<Vec<FilterSpec>>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(rename_all = "camelCase")]
pub struct SidecarSpec {
    pub name: Option<String>,
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
    pub target: String,
    pub components: Vec<ComponentStep>,
    pub role: String,
    pub is_first: bool,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(rename_all = "camelCase")]
pub enum ComponentAction {
    Start,
    Stop,
    Restart,
    Update,
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
