# 海光 DCU 支持简介

本组件支持复用海光 DCU 设备，并为此提供以下几种与 vGPU 类似的复用功能，包括：

***DCU 共享***: 每个任务可以只占用一部分显卡，多个任务可以共享一张显卡

***可限制分配的显存大小***: 你现在可以用显存值（例如3000M）来分配DCU，本组件会确保任务使用的显存不会超过分配数值

***可限制计算单元数量***: 你现在可以指定任务使用的算力比例（例如60即代表使用60%算力）来分配DCU，本组件会确保任务使用的算力不会超过分配数值

***指定DCU型号***：当前任务可以通过设置annotation("hygon.com/use-dcutype","hygon.com/nouse-dcutype")的方式，来选择使用或者不使用某些具体型号的DCU

## 节点需求

* dtk driver >= 24.04
* hy-smi v1.6.0

## 开启DCU复用

* 部署[dcu-vgpu-device-plugin](https://github.com/Project-HAMi/dcu-vgpu-device-plugin)

## 自定义 DCU 虚拟化参数

HAMi 支持通过以下方式自定义 DCU 虚拟化参数:

<details>
  <summary>自定义配置</summary>

  ### 在 HAMi charts 创建 files 的目录

  创建后的目录架构应为如下所示：

  ```bash
  tree -L 1
  .
  ├── Chart.yaml
  ├── files
  ├── templates
  └── values.yaml
  ```

  ### 在 files 目录下创建 device-config.yaml

  配置文件如下所示，可以按需调整：

  ```yaml
  hygon:
    resourceCountName: hygon.com/dcunum
    resourceMemoryName: hygon.com/dcumem
    resourceCoreName: hygon.com/dcucores
  ```

  ### Helm 安装和更新

  Helm 安装、更新将基于该配置文件，覆盖默认的配置文件

</details>

## 运行DCU任务

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: alexnet-tf-gpu-pod-mem
  labels:
    purpose: demo-tf-amdgpu
spec:
  containers:
    - name: alexnet-tf-gpu-container
      image: pytorch:resnet50
      workingDir: /root
      command: ["sleep","infinity"]
      resources:
        limits:
          hygon.com/dcunum: 1 # requesting a GPU
          hygon.com/dcumem: 2000 # each dcu require 2000 MiB device memory
          hygon.com/dcucores: 60 # each dcu use 60% of total compute cores

```

## 容器内开启虚拟DCU功能

使用vDCU首先需要激活虚拟环境
```
source /opt/hygondriver/env.sh
```

随后，使用hdmcli指令查看虚拟设备是否已经激活
```
hy-virtual -show-device-info
```

若输出如下，则代表虚拟设备已经成功激活
```
Device 0:
	Actual Device: 0
	Compute units: 60
	Global memory: 2097152000 bytes
```

接下来正常启动DCU任务即可

## 设备健康检查

HAMi 支持对海光 DCU 设备进行健康检查，确保只有健康的设备被分配给 Pod。健康检查包括以下内容：

- 设备状态检查
- 设备资源可用性检查
- 设备驱动状态检查

## 资源使用统计

HAMi 支持对海光 DCU 设备的资源使用情况进行统计，包括：

- 设备内存使用情况
- 计算核心使用情况
- 设备利用率

这些统计信息可以用于资源调度决策和性能优化。

## 节点锁定机制

HAMi 实现了节点锁定机制，确保在分配设备资源时不会发生冲突。当 Pod 请求海光 DCU 资源时，系统会锁定相应的节点，防止其他 Pod 同时使用相同的设备资源。

## 设备 UUID 选择

你可以通过 Pod 注解来指定要使用或排除特定的海光 DCU 设备：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: dcu-pod
  annotations:
    # 使用特定的 DCU 设备（逗号分隔的列表）
    hygon.com/use-gpuuuid: "device-uuid-1,device-uuid-2"
    # 或者排除特定的 DCU 设备（逗号分隔的列表）
    hygon.com/nouse-gpuuuid: "device-uuid-3,device-uuid-4"
spec:
  # ... 其余 Pod 配置
```

### 使用示例

以下是一个完整的示例，展示如何使用 UUID 选择功能：

<details>
  <summary>自定义配置</summary>

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: dcu-pod
  annotations:
    hygon.com/use-gpuuuid: "device-uuid-1,device-uuid-2"
spec:
  containers:
    - name: dcu-container
      image: pytorch:resnet50
      command: ["sleep", "infinity"]
      resources:
        limits:
          hygon.com/dcunum: 1
          hygon.com/dcumem: 2000
          hygon.com/dcucores: 60
```

在这个示例中，Pod 将只在 UUID 为 `device-uuid-1` 或 `device-uuid-2` 的海光 DCU 设备上运行。

#### 查找设备 UUID

你可以通过以下命令查找节点上的海光 DCU 设备 UUID：

```bash
kubectl describe node <node-name> | grep -A 10 "Allocated resources"
```

或者使用以下命令查看节点的注解：

```bash
kubectl get node <node-name> -o yaml | grep -A 10 "annotations:"
```

在节点注解中，查找 `hami.io/node-dcu-register` 或类似的注解，其中包含设备 UUID 信息。

</details>

## 注意事项

1. 在init container中无法使用DCU复用功能，否则该任务不会被调度

2. 每个容器最多只能使用一个虚拟DCU设备, 如果您希望在容器中挂载多个DCU设备，则不能使用`hygon.com/dcumem`和`hygon.com/dcucores`字段

3. `hygon.com/dcumem` 仅在 `hygon.com/dcunum=1` 时有效

4. 多设备请求（`hygon.com/dcunum > 1`）不支持 vDCU 模式
