# Symphony Multi-Cluster and Multi-Site Deployments

| Target Provider | On-target Agent | Agent mode| Protocol(s) |Agent Network |
|--------|--------|--------|--------|--------|
| K8s Provider | N/A | Push | TCP | LAN |
| Proxy Provider | Symphony Agent | Push | HTTPS/MQTT | LAN/WAN
| Staging Provider | Piccolo | Poll |  HTTPS | LAN/WAN |
| Staging Provider | Symphony Agent | Poll | HTTPS | LAN/WAN |
| Primary Symphony | Secondary Symphony | Poll | HTTPS | LAN/WAN |



