# Modeling Applications

You can assemble components from different artifact formats into an orchestrated application model using Symphony’s Solution object.

A [Solution](../uom/solution.md) describes an application. It consists of a list of [Component](../uom/solution.md#componentspec), which can be a container, a Helm chart, a Kubernetes artifact file, a security policy, a firmware, or anything else. Instead of forcing to adopt the Symphony [Component](../uom/solution.md#componentspec) artifact format, Symphony allows existing application artifacts to be directly embedded into Symphony solutions.

When modeling a microservice application, components are assumed to be independent from each other. However, in many legacy applications there are implicit or explicit dependencies among components. Symphony allows you to attach optional dependencies to components to build up a dependency tree. When Symphony deploys the Solution, it walks the dependency tree and ensures components are deployed in the correct order.

## Topics
* [Configuration Management](./configuration-management.md)