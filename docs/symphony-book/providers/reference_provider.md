# Reference Provider
A reference provider resolves an object reference to the actual object spec. For example, you can use a Kubernetes reference project to query a ```Solution``` object by name, or list ```Targets``` by a label selector or a field selector.

**Symphony-API** uses a reference manager to resolve references. The reference manager has a built-in cache so that it doesn't repeatedly query the underlying system. The lifespan of cached items is controlled by the ``` cacheLifespan``` setting, which is in seconds.

The following table summarizes how query parameters are interpreted by different providers:

| Parameter | HTTP Provider | K8s Provider | Custom Vision Provider|
|--------|--------|--------|--------|
| **field-selector**| field selector | field selector | - |
| **group** | object group | object group | model platform |
| **id** | object id | object id | Custom Vision project |
| **label-selector** | label selector | label selector | - |
| **kind** | object kind | object kind | model flavor |
| **ref** | v1alpha2.ReferenceHTTP | v1alpha2.ReferenceK8sCRD| v1alpha2.CustomVision | 
| **scope** | object namespace | object namespace | Custom Vision endpoint|
| **version** | object version | object version | model iteration |