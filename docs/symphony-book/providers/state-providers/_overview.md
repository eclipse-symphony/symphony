# State Provider
State provider provides state management on internal or external state stores including get, upsert, delete and list operations. 

A state provider can be persistent or volatile depending on whether the state store is crash consistency. It is essential to choose appropriate state provider for different managers to provide stable functionality and great performance.

Currently we support four types of state providers
| provider | Comment | persistent or volatile |
|---|---|---|
| providers.state.k8s | Use kubernetes etcd as state store | persistent |
| providers.state.memory | Use symphony in-memory dictionary as state store | volatile |
| providers.state.redis | Use external redis server as state store | depending on whether redis server is crash consistent |
| providers.state.http | Use external server accepting HTTP request | depending on whether external server is crash consistent |