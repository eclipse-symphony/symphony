# Sample Python Target Agent 


## Launch a local MQTT broker for testing
```bash
docker run -d --name mosquitto -p 1883:1883 -v $(pwd)/mosquitt
o.conf:/mosquitto/mosquitto.conf eclipse-mosquitto
```

## Install Symphony Target Agent Python SDK
```bash
pip install -e <path to Python SDK folder>
# for example
pip install -e ../../../sdks/python-sdk/
```

## Launch Your Target Agent
```bash
python provider.py
```