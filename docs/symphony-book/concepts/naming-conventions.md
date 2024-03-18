# Symphony Naming Rules and Conventions

_(last edit: 3/15/2024)_


Symphony prefers naming conventions over rigid schemas in many cases. And in some cases, due to the limitations of the underlying system, Symphony uses string-based annotations allowed by the corresponding system to carry some necessary metadata. 
This document captures naming rules and conventions that are used at different places of Symphony.


## Catalog Naming Conventions
* Topology catalogs are named after the objects they represent, with a ```-topology``` postfix. For example, a catalog that represents the topology of a solution ```solution-a``` is named as ```solution-a-topology```.

## Environment Variables

| Variable | Value |
|--------|--------|
| `SYMPHONY_SITE_ID` | Symphony Site Id |
| `SYMPHONY_API_BASE_URL` | Symphony API base Url (`http(s)://<address>:<port>/v1alpha2/`) |
| `SYMPHONY_API_USER` | Symphony API user |
| `SYMPHONY_API_PASSWORD` | Symphony API password |
| `PARENT_SYMPHONY_API_BASE_URL` | Parent Symphony API base Url (`http(s)://<address>:<port>/v1alpha2/`) |
| `PARENT_SYMPHONY_API_USER` | Parent Symphony API user |
| `PARENT_SYMPHONY_API_PASSWORD` | Parent Symphony API password |
| `SYMPHONY_TARGET_NAME` | Symphony Target Name (applicable to poll agent) |


## K8s Pod Labeling
* For each container in a Kubernetes pod, Symphony attaches a `label` pod metadata to indicate which container the sidecar is attached to. For example, a sidecar `sidecar_a` belonging to a component named `component_a` will be labeled as `sidecar_a.sidecar_of: component_a`.
