# HAMi (Heterogeneous AI Computing Virtualization Middleware)

## 项目概述

HAMi 是一个用于 Kubernetes 的异构设备管理中间件，前身为 'k8s-vGPU-scheduler'。它能够管理不同类型的异构设备（如 GPU、NPU 等），实现设备共享，并基于设备拓扑和调度策略做出更好的调度决策。

### 项目目标
- 消除不同异构设备之间的差距
- 为用户提供统一的接口，无需修改应用程序
- 支持多种异构设备的虚拟化和资源隔离
- 提供高效的设备共享机制
- 实现智能调度和资源优化

### 项目状态
- CNCF Sandbox 项目
- 被广泛用于互联网、公有云和私有云
- 在金融、证券、能源、电信、教育和制造业等多个垂直行业得到采用
- 超过 50 家公司或机构既是最终用户也是活跃贡献者
- 持续活跃开发和社区贡献

## 核心功能

### 1. 设备虚拟化
- 支持设备共享
  - 通过指定设备核心使用量进行部分设备分配
  - 通过指定设备内存进行部分设备分配
  - 对流式多处理器实施硬限制
  - 无需修改现有程序
  - 支持动态 MIG 功能
- 支持多种虚拟化模式
  - HAMi-core 模式：适用于大多数 GPU 设备
  - MIG 模式：适用于 NVIDIA Ampere 及更新架构的 GPU
  - 混合模式：支持在同一个集群中混合使用不同虚拟化模式

### 2. 设备资源隔离
- 提供设备内存隔离
  - 支持按百分比分配内存
  - 支持按固定大小分配内存
  - 硬限制确保资源使用不超过分配量
- 支持设备核心隔离
  - 支持按百分比分配计算核心
  - 支持动态调整核心使用率
- 确保资源使用的硬限制
  - 通过容器内控制机制实现
  - 支持 OOM 监控和自动终止超限进程

### 3. 智能调度策略
- 节点级调度策略
  - binpack：尽可能将任务分配到同一个 GPU 节点
  - spread：尽可能将任务分配到不同的 GPU 节点
- GPU 级调度策略
  - binpack：尽可能将任务分配到同一个 GPU 卡
  - spread：尽可能将任务分配到不同的 GPU 卡
- 支持通过 Pod 注解覆盖默认策略
- 支持基于设备拓扑的调度决策

### 4. 支持的设备
- NVIDIA GPU：支持所有型号，提供内存和核心隔离，支持多卡
- Cambricon MLU：支持 370、590 型号，提供内存和核心隔离
- Hygon DCU：支持 Z100、Z100L 型号，提供内存和核心隔离
- Ascend NPU：支持 910B、910B3、910B4、310P 型号，提供内存和核心隔离
- Iluvatar GPU：支持所有型号，提供内存和核心隔离
- Mthreads GPU：提供内存和核心隔离
- Metax GPU：提供内存和核心隔离

## 系统架构

HAMi 由以下组件组成：

### 1. MutatingWebhook
- 检查每个任务的有效性
- 如果资源请求被 HAMi 识别，则将 "schedulerName" 设置为 "HAMi scheduler"
- 否则，不做任何操作并将任务传递给默认调度器
- 支持自定义资源名称和默认值

### 2. Scheduler
- 支持默认 kube-scheduler 和 volcano-scheduler
- 实现扩展器并注册 'Filter' 和 'Score' 方法来处理可共享设备
- 当带有可共享设备请求的 Pod 到达时，'Filter' 搜索集群并返回"可用"节点列表
- 'Score' 对 'Filter' 返回的每个节点进行评分，并选择得分最高的节点来托管 Pod
- 在相应的 Pod 注解中修补调度决策
- 支持自定义调度策略和评分算法

### 3. DevicePlugin
- 当做出调度决策时，调度器调用该节点上的 devicePlugin 来根据 Pod 注解生成环境变量和挂载
- 这里使用的 DP 是定制版本，需要根据 README 文档安装
- 支持动态 MIG 实例管理
- 支持设备资源监控和报告

### 4. InContainer Control
- 不同设备的容器内硬限制实现不同
- HAMi-Core 负责 NVIDIA 设备
- libvgpu-control.so 负责 iluvatar 设备等
- HAMi 需要传递正确的环境变量以便其操作
- 支持资源使用监控和限制

## 设备管理架构

HAMi 的设备管理架构位于 `pkg/device` 目录，提供了统一的设备管理接口和多种设备类型的实现。

### 1. 核心接口
- `Devices` 接口定义了设备管理的标准方法：
  - `CommonWord()`：返回设备的通用名称
  - `MutateAdmission()`：修改容器配置以适应设备需求
  - `CheckHealth()`：检查设备健康状态
  - `NodeCleanUp()`：清理节点上的设备资源
  - `GetNodeDevices()`：获取节点上的设备信息
  - `CheckType()`：检查设备类型是否匹配
  - `CheckUUID()`：检查设备 UUID 是否匹配
  - `LockNode()`：锁定节点资源
  - `ReleaseNodeLock()`：释放节点锁
  - `GenerateResourceRequests()`：生成资源请求
  - `PatchAnnotations()`：修补 Pod 注解
  - `CustomFilterRule()`：自定义过滤规则
  - `ScoreNode()`：节点评分
  - `AddResourceUsage()`：添加资源使用记录

### 2. 设备类型实现
每种设备类型都有独立的实现目录，包含以下文件：
- `device.go`：设备管理的核心实现
- `device_test.go`：设备管理的测试用例
- 其他特定文件（如配置、协议等）

#### 2.1 NVIDIA 设备
- 位置：`pkg/device/nvidia/`
- 主要文件：
  - `device.go`：NVIDIA GPU 设备管理的核心实现
  - `device_test.go`：NVIDIA 设备管理的测试用例
- 特性：
  - 支持 MIG 模式
  - 支持多卡管理
  - 提供内存和核心隔离
  - 支持动态资源分配

#### 2.2 Ascend NPU
- 位置：`pkg/device/ascend/`
- 主要文件：
  - `device.go`：Ascend NPU 设备管理的核心实现
  - `vnpu.go`：虚拟 NPU 配置
  - `device_test.go`：Ascend 设备管理的测试用例
- 特性：
  - 支持 910B、910B3、910B4、310P 型号
  - 提供内存和核心隔离
  - 支持虚拟 NPU 配置
  - 支持设备模板配置
  - 支持设备 UUID 过滤
  - 支持设备健康检查
  - 支持资源使用统计
  - 支持节点锁定机制
  - 支持内存自动对齐到模板

#### 2.3 Cambricon MLU
- 位置：`pkg/device/cambricon/`
- 主要文件：
  - `device.go`：Cambricon MLU 设备管理的核心实现
  - `device_test.go`：Cambricon 设备管理的测试用例
- 特性：
  - 支持 370、590 型号
  - 提供内存和核心隔离

#### 2.4 Hygon DCU
- 位置：`pkg/device/hygon/`
- 主要文件：
  - `device.go`：Hygon DCU 设备管理的核心实现
  - `device_test.go`：Hygon 设备管理的测试用例
- 特性：
  - 支持 Z100、Z100L 型号
  - 提供内存和核心隔离

#### 2.5 Iluvatar GPU
- 位置：`pkg/device/iluvatar/`
- 主要文件：
  - `device.go`：Iluvatar GPU 设备管理的核心实现
  - `device_test.go`：Iluvatar 设备管理的测试用例
- 特性：
  - 支持所有型号
  - 提供内存和核心隔离
  - 支持设备 UUID 选择
  - 支持设备共享
  - 支持多 GPU 请求
  - 支持资源请求和限制
  - 支持设备健康检查
  - 支持资源使用统计
  - 支持节点锁定机制
  - 支持内存自动对齐到模板

#### 2.6 Metax GPU
- 位置：`pkg/device/metax/`
- 主要文件：
  - `device.go`：Metax GPU 设备管理的核心实现
  - `sdevice.go`：共享设备管理
  - `protocol.go`：设备通信协议
  - `config.go`：设备配置
  - 测试文件：`device_test.go`、`sdevice_test.go`、`protocol_test.go`
- 特性：
  - 提供内存和核心隔离
  - 支持设备共享
  - 自定义通信协议

#### 2.7 Mthreads GPU
- 位置：`pkg/device/mthreads/`
- 主要文件：
  - `device.go`：Mthreads GPU 设备管理的核心实现
  - `device_test.go`：Mthreads 设备管理的测试用例
- 特性：
  - 提供内存和核心隔离

### 3. 配置管理
- `Config` 结构体定义了所有设备类型的配置：
  ```go
  type Config struct {
      NvidiaConfig    nvidia.NvidiaConfig
      MetaxConfig     metax.MetaxConfig
      HygonConfig     hygon.HygonConfig
      CambriconConfig cambricon.CambriconConfig
      MthreadsConfig  mthreads.MthreadsConfig
      IluvatarConfig  iluvatar.IluvatarConfig
      VNPUs           []ascend.VNPUConfig
  }
  ```
- 支持通过 YAML 配置文件加载设备配置
- 提供配置验证功能

### 4. 设备初始化
- `InitDevicesWithConfig()`：使用配置初始化设备
- `InitDevices()`：初始化所有设备
- `InitDefaultDevices()`：初始化默认设备
- 支持动态加载和卸载设备

### 5. Pod 分配管理
- `PodAllocationTrySuccess()`：尝试分配 Pod 成功
- `PodAllocationSuccess()`：Pod 分配成功
- `PodAllocationFailed()`：Pod 分配失败
- 支持 Pod 注解的更新和锁的释放

## 设备插件实现

HAMi 的设备插件实现位于 `cmd/device-plugin` 目录，目前主要支持 NVIDIA GPU 设备。设备插件是 Kubernetes 设备插件框架的一部分，负责在节点上发现、分配和管理设备资源。

### 1. NVIDIA 设备插件

#### 1.1 主要文件
- 位置：`cmd/device-plugin/nvidia/`
- 主要文件：
  - `main.go`：设备插件的主入口，负责初始化和启动设备插件
  - `plugin-manager.go`：管理设备插件的生命周期
  - `vgpucfg.go`：处理 vGPU 配置
  - `watchers.go`：监控设备状态变化

#### 1.2 核心功能
- **设备发现**：使用 NVML 库发现节点上的 NVIDIA GPU 设备
- **资源分配**：根据 Pod 请求分配 GPU 资源
- **MIG 支持**：支持 NVIDIA MIG 模式，可以创建和管理 MIG 实例
- **设备监控**：监控设备状态和健康情况
- **配置管理**：支持通过命令行参数和配置文件进行配置

#### 1.3 配置选项
- **MIG 策略**：
  - `none`：不使用 MIG 模式
  - `single`：所有 GPU 使用相同的 MIG 配置
  - `mixed`：允许不同的 GPU 使用不同的 MIG 配置
- **设备分割**：
  - `device-split-count`：设备分割数量
  - `device-memory-scaling`：设备内存缩放比例
  - `device-cores-scaling`：设备核心缩放比例
- **资源限制**：
  - `disable-core-limit`：是否禁用核心限制
- **资源名称**：
  - `resource-name`：容器中可见的 GPU 数量字段名称

#### 1.4 插件管理器
- `NewPluginManager()`：创建基于 NVML 的插件管理器
- 支持 CDI（Container Device Interface）规范
- 提供设备列表策略配置
- 支持 GDS（GPU Direct Storage）和 MOFED 功能

#### 1.5 文件监控
- `newFSWatcher()`：创建文件系统监视器，用于监控配置文件变化
- `newOSWatcher()`：创建操作系统信号监视器，用于处理终止信号

#### 1.6 设备配置生成
- `generateDeviceConfigFromNvidia()`：从 NVIDIA 配置生成设备配置
- 支持从命令行参数和配置文件加载配置
- 提供配置验证和错误处理

### 2. 设备插件工作流程

1. **初始化**：
   - 加载配置文件和命令行参数
   - 初始化 NVML 库
   - 创建插件管理器
   - 设置文件监视器

2. **设备发现**：
   - 使用 NVML 发现节点上的 GPU 设备
   - 根据 MIG 策略配置 MIG 实例
   - 注册设备资源

3. **资源分配**：
   - 接收来自 kubelet 的分配请求
   - 根据请求分配 GPU 资源
   - 生成设备规格和环境变量
   - 返回分配结果

4. **设备监控**：
   - 定期检查设备健康状态
   - 监控设备资源使用情况
   - 报告设备状态变化

5. **配置更新**：
   - 监控配置文件变化
   - 重新加载配置
   - 更新设备状态

### 3. 与其他组件的交互

- **与 kubelet 交互**：通过 gRPC 接口与 kubelet 通信
- **与调度器交互**：通过 Pod 注解传递设备分配信息
- **与容器运行时交互**：通过 CDI 规范与容器运行时交互
- **与监控系统交互**：提供 Prometheus 格式的指标

## Ascend 910B 支持

HAMi 支持华为 Ascend 910B 系列 NPU 设备，包括 910B、910B3、910B4 和 310P 型号。通过实现与 NVIDIA GPU 类似的设备共享功能，HAMi 使 Ascend NPU 能够在 Kubernetes 集群中高效共享。

### 1. 支持的功能

#### 1.1 NPU 共享
- 支持将单个 Ascend NPU 分配给多个任务
- 通过内存和核心隔离实现资源隔离
- 支持动态资源分配和回收
- 无需修改现有应用程序

#### 1.2 设备内存控制
- 支持按固定大小分配设备内存
- 支持内存使用硬限制
- 自动将请求的内存大小对齐到预定义的模板
- 支持内存使用统计和监控

#### 1.3 设备核心控制
- 支持按固定数量分配 AI 核心
- 支持 AI CPU 核心分配
- 支持核心使用统计和监控

#### 1.4 设备模板
- 支持预定义的设备模板配置
- 每个模板包含内存大小、AI 核心数量和 AI CPU 核心数量
- 支持多种模板配置，适应不同工作负载需求
- 自动选择最接近请求大小的模板

### 2. 前提条件

- Ascend 设备类型：910B、910B3、910B4、310P
- 驱动版本 >= 24.1.rc1
- Ascend docker runtime

### 3. 启用 Ascend 共享支持

1. 使用 Helm 安装 HAMi 图表
2. 为 Ascend-910B 节点添加标签：
   ```bash
   kubectl label node {ascend-node} accelerator=huawei-Ascend910
   ```
3. 安装 Ascend docker runtime
4. 部署 Ascend-vgpu-device-plugin：
   ```bash
   wget https://raw.githubusercontent.com/Project-HAMi/ascend-device-plugin/master/build/ascendplugin-910-hami.yaml
   kubectl apply -f ascendplugin-910-hami.yaml
   ```

### 4. 自定义 Ascend 共享配置

HAMi 提供了内置的 Ascend 共享配置，但用户可以根据需要自定义配置。配置包括以下内容：

- 芯片名称（chipName）
- 通用名称（commonWord）
- 资源名称（resourceName）
- 资源内存名称（resourceMemoryName）
- 可分配内存（memoryAllocatable）
- 内存容量（memoryCapacity）
- AI 核心数量（aiCore）
- AI CPU 核心数量（aiCPU）
- 设备模板（templates）

每个模板包含：
- 模板名称（name）
- 内存大小（memory）
- AI 核心数量（aiCore）
- AI CPU 核心数量（aiCPU）

### 5. 运行 Ascend 作业

用户可以通过以下方式请求 Ascend 910B 资源：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: gpu-pod
spec:
  containers:
    - name: ubuntu-container
      image: ascendhub.huawei.com/public-ascendhub/ascend-mindspore:23.0.RC3-centos7
      command: ["bash", "-c", "sleep 86400"]
      resources:
        limits:
          huawei.com/Ascend910: 1 # 请求 1 个 vNPU
          huawei.com/Ascend910-memory: 2000 # 请求 2000m 设备内存
```

### 6. 注意事项

1. 内存请求会自动对齐到最接近的模板大小
2. 不支持在 init 容器中使用 Ascend-910B 共享
3. `huawei.com/Ascend910-memory` 仅在 `huawei.com/Ascend910=1` 时有效
4. 多设备请求（`huawei.com/Ascend910 > 1`）不支持 vNPU 模式
5. 支持通过注解指定使用特定 UUID 的设备
6. 支持通过注解排除特定 UUID 的设备
7. 支持设备健康检查和监控

## 命令行工具

HAMi 项目包含三个主要的命令行工具，分别对应系统的不同组件：

### 1. device-plugin
- 位置：`cmd/device-plugin/nvidia/`
- 主要文件：
  - `main.go`：设备插件的主入口，负责初始化和启动设备插件
  - `plugin-manager.go`：管理设备插件的生命周期
  - `vgpucfg.go`：处理 vGPU 配置
  - `watchers.go`：监控设备状态变化
- 功能：
  - 发现和管理 NVIDIA GPU 设备
  - 实现设备资源的分配和回收
  - 支持动态 MIG 实例管理
  - 提供设备状态监控和报告
  - 生成容器所需的环境变量和挂载点

### 2. scheduler
- 位置：`cmd/scheduler/`
- 主要文件：
  - `main.go`：调度器的主入口，负责初始化和启动调度器
  - `metrics.go`：定义和收集调度器相关的指标
- 功能：
  - 实现 Kubernetes 调度器扩展
  - 提供设备感知的调度决策
  - 支持自定义调度策略（binpack/spread）
  - 收集和报告调度相关的指标
  - 提供 HTTP API 接口

### 3. vGPUmonitor
- 位置：`cmd/vGPUmonitor/`
- 主要文件：
  - `main.go`：监控工具的主入口
  - `metrics.go`：定义和收集 vGPU 相关的指标
  - `feedback.go`：提供资源使用反馈
  - `validation.go`：验证配置和状态
- 功能：
  - 监控 vGPU 资源使用情况
  - 收集性能指标和资源利用率
  - 提供 Prometheus 格式的指标
  - 支持资源使用反馈和优化建议
  - 监控容器级别的资源使用情况

## 技术栈

- 编程语言：Go 1.22.2
- 主要依赖：
  - Kubernetes 相关组件 (v0.28.3)
  - NVIDIA 相关库
    - go-gpuallocator
    - go-nvlib
    - go-nvml
    - nvidia-container-toolkit
  - Prometheus 监控
  - gRPC
  - 其他云原生工具和库
- 容器运行时支持：
  - containerd
  - docker
  - cri-o

## 部署要求

### 前提条件
- NVIDIA 驱动 >= 440
- nvidia-docker 版本 > 2.0
- containerd/docker/cri-o 容器运行时配置 nvidia 为默认运行时
- Kubernetes 版本 >= 1.16
- glibc >= 2.17 & glibc < 2.30
- 内核版本 >= 3.10
- helm > 3.0

### 安装方式
- 支持通过 Helm 进行部署
- 提供 WebUI 界面和监控功能
- 支持离线安装
- 支持自定义配置

### 配置选项
- 设备配置：
  - 设备内存缩放比例
  - 设备分割数量
  - MIG 策略
  - 核心限制
  - 默认内存和核心设置
  - 资源名称自定义
- 调度器配置：
  - 节点调度策略
  - GPU 调度策略
- Pod 配置：
  - GPU UUID 选择
  - GPU 类型选择
  - 调度策略覆盖

## 监控和可视化

- 支持 Prometheus 和 Grafana 集成
- 提供预配置的仪表盘
- 监控 GPU 使用情况、内存使用情况、核心使用情况等
- 支持 MIG 实例监控
- 支持自定义监控指标
- 提供资源利用率分析

## 项目特点

1. 开源协议：Apache 2.0
2. 活跃的社区支持
3. 完善的文档和示例
4. 支持多种异构设备
5. 提供监控和可视化界面
6. 支持动态资源分配
7. 灵活的调度策略
8. 高效的资源隔离
9. 无需修改应用程序
10. 支持多种虚拟化模式

## 使用场景

1. 大规模 AI 训练集群
2. 云计算平台
3. 企业私有云
4. 科研计算环境
5. 工业应用场景
6. 高性能计算
7. 机器学习推理服务
8. 图形渲染农场
9. 视频处理集群
10. 科学计算环境

## 项目规划

1. 持续优化设备虚拟化性能
2. 扩展支持更多类型的异构设备
   - 支持视频编解码处理
   - 支持 Intel GPU 设备
   - 支持 AMD GPU 设备
3. 增强调度策略的灵活性
   - 支持 NUMA 亲和性
   - 支持 DRA (Dynamic Resource Allocation)
4. 改进监控和告警功能
   - 集成 gpu-operator
   - 丰富的可观察性支持
5. 优化资源利用率
   - 支持更细粒度的资源分配
   - 支持动态资源调整

## 贡献指南

项目欢迎社区贡献，包括但不限于：
1. 代码贡献
2. 文档改进
3. 问题报告
4. 功能建议
5. 测试用例
6. 性能优化
7. 新设备支持
8. 使用案例分享

详细的贡献指南请参考 CONTRIBUTING.md 文件。

## 社区和资源

- 官方网站：http://project-hami.io
- GitHub 仓库：https://github.com/Project-HAMi/HAMi
- 社区会议：每周五 16:00 UTC+8 (中文)
- 邮件列表：https://groups.google.com/forum/#!forum/hami-project
- Slack 频道：https://cloud-native.slack.com/archives/C07T10BU4R2
- 讨论区：https://github.com/Project-HAMi/HAMi/discussions 

## vGPU利用率监控功能

HAMi最近增强了vGPU利用率监控功能，使用户能够更精细地监控Pod级别的GPU使用情况。本节详细介绍此功能的设计和使用方法。

### 1. 功能概述

vGPU利用率监控功能允许用户观察每个Pod对其分配的vGPU的使用情况，主要特性包括：

- 监控Pod级别的vGPU利用率数据
- 支持多种利用率指标（SM、编码器、解码器）
- 提供Prometheus格式的指标
- 支持细粒度的资源使用分析
- 无需修改现有应用程序

### 2. 数据结构设计

vGPU利用率监控功能的核心数据结构设计如下：

```go
// 设备利用率结构
type deviceUtilization struct {
    DecUtil uint64 // 解码器利用率
    EncUtil uint64 // 编码器利用率
    SmUtil  uint64 // SM利用率
}

// GPU利用率数据
type GPUUtilization struct {
    DecUtil   uint64 // 解码器利用率
    EncUtil   uint64 // 编码器利用率
    SmUtil    uint64 // SM利用率
    Timestamp int64  // 时间戳
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

### 3. 关键方法实现

#### 3.1 进程GPU利用率更新

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

    // 更新进程统计信息
    // 同步到共享内存
    // ...
}
```

#### 3.2 Pod GPU利用率更新

```go
func (s *Spec) UpdatePodGPUUtilization(podUID string, namespace string, name string, vgpuIndex int, vgpuUUID string, util DeviceUtilization) error {
    // 参数验证
    // 更新Pod统计信息
    // 更新vGPU统计信息，包括UUID
    // 更新利用率数据
    // 同步到设备利用率
    // ...
}
```

### 4. 指标暴露

HAMi通过Prometheus格式暴露以下vGPU利用率指标：

1. `pod_vgpu_sm_utilization`: Pod的vGPU SM利用率
2. `pod_vgpu_dec_utilization`: Pod的vGPU解码器利用率
3. `pod_vgpu_enc_utilization`: Pod的vGPU编码器利用率

用户可以通过以下方式查询这些指标：

```
pod_vgpu_sm_utilization{podnamespace="default", podname="my-gpu-pod"}
```

### 5. 使用方法

vGPU利用率监控功能已内置于HAMi，无需额外配置即可使用。要查看指标，您可以：

1. 部署Prometheus和Grafana
2. 配置Prometheus抓取HAMi暴露的指标
3. 创建自定义仪表盘以可视化数据
4. 使用以下示例查询分析Pod的GPU利用率：

```
# 获取所有Pod的SM利用率
pod_vgpu_sm_utilization

# 获取特定命名空间的Pod SM利用率
pod_vgpu_sm_utilization{podnamespace="ml-training"}

# 按Pod名称过滤
pod_vgpu_sm_utilization{podname=~"training-.*"}

# 计算命名空间平均利用率
avg(pod_vgpu_sm_utilization) by (podnamespace)
```

### 6. 使用场景

vGPU利用率监控功能适用于以下场景：

1. **资源优化**: 识别GPU资源使用不足或过度分配的场景
2. **容量规划**: 基于实际使用情况进行集群扩容规划
3. **性能调优**: 发现应用程序的GPU使用瓶颈
4. **成本分析**: 按团队或项目分析GPU资源使用成本
5. **异常检测**: 识别异常的GPU使用模式

### 7. 最佳实践

1. **设置告警**: 为低利用率和高利用率配置Prometheus告警规则
2. **创建仪表盘**: 设计直观的Grafana仪表盘展示利用率趋势
3. **定期审计**: 定期审查利用率数据，调整资源分配
4. **基准测试**: 建立应用程序的基准利用率模式
5. **趋势分析**: 分析长期利用率趋势，预测未来资源需求

### 8. 示例仪表盘

创建名为"HAMi vGPU Utilization"的Grafana仪表盘，包含以下面板：

1. Pod vGPU SM利用率热图
2. 每个命名空间的平均利用率
3. 利用率最高的前5个Pod
4. 利用率最低的前5个Pod
5. 各类型利用率(SM/Dec/Enc)对比
6. 利用率历史趋势

通过此功能，用户可以全面了解其工作负载的GPU使用情况，实现更高效的资源分配和成本优化。 