---
type: docs
title: "Object Metadata"
linkTitle: "ObjectMeta"
description: ""
weight: 60
---

| Field | Type | Description |
|--------|--------|--------|
| Namespace | string | Namespace the object belongs to. |
| Name | string | Name of the object. |
| ETag | string | String representation of the object version. |
| ObjGeneration | int64 | Changes when the object spec changes. |
| Labels | map[string]string | Labels on the object. |
| Annotations | map[string]string | Annotations on the object. |
| UID | UID | Unique identifier for the object. |
| OwnerReferences | []metav1.OwnerReference | Owner references for the object. |