# Design for Ephemeral Storage Pod Settings

## Abstract
This design describes how ephemeral-storage requests and limit to be incoporated into existing Velero pods requests and limits.

## Background

All velero, node-agent, and datamover Pods have minimum, and sometimes variable local node storage requirements to execute.  VBDM Pods host local kopia caches that can be gibibytes in size.

On nodes that share local container storage with the underlying operating system, filling the local node storage can cause behavioral problems to the node. The available storage on Kubernetes nodes can be extremely limited above the OS requirements. Setting ephemeral-storage requests and limits enables safer scheduling of pods onto nodes and protecting the underlying node from disruption.

## Goals
- Enable setting ephemeral-storage requests and limits across Velero Pods that use ephemeral-storage: velero, node-agent, VBDM data movers, repository maintenance.

## Non Goals
- Modifying the behavior of ephemeral-storage usage.


## High-Level Design
Extend the existing mechanism for setting Pod resources - CPU and memory - and their requests and limits to a 3rd resource type of ephemeral-storage.

The default behavior of Velero Quality of Service `BestEffort` shall be maintained. Existing behavior of ephemeral-storage resources request and limit will not change unless the values are set.

The behavior of Pod Volume Backup will be unchanged. Ephemeral storage requests and limits will not be available for Pod Volume Backup.

## Detailed Design

### PodResources

The underlying struct for all Velero resource input is the PodResources struct. It is extended to add `EphemeralStorageRequest` and `EphemeralStorageLimit` fields.

```
type PodResources struct {
	CPURequest              string `json:"cpuRequest,omitempty"`
	MemoryRequest           string `json:"memoryRequest,omitempty"`
	CPULimit                string `json:"cpuLimit,omitempty"`
	MemoryLimit             string `json:"memoryLimit,omitempty"`
+	EphemeralStorageRequest string `json:"ephemeralStorageRequest,omitempty"`
+	EphemeralStorageLimit   string `json:"ephemeralStorageLimit,omitempty"`
}
```

[ParseResourceRequirements](../pkg/util/kube/resource_requirements.go) is extended to accept ephemeral-storage request and limits arguments. The default values if unset is "0".

The ephemeral-storage values can be set in the following places:

### Velero Command Line Install

The additional CLI options control the Velero Deployment and Node-Agent DaemonSet resources.

`--velero-pod-ephemeral-storage-request` to set Velero Pod ephemeral-storage requests.

`--velero-pod-ephemeral-storage-limit` to set Velero Pod ephemeral-storage limits.

`--node-agent-pod-ephemeral-storage-request` to set Node-Agent Pod ephemeral-storage requests.

`--node-agent-pod-ephemeral-storage-limit` to set Node-Agent Pod ephemeral-storage limits.

The default value is "0", for unregulated ephemeral-storage usage.

### VBDM Datamover Pods

Configuration of VBDM pods is through the Node-Agent-Config design under the podResources object.

The podResources object is extended for additional `ephemeralStorageRequest` and `ephemeralStorageLimit` fields.

```
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: velero-node-agent-config
      namespace: velero
    data:
      podResources: |
        {  
           "podResources": {
             "cpuRequest": "1000m",
             "cpuLimit": "1000m",
             "memoryRequest": "512Mi",
             "memoryLimit": "1Gi",
             "ephemeralStorageRequest": "2Gi",
             "ephemeralStorageLimit": "4Gi"        
           }
        }
```

### BackupRepository Maintenance Jobs

The maintenance jobs inherit the ephemeral-storage options from the PodResources struct similar to the node-agent-config. The podResources object will have additional `ephemeralStorageRequest` and `ephemeralStorageLimit` fields.

```
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: maintenance-config
      namespace: velero
    data:
    "global": |
      {  
         "podResources": {
           "cpuRequest": "1000m",
           "cpuLimit": "1000m",
           "memoryRequest": "512Mi",
           "memoryLimit": "1Gi",
           "ephemeralStorageRequest": "2Gi",
           "ephemeralStorageLimit": "4Gi"        
         }
      }
```

## Alternatives Considered
The Kubernetes native Pod resources object is `corev1api.ResourceRequirements`. Adopting this would allow for all existing accepting resource types and naturally extend over time with changes in Kubernetes. However, doing so would break existing VBDM and maintenance configs without a migration process. This alternative was rejected as too large and outside the scope of this design.

## Security Considerations
No new security concerns.

## Compatibility
The new resource fields must not interfere with the exiating fields or change existing behaviors unless the new values are set.

## Implementation
A proposed implementation via pull request based on the design if accepted.

## Open Issues
If there is a requirement to migrate off the existing PodResources struct the proposed changes will grow substantially in size and complexity and cannot be contained in the scope of this design.
