/*
Copyright 2024 The HAMi Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Project-HAMi/HAMi/pkg/monitor/nvidia"
	v1 "github.com/Project-HAMi/HAMi/pkg/monitor/nvidia/v1"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	////v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

var cgroupDriver int

//type hostGPUPid struct {
//	hostGPUPid int
//	mtime      uint64
//}

type UtilizationPerDevice []int

func setcGgroupDriver() int {
	// 1 for cgroupfs 2 for systemd
	kubeletconfig, err := os.ReadFile("/hostvar/lib/kubelet/config.yaml")
	if err != nil {
		return 0
	}
	content := string(kubeletconfig)
	pos := strings.LastIndex(content, "cgroupDriver:")
	if pos < 0 {
		return 0
	}
	if strings.Contains(content, "systemd") {
		return 2
	}
	if strings.Contains(content, "cgroupfs") {
		return 1
	}
	return 0
}

func getUsedGPUPid() ([]uint, nvml.Return) {
	tmp := []nvml.ProcessInfo{}
	count, err := nvml.DeviceGetCount()
	if err != nvml.SUCCESS {
		return []uint{}, err
	}
	for i := 0; i < count; i++ {
		device, err := nvml.DeviceGetHandleByIndex(i)
		if err != nvml.SUCCESS {
			return []uint{}, err
		}
		ids, err := device.GetComputeRunningProcesses()
		if err != nvml.SUCCESS {
			return []uint{}, err
		}
		tmp = append(tmp, ids...)
	}
	result := make([]uint, 0)
	m := make(map[uint]bool)
	for _, v := range tmp {
		if _, ok := m[uint(v.Pid)]; !ok {
			result = append(result, uint(v.Pid))
			m[uint(v.Pid)] = true
		}
	}
	sort.Slice(tmp, func(i, j int) bool { return tmp[i].Pid > tmp[j].Pid })
	return result, nvml.SUCCESS
}

//func setHostPid(pod corev1.Pod, ctr corev1.ContainerStatus, sr *podusage) error {
//	var pids []string
//	mutex.Lock()
//	defer mutex.Unlock()
//
//	if cgroupDriver == 0 {
//		cgroupDriver = setcGgroupDriver()
//	}
//	if cgroupDriver == 0 {
//		return errors.New("can not identify cgroup driver")
//	}
//	usedGPUArray, err := getUsedGPUPid()
//	if err != nvml.SUCCESS {
//		return errors.New("get usedGPUID failed, ret:" + nvml.ErrorString(err))
//	}
//	if len(usedGPUArray) == 0 {
//		return nil
//	}
//	qos := strings.ToLower(string(pod.Status.QOSClass))
//	var filename string
//	if cgroupDriver == 1 {
//		/* Cgroupfs */
//		filename = fmt.Sprintf("/sysinfo/fs/cgroup/memory/kubepods/%s/pod%s/%s/tasks", qos, pod.UID, strings.TrimPrefix(ctr.ContainerID, "docker://"))
//	}
//	if cgroupDriver == 2 {
//		/* Systemd */
//		cgroupuid := strings.ReplaceAll(string(pod.UID), "-", "_")
//		filename = fmt.Sprintf("/sysinfo/fs/cgroup/systemd/kubepods.slice/kubepods-%s.slice/kubepods-%s-pod%s.slice/docker-%s.scope/tasks", qos, qos, cgroupuid, strings.TrimPrefix(ctr.ContainerID, "docker://"))
//	}
//	fmt.Println("filename=", filename)
//	content, ferr := os.ReadFile(filename)
//	if ferr != nil {
//		return ferr
//	}
//	pids = strings.Split(string(content), "\n")
//	hostPidArray := []hostGPUPid{}
//	for _, val := range pids {
//		tmp, _ := strconv.Atoi(val)
//		if tmp != 0 {
//			var stat os.FileInfo
//			var err error
//			if stat, err = os.Lstat(fmt.Sprintf("/proc/%v", tmp)); err != nil {
//				return err
//			}
//			mtime := stat.ModTime().Unix()
//			hostPidArray = append(hostPidArray, hostGPUPid{
//				hostGPUPid: tmp,
//				mtime:      uint64(mtime),
//			})
//		}
//	}
//	usedGPUHostArray := []hostGPUPid{}
//	for _, val := range usedGPUArray {
//		for _, hostpid := range hostPidArray {
//			if uint(hostpid.hostGPUPid) == val {
//				usedGPUHostArray = append(usedGPUHostArray, hostpid)
//			}
//		}
//	}
//	//fmt.Println("usedHostGPUArray=", usedGPUHostArray)
//	sort.Slice(usedGPUHostArray, func(i, j int) bool { return usedGPUHostArray[i].mtime > usedGPUHostArray[j].mtime })
//	if sr == nil || sr.sr == nil {
//		return nil
//	}
//	for idx, val := range sr.sr.procs {
//		//fmt.Println("pid=", val.pid)
//		if val.pid == 0 {
//			break
//		}
//		if idx < len(usedGPUHostArray) {
//			if val.hostpid == 0 || val.hostpid != int32(usedGPUHostArray[idx].hostGPUPid) {
//				fmt.Println("Assign host pid to pid instead", usedGPUHostArray[idx].hostGPUPid, val.pid, val.hostpid)
//				sr.sr.procs[idx].hostpid = int32(usedGPUHostArray[idx].hostGPUPid)
//				fmt.Println("val=", val.hostpid, sr.sr.procs[idx].hostpid)
//			}
//		}
//	}
//	return nil
//
//}

func CheckBlocking(utSwitchOn map[string]UtilizationPerDevice, p int, c *nvidia.ContainerUsage) bool {
	for i := 0; i < c.Info.DeviceMax(); i++ {
		uuid := c.Info.DeviceUUID(i)
		_, ok := utSwitchOn[uuid]
		if ok {
			for i := 0; i < p; i++ {
				if utSwitchOn[uuid][i] > 0 {
					return true
				}
			}
			return false
		}
	}
	return false
}

// Check whether task with higher priority use GPU or there are other tasks with the same priority.
func CheckPriority(utSwitchOn map[string]UtilizationPerDevice, p int, c *nvidia.ContainerUsage) bool {
	for i := 0; i < c.Info.DeviceMax(); i++ {
		uuid := c.Info.DeviceUUID(i)
		_, ok := utSwitchOn[uuid]
		if ok {
			for i := 0; i < p; i++ {
				if utSwitchOn[uuid][i] > 0 {
					return true
				}
			}
			if utSwitchOn[uuid][p] > 1 {
				return true
			}
		}
	}
	return false
}

func Observe(lister *nvidia.ContainerLister) {
	utSwitchOn := map[string]UtilizationPerDevice{}
	containers := lister.ListContainers()

	// 新增：清理不活跃的进程和Pod数据
	for _, c := range containers {
		// 如果容器包含Spec指针，清理不活跃的进程和Pod
		if spec, ok := c.Info.(*v1.Spec); ok {
			spec.CleanupInactiveProcesses()
			spec.CleanupInactivePods()
		}
	}

	// 首先收集每个容器的GPU利用率信息，并关联到对应的Pod
	for _, c := range containers {
		if c.PodUID == "" {
			continue
		}

		// 获取Pod信息
		pods, err := lister.Clientset().CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.uid=%s", c.PodUID),
		})
		if err != nil || len(pods.Items) == 0 {
			klog.V(4).Infof("Failed to find Pod with UID %s: %v", c.PodUID, err)
			continue
		}
		pod := pods.Items[0]

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

			// 将数据转换为deviceUtilization格式
			if spec, ok := c.Info.(*v1.Spec); ok {
				// 创建DeviceUtilization结构体
				util := v1.DeviceUtilization{
					SmUtil:  smUtil,
					DecUtil: decUtil,
					EncUtil: encUtil,
				}

				// 更新Pod的vGPU利用率
				err := spec.UpdatePodGPUUtilization(
					c.PodUID,
					pod.Namespace,
					pod.Name,
					i,
					uuid,
					util,
				)
				if err != nil {
					klog.Errorf("Failed to update Pod GPU utilization for Pod %s/%s, device %d: %v",
						pod.Namespace, pod.Name, i, err)
				} else {
					klog.V(5).Infof("Updated vGPU utilization for Pod %s/%s, device %d: SM=%d%%, Dec=%d%%, Enc=%d%%",
						pod.Namespace, pod.Name, i, smUtil, decUtil, encUtil)
				}
			}
		}
	}

	// 原有的观察逻辑保持不变
	for _, c := range containers {
		recentKernel := c.Info.GetRecentKernel()
		if recentKernel > 0 {
			recentKernel--
			if recentKernel > 0 {
				for i := 0; i < c.Info.DeviceMax(); i++ {
					//for _, devuuid := range val.sr.uuids {
					// Null device condition
					if !c.Info.IsValidUUID(i) {
						continue
					}
					uuid := c.Info.DeviceUUID(i)
					if len(utSwitchOn[uuid]) == 0 {
						utSwitchOn[uuid] = []int{0, 0}
					}
					utSwitchOn[uuid][c.Info.GetPriority()]++
				}
			}
			c.Info.SetRecentKernel(recentKernel)
		}
	}
	for idx, c := range containers {
		priority := c.Info.GetPriority()
		recentKernel := c.Info.GetRecentKernel()
		utilizationSwitch := c.Info.GetUtilizationSwitch()
		if CheckBlocking(utSwitchOn, priority, c) {
			if recentKernel >= 0 {
				klog.V(5).Infof("utSwitchon=%v", utSwitchOn)
				klog.V(5).Infof("Setting Blocking to on %v", idx)
				c.Info.SetRecentKernel(-1)
			}
		} else {
			if recentKernel < 0 {
				klog.V(5).Infof("utSwitchon=%v", utSwitchOn)
				klog.V(5).Infof("Setting Blocking to off %v", idx)
				c.Info.SetRecentKernel(0)
			}
		}
		if CheckPriority(utSwitchOn, priority, c) {
			if utilizationSwitch != 1 {
				klog.V(5).Infof("utSwitchon=%v", utSwitchOn)
				klog.V(5).Infof("Setting UtilizationSwitch to on %v", idx)
				c.Info.SetUtilizationSwitch(1)
			}
		} else {
			if utilizationSwitch != 0 {
				klog.V(5).Infof("utSwitchon=%v", utSwitchOn)
				klog.V(5).Infof("Setting UtilizationSwitch to off %v", idx)
				c.Info.SetUtilizationSwitch(0)
			}
		}
	}
}
