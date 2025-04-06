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
Currently, HMAi supports dividing a Nvidia GPU card into several vGPU cards to efficiently utilize the 
efficiency of the GPU. When I assign a vGPU to a Pod, HAMi cannot provide information on the Pod's 
utilization of the vGPU. This results in users being unable to observe the usage situation of the Pod's vGPU.

This KEP proposes support for monitoring vGPU utilization.


## Motivation

### Goals
- Support for monitoring vGPU utilization

### Non-Goals

Does not support monitoring of GPU utilization for non-Nvidia GPUs.

## Proposal

### User Stories (Optional)


#### Story 1

I have partitioned a Nvidia GPU card into 4 parts and deployed 2 Pods on this card.
Currently, I want to observe the GPU usage of these two Pods separately, in order to assess 
whether my business logic is reasonable.

Currently, HAMi provides a `HostCoreUtilization` usage rate for the entire GPU card, 
but it still cannot observe the use of GPUs from each Pod's perspective.

### Notes/Constraints/Caveats (Optional)


### Risks and Mitigations

Because the design scheme will expand the fields of the `struct shared_region` structure, there may be potential incompatibilities.

## Design Details

### 1. 数据结构设计

我们在v1包中设计了三层数据结构来支持vGPU利用率监控：

#### 1.1 设备利用率结构

```go
// 内部使用的设备利用率结构
type deviceUtilization struct {
    DecUtil uint64 // 解码器利用率
    EncUtil uint64 // 编码器利用率
    SmUtil  uint64 // SM利用率
    unused  [3]uint64
}

// 公开的设备利用率结构，供外部包使用
type DeviceUtilization struct {
    DecUtil uint64 // 解码器利用率
    EncUtil uint64 // 编码器利用率
    SmUtil  uint64 // SM利用率
}
```

#### 1.2 进程和Pod的GPU利用率统计

```go
// GPU利用率数据
type GPUUtilization struct {
    DecUtil   uint64 // 解码器利用率
    EncUtil   uint64 // 编码器利用率
    SmUtil    uint64 // SM利用率
    Timestamp int64  // 时间戳
}

// 进程的GPU统计信息
type ProcessGPUStats struct {
    PID         int32
    Status      int32
    Utilization GPUUtilization
    LastUpdate  int64
}

// vGPU的统计信息
type VGPUStats struct {
    Index       int    // vGPU索引
    UUID        string // vGPU UUID
    Utilization GPUUtilization
    LastUpdate  int64
}

// Pod的GPU统计信息
type PodGPUStats struct {
    PodUID     string
    Namespace  string
    Name       string
    VGPUs      map[int]*VGPUStats // vGPU索引 -> 统计信息
    LastUpdate int64
}
```

#### 1.3 Spec结构中的映射关系

```go
// Spec表示GPU规格和状态
type Spec struct {
    sr    *sharedRegionT
    lock  sync.RWMutex
    stats map[int32]*ProcessGPUStats // 进程GPU统计信息缓存
    pods  map[string]*PodGPUStats    // Pod UID -> Pod GPU统计信息
}
```

### 2. 数据收集流程

#### 2.1 进程级别利用率收集

在`UpdateProcessUtilization`方法中实现进程GPU利用率更新：

```go
func (s *Spec) UpdateProcessUtilization(pid int32, idx int, util deviceUtilization) error {
    // 参数验证
    if idx < 0 || idx >= maxDevices {
        return fmt.Errorf("invalid device index: %d", idx)
    }

    // 验证利用率值
    if util.DecUtil > 100 || util.EncUtil > 100 || util.SmUtil > 100 {
        return fmt.Errorf("invalid utilization values: dec=%d, enc=%d, sm=%d",
            util.DecUtil, util.EncUtil, util.SmUtil)
    }

    // 更新或创建进程统计信息
    stats, exists := s.stats[pid]
    if !exists {
        stats = &ProcessGPUStats{
            PID:        pid,
            Status:     1,
            LastUpdate: time.Now().Unix(),
        }
        s.stats[pid] = stats
    }

    // 更新利用率数据
    stats.Utilization = GPUUtilization{
        DecUtil:   util.DecUtil,
        EncUtil:   util.EncUtil,
        SmUtil:    util.SmUtil,
        Timestamp: time.Now().Unix(),
    }
    stats.LastUpdate = time.Now().Unix()

    // 同步到共享内存
    // ...
}
```

#### 2.2 Pod级别利用率收集

在`UpdatePodGPUUtilization`方法中实现Pod GPU利用率更新：

```go
func (s *Spec) UpdatePodGPUUtilization(podUID string, namespace string, name string, vgpuIndex int, vgpuUUID string, util DeviceUtilization) error {
    // 参数验证
    // ...

    // 更新或创建Pod统计信息
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

    // 更新或创建vGPU统计信息
    vgpuStats, exists := podStats.VGPUs[vgpuIndex]
    if !exists {
        vgpuStats = &VGPUStats{
            Index:      vgpuIndex,
            UUID:       vgpuUUID,
            LastUpdate: time.Now().Unix(),
        }
        podStats.VGPUs[vgpuIndex] = vgpuStats
    } else {
        // 更新现有vGPU的UUID
        vgpuStats.UUID = vgpuUUID
    }

    // 更新利用率数据
    vgpuStats.Utilization = GPUUtilization{
        DecUtil:   util.DecUtil,
        EncUtil:   util.EncUtil,
        SmUtil:    util.SmUtil,
        Timestamp: time.Now().Unix(),
    }
    vgpuStats.LastUpdate = time.Now().Unix()
    podStats.LastUpdate = time.Now().Unix()

    // 同步到设备利用率
    // ...
}
```

#### 2.3 数据收集逻辑

在`cmd/vGPUmonitor/feedback.go`中的`Observe`函数中实现数据收集：

```go
func Observe(lister *nvidia.ContainerLister) {
    // 清理不活跃的进程和Pod数据
    for _, c := range containers {
        if spec, ok := c.Info.(*v1.Spec); ok {
            spec.CleanupInactiveProcesses()
            spec.CleanupInactivePods()
        }
    }

    // 收集每个容器的GPU利用率信息，并关联到对应的Pod
    for _, c := range containers {
        if c.PodUID == "" {
            continue
        }

        // 获取Pod信息
        // ...

        // 针对每个vGPU设备，收集并更新利用率信息
        for i := 0; i < c.Info.DeviceNum(); i++ {
            if !c.Info.IsValidUUID(i) {
                continue
            }

            uuid := c.Info.DeviceUUID(i)
            if uuid == "" {
                continue
            }

            // 获取设备利用率
            smUtil := c.Info.DeviceSmUtil(i)
            decUtil := c.Info.DeviceDecUtil(i)
            encUtil := c.Info.DeviceEncUtil(i)

            // 更新Pod的vGPU利用率
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
    
    // 其他观察逻辑
    // ...
}
```

### 3. 指标暴露设计

在`cmd/vGPUmonitor/metrics.go`中添加新的指标描述符：

```go
// Pod vGPU利用率指标
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

在`collectPodAndContainerInfo`方法中收集和暴露这些指标：

```go
// 收集vGPU指标
podStats, err := cc.ClusterManager.Spec.GetPodGPUUtilization(string(pod.UID))
if err != nil {
    klog.V(5).Infof("No GPU stats found for pod %s/%s: %v", pod.Namespace, pod.Name, err)
    continue
}

for vgpuIndex, vgpuStats := range podStats.VGPUs {
    // 使用新的Pod级别vGPU利用率指标
    labels := []string{
        pod.Namespace,
        pod.Name,
        string(pod.UID),
        fmt.Sprint(vgpuIndex),
        vgpuStats.UUID,
    }
    
    // SM (Streaming Multiprocessor) 利用率
    if err := sendMetric(ch, podGPUSmUtilizationDesc, prometheus.GaugeValue, 
        float64(vgpuStats.Utilization.SmUtil), labels...); err != nil {
        klog.Errorf("Failed to send SM utilization metric for vGPU %d in Pod %s/%s: %v", 
            vgpuIndex, pod.Namespace, pod.Name, err)
    }
    
    // 解码器利用率
    if err := sendMetric(ch, podGPUDecUtilizationDesc, prometheus.GaugeValue, 
        float64(vgpuStats.Utilization.DecUtil), labels...); err != nil {
        klog.Errorf("Failed to send decoder utilization metric for vGPU %d in Pod %s/%s: %v", 
            vgpuIndex, pod.Namespace, pod.Name, err)
    }
    
    // 编码器利用率
    if err := sendMetric(ch, podGPUEncUtilizationDesc, prometheus.GaugeValue, 
        float64(vgpuStats.Utilization.EncUtil), labels...); err != nil {
        klog.Errorf("Failed to send encoder utilization metric for vGPU %d in Pod %s/%s: %v", 
            vgpuIndex, pod.Namespace, pod.Name, err)
    }

    // 为兼容性保留现有的容器级别vGPU指标收集
    // ...
}
```

### 4. 数据清理机制

为确保内存不会无限增长，实现了进程和Pod的清理机制：

```go
// 清理不活跃的进程数据
func (s *Spec) CleanupInactiveProcesses() {
    // ...
    // 清理超过5分钟未更新的进程
    if now-stats.LastUpdate > 300 {
        delete(s.stats, pid)
        // 同步清理共享内存
        // ...
    }
    // ...
}

// 清理不活跃的Pod数据
func (s *Spec) CleanupInactivePods() {
    // ...
    // 清理超过5分钟未更新的Pod
    if now-podStats.LastUpdate > 300 {
        delete(s.pods, podUID)
    }
    // ...
}
```

### 测试效果

使用Prometheus查询新增的指标，用户可以获取以下数据：

1. **Pod级SM利用率**：`pod_vgpu_sm_utilization{podnamespace="default", podname="my-gpu-pod"}`
2. **Pod级解码器利用率**：`pod_vgpu_dec_utilization{podnamespace="default", podname="my-gpu-pod"}`
3. **Pod级编码器利用率**：`pod_vgpu_enc_utilization{podnamespace="default", podname="my-gpu-pod"}`

这样用户可以清晰地看到每个Pod对分配给它的vGPU的使用情况，有助于分析业务逻辑是否合理，并优化资源分配。

## Test Plan

### 单元测试

我们添加了全面的单元测试，涵盖了以下功能：

1. `Test_UpdateProcessUtilization`：测试进程GPU利用率更新，包含多种情况
   - 正常更新利用率
   - 无效的设备索引（负数或超出范围）
   - 无效的利用率值（超过100%）

2. `Test_UpdatePodGPUUtilization`：测试Pod GPU利用率更新，包含多种情况
   - 正常更新利用率
   - 更新现有vGPU的信息
   - 无效的设备索引和利用率值

### 集成测试

部署多个使用GPU的Pod到节点，观察HAMi是否提供了这些Pod的实际GPU利用率：

1. 使用GPU密集型负载（如TensorFlow训练）部署Pod
2. 从Prometheus查询`pod_vgpu_sm_utilization`等指标
3. 验证指标值是否与期望的负载模式相符
4. 验证多个Pod的vGPU隔离效果

## Risks and Mitigations

由于我们修改了数据结构字段名称（从小写改为大写以导出字段），可能导致与旧版本不兼容。

缓解措施：
1. 保持内部数据结构的稳定性
2. 提供额外的公开接口，使外部调用与内部实现解耦
3. 添加全面的单元测试，确保功能正确