#!/usr/bin/env bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

# Function to print usage
usage() {
    echo -e "\e[31mUsage for HTTP mode:\e[0m"
    echo -e "\e[31m  $0 http <endpoint> <cert_path> <key_path> <target_name> <namespace> <topology> <user> <group> [symphony_ca_cert_path]\e[0m"
    echo -e "\e[31mUsage for MQTT mode:\e[0m"
    echo -e "\e[31m  $0 mqtt <broker_address> <broker_port> <cert_path> <key_path> <target_name> <namespace> <topology> <user> <group> [binary_path] [ca_cert_path] [use_cert_subject]\e[0m"
    echo -e "\e[31mNote: binary_path is required when protocol is 'mqtt'\e[0m"
    echo -e "\e[31m      use_cert_subject: true/false, optional, use cert subject as topic suffix for MQTT\e[0m"
    echo -e "\e[31m      symphony_ca_cert_path: optional, Symphony server CA certificate for HTTPS verification\e[0m"
    exit 1
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "\e[33mjq is not installed. Installing jq...\e[0m"
    sudo apt-get update && sudo apt-get install -y jq
    if [ $? -ne 0 ]; then
        echo -e "\e[31mError: Failed to install jq. Exiting...\e[0m"
        exit 1
    else
        echo -e "\e[32mjq installed successfully.\e[0m"
    fi
fi

# Assign parameters to variables
protocol=$1

if [ "$protocol" = "http" ]; then
    endpoint=$2
    cert_path=$3
    key_path=$4
    target_name=$5
    namespace=$6
    topology=$7
    user=$8
    group=$9
    symphony_ca_cert_path=${10}
    
    # Validate the endpoint (basic URL validation)
    if ! [[ $endpoint =~ ^https?:// ]]; then
        echo -e "\e[31mError: Invalid endpoint. Must be a valid URL starting with http:// or https://\e[0m"
        usage
    fi
    
    # Check for Symphony CA certificate
    if [ ! -z "$symphony_ca_cert_path" ] && [ ! -f "$symphony_ca_cert_path" ]; then
        echo -e "\e[31mError: Symphony CA certificate file not found at path: $symphony_ca_cert_path\e[0m"
        usage
    fi
    
    # Create the JSON configuration for HTTP
    echo -e "\e[32mCreating JSON configuration for HTTP mode...\e[0m"
    config_json=$(cat <<EOF
{
    "requestEndpoint": "$endpoint/solution/tasks",
    "responseEndpoint": "$endpoint/solution/task/getResult",
    "baseUrl": "$endpoint"
}
EOF
    )
elif [ "$protocol" = "mqtt" ]; then
    broker_address=$2
    broker_port=$3
    cert_path=$4
    key_path=$5
    target_name=$6
    namespace=$7
    topology=$8
    user=$9
    group=${10}
    binary_path=${11}
    ca_cert_path=${12}
    use_cert_subject=${13}
    
    # Validate MQTT broker address
    if [ -z "$broker_address" ]; then
        echo -e "\e[31mError: MQTT broker address cannot be empty\e[0m"
        usage
    fi
    
    # Validate MQTT broker port
    if [ -z "$broker_port" ] || ! [[ "$broker_port" =~ ^[0-9]+$ ]]; then
        echo -e "\e[31mError: MQTT broker port must be a valid number\e[0m"
        usage
    fi
    
    # Create the JSON configuration for MQTT
    echo -e "\e[32mCreating JSON configuration for MQTT mode...\e[0m"
    config_json=$(cat <<EOF
{
    "mqttBroker": "$broker_address",
    "mqttPort": $broker_port
}
EOF
    )
    
    # Check for CA certificate
    if [ -z "$ca_cert_path" ]; then
        echo -e "\e[33mWarning: CA certificate path not provided for MQTT. You may need this for secure MQTT connections.\e[0m"
        read -p "Do you want to provide a CA certificate path? [y/N]: " provide_ca
        if [[ "$provide_ca" =~ ^[Yy]$ ]]; then
            read -p "Please enter the CA certificate path: " ca_cert_path
        fi
    fi
    
    if [ ! -z "$ca_cert_path" ] && [ ! -f "$ca_cert_path" ]; then
        echo -e "\e[31mError: CA certificate file not found at path: $ca_cert_path\e[0m"
        usage
    fi
else
    echo -e "\e[31mError: Protocol must be either 'http' or 'mqtt'.\e[0m"
    usage
fi

# Common validations regardless of protocol
# Validate the certificate path (check if the file exists)
if [ ! -f "$cert_path" ]; then
    echo -e "\e[31mError: Certificate file not found at path: $cert_path\e[0m"
    usage
fi

# Validate the key path (check if the file exists)
if [ ! -f "$key_path" ]; then
    echo -e "\e[31mError: Key file not found at path: $key_path\e[0m"
    usage
fi

# Validate the target name (non-empty string)
if [ -z "$target_name" ]; then
    echo -e "\e[31mError: Target name must be a non-empty string.\e[0m"
    usage
fi

# Validate the namespace (default if not provided)
if [ -z "$namespace" ]; then
    echo -e "\e[31mError: Namespace must be a non-empty string.\e[0m"
    namespace="default"
fi

# Validate the topology file (non-empty string)
if [ -z "$topology" ]; then
    echo -e "\e[31mError: Topology file must be a non-empty string.\e[0m"
    usage
fi

# Validate the user (non-empty string)
if [ -z "$user" ]; then
    echo -e "\e[31mError: User must be a non-empty string.\e[0m"
    usage
fi

# Validate the group (non-empty string)
if [ -z "$group" ]; then
    echo -e "\e[31mError: Group must be a non-empty string.\e[0m"
    usage
fi

# Save the JSON configuration to a file
config_file="config.json"
echo "$config_json" > "$config_file"
echo -e "\e[32mJSON configuration saved to $config_file\e[0m"

# Convert cert_path, key_path, topology_path, config to absolute paths
cert_path=$(realpath $cert_path)
key_path=$(realpath $key_path)
topology=$(realpath $topology)
config=$(realpath $config_file)

# Protocol-specific handling
if [ "$protocol" = "http" ]; then
    # HTTP mode: Call the certificate endpoint to get the public and private keys
    bootstarpCertEndpoint="$endpoint/targets/bootstrap/$target_name?namespace=$namespace&osPlatform=linux"
    echo -e "\e[32mCalling certificate endpoint: $bootstarpCertEndpoint\e[0m"

    # Build curl command with optional CA certificate
    curl_cmd="curl --cert \"$cert_path\" --key \"$key_path\""
    if [ ! -z "$symphony_ca_cert_path" ]; then
        symphony_ca_cert_path=$(realpath "$symphony_ca_cert_path")
        curl_cmd="$curl_cmd --cacert \"$symphony_ca_cert_path\""
        echo -e "\e[32mUsing Symphony CA certificate: $symphony_ca_cert_path\e[0m"
    else
        echo -e "\e[33mWarning: No Symphony CA certificate provided. Using insecure mode for testing.\e[0m"
        curl_cmd="$curl_cmd -k"
    fi

    # Get certificate
    result=$(eval "$curl_cmd -X POST \"$bootstarpCertEndpoint\" -H \"Content-Type: application/json\"")

    if [ $? -ne 0 ]; then
        echo -e "\e[31mError: Failed to call certificate endpoint. Please check the endpoint and try again.\e[0m"
        exit 1
    else
        echo -e "\e[32mCertificate endpoint response received\e[0m"
    fi

    # Parse JSON response and extract certificates
    public=$(echo $result | jq -r '.public')
    private=$(echo $result | jq -r '.private')

    # Check if we got valid certificates
    if [ "$public" = "null" ] || [ "$private" = "null" ] || [ -z "$public" ] || [ -z "$private" ]; then
        echo -e "\e[31mError: Failed to extract certificates from response. Response: $result\e[0m"
        exit 1
    fi

    # Reconstruct PEM format properly (Symphony converts \n to spaces for transmission)
    # Convert to word arrays and reconstruct with proper headers/footers
    public_words=($public)
    private_words=($private)
    
    # Reconstruct public certificate
    {
        echo "-----BEGIN CERTIFICATE-----"
        # Skip the header words (-----BEGIN CERTIFICATE-----) and footer words (-----END CERTIFICATE-----)
        for ((i=2; i<${#public_words[@]}-2; i++)); do
            echo "${public_words[i]}"
        done
        echo "-----END CERTIFICATE-----"
    } > public.pem
    
    if [ $? -ne 0 ]; then
        echo -e "\e[31mError: Failed to save public certificate to public.pem. Exiting...\e[0m"
        exit 1
    else
        echo -e "\e[32mPublic certificate saved to public.pem\e[0m"
    fi

    # Reconstruct private key
    {
        echo "-----BEGIN RSA PRIVATE KEY-----"
        # Skip the header words (-----BEGIN RSA PRIVATE KEY-----) and footer words (-----END RSA PRIVATE KEY-----)
        for ((i=4; i<${#private_words[@]}-4; i++)); do
            echo "${private_words[i]}"
        done
        echo "-----END RSA PRIVATE KEY-----"
    } > private.pem
    
    if [ $? -ne 0 ]; then
        echo -e "\e[31mError: Failed to save private key to private.pem. Exiting...\e[0m"
        exit 1
    else
        echo -e "\e[32mPrivate key saved to private.pem\e[0m"
    fi

    # No longer update the topology here, it's handled when remote-agent starts
    echo -e "\e[32mCertificates prepared. Topology will be updated when remote-agent starts.\e[0m"
    # Download the remote-agent binary
    echo -e "\e[32mDownloading remote-agent binary...\e[0m"
    download_result=$(eval "$curl_cmd -X GET \"$endpoint/files/remote-agent\" -o remote-agent")
    if [ $? -ne 0 ]; then
        echo -e "\e[31mError: Failed to download remote-agent binary. Exiting...\e[0m"
        exit 1
    else
        echo -e "\e[32mRemote-agent binary downloaded successfully\e[0m"
    fi
    
    # Set the agent path for HTTP mode
    agent_path=$(realpath "./remote-agent")
    echo -e "\e[32mUsing remote-agent binary: $agent_path\e[0m"
else
    # MQTT mode: Use original certificates and binary
    echo -e "\e[32mMQTT mode: Using original certificates directly\e[0m"
    echo -e "\e[32mCertificate path: $cert_path\e[0m"
    echo -e "\e[32mKey path: $key_path\e[0m"
    echo -e "\e[32mBroker: $broker_address:$broker_port\e[0m"
    echo -e "\e[32mTopology will be sent via MQTT topic 'symphony/request/$target_name' after agent starts\e[0m"
    
    # Handle binary path for MQTT mode
    if [ -z "$binary_path" ]; then
        read -p "Please input the full path to your remote-agent binary: " agent_path
    else
        agent_path="$binary_path"
        echo -e "\e[32mUsing provided binary path: $agent_path\e[0m"
    fi
    
    agent_path=$(realpath "$agent_path")
    if [ ! -f "$agent_path" ]; then
        echo -e "\e[31mError: remote-agent binary not found at $agent_path. Exiting...\e[0m"
        exit 1
    fi
    chmod +x "$agent_path"
    echo -e "\e[32mUsing remote-agent binary: $agent_path\e[0m"
fi

# Make the remote-agent binary executable (for HTTP mode, chmod was already done above)
if [ "$protocol" = "http" ]; then
    chmod +x "./remote-agent"
    if [ $? -ne 0 ]; then
        echo -e "\e[31mError: Failed to make remote-agent binary executable. Exiting...\e[0m"
        exit 1
    else
        echo -e "\e[32mRemote-agent binary made executable\e[0m"
    fi
fi

echo -e "\e[32mFiles created successfully:\e[0m"
if [ "$protocol" = "http" ]; then
    echo -e "\e[32mpublic.pem\e[0m"
    echo -e "\e[32mprivate.pem\e[0m"
fi
echo -e "\e[32mremote-agent\e[0m"

# Set certificate paths appropriately based on protocol
if [ "$protocol" = "http" ]; then
    # HTTP mode: Use the generated public.pem and private.pem
    public_path=$(realpath "./public.pem")
    private_path=$(realpath "./private.pem")
    # agent_path was already set above for HTTP mode, no need to reset it
else
    # MQTT mode: Use the original certificate paths
    public_path=$cert_path
    private_path=$key_path
fi

# Build the service command
service_command="$agent_path -config=$config -client-cert=$public_path -client-key=$private_path -target-name=$target_name -namespace=$namespace -topology=$topology -protocol=$protocol"

# Add CA certificate parameter if available
if [ "$protocol" = "mqtt" ] && [ ! -z "$ca_cert_path" ]; then
    ca_cert_path=$(realpath "$ca_cert_path")
    service_command="$service_command -ca-cert=$ca_cert_path"
    echo -e "\e[32mUsing CA certificate: $ca_cert_path\e[0m"
elif [ "$protocol" = "http" ] && [ ! -z "$symphony_ca_cert_path" ]; then
    symphony_ca_cert_path=$(realpath "$symphony_ca_cert_path")
    service_command="$service_command -ca-cert=$symphony_ca_cert_path"
    echo -e "\e[32mUsing Symphony CA certificate for remote-agent: $symphony_ca_cert_path\e[0m"
fi

# Add use-cert-subject parameter if set
if [ "$protocol" = "mqtt" ] && [ "$use_cert_subject" = "true" ]; then
    service_command="$service_command -use-cert-subject=true"
    echo -e "\e[32mUsing certificate subject as topic suffix for MQTT.\e[0m"
fi

echo -e "\e[32mService command: $service_command\e[0m"

# Debug: Check if the binary exists and is executable
echo -e "\e[32mDebugging: Checking binary at path: $agent_path\e[0m"
echo -e "\e[32mDebugging: Current working directory: $(pwd)\e[0m"
if [ -f "$agent_path" ]; then
    echo -e "\e[32mDebugging: Binary file exists\e[0m"
    ls -la "$agent_path"
    if [ -x "$agent_path" ]; then
        echo -e "\e[32mDebugging: Binary is executable\e[0m"
    else
        echo -e "\e[31mDebugging: Binary is NOT executable\e[0m"
    fi
    
    # Test if the binary can actually run
    echo -e "\e[32mDebugging: Testing binary execution...\e[0m"
    if "$agent_path" --help >/dev/null 2>&1 || "$agent_path" -h >/dev/null 2>&1 || "$agent_path" --version >/dev/null 2>&1; then
        echo -e "\e[32mDebugging: Binary executes successfully\e[0m"
    else
        echo -e "\e[31mDebugging: Binary failed to execute (exit code: $?)\e[0m"
    fi
else
    echo -e "\e[31mDebugging: Binary file does NOT exist at path: $agent_path\e[0m"
    echo -e "\e[31mDebugging: Current directory contents:\e[0m"
    ls -la .
fi

# Create the remote-agent.service file
echo -e "\e[32mCreating remote-agent.service file...\e[0m"
sudo bash -c "cat <<EOF > /etc/systemd/system/remote-agent.service
[Unit]
Description=Remote Agent Service
After=network.target

[Service]
ExecStart=$service_command
WorkingDirectory=$(pwd)
Restart=always
User=$user
Group=$group

[Install]
WantedBy=multi-user.target
EOF"
if [ $? -ne 0 ]; then
    echo -e "\e[31mError: Failed to create remote-agent.service file. Exiting...\e[0m"
    exit 1
else
    echo -e "\e[32mremote-agent.service file created\e[0m"
fi

# Reload systemd to recognize the new service
sudo systemctl daemon-reload
if [ $? -ne 0 ]; then
    echo -e "\e[31mError: Failed to reload systemd daemon. Exiting...\e[0m"
    exit 1
else
    echo -e "\e[32mSystemd daemon reloaded\e[0m"
fi

# Enable the service to start on boot
sudo systemctl enable remote-agent.service
if [ $? -ne 0 ]; then
    echo -e "\e[31mError: Failed to enable remote-agent.service. Exiting...\e[0m"
    exit 1
else
    echo -e "\e[32mremote-agent.service enabled to start on boot\e[0m"
fi

# Start the service
sudo systemctl start remote-agent.service
if [ $? -ne 0 ]; then
    echo -e "\e[31mError: Failed to start remote-agent.service. Exiting...\e[0m"
    exit 1
else
    echo -e "\e[32mremote-agent.service started\e[0m"
fi

# Check the status of the service
sudo systemctl status remote-agent.service
