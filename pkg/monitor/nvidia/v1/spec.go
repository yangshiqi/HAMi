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

package v1

import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

const maxDevices = 16
const maxProcesses = 1024

type deviceMemory struct {
	contextSize uint64
	moduleSize  uint64
	bufferSize  uint64
	offset      uint64
	total       uint64
	unused      [3]uint64
}

type deviceUtilization struct {
	DecUtil uint64
	EncUtil uint64
	SmUtil  uint64
	unused  [3]uint64
}

// DeviceUtilization 是公开的设备利用率结构，供外部包使用
type DeviceUtilization struct {
	DecUtil uint64
	EncUtil uint64
	SmUtil  uint64
}

type shrregProcSlotT struct {
	pid         int32
	hostpid     int32
	used        [16]deviceMemory
	monitorused [16]uint64
	deviceUtil  [16]deviceUtilization
	status      int32
	unused      [3]uint64
}

type uuid struct {
	uuid [96]byte
}

type semT struct {
	sem [32]byte
}

type sharedRegionT struct {
	initializedFlag int32
	majorVersion    int32
	minorVersion    int32
	smInitFlag      int32
	ownerPid        uint32
	sem             semT
	num             uint64
	uuids           [16]uuid

	limit   [16]uint64
	smLimit [16]uint64
	procs   [1024]shrregProcSlotT

	procnum           int32
	utilizationSwitch int32
	recentKernel      int32
	priority          int32
	lastKernelTime    int64
	unused            [4]uint64
}

// GPUUtilization 表示 GPU 利用率数据
type GPUUtilization struct {
	DecUtil   uint64 // 解码器利用率
	EncUtil   uint64 // 编码器利用率
	SmUtil    uint64 // SM 利用率
	Timestamp int64  // 时间戳
}

// ProcessGPUStats 表示进程的 GPU 统计信息
type ProcessGPUStats struct {
	PID         int32
	Status      int32
	Utilization GPUUtilization
	LastUpdate  int64
}

// VGPUStats 表示 vGPU 的统计信息
type VGPUStats struct {
	Index       int    // vGPU 索引
	UUID        string // vGPU UUID
	Utilization GPUUtilization
	LastUpdate  int64
}

// PodGPUStats 表示 Pod 的 GPU 统计信息
type PodGPUStats struct {
	PodUID     string
	Namespace  string
	Name       string
	VGPUs      map[int]*VGPUStats // vGPU 索引 -> 统计信息
	LastUpdate int64
}

// Spec 表示 GPU 规格和状态
type Spec struct {
	sr    *sharedRegionT
	lock  sync.RWMutex
	stats map[int32]*ProcessGPUStats // 进程 GPU 统计信息缓存
	pods  map[string]*PodGPUStats    // Pod UID -> Pod GPU 统计信息
}

func (s Spec) DeviceMax() int {
	return maxDevices
}

func (s Spec) DeviceNum() int {
	return int(s.sr.num)
}

func (s Spec) DeviceMemoryContextSize(idx int) uint64 {
	v := uint64(0)
	for _, p := range s.sr.procs {
		v += p.used[idx].contextSize
	}
	return v
}

func (s Spec) DeviceMemoryModuleSize(idx int) uint64 {
	v := uint64(0)
	for _, p := range s.sr.procs {
		v += p.used[idx].moduleSize
	}
	return v
}

func (s Spec) DeviceMemoryBufferSize(idx int) uint64 {
	v := uint64(0)
	for _, p := range s.sr.procs {
		v += p.used[idx].bufferSize
	}
	return v
}

func (s Spec) DeviceMemoryOffset(idx int) uint64 {
	v := uint64(0)
	for _, p := range s.sr.procs {
		v += p.used[idx].offset
	}
	return v
}

func (s Spec) DeviceMemoryTotal(idx int) uint64 {
	v := uint64(0)
	for _, p := range s.sr.procs {
		v += p.used[idx].total
	}
	return v
}

func (s Spec) DeviceSmUtil(idx int) uint64 {
	if idx < 0 || idx >= maxDevices {
		return 0
	}

	var sum uint64
	for _, p := range s.sr.procs {
		sum += p.deviceUtil[idx].SmUtil
	}
	return sum
}

func (s Spec) DeviceDecUtil(idx int) uint64 {
	if idx < 0 || idx >= maxDevices {
		return 0
	}

	var sum uint64
	for _, p := range s.sr.procs {
		if p.status == 1 && p.pid > 0 {
			sum += p.deviceUtil[idx].DecUtil
		}
	}
	return sum
}

func (s Spec) DeviceEncUtil(idx int) uint64 {
	if idx < 0 || idx >= maxDevices {
		return 0
	}

	var sum uint64
	for _, p := range s.sr.procs {
		if p.status == 1 && p.pid > 0 {
			sum += p.deviceUtil[idx].EncUtil
		}
	}
	return sum
}

func (s Spec) SetDeviceSmLimit(l uint64) {
	idx := uint64(0)
	for idx < s.sr.num {
		s.sr.smLimit[idx] = l
		idx += 1
	}
}

func (s Spec) IsValidUUID(idx int) bool {
	return s.sr.uuids[idx].uuid[0] != 0
}

func (s Spec) DeviceUUID(idx int) string {
	return string(s.sr.uuids[idx].uuid[:])
}

func (s Spec) DeviceMemoryLimit(idx int) uint64 {
	return s.sr.limit[idx]
}

func (s Spec) SetDeviceMemoryLimit(l uint64) {
	idx := uint64(0)
	for idx < s.sr.num {
		s.sr.limit[idx] = l
		idx += 1
	}
}

func (s Spec) LastKernelTime() int64 {
	return s.sr.lastKernelTime
}

func CastSpec(data []byte) Spec {
	return Spec{
		sr: (*sharedRegionT)(unsafe.Pointer(&data[0])),
	}
}

//	func (s *SharedRegionT) UsedMemory(idx int) (uint64, error) {
//		return 0, nil
//	}

func (s Spec) GetPriority() int {
	return int(s.sr.priority)
}

func (s Spec) GetRecentKernel() int32 {
	return s.sr.recentKernel
}

func (s Spec) SetRecentKernel(v int32) {
	s.sr.recentKernel = v
}

func (s Spec) GetUtilizationSwitch() int32 {
	return s.sr.utilizationSwitch
}

func (s Spec) SetUtilizationSwitch(v int32) {
	s.sr.utilizationSwitch = v
}

// NewSpec 创建新的 Spec 实例
func NewSpec() *Spec {
	return &Spec{
		stats: make(map[int32]*ProcessGPUStats),
		pods:  make(map[string]*PodGPUStats),
	}
}

// UpdateProcessUtilization 更新进程的 GPU 利用率
func (s *Spec) UpdateProcessUtilization(pid int32, idx int, util deviceUtilization) error {
	// 参数验证
	if idx < 0 || idx >= maxDevices {
		return fmt.Errorf("invalid device index: %d", idx)
	}

	if s.sr == nil {
		return fmt.Errorf("shared region is nil")
	}

	// 验证利用率值
	if util.DecUtil > 100 || util.EncUtil > 100 || util.SmUtil > 100 {
		return fmt.Errorf("invalid utilization values: dec=%d, enc=%d, sm=%d",
			util.DecUtil, util.EncUtil, util.SmUtil)
	}

	s.lock.Lock()
	defer s.lock.Unlock()

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
	slot := -1
	for i := 0; i < int(s.sr.procnum); i++ {
		if s.sr.procs[i].pid == pid {
			slot = i
			break
		}
	}

	if slot == -1 {
		if int(s.sr.procnum) >= maxProcesses {
			return fmt.Errorf("no available process slot")
		}
		slot = int(s.sr.procnum)
		s.sr.procnum++
	}

	s.sr.procs[slot].pid = pid
	s.sr.procs[slot].status = 1
	s.sr.procs[slot].deviceUtil[idx] = util

	return nil
}

// GetProcessUtilization 获取指定进程的 GPU 利用率
func (s *Spec) GetProcessUtilization(pid int32) (*GPUUtilization, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	stats, exists := s.stats[pid]
	if !exists {
		return nil, fmt.Errorf("process %d not found", pid)
	}

	return &stats.Utilization, nil
}

// GetTotalUtilization 获取所有活动进程的总 GPU 利用率
func (s *Spec) GetTotalUtilization() GPUUtilization {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var total GPUUtilization
	for _, stats := range s.stats {
		if stats.Status == 1 {
			total.DecUtil += stats.Utilization.DecUtil
			total.EncUtil += stats.Utilization.EncUtil
			total.SmUtil += stats.Utilization.SmUtil
		}
	}
	total.Timestamp = time.Now().Unix()
	return total
}

// GetActiveProcesses 获取所有活动进程的 GPU 利用率
func (s *Spec) GetActiveProcesses() map[int32]GPUUtilization {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := make(map[int32]GPUUtilization)
	for pid, stats := range s.stats {
		if stats.Status == 1 {
			result[pid] = stats.Utilization
		}
	}
	return result
}

// CleanupInactiveProcesses 清理不活跃的进程数据
func (s *Spec) CleanupInactiveProcesses() {
	s.lock.Lock()
	defer s.lock.Unlock()

	now := time.Now().Unix()

	// 创建进程状态映射表
	activeProcs := make(map[int32]bool)

	// 首先清理 stats 中的不活跃进程
	for pid, stats := range s.stats {
		// 清理超过 5 分钟未更新的进程
		if now-stats.LastUpdate > 300 {
			delete(s.stats, pid)
			// 同步清理共享内存
			for i := 0; i < int(s.sr.procnum); i++ {
				if s.sr.procs[i].pid == pid {
					s.sr.procs[i].status = 0
					s.sr.procs[i].deviceUtil = [16]deviceUtilization{}
					break
				}
			}
		} else if stats.Status == 1 {
			activeProcs[pid] = true
		}
	}

	// 检查共享内存中的活跃进程
	activeCount := 0
	for i := 0; i < int(s.sr.procnum); i++ {
		proc := s.sr.procs[i]
		if proc.status == 1 && proc.pid > 0 {
			activeProcs[proc.pid] = true
		}
	}

	// 计算活跃进程数量
	for range activeProcs {
		activeCount++
	}

	// 更新 procnum 为活跃进程的数量
	s.sr.procnum = int32(activeCount)
}

// UpdatePodGPUUtilization 更新 Pod 的 GPU 利用率
func (s *Spec) UpdatePodGPUUtilization(podUID string, namespace string, name string, vgpuIndex int, vgpuUUID string, util DeviceUtilization) error {
	// 参数验证
	if vgpuIndex < 0 || vgpuIndex >= maxDevices {
		return fmt.Errorf("invalid vGPU index: %d", vgpuIndex)
	}

	if s.sr == nil {
		return fmt.Errorf("shared region is nil")
	}

	// 验证利用率值
	if util.DecUtil > 100 || util.EncUtil > 100 || util.SmUtil > 100 {
		return fmt.Errorf("invalid utilization values: dec=%d, enc=%d, sm=%d",
			util.DecUtil, util.EncUtil, util.SmUtil)
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	// 更新或创建 Pod 统计信息
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

	// 更新或创建 vGPU 统计信息
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
	devUtil := deviceUtilization{
		DecUtil: util.DecUtil,
		EncUtil: util.EncUtil,
		SmUtil:  util.SmUtil,
	}

	// 尝试查找进程槽位并更新，但不会报错，因为这是从Pod到进程级别的映射，可能不存在
	for i := 0; i < int(s.sr.procnum); i++ {
		if s.sr.procs[i].status == 1 && s.sr.procs[i].pid > 0 {
			s.sr.procs[i].deviceUtil[vgpuIndex] = devUtil
		}
	}

	return nil
}

// GetPodGPUUtilization 获取指定 Pod 的 GPU 利用率
func (s *Spec) GetPodGPUUtilization(podUID string) (*PodGPUStats, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	podStats, exists := s.pods[podUID]
	if !exists {
		return nil, fmt.Errorf("pod %s not found", podUID)
	}

	return podStats, nil
}

// GetAllPodsGPUUtilization 获取所有 Pod 的 GPU 利用率
func (s *Spec) GetAllPodsGPUUtilization() map[string]*PodGPUStats {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := make(map[string]*PodGPUStats)
	for podUID, podStats := range s.pods {
		result[podUID] = podStats
	}
	return result
}

// CleanupInactivePods 清理不活跃的 Pod 数据
func (s *Spec) CleanupInactivePods() {
	s.lock.Lock()
	defer s.lock.Unlock()

	now := time.Now().Unix()
	for podUID, podStats := range s.pods {
		// 清理超过 5 分钟未更新的 Pod
		if now-podStats.LastUpdate > 300 {
			delete(s.pods, podUID)
		}
	}
}
