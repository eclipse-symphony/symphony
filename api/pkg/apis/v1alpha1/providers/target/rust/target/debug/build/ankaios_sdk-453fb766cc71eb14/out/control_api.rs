/// *
/// Messages to the Ankaios server.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ToAnkaios {
    #[prost(oneof = "to_ankaios::ToAnkaiosEnum", tags = "1, 3")]
    pub to_ankaios_enum: ::core::option::Option<to_ankaios::ToAnkaiosEnum>,
}
/// Nested message and enum types in `ToAnkaios`.
pub mod to_ankaios {
    #[derive(serde::Deserialize, serde::Serialize)]
    #[serde(rename_all = "camelCase")]
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum ToAnkaiosEnum {
        /// / The fist message sent when a connection is established. The message is needed to make sure the connected components are compatible.
        #[prost(message, tag = "1")]
        Hello(super::Hello),
        /// / A request to Ankaios
        #[prost(message, tag = "3")]
        Request(super::super::ank_base::Request),
    }
}
/// *
/// This message is the first one that needs to be sent when a new connection to the Ankaios cluster is established. Without this message being sent all further request are rejected.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Hello {
    /// / The protocol version used by the calling component.
    #[prost(string, tag = "2")]
    pub protocol_version: ::prost::alloc::string::String,
}
/// *
/// Messages from the Ankaios server to e.g. the Ankaios agent.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FromAnkaios {
    #[prost(oneof = "from_ankaios::FromAnkaiosEnum", tags = "3, 5")]
    pub from_ankaios_enum: ::core::option::Option<from_ankaios::FromAnkaiosEnum>,
}
/// Nested message and enum types in `FromAnkaios`.
pub mod from_ankaios {
    #[derive(serde::Deserialize, serde::Serialize)]
    #[serde(rename_all = "camelCase")]
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum FromAnkaiosEnum {
        /// / A message containing a response to a previous request.
        #[prost(message, tag = "3")]
        Response(::prost::alloc::boxed::Box<super::super::ank_base::Response>),
        /// / A message sent by Ankaios to inform a workload that the connection to Anakios was closed.
        #[prost(message, tag = "5")]
        ConnectionClosed(super::ConnectionClosed),
    }
}
/// *
/// This message informs the user of the Control Interface that the connection was closed by Ankaios.
/// No more messages will be processed by Ankaios after this message is sent.
#[derive(serde::Deserialize, serde::Serialize)]
#[serde(rename_all = "camelCase")]
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ConnectionClosed {
    /// / A string containing the reason for closing the connection.
    #[prost(string, tag = "1")]
    pub reason: ::prost::alloc::string::String,
}
