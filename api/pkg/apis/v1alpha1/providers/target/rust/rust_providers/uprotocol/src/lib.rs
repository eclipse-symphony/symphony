/*
 * Copyright (c) Contributors to the Eclipse Foundation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

use std::collections::HashMap;
use std::ffi::{c_char, CStr};
use std::ptr;
use std::sync::Arc;
use std::time::Duration;

use backon::{ExponentialBuilder, Retryable};
use serde_json::{json, Value};
use symphony::models::{
    ComponentAction, ComponentResultSpec, ComponentSpec, ComponentStep, DeploymentSpec,
    DeploymentStep, ProviderConfig, ValidationRule,
};
use symphony::{ITargetProvider, ProviderWrapper};
use tokio::runtime::Runtime;
use tracing::{debug, error, info, warn};
use up_rust::communication::{CallOptions, InMemoryRpcClient, RpcClient, UPayload};
use up_rust::local_transport::LocalTransport;
use up_rust::{StaticUriProvider, UCode, UTransport, UUri};
use up_transport_mqtt5::{Mqtt5Transport, Mqtt5TransportOptions, MqttClientOptions};
use up_transport_zenoh::{zenoh_config, UPTransportZenoh};

const CONFIG_PROPERTY_ZENOH_CONFIG_FILE: &str = "zenohConfig";
const CONFIG_PROPERTY_MQTT_BROKER_ADDRESS: &str = "brokerAddress";
const CONFIG_PROPERTY_MQTT_CLIENT_ID: &str = "clientID";
const CONFIG_PROPERTY_LOCAL_ENTITY: &str = "localEntity";
const CONFIG_PROPERTY_GET_METHOD_URI: &str = "getMethodUri";
const CONFIG_PROPERTY_GET_METHOD_TIMEOUT: &str = "getMethodTimeoutMillis";
const CONFIG_PROPERTY_APPLY_METHOD_TIMEOUT: &str = "applyMethodTimeoutMillis";
const CONFIG_PROPERTY_VALIDATION_RULE: &str = "validationRule";

const RESOURCE_ID_GET: u16 = 0x0001;
const RESOURCE_ID_UPDATE: u16 = 0x0002;
const RESOURCE_ID_DELETE: u16 = 0x0003;

const PROPERTY_NAME_DEPLOYMENT: &str = "deployment";
const PROPERTY_NAME_COMPONENTS: &str = "components";

fn get_string_property(provider_config: &ProviderConfig, property_name: &str) -> Option<String> {
    provider_config
        .as_object()
        .and_then(|object| object.get(property_name))
        .and_then(Value::as_str)
        .map(str::to_string)
}

fn extract_string_property(
    provider_config: &ProviderConfig,
    property_name: &str,
) -> Result<String, Box<dyn core::error::Error + Send + Sync>> {
    get_string_property(provider_config, property_name).ok_or_else(|| {
        Box::from(format!(
            "JSON config does not contain required {property_name} String property"
        ))
    })
}

fn get_zenoh_config(
    path: Option<String>,
) -> Result<zenoh_config::Config, Box<dyn core::error::Error>> {
    if let Some(zenoh_config_file) = path {
        debug!("using Zenoh config from file: {}", zenoh_config_file);
        up_transport_zenoh::zenoh_config::Config::from_file(&zenoh_config_file)
            .map_err(|_err| Box::from("failed to create Zenoh config"))
    } else {
        debug!("using default Zenoh config");
        Ok(up_transport_zenoh::zenoh_config::Config::default())
    }
}

async fn get_zenoh_transport(
    local_authority: String,
    path: Option<String>,
) -> Result<Arc<dyn UTransport>, Box<dyn core::error::Error>> {
    let zenoh_config = get_zenoh_config(path)?;
    let builder = UPTransportZenoh::builder(local_authority)?;
    Ok(builder
        .with_config(zenoh_config)
        .build()
        .await
        .inspect_err(|err| error!("failed to create Zenoh transport: {err}"))
        .map(Arc::new)?)
}

async fn get_mqtt5_transport(
    local_authority: String,
    config: &ProviderConfig,
) -> Result<Arc<dyn UTransport>, Box<dyn core::error::Error>> {
    let Some(broker_uri) = get_string_property(&config, CONFIG_PROPERTY_MQTT_BROKER_ADDRESS) else {
        return Err(Box::from(format!(
            "JSON config does not contain required {} String property",
            CONFIG_PROPERTY_MQTT_BROKER_ADDRESS
        )));
    };
    let client_id = get_string_property(&config, CONFIG_PROPERTY_MQTT_CLIENT_ID);
    let transport_options = Mqtt5TransportOptions {
        mode: up_transport_mqtt5::TransportMode::InVehicle,
        mqtt_client_options: MqttClientOptions {
            client_id,
            broker_uri: broker_uri.clone(),
            ..Default::default()
        },
        ..Default::default()
    };
    let transport = Mqtt5Transport::new(transport_options, local_authority)
        .await
        .map(Arc::new)
        .map_err(|err| {
            error!("failed to create MQTT5 transport: {err}");
            Box::new(err)
        })?;

    (|| {
        info!("Connecting to MQTT5 broker...");
        transport.connect()
    })
    .retry(ExponentialBuilder::default().with_total_delay(Some(Duration::from_secs(5))))
    .notify(|error, sleep_duration| {
        error!("{}, retrying in {sleep_duration:?}", error.get_message());
    })
    .when(|err| {
        // no need to keep retrying if authentication or permission is denied
        err.get_code() != UCode::UNAUTHENTICATED && err.get_code() != UCode::PERMISSION_DENIED
    })
    .await?;

    Ok(transport)
}

async fn get_transport(
    local_authority: String,
    config: &ProviderConfig,
) -> Result<Arc<dyn UTransport>, Box<dyn core::error::Error>> {
    match (
        get_string_property(&config, CONFIG_PROPERTY_ZENOH_CONFIG_FILE),
        get_string_property(&config, CONFIG_PROPERTY_MQTT_BROKER_ADDRESS),
    ) {
        (Some(zenoh_config_file), None) => {
            debug!("creating Zenoh transport");
            get_zenoh_transport(local_authority, Some(zenoh_config_file)).await
        }
        (None, Some(_)) => {
            debug!("creating MQTT5 transport");
            get_mqtt5_transport(local_authority, config).await
        }
        (None, None) => {
            warn!("no transport configured, no messages will be sent to remote target provider, using LocalTransport instead");
            Ok(Arc::new(LocalTransport::default()))
        }
        (Some(_), Some(_)) => {
            error!("both Zenoh and MQTT5 transport configured");
            Err(Box::from(
                "both Zenoh and MQTT5 transport configured, only one transport can be used",
            ))
        }
    }
}

/// A builder for configuring and creating a uProtocol target provider.
struct UProtocolTargetProviderBuilder {
    local_uuri: UUri,
    get_method_uuri: UUri,
    get_method_timeout: u32,
    apply_method_timeout: u32,
    validation_rule: Option<String>,
    transport: Arc<dyn UTransport>,
    tokio_runtime: Arc<Runtime>,
}

impl UProtocolTargetProviderBuilder {
    fn new(
        tokio_runtime: Arc<Runtime>,
        transport: Arc<dyn UTransport>,
        local_uuri: UUri,
        get_method_uuri: UUri,
    ) -> Self {
        Self {
            local_uuri,
            get_method_uuri,
            get_method_timeout: 120_000,
            apply_method_timeout: 300_000,
            validation_rule: None,
            transport,
            tokio_runtime,
        }
    }

    /// Sets the timeout for the Get method in milliseconds.
    /// Default is 120000 (2 minutes).
    ///
    /// # Arguments
    /// * `timeout` - The timeout in milliseconds.
    fn with_get_method_timeout(mut self, timeout: u32) -> Self {
        self.get_method_timeout = timeout;
        self
    }

    /// Sets the timeout for the Update and Delete methods in milliseconds.
    /// Default is 300000 (5 minutes).
    /// # Arguments
    /// * `timeout` - The timeout in milliseconds.
    fn with_apply_method_timeout(mut self, timeout: u32) -> Self {
        self.apply_method_timeout = timeout;
        self
    }

    /// Sets the validation rule to be used by the target provider.
    /// If not set, the default validation rule will be used.
    /// # Arguments
    /// * `rule` - The validation rule as a JSON string.
    fn with_validation_rule<S: Into<String>>(mut self, rule: S) -> Self {
        self.validation_rule = Some(rule.into());
        self
    }

    /// Creates the target provider based on the configured options.
    ///
    /// # Errors
    /// Returns an error if any of the configured options are invalid.
    fn build(self) -> Result<UProtocolTargetProvider, Box<dyn core::error::Error>> {
        if !self.local_uuri.is_rpc_response() {
            return Err(Box::from("Local Entity URI must have resource ID 0x0000"));
        }
        debug!("using Local Entity URI: {}", self.local_uuri.to_uri(true));

        if self.get_method_uuri.resource_id() != RESOURCE_ID_GET {
            return Err(Box::from(format!(
                "Get URI must have resource ID {RESOURCE_ID_GET}"
            )));
        }
        debug!(
            "using Get method [URI: {}, timeout: {}ms]",
            self.get_method_uuri.to_uri(true),
            self.get_method_timeout
        );

        let update_method_uuri = UUri::try_from_parts(
            self.get_method_uuri.authority_name().as_str(),
            self.get_method_uuri.ue_id,
            self.get_method_uuri.uentity_major_version(),
            RESOURCE_ID_UPDATE,
        )
        .inspect_err(|err| {
            error!("failed to create Update method URI: {err}");
        })?;
        debug!(
            "using Update method [URI: {}, timeout: {}ms]",
            update_method_uuri.to_uri(true),
            self.apply_method_timeout
        );

        let delete_method_uuri = UUri::try_from_parts(
            self.get_method_uuri.authority_name().as_str(),
            self.get_method_uuri.ue_id,
            self.get_method_uuri.uentity_major_version(),
            RESOURCE_ID_DELETE,
        )
        .inspect_err(|err| {
            error!("failed to create Delete method URI: {err}");
        })?;
        debug!(
            "using Delete method [URI: {}, timeout: {}ms]",
            delete_method_uuri.to_uri(true),
            self.apply_method_timeout
        );

        let validation_rule = match self.validation_rule {
            Some(rule) => {
                debug!("using custom ValidationRule from JSON config");
                serde_json::from_str(rule.as_str()).inspect_err(|err| {
                    debug!("failed to create ValidationRule from JSON config: {err}")
                })
            }
            None => {
                debug!("using default validation rule");
                Ok(ValidationRule::default())
            }
        }?;

        let rpc_client = self
            .tokio_runtime
            .block_on(async {
                let uri_provider = Arc::new(StaticUriProvider::try_from(self.local_uuri)?);
                InMemoryRpcClient::new(self.transport, uri_provider)
                    .await
                    .map_err(Box::<dyn core::error::Error>::from)
            })
            .inspect_err(|err| error!("failed to create RpcClient: {err}"))
            .map(Arc::new)?;

        Ok(UProtocolTargetProvider {
            tokio_runtime: self.tokio_runtime,
            rpc_client,
            get_method_uuri: self.get_method_uuri,
            get_method_timeout: self.get_method_timeout,
            apply_method_timeout: self.apply_method_timeout,
            update_method_uuri,
            delete_method_uuri,
            validation_rule,
        })
    }
}

struct UProtocolTargetProvider {
    tokio_runtime: Arc<Runtime>,
    rpc_client: Arc<dyn RpcClient>,
    get_method_uuri: UUri,
    get_method_timeout: u32,
    apply_method_timeout: u32,
    update_method_uuri: UUri,
    delete_method_uuri: UUri,
    validation_rule: ValidationRule,
}

impl UProtocolTargetProvider {
    /// Gets a builder for configuring and creating a uProtocol target provider.
    fn builder(
        tokio_runtime: Arc<Runtime>,
        transport: Arc<dyn UTransport>,
        local_uuri: UUri,
        get_method_uuri: UUri,
    ) -> UProtocolTargetProviderBuilder {
        UProtocolTargetProviderBuilder::new(tokio_runtime, transport, local_uuri, get_method_uuri)
    }

    async fn invoke_apply_method(
        &self,
        deployment: &DeploymentSpec,
        affected_components: Vec<ComponentSpec>,
        method_to_invoke: &UUri,
    ) -> Result<HashMap<String, ComponentResultSpec>, String> {
        if affected_components.is_empty() {
            return Ok(HashMap::new());
        }

        let request_data = json!({
            PROPERTY_NAME_DEPLOYMENT: deployment,
            PROPERTY_NAME_COMPONENTS: affected_components
        });

        let data = serde_json::to_vec(&request_data).map_err(|err| err.to_string())?;
        let payload = UPayload::new(data, up_rust::UPayloadFormat::UPAYLOAD_FORMAT_JSON);
        let response = self
            .rpc_client
            .invoke_method(
                method_to_invoke.to_owned(),
                // execution of the operation might take some time ...
                CallOptions::for_rpc_request(self.apply_method_timeout, None, None, None),
                Some(payload),
            )
            .await
            .map_err(|err| err.to_string())?;
        let Some(response_payload) = response else {
            return Err("target provider returned empty response to request".to_string());
        };
        let result: HashMap<String, ComponentResultSpec> = serde_json::from_slice(
            response_payload.payload().to_vec().as_slice(),
        )
        .map_err(|e| {
            debug!(
                "failed to deserialize response from remote target provider [operation: {}]: {:?}",
                method_to_invoke.to_uri(true),
                e
            );
            e.to_string()
        })?;
        Ok(result)
    }

    async fn get_components(
        &self,
        deployment: DeploymentSpec,
        component_specs: Vec<ComponentSpec>,
    ) -> Result<Vec<ComponentSpec>, String> {
        let request_data = json!({
            PROPERTY_NAME_DEPLOYMENT: deployment,
            PROPERTY_NAME_COMPONENTS: component_specs
        });
        let data = serde_json::to_vec(&request_data).map_err(|err| err.to_string())?;
        let payload = UPayload::new(data, up_rust::UPayloadFormat::UPAYLOAD_FORMAT_JSON);
        let response = self
            .rpc_client
            .invoke_method(
                self.get_method_uuri.clone(),
                CallOptions::for_rpc_request(self.get_method_timeout, None, None, None),
                Some(payload),
            )
            .await
            .map_err(|err| err.to_string())?;
        let Some(response_payload) = response else {
            return Err("target provider returned empty response to Get request".to_string());
        };
        let spec_array: Vec<ComponentSpec> = serde_json::from_slice(
            response_payload.payload().to_vec().as_slice(),
        )
        .map_err(|e| {
            error!(
                "failed to deserialize response from remote target provider: {:?}",
                e
            );
            e.to_string()
        })?;
        Ok(spec_array)
    }
}

/// Creates a new uProtocol target provider instance.
///
/// The target provider is stateless and thus a new instance is created for each and every
/// invocation. It supports both Eclipse Zenoh and MQTT 5 as transport protocols to communicate with
/// the remote target provider. If neither transport is configured, a local transport is used which
/// is mainly intended for testing purposes, as it does only support in-process communication.
///
/// # Arguments
/// * `config_json` - A pointer to a null-terminated JSON string containing the configuration.
///                   The following properties are required/supported:
///                   - `localEntity`: The Local Entity URI of this target provider (required).
///                   - `getMethodUri`: The URI of the Get method on the remote target provider (required).
///                   - `getMethodTimeoutMillis`: The timeout for the Get method in milliseconds (optional, default: 120000).
///                   - `applyMethodTimeoutMillis`: The timeout for the Update and Delete methods in milliseconds (optional, default: 300000).
///                   - `zenohConfig`: The (absolute) path to a Zenoh configuration file (optional). Either this property or `brokerAddress` must be set to configure a usable transport.
///                   - `brokerAddress`: The address of the MQTT 5 broker (optional). Either this property or `zenohConfig` must be set to configure a usable transport.
///                   - `clientID`: The client ID to use when connecting to the MQTT 5 broker (optional).
///
/// # Errors
/// Returns a null pointer if the provider could not be created, e.g. because the configuration
/// is invalid or missing required fields. In particular, it is an error to set both
/// `zenohConfig` and `brokerAddress`.
///
/// # Safety
///
/// Client code needs to make sure that the passed in pointer is valid.
#[no_mangle]
pub unsafe extern "C" fn create_provider(config_json: *const c_char) -> *mut ProviderWrapper {
    // try to configure tracing to output Events to stdout
    if tracing_subscriber::fmt::try_init().is_err() {
        // This will fail on subsequent invocations of the target provider's functions
        // because the provider is stateless and thus newly created for each and every invocation.
        // Consequently, we do not need to clutter stdout with corresponding warning
        // but can assume that the subscriber has been initialized one way or the other ...
    }
    if config_json.is_null() {
        error!("Pointer to configuration JSON string is null");
        return ptr::null_mut();
    }

    let config_str = match unsafe { CStr::from_ptr(config_json).to_str() } {
        Ok(str) => str,
        Err(err) => {
            error!("Error converting pointer to JSON string: {:?}", err);
            return ptr::null_mut();
        }
    };
    debug!(
        "creating UProtocolTargetProvider using config: {}",
        config_str
    );

    let config: ProviderConfig = match serde_json::from_str(config_str) {
        Ok(cfg) => cfg,
        Err(err) => {
            error!("Error deserializing configuration JSON string: {:?}", err);
            return ptr::null_mut();
        }
    };

    let Ok(local_uuri) = extract_string_property(&config, CONFIG_PROPERTY_LOCAL_ENTITY)
        .inspect_err(|err| {
            error!("{err}");
        })
        .and_then(|s| {
            s.parse::<UUri>().map_err(|err| {
                error!("failed to create Local Entity URI: {err}");
                Box::from(err)
            })
        })
    else {
        return ptr::null_mut();
    };

    let Ok(get_method_uuri) = extract_string_property(&config, CONFIG_PROPERTY_GET_METHOD_URI)
        .inspect_err(|err| {
            error!("{err}");
        })
        .and_then(|s| {
            s.parse::<UUri>().map_err(|err| {
                error!("failed to create Get method URI: {err}");
                Box::from(err)
            })
        })
    else {
        return ptr::null_mut();
    };

    // we need to create a multi-threaded Tokio runtime because Zenoh does
    // not work with the current-thread runtime
    let Ok(rt) = tokio::runtime::Builder::new_multi_thread()
        .enable_time()
        .build()
        .map(Arc::new)
    else {
        error!("Failed to create Tokio runtime");
        return ptr::null_mut();
    };

    let Ok(transport) = rt.block_on(get_transport(local_uuri.authority_name(), &config)) else {
        error!("Failed to create uProtocol transport");
        return ptr::null_mut();
    };

    let mut provider_builder =
        UProtocolTargetProvider::builder(rt.clone(), transport, local_uuri, get_method_uuri);

    if let Some(v) = get_string_property(&config, CONFIG_PROPERTY_GET_METHOD_TIMEOUT) {
        if let Ok(millis) = v.parse::<u32>() {
            provider_builder = provider_builder.with_get_method_timeout(millis);
        } else {
            error!("{CONFIG_PROPERTY_GET_METHOD_TIMEOUT} must not exceed 2^32");
            return ptr::null_mut();
        }
    }

    if let Some(v) = get_string_property(&config, CONFIG_PROPERTY_APPLY_METHOD_TIMEOUT) {
        if let Ok(millis) = v.parse::<u32>() {
            provider_builder = provider_builder.with_apply_method_timeout(millis);
        } else {
            error!("{CONFIG_PROPERTY_APPLY_METHOD_TIMEOUT} must not exceed 2^32");
            return ptr::null_mut();
        }
    }

    if let Some(validation_rule) = get_string_property(&config, CONFIG_PROPERTY_VALIDATION_RULE) {
        provider_builder = provider_builder.with_validation_rule(validation_rule);
    }

    let provider = match provider_builder.build() {
        Ok(provider) => Box::new(provider),
        Err(e) => {
            error!("Error creating UProtocolTargetProvider: {e}");
            return ptr::null_mut();
        }
    };
    let wrapper = Box::new(ProviderWrapper { inner: provider });
    Box::into_raw(wrapper)
}

impl ITargetProvider for UProtocolTargetProvider {
    fn get_validation_rule(&self) -> Result<ValidationRule, String> {
        Ok(self.validation_rule.clone())
    }

    // we simply forward the passed in args to the remote target provider's Get operation.
    fn get(
        &self,
        deployment_spec: DeploymentSpec,
        references: Vec<ComponentStep>,
    ) -> Result<Vec<ComponentSpec>, String> {
        let component_specs: Vec<ComponentSpec> = references
            .iter()
            .map(|step| step.component.clone())
            .collect();
        self.tokio_runtime
            .block_on(self.get_components(deployment_spec, component_specs))
    }

    fn apply(
        &self,
        deployment: DeploymentSpec,
        step: DeploymentStep,
        is_dry_run: bool,
    ) -> Result<HashMap<String, ComponentResultSpec>, String> {
        if is_dry_run {
            info!("dryRun is enabled, skipping Apply");
            return Ok(HashMap::new());
        }

        // determine the components that need to be updated/deleted
        let mut components_to_update: Vec<ComponentSpec> = vec![];
        let mut components_to_delete: Vec<ComponentSpec> = vec![];
        step.components.iter().for_each(|step| match step.action {
            ComponentAction::Update => {
                components_to_update.push(step.component.clone());
            }
            ComponentAction::Delete => {
                components_to_delete.push(step.component.clone());
            }
        });

        // and then invoke dedicated operations on the target provider
        let deletion_outcome = self
            .tokio_runtime
            .block_on(self.invoke_apply_method(
                &deployment,
                components_to_delete,
                &self.delete_method_uuri,
            ))
            .map_err(|err| err.to_string())?;
        let mut result = self
            .tokio_runtime
            .block_on(self.invoke_apply_method(
                &deployment,
                components_to_update,
                &self.update_method_uuri,
            ))
            .map_err(|err| err.to_string())?;
        result.extend(deletion_outcome);
        Ok(result)
    }
}

#[cfg(test)]
mod tests {
    use std::{ffi::CString, str::FromStr};

    use super::*;
    use async_trait::async_trait;
    use serde_json::json;
    use up_rust::{
        local_transport::LocalTransport, MockTransport, UListener, UMessage, UMessageBuilder,
        UMessageError,
    };

    #[test]
    fn test_extract_string_property() {
        let value = json!({
            "zenoh_config": "/path/to/zenoh/config"
        });
        assert!(extract_string_property(&value, "non_existing").is_err());
        assert!(extract_string_property(&value, "zenoh_config")
            .is_ok_and(|v| v == "/path/to/zenoh/config"));
    }

    #[test]
    fn test_get_string_property() {
        let value = json!({
            "zenoh_config": "/path/to/zenoh/config"
        });
        assert!(get_string_property(&value, "non_existing").is_none());
        assert!(get_string_property(&value, "zenoh_config")
            .is_some_and(|v| v == *"/path/to/zenoh/config"));
    }

    #[test_case::test_case("//symphony/1DA/2/0", "//updater/BBC/1/1" => true; "succeeds")]
    #[test_case::test_case("//symphony/1DA/2/BA", "//updater/BBC/1/1" => false; "fails for invalid local entity URI")]
    #[test_case::test_case("//symphony/1DA/2/0", "//updater/BBC/1/BA" => false; "fails for invalid Get method URI")]
    fn test_target_provider_builder(local_uri: &str, get_method_uri: &str) -> bool {
        let local_uuri = UUri::from_str(local_uri).expect("invalid local URI");
        let get_method_uuri = UUri::from_str(get_method_uri).expect("invalid GET method URI");
        let mut transport = MockTransport::new();
        transport
            .expect_do_register_listener()
            .returning(|_source_filter, _sink_filter, _listener| Ok(()));

        let tokio_runtime = tokio::runtime::Builder::new_current_thread()
            .enable_time()
            .build()
            .map(Arc::new)
            .expect("failed to create tokio runtime");
        UProtocolTargetProvider::builder(
            tokio_runtime.clone(),
            Arc::new(transport),
            local_uuri,
            get_method_uuri,
        )
        .build()
        .is_ok()
    }

    #[test_case::test_case(json!({
        "localEntity": "//symphony/1DA/2/0",
        "getMethodUri": "//updater/BBC/1/1"
    }) => true; "succeeds")]
    #[test_case::test_case(json!({
        "getMethodUri": "//updater/BBC/1/1"
    }) => false; "fails for missing local URI")]
    #[test_case::test_case(json!({
        "localEntity": "//symphony/1DA/2/0",
    }) => false; "fails for missing GET method URI")]
    #[test_case::test_case(json!({
        "localEntity": "//symphony/1DA/2/0",
        "getMethodUri": "//updater/BBC/1/1",
        "zenohConfig": "/path/to/zenoh/config",
        "brokerAddress": "mqtt://broker.address:1883"
    }) => false; "fails for both transports configured")]
    fn test_create_provider(provider_config: ProviderConfig) -> bool {
        let s = CString::new(provider_config.to_string())
            .expect("failed to create C string from JSON config");
        unsafe {
            let v = create_provider(s.as_ptr());
            !v.is_null()
        }
    }

    struct RequestHandler {
        response_transport: Arc<dyn UTransport>,
        response_factory: Box<dyn Fn(UMessage) -> Result<UMessage, UMessageError> + Sync + Send>,
    }
    impl RequestHandler {
        fn new(
            response_transport: Arc<dyn UTransport>,
            response_factory: Box<
                dyn Fn(UMessage) -> Result<UMessage, UMessageError> + Sync + Send,
            >,
        ) -> Self {
            Self {
                response_transport,
                response_factory,
            }
        }
    }
    #[async_trait]
    impl UListener for RequestHandler {
        async fn on_receive(&self, msg: UMessage) {
            let f = self.response_factory.as_ref();
            let response = f(msg).expect("failed to create response message");
            let _ = self.response_transport.send(response).await;
        }
    }

    #[test_log::test]
    fn test_target_provider_get_succeeds() {
        let tokio_runtime = tokio::runtime::Builder::new_current_thread()
            .enable_time()
            .build()
            .map(Arc::new)
            .expect("failed to create tokio runtime");

        let local_uuri = UUri::from_str("//symphony/1DA/2/0").expect("invalid local URI");
        let get_method_uuri = UUri::from_str("//updater/BBC/1/1").expect("invalid GET URI");
        let transport = Arc::new(LocalTransport::default());

        let provider = UProtocolTargetProvider::builder(
            tokio_runtime.clone(),
            transport.clone(),
            local_uuri.clone(),
            get_method_uuri.clone(),
        )
        .build()
        .expect("failed to create target provider");

        let response_factory = |msg: UMessage| {
            debug!(
                "received request [method-to-invoke: {}, reply-to: {}]",
                msg.attributes
                    .get_or_default()
                    .sink
                    .get_or_default()
                    .to_uri(false),
                msg.attributes
                    .get_or_default()
                    .source
                    .get_or_default()
                    .to_uri(false)
            );
            let response_payload = serde_json::to_vec::<Vec<ComponentSpec>>(&vec![])
                .expect("invalid response payload");
            UMessageBuilder::response_for_request(&msg.attributes).build_with_payload(
                response_payload,
                up_rust::UPayloadFormat::UPAYLOAD_FORMAT_JSON,
            )
        };
        let request_listener = RequestHandler::new(transport.clone(), Box::new(response_factory));

        tokio_runtime.block_on(async {
            transport
                .register_listener(
                    &local_uuri,
                    Some(&get_method_uuri),
                    Arc::new(request_listener),
                )
                .await
                .expect("failed to register request listener");
        });
        assert!(provider.get(DeploymentSpec::empty(), vec![]).is_ok());
    }

    #[test_log::test]
    fn test_target_provider_apply_succeeds() {
        let tokio_runtime = tokio::runtime::Builder::new_current_thread()
            .enable_time()
            .build()
            .map(Arc::new)
            .expect("failed to create tokio runtime");

        let local_uuri = UUri::from_str("//symphony/1DA/2/0").expect("invalid local URI");
        let get_method_uuri = UUri::from_str("//updater/BBC/1/1").expect("invalid GET URI");
        let sink_filter_uri = UUri::try_from_parts(
            &get_method_uuri.authority_name(),
            get_method_uuri.ue_id,
            get_method_uuri.uentity_major_version(),
            0xFFFF,
        )
        .expect("invalid sink filter URI");
        let transport = Arc::new(LocalTransport::default());

        let provider = UProtocolTargetProvider::builder(
            tokio_runtime.clone(),
            transport.clone(),
            local_uuri.clone(),
            get_method_uuri.clone(),
        )
        .build()
        .expect("failed to create target provider");

        let response_factory = |msg: UMessage| {
            debug!(
                "received request [method-to-invoke: {}, reply-to: {}]",
                msg.attributes
                    .get_or_default()
                    .sink
                    .get_or_default()
                    .to_uri(false),
                msg.attributes
                    .get_or_default()
                    .source
                    .get_or_default()
                    .to_uri(false)
            );
            let response_payload =
                serde_json::to_vec::<HashMap<String, ComponentSpec>>(&HashMap::new())
                    .expect("invalid response payload");
            UMessageBuilder::response_for_request(&msg.attributes).build_with_payload(
                response_payload,
                up_rust::UPayloadFormat::UPAYLOAD_FORMAT_JSON,
            )
        };
        let request_listener = RequestHandler::new(transport.clone(), Box::new(response_factory));

        tokio_runtime.block_on(async {
            transport
                .register_listener(
                    &local_uuri,
                    Some(&sink_filter_uri),
                    Arc::new(request_listener),
                )
                .await
                .expect("failed to register request listener")
        });
        let component_step = ComponentStep {
            action: ComponentAction::Update,
            component: ComponentSpec {
                component_type: None,
                name: "Test Component".to_string(),
                metadata: None,
                properties: None,
                parameters: None,
                routes: None,
                constraints: None,
                dependencies: None,
                skills: None,
                sidecars: None,
            },
        };
        let deployment_step = DeploymentStep {
            components: vec![component_step],
            target: None,
            role: "".to_string(),
            is_first: false,
        };
        assert!(provider
            .apply(DeploymentSpec::empty(), deployment_step, false)
            .is_ok());
    }
}
