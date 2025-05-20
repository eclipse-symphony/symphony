# CatalogEvalExpression

This document provides details on the `CatalogEvalExpression` Custom Resource (CR). Update of a CatalogEvalExpression not allowed. Deletion would happen automatically every 10 hrs.

## Metadata

- `metadata.name`: As with all CRs, this is provided by users.
- `annotations.management.azure.com/operationId`: A tracking field that must be a string value.

## Spec

The `resourceRef` object under `spec` is the catalog you want to be evaluated. The `apiGroup` and `kind` should be `federation.symphony` and `Catalog`, respectively. The `name` and `namespace` should be the identifiers.

### Example

```yaml
apiVersion: federation.symphony/v1
kind: CatalogEvalExpression
metadata:
  name: evaluateevalcatalog01
  annotations:
    management.azure.com/operationId: "1"
spec:
  resourceRef:
    apiGroup: federation.symphony
    kind: Catalog
    name: evalcatalog-v-version1
    namespace: default
```
## Clean Evaluation
### A clean evaluation should look like this:

```yaml
apiVersion: federation.symphony/v1
kind: CatalogEvalExpression
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"federation.symphony/v1","kind":"CatalogEvalExpression","metadata":{"annotations":{"management.azure.com/operationId":"3"},"name":"evaluateevalcatalog03","namespace":"default"},"spec":{"resourceRef":{"apiGroup":"federation.symphony","kind":"Catalog","name":"evalcatalog-v-version3","namespace":"default"}}}
    management.azure.com/operationId: "3"
  creationTimestamp: "2024-09-06T05:35:50Z"
  generation: 2
  name: evaluateevalcatalog03
  namespace: default
  resourceVersion: "565765"
  uid: 823b68e8-0234-4942-be46-6b95b65ce004
spec:
  resourceRef:
    apiGroup: federation.symphony
    kind: Catalog
    name: evalcatalog-v-version3
    namespace: default
status:
  actionStatus:
    operationID: "3"
    output:
      city: Sydney
      country: Australia
      evaluationStatus: Succeeded
    status: Succeeded
```
You will see the evaluation output in status.actionStatus.output. status.actionStatus.status indicates whether the evaluation call succeeds or not. It should always be successful unless the catalog reference is wrong. status.actionStatus.evaluationStatus indicates if there is any field that failed to be evaluated during the process.

## Failed Evaluation
### A failed status.actionStatus.evaluationStatus would look like this:
```yaml
apiVersion: federation.symphony/v1
kind: CatalogEvalExpression
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"federation.symphony/v1","kind":"CatalogEvalExpression","metadata":{"annotations":{"management.azure.com/operationId":"1"},"name":"evaluateevalcatalog01","namespace":"default"},"spec":{"resourceRef":{"apiGroup":"federation.symphony","kind":"Catalog","name":"evalcatalog-v-version1","namespace":"default"}}}
    management.azure.com/operationId: "1"
  creationTimestamp: "2024-09-06T05:39:59Z"
  generation: 2
  name: evaluateevalcatalog01
  namespace: default
  resourceVersion: "566372"
  uid: 5c693730-c61c-48c6-9297-24cba7cebdbd
spec:
  resourceRef:
    apiGroup: federation.symphony
    kind: Catalog
    name: evalcatalog-v-version1
    namespace: default
status:
  actionStatus:
    operationID: "1"
    output:
      address: 1st Avenue
      city: Sydney
      country: 'Bad Config: invalid function name: ''wrongexpression'''
      county: 'Not Found: object not found'
      evaluationStatus: Failed
      from:
        country: Australia
      zipcode: 'Not Found: field ''zipcode'' is not found in configuration ''evalcatalog-v-version2'''
    status: Succeeded
```