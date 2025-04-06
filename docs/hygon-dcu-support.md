# Hygon DCU Support Introduction

This component supports Hygon DCU device sharing and provides several features similar to vGPU, including:

***DCU sharing***: Each task can allocate a portion of DCU instead of a whole DCU card, thus DCU can be shared among multiple tasks.

***Device Memory Control***: DCUs can be allocated with certain device memory size (e.g., 3000M) and the component ensures that the task's memory usage does not exceed the allocated value.

***Device compute core limitation***: DCUs can be allocated with certain percentage of device core (e.g., hygon.com/dcucores:60 indicates this container uses 60% compute cores of this device).

***DCU Type Specification***: You can specify which type of DCU to use or to avoid for a certain task, by setting "hygon.com/use-dcutype" or "hygon.com/nouse-dcutype" annotations.

## Prerequisites

* dtk driver >= 24.04
* hy-smi v1.6.0

## Enabling DCU-sharing Support

* Deploy the dcu-vgpu-device-plugin [here](https://github.com/Project-HAMi/dcu-vgpu-device-plugin)

## Customizing DCU Virtualization Parameters

HAMi supports customizing DCU virtualization parameters through the following methods:

<details>
  <summary>Custom Configuration</summary>

  ### Create a files directory in HAMi charts

  The directory structure should look like this:

  ```bash
  tree -L 1
  .
  ├── Chart.yaml
  ├── files
  ├── templates
  └── values.yaml
  ```

  ### Create device-config.yaml in the files directory

  The configuration file is as follows, which can be adjusted as needed:

  ```yaml
  hygon:
    resourceCountName: hygon.com/dcunum
    resourceMemoryName: hygon.com/dcumem
    resourceCoreName: hygon.com/dcucores
  ```

  ### Helm Installation and Update

  Helm installation and updates will be based on this configuration file, overriding the default configuration file.

</details>

## Running DCU jobs

Hygon DCUs can now be requested by a container
using the `hygon.com/dcunum` , `hygon.com/dcumem` and `hygon.com/dcucores` resource type:

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

## Enable vDCU inside container

You need to enable vDCU inside container in order to use it.
```
source /opt/hygondriver/env.sh
```

check if you have successfully enabled vDCU by using following command

```
hy-virtual -show-device-info
```

If you have an output like this, then you have successfully enabled vDCU inside container.

```
Device 0:
	Actual Device: 0
	Compute units: 60
	Global memory: 2097152000 bytes
```

Launch your DCU tasks like you usually do

## Device Health Check

HAMi supports health checks for Hygon DCU devices to ensure that only healthy devices are allocated to Pods. Health checks include:

- Device status check
- Device resource availability check
- Device driver status check

## Resource Usage Statistics

HAMi supports statistics on Hygon DCU device resource usage, including:

- Device memory usage
- Compute core usage
- Device utilization

These statistics can be used for resource scheduling decisions and performance optimization.

## Node Locking Mechanism

HAMi implements a node locking mechanism to ensure that there are no conflicts when allocating device resources. When a Pod requests Hygon DCU resources, the system locks the corresponding node to prevent other Pods from using the same device resources simultaneously.

## Device UUID Selection

You can specify which Hygon DCU devices to use or exclude through Pod annotations:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: dcu-pod
  annotations:
    # Use specific DCU devices (comma-separated list)
    hygon.com/use-gpuuuid: "device-uuid-1,device-uuid-2"
    # Or exclude specific DCU devices (comma-separated list)
    hygon.com/nouse-gpuuuid: "device-uuid-3,device-uuid-4"
spec:
  # ... rest of Pod configuration
```

### Usage Example

Here's a complete example showing how to use the UUID selection feature:

<details>
  <summary>Custom Configuration</summary>

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

In this example, the Pod will only run on Hygon DCU devices with UUIDs `device-uuid-1` or `device-uuid-2`.

#### Finding Device UUIDs

You can find Hygon DCU device UUIDs on a node using the following command:

```bash
kubectl describe node <node-name> | grep -A 10 "Allocated resources"
```

Or use the following command to view node annotations:

```bash
kubectl get node <node-name> -o yaml | grep -A 10 "annotations:"
```

In the node annotations, look for `hami.io/node-dcu-register` or similar annotations, which contain device UUID information.

</details>

## Notes

1. DCU-sharing in init container is not supported, pods with "hygon.com/dcumem" in init container will never be scheduled.

2. Only one vdcu can be aquired per container. If you want to mount multiple dcu devices, then you shouldn't set `hygon.com/dcumem` or `hygon.com/dcucores`

3. `hygon.com/dcumem` is only valid when `hygon.com/dcunum=1`

4. Multiple device requests (`hygon.com/dcunum > 1`) do not support vDCU mode
