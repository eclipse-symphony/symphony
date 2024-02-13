# Symphony Naming Rules and Conventions

_(last edit: 2/2/2024)_


Symphony prefers naming conventions over rigid schemas in many cases. And in some cases, due to the limitations of the underlying system, Symphony uses string-based annotations allowed by the corresponding system to carry some necessary metadata. 
This document captures naming rules and conventions that are used at different places of Symphony.


## Catalog Naming Conventions
* Topology catalogs are named after the objects they represent, with a ```-topology``` postfix. For example, a catalog that represents the topology of a solution ```solution-a``` is named as ```solution-a-topology```.

## K8s Pod Labeling
* For each container in a Kubernetes pod, Symphony attaches a `label` pod metadata to indicate which container the sidecar is attached to. For example, a sidecar `sidecar_a` belonging to a component named `component_a` will be labeled as `sidecar_a.sidecar_of: component_a`.
