/// *
/// A message containing a request to the Ankaios server to update the state or to request the complete state of the Ankaios system.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Request {
    #[prost(string, tag = "1")]
    pub request_id: ::prost::alloc::string::String,
    #[prost(oneof = "request::RequestContent", tags = "2, 3")]
    pub request_content: ::core::option::Option<request::RequestContent>,
}
/// Nested message and enum types in `Request`.
pub mod request {
    #[derive(serde::Deserialize, serde::Serialize)]
    #[serde(rename_all = "camelCase")]
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum RequestContent {
        /// / A message to Ankaios server to update the state of one or more agent(s).
        #[prost(message, tag = "2")]
        UpdateStateRequest(::prost::alloc::boxed::Box<super::UpdateStateRequest>),
        /// / A message to Ankaios server to request the complete state by the given request id and the optional field mask.
        #[prost(message, tag = "3")]
        CompleteStateRequest(super::CompleteStateRequest),
    }
}
/// *
/// A message containing a response from the Ankaios server to a particular request.
/// The response content depends on the request content previously sent to the Ankaios server.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Response {
    #[prost(string, tag = "1")]
    pub request_id: ::prost::alloc::string::String,
    #[prost(oneof = "response::ResponseContent", tags = "3, 4, 5")]
    pub response_content: ::core::option::Option<response::ResponseContent>,
}
/// Nested message and enum types in `Response`.
pub mod response {
    #[derive(serde::Deserialize, serde::Serialize)]
    #[serde(rename_all = "camelCase")]
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum ResponseContent {
        #[prost(message, tag = "3")]
        Error(super::Error),
        #[prost(message, tag = "4")]
        CompleteState(super::CompleteState),
        #[prost(message, tag = "5")]
        UpdateStateSuccess(super::UpdateStateSuccess),
    }
}
/// *
/// A message containing a request for the complete/partial state of the Ankaios system.
/// This is usually answered with a \[`CompleteState`\](#completestate) message.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CompleteStateRequest {
    /// / A list of symbolic field paths within the State message structure e.g. 'desiredState.workloads.nginx'.
    #[prost(string, repeated, tag = "1")]
    pub field_mask: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// *
/// A message containing a request to update the state of the Ankaios system.
/// The new state is provided as state object.
/// To specify which part(s) of the new state object should be updated
/// a list of update mask (same as field mask) paths needs to be provided.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UpdateStateRequest {
    /// / The new state of the Ankaios system.
    #[prost(message, optional, tag = "1")]
    pub new_state: ::core::option::Option<CompleteState>,
    /// / A list of symbolic field paths within the state message structure e.g. 'desiredState.workloads.nginx' to specify what to be updated.
    #[prost(string, repeated, tag = "2")]
    pub update_mask: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// *
/// A message from the server containing the ids of the workloads that have been started and stopped in response to a previously sent UpdateStateRequest.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UpdateStateSuccess {
    /// / Workload istance names of workloads which will be started
    #[prost(string, repeated, tag = "1")]
    pub added_workloads: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// / Workload instance names of workloads which will be stopped
    #[prost(string, repeated, tag = "2")]
    pub deleted_workloads: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// *
/// A message containing the complete state of the Ankaios system.
/// This is a response to the \[CompleteStateRequest\](#completestaterequest) message.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CompleteState {
    /// / The state the user wants to reach.
    #[prost(message, optional, tag = "1")]
    pub desired_state: ::core::option::Option<State>,
    /// / The current execution states of the workloads.
    #[prost(message, optional, tag = "2")]
    pub workload_states: ::core::option::Option<WorkloadStatesMap>,
    /// / The agents currently connected to the Ankaios cluster.
    #[prost(message, optional, tag = "3")]
    pub agents: ::core::option::Option<AgentMap>,
}
/// *
/// A nested map that provides the execution state of a workload in a structured way.
/// The first level allows searches by agent.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkloadStatesMap {
    #[prost(map = "string, message", tag = "1")]
    #[serde(flatten)]
    pub agent_state_map: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        ExecutionsStatesOfWorkload,
    >,
}
/// *
/// A map providing the execution state of a workload for a given name.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionsStatesOfWorkload {
    #[prost(map = "string, message", tag = "1")]
    #[serde(flatten)]
    pub wl_name_state_map: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        ExecutionsStatesForId,
    >,
}
/// *
/// A map providing the execution state of a specific workload for a given id.
/// This level is needed as a workload could be running more than once on one agent in different versions.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionsStatesForId {
    #[prost(map = "string, message", tag = "1")]
    #[serde(flatten)]
    pub id_state_map: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        ExecutionState,
    >,
}
/// *
/// A message containing information about the detailed state of a workload in the Ankaios system.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExecutionState {
    /// / The additional info contains more detailed information from the runtime regarding the execution state.
    #[prost(string, tag = "1")]
    pub additional_info: ::prost::alloc::string::String,
    #[prost(
        oneof = "execution_state::ExecutionStateEnum",
        tags = "2, 3, 4, 5, 6, 7, 8, 9"
    )]
    #[serde(flatten)]
    pub execution_state_enum: ::core::option::Option<
        execution_state::ExecutionStateEnum,
    >,
}
/// Nested message and enum types in `ExecutionState`.
pub mod execution_state {
    #[derive(serde::Deserialize, serde::Serialize)]
    #[serde(rename_all = "camelCase")]
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum ExecutionStateEnum {
        /// / The exact state of the workload cannot be determined, e.g., because of a broken connection to the responsible agent.
        #[prost(enumeration = "super::AgentDisconnected", tag = "2")]
        AgentDisconnected(i32),
        /// / The workload is going to be started eventually.
        #[prost(enumeration = "super::Pending", tag = "3")]
        Pending(i32),
        /// / The workload is operational.
        #[prost(enumeration = "super::Running", tag = "4")]
        Running(i32),
        /// / The workload is scheduled for stopping.
        #[prost(enumeration = "super::Stopping", tag = "5")]
        Stopping(i32),
        /// / The workload has successfully finished its operation.
        #[prost(enumeration = "super::Succeeded", tag = "6")]
        Succeeded(i32),
        /// / The workload has failed or is in a degraded state.
        #[prost(enumeration = "super::Failed", tag = "7")]
        Failed(i32),
        /// / The workload is not scheduled to run at any agent. This is signalized with an empty agent in the workload specification.
        #[prost(enumeration = "super::NotScheduled", tag = "8")]
        NotScheduled(i32),
        /// / The workload was removed from Ankaios. This state is used only internally in Ankaios. The outside world removed states are just not there.
        #[prost(enumeration = "super::Removed", tag = "9")]
        Removed(i32),
    }
}
/// *
/// A nested map that provides the names of the connected agents and their optional attributes.
/// The first level allows searches by agent name.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AgentMap {
    #[prost(map = "string, message", tag = "1")]
    #[serde(flatten)]
    pub agents: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        AgentAttributes,
    >,
}
/// *
/// A message containing the CPU usage information of the agent.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CpuUsage {
    /// expressed in percent, the formula for calculating: cpu_usage = (new_work_time - old_work_time) / (new_total_time - old_total_time) * 100
    #[prost(uint32, tag = "1")]
    pub cpu_usage: u32,
}
/// *
/// A message containing the amount of free memory of the agent.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FreeMemory {
    /// expressed in bytes
    #[prost(uint64, tag = "1")]
    pub free_memory: u64,
}
/// *
/// A message that contains attributes of the agent.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AgentAttributes {
    /// / The cpu usage of the agent.
    #[prost(message, optional, tag = "1")]
    pub cpu_usage: ::core::option::Option<CpuUsage>,
    /// / The amount of free memory of the agent.
    #[prost(message, optional, tag = "2")]
    pub free_memory: ::core::option::Option<FreeMemory>,
}
/// *
/// A message containing the information about the workload state.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkloadState {
    #[prost(message, optional, tag = "1")]
    pub instance_name: ::core::option::Option<WorkloadInstanceName>,
    /// / The workload execution state.
    #[prost(message, optional, tag = "2")]
    pub execution_state: ::core::option::Option<ExecutionState>,
}
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkloadInstanceName {
    /// / The name of the workload.
    #[prost(string, tag = "1")]
    pub workload_name: ::prost::alloc::string::String,
    /// / The name of the owning Agent.
    #[prost(string, tag = "2")]
    pub agent_name: ::prost::alloc::string::String,
    /// A unique identifier of the workload.
    #[prost(string, tag = "3")]
    pub id: ::prost::alloc::string::String,
}
/// *
/// A message containing the state information.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct State {
    /// / The current version of the API.
    #[prost(string, tag = "1")]
    pub api_version: ::prost::alloc::string::String,
    /// / A mapping from workload names to workload configurations.
    #[prost(message, optional, tag = "2")]
    pub workloads: ::core::option::Option<WorkloadMap>,
    /// / Configuration values which can be referenced in workload configurations.
    #[prost(message, optional, tag = "3")]
    pub configs: ::core::option::Option<ConfigMap>,
}
/// *
/// This is a workaround for proto not supporing optional maps
/// Workload names shall not be shorter than 1 symbol longer then 63 symbols and can contain only regular characters, digits, the "-" and "_" symbols.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkloadMap {
    #[prost(map = "string, message", tag = "1")]
    #[serde(flatten)]
    pub workloads: ::std::collections::HashMap<::prost::alloc::string::String, Workload>,
}
/// *
/// A message containing the configuration of a workload.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Workload {
    /// / The name of the owning Agent.
    #[prost(string, optional, tag = "1")]
    pub agent: ::core::option::Option<::prost::alloc::string::String>,
    /// / An enum value that defines the condition under which a workload is restarted.
    #[prost(enumeration = "RestartPolicy", optional, tag = "2")]
    pub restart_policy: ::core::option::Option<i32>,
    /// / A map of workload names and expected states to enable a synchronized start of the workload.
    #[prost(message, optional, tag = "3")]
    #[serde(flatten)]
    pub dependencies: ::core::option::Option<Dependencies>,
    /// / A list of tag names.
    #[prost(message, optional, tag = "4")]
    #[serde(flatten)]
    pub tags: ::core::option::Option<Tags>,
    /// / The name of the runtime e.g. podman.
    #[prost(string, optional, tag = "5")]
    pub runtime: ::core::option::Option<::prost::alloc::string::String>,
    /// / The configuration information specific to the runtime.
    #[prost(string, optional, tag = "6")]
    pub runtime_config: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(message, optional, tag = "7")]
    pub control_interface_access: ::core::option::Option<ControlInterfaceAccess>,
    /// / A mapping containing the configurations assigned to the workload.
    #[prost(message, optional, tag = "8")]
    #[serde(flatten)]
    pub configs: ::core::option::Option<ConfigMappings>,
}
/// *
/// This is a workaround for proto not supporing optional repeated values
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Tags {
    #[prost(message, repeated, tag = "1")]
    pub tags: ::prost::alloc::vec::Vec<Tag>,
}
/// *
/// This is a workaround for proto not supporing optional maps
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Dependencies {
    #[prost(map = "string, enumeration(AddCondition)", tag = "1")]
    pub dependencies: ::std::collections::HashMap<::prost::alloc::string::String, i32>,
}
/// *
/// A message to store a tag.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Tag {
    /// / The key of the tag.
    #[prost(string, tag = "1")]
    pub key: ::prost::alloc::string::String,
    /// / The value of the tag.
    #[prost(string, tag = "2")]
    pub value: ::prost::alloc::string::String,
}
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Error {
    #[prost(string, tag = "1")]
    pub message: ::prost::alloc::string::String,
}
/// *
/// A message containing the parts of the control interface the workload as authorized to access.
/// By default, all access is denied.
/// Only if a matching allow rule is found, and no matching deny rules is found, the access is allowed.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ControlInterfaceAccess {
    /// Rules allow the access
    #[prost(message, repeated, tag = "1")]
    #[serde(with = "serde_yaml::with::singleton_map_recursive")]
    #[serde(default)]
    pub allow_rules: ::prost::alloc::vec::Vec<AccessRightsRule>,
    /// Rules denying the access
    #[prost(message, repeated, tag = "2")]
    #[serde(with = "serde_yaml::with::singleton_map_recursive")]
    #[serde(default)]
    pub deny_rules: ::prost::alloc::vec::Vec<AccessRightsRule>,
}
/// *
/// A message containing an allow or deny rule.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AccessRightsRule {
    #[prost(oneof = "access_rights_rule::AccessRightsRuleEnum", tags = "1")]
    pub access_rights_rule_enum: ::core::option::Option<
        access_rights_rule::AccessRightsRuleEnum,
    >,
}
/// Nested message and enum types in `AccessRightsRule`.
pub mod access_rights_rule {
    #[derive(serde::Deserialize, serde::Serialize)]
    #[serde(rename_all = "camelCase")]
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum AccessRightsRuleEnum {
        /// Rule for getting or setting the state
        #[prost(message, tag = "1")]
        StateRule(super::StateRule),
    }
}
/// *
/// Message containing a rule for getting or setting the state
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StateRule {
    /// Defines which actions are allowed
    #[prost(enumeration = "ReadWriteEnum", tag = "1")]
    pub operation: i32,
    /// Pathes definind what can be accessed. Segements of path can be a wildcare "*".
    #[prost(string, repeated, tag = "2")]
    pub filter_masks: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// *
/// This is a workaround for proto not supporing optional maps
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ConfigMappings {
    #[prost(map = "string, string", tag = "1")]
    pub configs: ::std::collections::HashMap<
        ::prost::alloc::string::String,
        ::prost::alloc::string::String,
    >,
}
/// *
/// This is a workaround for proto not supporing optional maps
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ConfigMap {
    #[prost(map = "string, message", tag = "1")]
    #[serde(flatten)]
    pub configs: ::std::collections::HashMap<::prost::alloc::string::String, ConfigItem>,
}
/// *
/// An enum type describing possible configuration objects.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[serde(into = "serde_yaml::Value")]
#[serde(try_from = "serde_yaml::Value")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ConfigItem {
    #[prost(oneof = "config_item::ConfigItem", tags = "1, 2, 3")]
    pub config_item: ::core::option::Option<config_item::ConfigItem>,
}
/// Nested message and enum types in `ConfigItem`.
pub mod config_item {
    #[derive(serde::Deserialize, serde::Serialize)]
    #[serde(rename_all = "camelCase")]
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum ConfigItem {
        #[prost(string, tag = "1")]
        String(::prost::alloc::string::String),
        #[prost(message, tag = "2")]
        Array(super::ConfigArray),
        #[prost(message, tag = "3")]
        Object(super::ConfigObject),
    }
}
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ConfigArray {
    #[prost(message, repeated, tag = "1")]
    pub values: ::prost::alloc::vec::Vec<ConfigItem>,
}
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ConfigObject {
    #[prost(map = "string, message", tag = "1")]
    pub fields: ::std::collections::HashMap<::prost::alloc::string::String, ConfigItem>,
}
/// *
/// An enum type describing the expected workload state. Used for dependency management.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum AddCondition {
    /// / The workload is operational.
    AddCondRunning = 0,
    /// / The workload has successfully exited.
    AddCondSucceeded = 1,
    /// / The workload has exited with an error or could not be started.
    AddCondFailed = 2,
}
impl AddCondition {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            AddCondition::AddCondRunning => "ADD_COND_RUNNING",
            AddCondition::AddCondSucceeded => "ADD_COND_SUCCEEDED",
            AddCondition::AddCondFailed => "ADD_COND_FAILED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "ADD_COND_RUNNING" => Some(Self::AddCondRunning),
            "ADD_COND_SUCCEEDED" => Some(Self::AddCondSucceeded),
            "ADD_COND_FAILED" => Some(Self::AddCondFailed),
            _ => None,
        }
    }
}
/// *
/// The workload was removed from Ankaios. This state is used only internally in Ankaios. The outside world removed states are just not there.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum Removed {
    Removed = 0,
}
impl Removed {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            Removed::Removed => "REMOVED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "REMOVED" => Some(Self::Removed),
            _ => None,
        }
    }
}
/// *
/// The exact state of the workload cannot be determined, e.g., because of a broken connection to the responsible agent.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum AgentDisconnected {
    AgentDisconnected = 0,
}
impl AgentDisconnected {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            AgentDisconnected::AgentDisconnected => "AGENT_DISCONNECTED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "AGENT_DISCONNECTED" => Some(Self::AgentDisconnected),
            _ => None,
        }
    }
}
/// *
/// The workload is not scheduled to run at any agent. This is signalized with an empty agent in the workload specification.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum NotScheduled {
    NotScheduled = 0,
}
impl NotScheduled {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            NotScheduled::NotScheduled => "NOT_SCHEDULED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "NOT_SCHEDULED" => Some(Self::NotScheduled),
            _ => None,
        }
    }
}
/// *
/// The workload is going to be started eventually.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum Pending {
    /// / The workload specification has not yet being scheduled
    Initial = 0,
    /// / The start of the workload will be triggered once all its dependencies are met.
    WaitingToStart = 1,
    /// / Starting the workload was scheduled at the corresponding runtime.
    Starting = 2,
    /// / The starting of the workload by the runtime failed.
    StartingFailed = 8,
}
impl Pending {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            Pending::Initial => "PENDING_INITIAL",
            Pending::WaitingToStart => "PENDING_WAITING_TO_START",
            Pending::Starting => "PENDING_STARTING",
            Pending::StartingFailed => "PENDING_STARTING_FAILED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "PENDING_INITIAL" => Some(Self::Initial),
            "PENDING_WAITING_TO_START" => Some(Self::WaitingToStart),
            "PENDING_STARTING" => Some(Self::Starting),
            "PENDING_STARTING_FAILED" => Some(Self::StartingFailed),
            _ => None,
        }
    }
}
/// *
/// The workload is operational.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum Running {
    /// / The workload is operational.
    Ok = 0,
}
impl Running {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            Running::Ok => "RUNNING_OK",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "RUNNING_OK" => Some(Self::Ok),
            _ => None,
        }
    }
}
/// *
/// The workload is scheduled for stopping.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum Stopping {
    /// / The workload is being stopped.
    Stopping = 0,
    /// / The deletion of the workload will be triggered once neither 'pending' nor 'running' workload depending on it exists.
    WaitingToStop = 1,
    /// / This is an Ankaios generated state returned when the stopping was explicitly trigged by the user and the request was sent to the runtime.
    RequestedAtRuntime = 2,
    /// / The deletion of the workload by the runtime failed.
    DeleteFailed = 8,
}
impl Stopping {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            Stopping::Stopping => "STOPPING",
            Stopping::WaitingToStop => "STOPPING_WAITING_TO_STOP",
            Stopping::RequestedAtRuntime => "STOPPING_REQUESTED_AT_RUNTIME",
            Stopping::DeleteFailed => "STOPPING_DELETE_FAILED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "STOPPING" => Some(Self::Stopping),
            "STOPPING_WAITING_TO_STOP" => Some(Self::WaitingToStop),
            "STOPPING_REQUESTED_AT_RUNTIME" => Some(Self::RequestedAtRuntime),
            "STOPPING_DELETE_FAILED" => Some(Self::DeleteFailed),
            _ => None,
        }
    }
}
/// *
/// The workload has successfully finished operation.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum Succeeded {
    /// / The workload has successfully finished operation.
    Ok = 0,
}
impl Succeeded {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            Succeeded::Ok => "SUCCEEDED_OK",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "SUCCEEDED_OK" => Some(Self::Ok),
            _ => None,
        }
    }
}
/// *
/// The workload has failed or is in a degraded state.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum Failed {
    /// / The workload has failed during operation
    ExecFailed = 0,
    /// / The workload is in an unsupported by Ankaios runtime state. The workload was possibly altered outside of Ankaios.
    Unknown = 1,
    /// / The workload cannot be found anymore. The workload was possibly altered outside of Ankaios or was auto-removed by the runtime.
    Lost = 2,
}
impl Failed {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            Failed::ExecFailed => "FAILED_EXEC_FAILED",
            Failed::Unknown => "FAILED_UNKNOWN",
            Failed::Lost => "FAILED_LOST",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "FAILED_EXEC_FAILED" => Some(Self::ExecFailed),
            "FAILED_UNKNOWN" => Some(Self::Unknown),
            "FAILED_LOST" => Some(Self::Lost),
            _ => None,
        }
    }
}
/// *
/// An enum type describing the restart behavior of a workload.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum RestartPolicy {
    /// / The workload is never restarted. Once the workload exits, it remains in the exited state.
    Never = 0,
    /// / If the workload exits with a non-zero exit code, it will be restarted.
    OnFailure = 1,
    /// / The workload is restarted upon termination, regardless of the exit code.
    Always = 2,
}
impl RestartPolicy {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            RestartPolicy::Never => "NEVER",
            RestartPolicy::OnFailure => "ON_FAILURE",
            RestartPolicy::Always => "ALWAYS",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "NEVER" => Some(Self::Never),
            "ON_FAILURE" => Some(Self::OnFailure),
            "ALWAYS" => Some(Self::Always),
            _ => None,
        }
    }
}
/// *
/// An enum type describing which action is allowed.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum ReadWriteEnum {
    /// Allow nothing
    RwNothing = 0,
    /// Allow read
    RwRead = 1,
    /// Allow write
    RwWrite = 2,
    /// Allow read and write
    RwReadWrite = 5,
}
impl ReadWriteEnum {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            ReadWriteEnum::RwNothing => "RW_NOTHING",
            ReadWriteEnum::RwRead => "RW_READ",
            ReadWriteEnum::RwWrite => "RW_WRITE",
            ReadWriteEnum::RwReadWrite => "RW_READ_WRITE",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "RW_NOTHING" => Some(Self::RwNothing),
            "RW_READ" => Some(Self::RwRead),
            "RW_WRITE" => Some(Self::RwWrite),
            "RW_READ_WRITE" => Some(Self::RwReadWrite),
            _ => None,
        }
    }
}
