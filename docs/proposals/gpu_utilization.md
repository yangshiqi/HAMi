---
title: Support GPU Utilization Metrics
authors:
- @yangshiqi
reviewers:
- TBD
approvers:
- TBD
creation-date: 2024-04-10
---

# Support GPU Utilization Metrics

## Summary
Currently, HAMi supports dividing a Nvidia GPU card into several vGPU cards to efficiently 
utilize GPU resources. When a vGPU is assigned to a Pod, HAMi previously could not provide 
information on the Pod's utilization of the vGPU. This resulted in users being unable to 
observe the usage patterns of their Pod's vGPU allocation.

This KEP proposes comprehensive support for monitoring vGPU utilization at the Pod level.

## Motivation

### Goals
- Support for monitoring vGPU utilization at the Pod level
- Provide detailed utilization metrics (SM, decoder, encoder)
- Enable Prometheus-compatible metrics collection
- Support utilization tracking for processes and Pods
- Implement efficient data structures for utilization tracking

### Non-Goals
- Does not support monitoring of GPU utilization for non-Nvidia GPUs
- Does not modify the existing device allocation mechanisms
- Does not enforce utilization-based scheduling decisions

## Proposal

### User Stories

#### Story 1

I have partitioned a Nvidia GPU card into 4 parts and deployed 2 Pods on this card.
I want to observe the GPU usage of these two Pods separately to assess 
whether my business logic and resource allocation are optimal.

Currently, HAMi provides a `HostCoreUtilization` metric for the entire GPU card, 
but I cannot observe GPU usage from each Pod's perspective, making it difficult to 
optimize workloads and properly allocate resources.

### Risks and Mitigations

Because the design scheme expands the fields of the `struct shared_region` structure, 
there may be potential compatibility issues with older versions.

Mitigation strategies include:
1. Maintaining internal data structure stability
2. Providing additional public interfaces to decouple external calls from internal implementation
3. Adding comprehensive unit tests to ensure functionality

## Design Details

### 1. Data Structure Design

We've designed a three-layer data structure in the v1 package to support vGPU utilization monitoring:

#### 1.1 Device Utilization Structures

```go
// Internal device utilization structure
type deviceUtilization struct {
    DecUtil uint64 // Decoder utilization
    EncUtil uint64 // Encoder utilization
    SmUtil  uint64 // SM utilization
    unused  [3]uint64
}

// Public device utilization structure for external package use
type DeviceUtilization struct {
    DecUtil uint64 // Decoder utilization
    EncUtil uint64 // Encoder utilization
    SmUtil  uint64 // SM utilization
}
```

#### 1.2 Process and Pod GPU Utilization Statistics

```go
// GPU utilization data
type GPUUtilization struct {
    DecUtil   uint64 // Decoder utilization
    EncUtil   uint64 // Encoder utilization
    SmUtil    uint64 // SM utilization
    Timestamp int64  // Timestamp
}

// Process GPU statistics
type ProcessGPUStats struct {
    PID         int32
    Status      int32
    Utilization GPUUtilization
    LastUpdate  int64
}

// vGPU statistics
type VGPUStats struct {
    Index       int    // vGPU index
    UUID        string // vGPU UUID
    Utilization GPUUtilization
    LastUpdate  int64
}

// Pod GPU statistics
type PodGPUStats struct {
    PodUID     string
    Namespace  string
    Name       string
    VGPUs      map[int]*VGPUStats // vGPU index -> statistics
    LastUpdate int64
}
```

#### 1.3 Mapping Relationships in Spec Structure

```go
// Spec represents GPU specifications and state
type Spec struct {
    sr    *sharedRegionT
    lock  sync.RWMutex
    stats map[int32]*ProcessGPUStats // Process GPU statistics cache
    pods  map[string]*PodGPUStats    // Pod UID -> Pod GPU statistics
}
```

### 2. Data Collection Flow

#### 2.1 Process-level Utilization Collection

Implementation of process GPU utilization update in the `UpdateProcessUtilization` method:

```go
func (s *Spec) UpdateProcessUtilization(pid int32, idx int, util deviceUtilization) error {
    // Parameter validation
    if idx < 0 || idx >= maxDevices {
        return fmt.Errorf("invalid device index: %d", idx)
    }

    // Validate utilization values
    if util.DecUtil > 100 || util.EncUtil > 100 || util.SmUtil > 100 {
        return fmt.Errorf("invalid utilization values: dec=%d, enc=%d, sm=%d",
            util.DecUtil, util.EncUtil, util.SmUtil)
    }

    // Update or create process statistics
    stats, exists := s.stats[pid]
    if !exists {
        stats = &ProcessGPUStats{
            PID:        pid,
            Status:     1,
            LastUpdate: time.Now().Unix(),
        }
        s.stats[pid] = stats
    }

    // Update utilization data
    stats.Utilization = GPUUtilization{
        DecUtil:   util.DecUtil,
        EncUtil:   util.EncUtil,
        SmUtil:    util.SmUtil,
        Timestamp: time.Now().Unix(),
    }
    stats.LastUpdate = time.Now().Unix()

    // Synchronize to shared memory
    // ...
}
```

#### 2.2 Pod-level Utilization Collection

Implementation of Pod GPU utilization update in the `UpdatePodGPUUtilization` method:

```go
func (s *Spec) UpdatePodGPUUtilization(podUID string, namespace string, name string, vgpuIndex int, vgpuUUID string, util DeviceUtilization) error {
    // Parameter validation
    // ...

    // Update or create Pod statistics
    podStats, exists := s.pods[podUID]
    if !exists {
        podStats = &PodGPUStats{
            PodUID:     podUID,
            Namespace:  namespace,
            Name:       name,
            VGPUs:      make(map[int]*VGPUStats),
            LastUpdate: time.Now().Unix(),
        }
        s.pods[podUID] = podStats
    }

    // Update or create vGPU statistics
    vgpuStats, exists := podStats.VGPUs[vgpuIndex]
    if !exists {
        vgpuStats = &VGPUStats{
            Index:      vgpuIndex,
            UUID:       vgpuUUID,
            LastUpdate: time.Now().Unix(),
        }
        podStats.VGPUs[vgpuIndex] = vgpuStats
    } else {
        // Update UUID of existing vGPU
        vgpuStats.UUID = vgpuUUID
    }

    // Update utilization data
    vgpuStats.Utilization = GPUUtilization{
        DecUtil:   util.DecUtil,
        EncUtil:   util.EncUtil,
        SmUtil:    util.SmUtil,
        Timestamp: time.Now().Unix(),
    }
    vgpuStats.LastUpdate = time.Now().Unix()
    podStats.LastUpdate = time.Now().Unix()

    // Synchronize to device utilization
    // ...
}
```

#### 2.3 Data Collection Logic

Implementation of data collection in the `Observe` function in `cmd/vGPUmonitor/feedback.go`:

```go
func Observe(lister *nvidia.ContainerLister) {
    // Clean up inactive processes and Pod data
    for _, c := range containers {
        if spec, ok := c.Info.(*v1.Spec); ok {
            spec.CleanupInactiveProcesses()
            spec.CleanupInactivePods()
        }
    }

    // Collect GPU utilization information for each container and associate with the corresponding Pod
    for _, c := range containers {
        if c.PodUID == "" {
            continue
        }

        // Get Pod information
        // ...

        // For each vGPU device, collect and update utilization information
        for i := 0; i < c.Info.DeviceNum(); i++ {
            if !c.Info.IsValidUUID(i) {
                continue
            }

            uuid := c.Info.DeviceUUID(i)
            if uuid == "" {
                continue
            }

            // Get device utilization
            smUtil := c.Info.DeviceSmUtil(i)
            decUtil := c.Info.DeviceDecUtil(i)
            encUtil := c.Info.DeviceEncUtil(i)

            // Update Pod's vGPU utilization
            if spec, ok := c.Info.(*v1.Spec); ok {
                util := v1.DeviceUtilization{
                    SmUtil:  smUtil,
                    DecUtil: decUtil,
                    EncUtil: encUtil,
                }

                spec.UpdatePodGPUUtilization(
                    c.PodUID,
                    pod.Namespace,
                    pod.Name,
                    i,
                    uuid,
                    util,
                )
            }
        }
    }
    
    // Other observation logic
    // ...
}
```

### 3. Metrics Exposure Design

Adding new metrics descriptors in `cmd/vGPUmonitor/metrics.go`:

```go
// Pod vGPU utilization metrics
podGPUSmUtilizationDesc = prometheus.NewDesc(
    "pod_vgpu_sm_utilization",
    "Pod vGPU SM (Streaming Multiprocessor) utilization",
    []string{"podnamespace", "podname", "poduid", "vdeviceid", "deviceuuid"}, nil,
)

podGPUDecUtilizationDesc = prometheus.NewDesc(
    "pod_vgpu_dec_utilization",
    "Pod vGPU decoder utilization",
    []string{"podnamespace", "podname", "poduid", "vdeviceid", "deviceuuid"}, nil,
)

podGPUEncUtilizationDesc = prometheus.NewDesc(
    "pod_vgpu_enc_utilization",
    "Pod vGPU encoder utilization",
    []string{"podnamespace", "podname", "poduid", "vdeviceid", "deviceuuid"}, nil,
)
```

Collecting and exposing these metrics in the `collectPodAndContainerInfo` method:

```go
// Collect vGPU metrics
podStats, err := cc.ClusterManager.Spec.GetPodGPUUtilization(string(pod.UID))
if err != nil {
    klog.V(5).Infof("No GPU stats found for pod %s/%s: %v", pod.Namespace, pod.Name, err)
    continue
}

for vgpuIndex, vgpuStats := range podStats.VGPUs {
    // Use new Pod-level vGPU utilization metrics
    labels := []string{
        pod.Namespace,
        pod.Name,
        string(pod.UID),
        fmt.Sprint(vgpuIndex),
        vgpuStats.UUID,
    }
    
    // SM (Streaming Multiprocessor) utilization
    if err := sendMetric(ch, podGPUSmUtilizationDesc, prometheus.GaugeValue, 
        float64(vgpuStats.Utilization.SmUtil), labels...); err != nil {
        klog.Errorf("Failed to send SM utilization metric for vGPU %d in Pod %s/%s: %v", 
            vgpuIndex, pod.Namespace, pod.Name, err)
    }
    
    // Decoder utilization
    if err := sendMetric(ch, podGPUDecUtilizationDesc, prometheus.GaugeValue, 
        float64(vgpuStats.Utilization.DecUtil), labels...); err != nil {
        klog.Errorf("Failed to send decoder utilization metric for vGPU %d in Pod %s/%s: %v", 
            vgpuIndex, pod.Namespace, pod.Name, err)
    }
    
    // Encoder utilization
    if err := sendMetric(ch, podGPUEncUtilizationDesc, prometheus.GaugeValue, 
        float64(vgpuStats.Utilization.EncUtil), labels...); err != nil {
        klog.Errorf("Failed to send encoder utilization metric for vGPU %d in Pod %s/%s: %v", 
            vgpuIndex, pod.Namespace, pod.Name, err)
    }

    // Retain existing container-level vGPU metric collection for compatibility
    // ...
}
```

### 4. Data Cleanup Mechanism

To ensure memory does not grow indefinitely, process and Pod cleanup mechanisms were implemented:

```go
// Clean up inactive process data
func (s *Spec) CleanupInactiveProcesses() {
    // ...
    // Clean up processes not updated for over 5 minutes
    if now-stats.LastUpdate > 300 {
        delete(s.stats, pid)
        // Synchronize cleanup to shared memory
        // ...
    }
    // ...
}

// Clean up inactive Pod data
func (s *Spec) CleanupInactivePods() {
    // ...
    // Clean up Pods not updated for over 5 minutes
    if now-podStats.LastUpdate > 300 {
        delete(s.pods, podUID)
    }
    // ...
}
```

### 5. Testing Results

Using Prometheus to query the new metrics, users can obtain the following data:

1. **Pod-level SM utilization**: `pod_vgpu_sm_utilization{podnamespace="default", podname="my-gpu-pod"}`
2. **Pod-level decoder utilization**: `pod_vgpu_dec_utilization{podnamespace="default", podname="my-gpu-pod"}`
3. **Pod-level encoder utilization**: `pod_vgpu_enc_utilization{podnamespace="default", podname="my-gpu-pod"}`

This allows users to clearly see how each Pod uses its allocated vGPU, helping to analyze whether the business logic is reasonable and optimize resource allocation.

## Test Plan

### Unit Tests

We have added comprehensive unit tests covering the following functionality:

1. `Test_UpdateProcessUtilization`: Testing process GPU utilization updates, including various scenarios:
   - Normal utilization update
   - Invalid device index (negative or out of range)
   - Invalid utilization values (over 100%)

2. `Test_UpdatePodGPUUtilization`: Testing Pod GPU utilization updates, including various scenarios:
   - Normal utilization update
   - Updating existing vGPU information
   - Invalid device indices and utilization values

### Integration Tests

Deploy multiple GPU-using Pods to a node and observe whether HAMi provides actual GPU utilization rates for these Pods:

1. Deploy Pods with GPU-intensive workloads (such as TensorFlow training)
2. Query `pod_vgpu_sm_utilization` and other metrics from Prometheus
3. Verify that metric values align with expected load patterns
4. Verify vGPU isolation effects across multiple Pods

## Alternatives

### Alternative Approach 1: Aggregating process metrics only

One alternative would be to simply aggregate process-level metrics at the container level without maintaining separate Pod-level statistics. This approach would be simpler but would limit the granularity of monitoring and make it harder to track Pod-level resource usage over time.

### Alternative Approach 2: Using external monitoring tools

Another alternative would be to rely on external monitoring tools to collect and analyze GPU utilization data. This approach would offload the responsibility from HAMi but would require additional setup and configuration, and might not provide the same level of integration with Kubernetes.

## Implementation Timeline

1. Phase 1: Implement basic vGPU utilization tracking at the process level
2. Phase 2: Add Pod-level aggregation and statistics
3. Phase 3: Expose Prometheus metrics
4. Phase 4: Add data cleanup mechanisms
5. Phase 5: Implement comprehensive testing 