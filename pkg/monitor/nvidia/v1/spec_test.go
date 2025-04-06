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
	"testing"

	"gotest.tools/v3/assert"
)

func Test_DeviceMax(t *testing.T) {
	tests := []struct {
		name string
		args Spec
		want int
	}{
		{
			name: "device max is 8",
			args: Spec{
				sr: &sharedRegionT{
					num: 8,
				},
			},
			want: maxDevices,
		},
		{
			name: "device max is 16",
			args: Spec{
				sr: &sharedRegionT{
					num: 16,
				},
			},
			want: maxDevices,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args
			result := s.DeviceMax()
			if result != test.want {
				t.Errorf("DeviceMax is %d, want is %d", result, test.want)
			}
		})
	}
}

func Test_DeviceNum(t *testing.T) {
	tests := []struct {
		name string
		args Spec
		want int
	}{
		{
			name: "device num is 2",
			args: Spec{
				sr: &sharedRegionT{
					num: 2,
				},
			},
			want: int(2),
		},
		{
			name: "device num is 4",
			args: Spec{
				sr: &sharedRegionT{
					num: 4,
				},
			},
			want: int(4),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args
			result := s.DeviceNum()
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_DeviceMemoryContextSize(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			idx  int
			spec Spec
		}
		want uint64
	}{
		{
			name: "device memory context size for idx 0",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(0),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								used: [16]deviceMemory{
									{
										contextSize: 100,
									},
									{
										contextSize: 200,
									},
								},
							},
							{
								used: [16]deviceMemory{
									{
										contextSize: 100,
									},
									{
										contextSize: 200,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(200),
		},
		{
			name: "device memory context size for idx 1",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(1),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								used: [16]deviceMemory{
									{
										contextSize: 100,
									},
									{
										contextSize: 200,
									},
								},
							},
							{
								used: [16]deviceMemory{
									{
										contextSize: 100,
									},
									{
										contextSize: 200,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(400),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			result := s.DeviceMemoryContextSize(test.args.idx)
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_DeviceMemoryModuleSize(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			idx  int
			spec Spec
		}
		want uint64
	}{
		{
			name: "device memory module size for idx 0",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(0),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								used: [16]deviceMemory{
									{
										moduleSize: 100,
									},
									{
										moduleSize: 200,
									},
								},
							},
							{
								used: [16]deviceMemory{
									{
										moduleSize: 100,
									},
									{
										moduleSize: 200,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(200),
		},
		{
			name: "device memory module size for idx 1",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(1),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								used: [16]deviceMemory{
									{
										moduleSize: 100,
									},
									{
										moduleSize: 200,
									},
								},
							},
							{
								used: [16]deviceMemory{
									{
										moduleSize: 100,
									},
									{
										moduleSize: 200,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(400),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			result := s.DeviceMemoryModuleSize(test.args.idx)
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_DeviceMemoryBufferSize(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			idx  int
			spec Spec
		}
		want uint64
	}{
		{
			name: "device memory buffer size for idx 0",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(0),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								used: [16]deviceMemory{
									{
										bufferSize: 100,
									},
									{
										bufferSize: 200,
									},
								},
							},
							{
								used: [16]deviceMemory{
									{
										bufferSize: 100,
									},
									{
										bufferSize: 200,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(200),
		},
		{
			name: "device memory buffer size for idx 1",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(1),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								used: [16]deviceMemory{
									{
										bufferSize: 100,
									},
									{
										bufferSize: 200,
									},
								},
							},
							{
								used: [16]deviceMemory{
									{
										bufferSize: 100,
									},
									{
										bufferSize: 200,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(400),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			result := s.DeviceMemoryBufferSize(test.args.idx)
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_DeviceMemoryOffset(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			idx  int
			spec Spec
		}
		want uint64
	}{
		{
			name: "device memory offset for idx 0",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(0),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								used: [16]deviceMemory{
									{
										offset: 100,
									},
									{
										offset: 200,
									},
								},
							},
							{
								used: [16]deviceMemory{
									{
										offset: 100,
									},
									{
										offset: 200,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(200),
		},
		{
			name: "device memory offset for idx 1",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(1),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								used: [16]deviceMemory{
									{
										offset: 100,
									},
									{
										offset: 200,
									},
								},
							},
							{
								used: [16]deviceMemory{
									{
										offset: 100,
									},
									{
										offset: 200,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(400),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			result := s.DeviceMemoryOffset(test.args.idx)
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_DeviceMemoryTotal(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			idx  int
			spec *Spec
		}
		want uint64
	}{
		{
			name: "device memory total for idx 0",
			args: struct {
				idx  int
				spec *Spec
			}{
				idx: int(0),
				spec: &Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								used: [16]deviceMemory{
									{
										total: 100,
									},
									{
										total: 200,
									},
								},
							},
							{
								used: [16]deviceMemory{
									{
										total: 100,
									},
									{
										total: 200,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(200),
		},
		{
			name: "device memory total for idx 1",
			args: struct {
				idx  int
				spec *Spec
			}{
				idx: int(1),
				spec: &Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								used: [16]deviceMemory{
									{
										total: 100,
									},
									{
										total: 200,
									},
								},
							},
							{
								used: [16]deviceMemory{
									{
										total: 100,
									},
									{
										total: 200,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(400),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			result := s.DeviceMemoryTotal(test.args.idx)
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_DeviceSmUtil(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			idx  int
			spec Spec
		}
		want uint64
	}{
		{
			name: "device sm util for idx 0",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(0),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								deviceUtil: [16]deviceUtilization{
									{
										SmUtil: 100,
									},
									{
										SmUtil: 200,
									},
								},
							},
							{
								deviceUtil: [16]deviceUtilization{
									{
										SmUtil: 100,
									},
									{
										SmUtil: 200,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(200),
		},
		{
			name: "device sm util for idx 1",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(1),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								deviceUtil: [16]deviceUtilization{
									{
										SmUtil: 100,
									},
									{
										SmUtil: 200,
									},
								},
							},
							{
								deviceUtil: [16]deviceUtilization{
									{
										SmUtil: 100,
									},
									{
										SmUtil: 200,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(400),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			result := s.DeviceSmUtil(test.args.idx)
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_DeviceDecUtil(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			idx  int
			spec Spec
		}
		want uint64
	}{
		{
			name: "device decoder utilization for idx 0",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(0),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								pid:    1, // 设置pid大于0
								status: 1, // 设置status为1
								deviceUtil: [16]deviceUtilization{
									{
										DecUtil: 50,
									},
									{
										DecUtil: 75,
									},
								},
							},
							{
								pid:    2, // 设置pid大于0
								status: 1, // 设置status为1
								deviceUtil: [16]deviceUtilization{
									{
										DecUtil: 25,
									},
									{
										DecUtil: 100,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(75),
		},
		{
			name: "device decoder utilization for idx 1",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(1),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								pid:    1, // 设置pid大于0
								status: 1, // 设置status为1
								deviceUtil: [16]deviceUtilization{
									{
										DecUtil: 50,
									},
									{
										DecUtil: 75,
									},
								},
							},
							{
								pid:    2, // 设置pid大于0
								status: 1, // 设置status为1
								deviceUtil: [16]deviceUtilization{
									{
										DecUtil: 25,
									},
									{
										DecUtil: 100,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(175),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			result := s.DeviceDecUtil(test.args.idx)
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_DeviceEncUtil(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			idx  int
			spec Spec
		}
		want uint64
	}{
		{
			name: "device encoder utilization for idx 0",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(0),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								pid:    1, // 设置pid大于0
								status: 1, // 设置status为1
								deviceUtil: [16]deviceUtilization{
									{
										EncUtil: 30,
									},
									{
										EncUtil: 45,
									},
								},
							},
							{
								pid:    2, // 设置pid大于0
								status: 1, // 设置status为1
								deviceUtil: [16]deviceUtilization{
									{
										EncUtil: 15,
									},
									{
										EncUtil: 60,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(45),
		},
		{
			name: "device encoder utilization for idx 1",
			args: struct {
				idx  int
				spec Spec
			}{
				idx: int(1),
				spec: Spec{
					sr: &sharedRegionT{
						procs: [1024]shrregProcSlotT{
							{
								pid:    1, // 设置pid大于0
								status: 1, // 设置status为1
								deviceUtil: [16]deviceUtilization{
									{
										EncUtil: 30,
									},
									{
										EncUtil: 45,
									},
								},
							},
							{
								pid:    2, // 设置pid大于0
								status: 1, // 设置status为1
								deviceUtil: [16]deviceUtilization{
									{
										EncUtil: 15,
									},
									{
										EncUtil: 60,
									},
								},
							},
						},
					},
				},
			},
			want: uint64(105),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			result := s.DeviceEncUtil(test.args.idx)
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_SetDeviceSmLimit(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			l    uint64
			spec *Spec
		}
		want [16]uint64
	}{
		{
			name: "set device sm limit to 300",
			args: struct {
				l    uint64
				spec *Spec
			}{
				l: uint64(300),
				spec: &Spec{
					sr: &sharedRegionT{
						num:     2,
						smLimit: [16]uint64{},
					},
				},
			},
			want: [16]uint64{300, 300},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			s.SetDeviceSmLimit(test.args.l)
			result := test.args.spec.sr.smLimit
			assert.DeepEqual(t, result, test.want)
		})
	}
}

func Test_IsValidUUID(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			idx  int
			spec *Spec
		}
		want bool
	}{
		{
			name: "set vaild uuid",
			args: struct {
				idx  int
				spec *Spec
			}{
				idx: 0,
				spec: &Spec{
					sr: &sharedRegionT{
						uuids: [16]uuid{
							{
								uuid: [96]byte{
									1,
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "set invaild uuid",
			args: struct {
				idx  int
				spec *Spec
			}{
				idx: 0,
				spec: &Spec{
					sr: &sharedRegionT{
						uuids: [16]uuid{
							{
								uuid: [96]byte{
									0,
								},
							},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			result := s.IsValidUUID(test.args.idx)
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_DeviceUUID(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			idx  int
			spec *Spec
		}
		want string
	}{
		{
			name: "device uuid for idx 0",
			args: struct {
				idx  int
				spec *Spec
			}{
				idx: 0,
				spec: &Spec{
					sr: &sharedRegionT{
						uuids: [16]uuid{
							{
								uuid: [96]byte{
									'a', '1', 'b', '2',
								},
							},
						},
					},
				},
			},
			want: "a1b2",
		},
		{
			name: "device uuid for idx 1",
			args: struct {
				idx  int
				spec *Spec
			}{
				idx: 1,
				spec: &Spec{
					sr: &sharedRegionT{
						uuids: [16]uuid{
							{
								uuid: [96]byte{
									'a', '1', 'b', '2',
								},
							},
							{
								uuid: [96]byte{
									'c', '3', 'd', '4',
								},
							},
						},
					},
				},
			},
			want: "c3d4",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			result := s.DeviceUUID(test.args.idx)
			assert.Equal(t, result[:4], test.want)
		})
	}
}

func Test_DeviceMemoryLimit(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			idx  int
			spec *Spec
		}
		want uint64
	}{
		{
			name: "device memory limit for idx 0",
			args: struct {
				idx  int
				spec *Spec
			}{
				idx: 0,
				spec: &Spec{
					sr: &sharedRegionT{
						limit: [16]uint64{
							100,
						},
					},
				},
			},
			want: uint64(100),
		},
		{
			name: "device memory limit for idx 1",
			args: struct {
				idx  int
				spec *Spec
			}{
				idx: 1,
				spec: &Spec{
					sr: &sharedRegionT{
						limit: [16]uint64{
							100, 200,
						},
					},
				},
			},
			want: uint64(200),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			result := s.DeviceMemoryLimit(test.args.idx)
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_SetDeviceMemoryLimit(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			l    uint64
			spec *Spec
		}
		want [16]uint64
	}{
		{
			name: "set device memory limit to 1024",
			args: struct {
				l    uint64
				spec *Spec
			}{
				l: uint64(1024),
				spec: &Spec{
					sr: &sharedRegionT{
						num: 1,
					},
				},
			},
			want: [16]uint64{1024},
		},
		{
			name: "set device memory limit to 2048",
			args: struct {
				l    uint64
				spec *Spec
			}{
				l: uint64(2048),
				spec: &Spec{
					sr: &sharedRegionT{
						num: 2,
					},
				},
			},
			want: [16]uint64{2048, 2048},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			s.SetDeviceMemoryLimit(test.args.l)
			result := test.args.spec.sr.limit
			assert.DeepEqual(t, result, test.want)
		})
	}
}

func Test_LastKernelTime(t *testing.T) {
	tests := []struct {
		name string
		args *Spec
		want int64
	}{
		{
			name: "last kernel time",
			args: &Spec{
				sr: &sharedRegionT{
					lastKernelTime: int64(1234),
				},
			},
			want: int64(1234),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args
			result := s.LastKernelTime()
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_GetPriority(t *testing.T) {
	tests := []struct {
		name string
		args Spec
		want int
	}{
		{
			name: "get priority",
			args: Spec{
				sr: &sharedRegionT{
					priority: int32(1),
				},
			},
			want: int(1),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args
			result := s.GetPriority()
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_GetRecentKernel(t *testing.T) {
	tests := []struct {
		name string
		args Spec
		want int32
	}{
		{
			name: "get recent kernel",
			args: Spec{
				sr: &sharedRegionT{
					recentKernel: int32(1234),
				},
			},
			want: int32(1234),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args
			result := s.GetRecentKernel()
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_SetRecentKernel(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			v    int32
			spec Spec
		}
		want int32
	}{
		{
			name: "get recent kernel",
			args: struct {
				v    int32
				spec Spec
			}{
				v: int32(1111),
				spec: Spec{
					sr: &sharedRegionT{},
				},
			},
			want: int32(1111),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.args.spec
			s.SetRecentKernel(test.args.v)
			result := test.args.spec.sr.recentKernel
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_GetUtilizationSwitch(t *testing.T) {
	tests := []struct {
		name string
		args Spec
		want int32
	}{
		{
			name: "get utilzation switch",
			args: Spec{
				sr: &sharedRegionT{
					utilizationSwitch: int32(1234),
				},
			},
			want: int32(1234),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.args.GetUtilizationSwitch()
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_SetUtilizationSwitch(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			v    int32
			spec Spec
		}
		want int32
	}{
		{
			name: "set utilzation switch",
			args: struct {
				v    int32
				spec Spec
			}{
				v: int32(3333),
				spec: Spec{
					sr: &sharedRegionT{},
				},
			},
			want: int32(3333),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.args.spec.SetUtilizationSwitch(test.args.v)
			result := test.args.spec.sr.utilizationSwitch
			assert.Equal(t, result, test.want)
		})
	}
}

func Test_CleanupInactiveProcesses(t *testing.T) {
	// 创建测试用的 Spec 实例
	s := &Spec{
		sr: &sharedRegionT{
			procs: [maxProcesses]shrregProcSlotT{},
		},
	}

	// 测试用例1：没有不活跃的进程
	t.Run("No inactive processes", func(t *testing.T) {
		// 设置一些活跃的进程
		s.sr.procs[0] = shrregProcSlotT{pid: 1001, status: 1}
		s.sr.procs[1] = shrregProcSlotT{pid: 1002, status: 1}
		s.sr.procnum = 2

		// 执行清理
		s.CleanupInactiveProcesses()

		// 验证结果
		if s.sr.procnum != 2 {
			t.Errorf("Expected procnum to remain 2, got %d", s.sr.procnum)
		}
		if s.sr.procs[0].pid != 1001 || s.sr.procs[1].pid != 1002 {
			t.Error("Active processes should not be cleaned up")
		}
	})

	// 测试用例2：有不活跃的进程
	t.Run("With inactive processes", func(t *testing.T) {
		// 设置混合的活跃和不活跃进程
		s.sr.procs[0] = shrregProcSlotT{pid: 1001, status: 1}
		s.sr.procs[1] = shrregProcSlotT{pid: 0, status: 0} // 不活跃进程
		s.sr.procs[2] = shrregProcSlotT{pid: 1003, status: 1}
		s.sr.procs[3] = shrregProcSlotT{pid: 0, status: 0} // 不活跃进程
		s.sr.procnum = 4

		// 执行清理
		s.CleanupInactiveProcesses()

		// 验证结果
		if s.sr.procnum != 2 {
			t.Errorf("Expected procnum to be 2 after cleanup, got %d", s.sr.procnum)
		}
		if s.sr.procs[0].pid != 1001 || s.sr.procs[2].pid != 1003 {
			t.Error("Active processes should remain unchanged")
		}
		if s.sr.procs[1].pid != 0 || s.sr.procs[3].pid != 0 {
			t.Error("Inactive processes should be cleaned up")
		}
	})

	// 测试用例3：全部都是不活跃的进程
	t.Run("All inactive processes", func(t *testing.T) {
		// 设置全部不活跃的进程
		for i := 0; i < 4; i++ {
			s.sr.procs[i] = shrregProcSlotT{pid: 0, status: 0}
		}
		s.sr.procnum = 4

		// 执行清理
		s.CleanupInactiveProcesses()

		// 验证结果
		if s.sr.procnum != 0 {
			t.Errorf("Expected procnum to be 0 after cleanup, got %d", s.sr.procnum)
		}
		for i := 0; i < 4; i++ {
			if s.sr.procs[i].pid != 0 || s.sr.procs[i].status != 0 {
				t.Errorf("Process at index %d should be cleaned up", i)
			}
		}
	})

	// 测试用例4：边界情况 - 空进程列表
	t.Run("Empty process list", func(t *testing.T) {
		s.sr.procnum = 0
		s.CleanupInactiveProcesses()
		if s.sr.procnum != 0 {
			t.Errorf("Expected procnum to remain 0, got %d", s.sr.procnum)
		}
	})
}

// Test_UpdateProcessUtilization 测试更新进程的 GPU 利用率功能
func Test_UpdateProcessUtilization(t *testing.T) {
	// 创建一个测试用的共享区域
	testSR := &sharedRegionT{
		procnum: 0,
		procs:   [1024]shrregProcSlotT{},
	}

	tests := []struct {
		name    string
		pid     int32
		idx     int
		util    deviceUtilization
		wantErr bool
	}{
		{
			name: "update process with valid utilization",
			pid:  1000,
			idx:  0,
			util: deviceUtilization{
				DecUtil: 25,
				EncUtil: 35,
				SmUtil:  75,
			},
			wantErr: false,
		},
		{
			name: "update process with invalid device index",
			pid:  1001,
			idx:  -1, // 负数索引
			util: deviceUtilization{
				DecUtil: 25,
				EncUtil: 35,
				SmUtil:  75,
			},
			wantErr: true,
		},
		{
			name: "update process with excessive device index",
			pid:  1002,
			idx:  maxDevices + 1, // 超过最大设备数
			util: deviceUtilization{
				DecUtil: 25,
				EncUtil: 35,
				SmUtil:  75,
			},
			wantErr: true,
		},
		{
			name: "update process with invalid utilization values",
			pid:  1003,
			idx:  0,
			util: deviceUtilization{
				DecUtil: 101, // 超过100%
				EncUtil: 35,
				SmUtil:  75,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为每个测试用例创建新的Spec实例
			s := &Spec{
				sr:    testSR,
				stats: make(map[int32]*ProcessGPUStats),
				pods:  make(map[string]*PodGPUStats),
			}

			// 调用被测试的方法
			err := s.UpdateProcessUtilization(tt.pid, tt.idx, tt.util)

			// 验证错误结果
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateProcessUtilization() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 如果不期望错误，则验证数据是否正确更新
			if !tt.wantErr {
				// 1. 验证进程统计信息是否存在
				stats, exists := s.stats[tt.pid]
				if !exists {
					t.Errorf("Process stats not found for PID %d", tt.pid)
					return
				}

				// 2. 验证利用率数据是否正确
				if stats.Utilization.DecUtil != tt.util.DecUtil ||
					stats.Utilization.EncUtil != tt.util.EncUtil ||
					stats.Utilization.SmUtil != tt.util.SmUtil {
					t.Errorf("Utilization mismatch: got %+v, want DecUtil=%d, EncUtil=%d, SmUtil=%d",
						stats.Utilization, tt.util.DecUtil, tt.util.EncUtil, tt.util.SmUtil)
				}

				// 3. 验证Status和PID是否正确设置
				if stats.Status != 1 || stats.PID != tt.pid {
					t.Errorf("Stats attributes mismatch: got Status=%d, PID=%d, want Status=1, PID=%d",
						stats.Status, stats.PID, tt.pid)
				}

				// 4. 验证共享内存是否正确更新
				found := false
				for i := 0; i < int(s.sr.procnum); i++ {
					if s.sr.procs[i].pid == tt.pid {
						found = true
						if s.sr.procs[i].deviceUtil[tt.idx].DecUtil != tt.util.DecUtil ||
							s.sr.procs[i].deviceUtil[tt.idx].EncUtil != tt.util.EncUtil ||
							s.sr.procs[i].deviceUtil[tt.idx].SmUtil != tt.util.SmUtil {
							t.Errorf("Shared memory utilization mismatch: got %+v, want %+v",
								s.sr.procs[i].deviceUtil[tt.idx], tt.util)
						}
						break
					}
				}
				if !found {
					t.Errorf("Process with PID %d not found in shared memory", tt.pid)
				}
			}
		})
	}
}

// Test_UpdatePodGPUUtilization 测试更新Pod的GPU利用率功能
func Test_UpdatePodGPUUtilization(t *testing.T) {
	// 创建一个测试用的共享区域
	testSR := &sharedRegionT{
		procnum: 1,
		procs: [1024]shrregProcSlotT{
			{
				pid:    1234,
				status: 1,
			},
		},
	}

	tests := []struct {
		name       string
		podUID     string
		namespace  string
		podName    string
		vgpuIndex  int
		vgpuUUID   string
		util       DeviceUtilization
		wantErr    bool
		setupProcs func(s *Spec) // 用于设置额外的进程数据
	}{
		{
			name:      "update pod with valid utilization",
			podUID:    "pod-123",
			namespace: "default",
			podName:   "test-pod",
			vgpuIndex: 0,
			vgpuUUID:  "gpu-uuid-001",
			util: DeviceUtilization{
				DecUtil: 25,
				EncUtil: 35,
				SmUtil:  75,
			},
			wantErr: false,
		},
		{
			name:      "update pod with invalid device index",
			podUID:    "pod-124",
			namespace: "default",
			podName:   "test-pod-2",
			vgpuIndex: -1, // 负数索引
			vgpuUUID:  "gpu-uuid-002",
			util: DeviceUtilization{
				DecUtil: 25,
				EncUtil: 35,
				SmUtil:  75,
			},
			wantErr: true,
		},
		{
			name:      "update pod with excessive device index",
			podUID:    "pod-125",
			namespace: "default",
			podName:   "test-pod-3",
			vgpuIndex: maxDevices + 1, // 超过最大设备数
			vgpuUUID:  "gpu-uuid-003",
			util: DeviceUtilization{
				DecUtil: 25,
				EncUtil: 35,
				SmUtil:  75,
			},
			wantErr: true,
		},
		{
			name:      "update pod with invalid utilization values",
			podUID:    "pod-126",
			namespace: "default",
			podName:   "test-pod-4",
			vgpuIndex: 0,
			vgpuUUID:  "gpu-uuid-004",
			util: DeviceUtilization{
				DecUtil: 101, // 超过100%
				EncUtil: 35,
				SmUtil:  75,
			},
			wantErr: true,
		},
		{
			name:      "update existing pod",
			podUID:    "pod-127",
			namespace: "default",
			podName:   "test-pod-5",
			vgpuIndex: 0,
			vgpuUUID:  "gpu-uuid-005",
			util: DeviceUtilization{
				DecUtil: 50,
				EncUtil: 60,
				SmUtil:  70,
			},
			wantErr: false,
			setupProcs: func(s *Spec) {
				// 预先设置一个Pod
				s.pods["pod-127"] = &PodGPUStats{
					PodUID:     "pod-127",
					Namespace:  "default",
					Name:       "test-pod-5",
					VGPUs:      make(map[int]*VGPUStats),
					LastUpdate: 100, // 使用固定时间戳
				}
				// 预先设置一个vGPU
				s.pods["pod-127"].VGPUs[0] = &VGPUStats{
					Index: 0,
					UUID:  "gpu-uuid-005-old", // 旧的UUID
					Utilization: GPUUtilization{
						DecUtil:   10,
						EncUtil:   20,
						SmUtil:    30,
						Timestamp: 100,
					},
					LastUpdate: 100,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为每个测试用例创建新的Spec实例
			s := &Spec{
				sr:    testSR,
				stats: make(map[int32]*ProcessGPUStats),
				pods:  make(map[string]*PodGPUStats),
			}

			// 如果有设置函数，调用它
			if tt.setupProcs != nil {
				tt.setupProcs(s)
			}

			// 记录更新前的时间戳
			var initialLastUpdate int64
			if existingPod, exists := s.pods[tt.podUID]; exists {
				initialLastUpdate = existingPod.LastUpdate
			}

			// 调用被测试的方法
			err := s.UpdatePodGPUUtilization(tt.podUID, tt.namespace, tt.podName, tt.vgpuIndex, tt.vgpuUUID, tt.util)

			// 验证错误结果
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePodGPUUtilization() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 如果不期望错误，则验证数据是否正确更新
			if !tt.wantErr {
				// 1. 验证Pod统计信息是否存在
				podStats, exists := s.pods[tt.podUID]
				if !exists {
					t.Errorf("Pod stats not found for UID %s", tt.podUID)
					return
				}

				// 2. 验证Pod基本信息是否正确
				if podStats.Namespace != tt.namespace || podStats.Name != tt.podName {
					t.Errorf("Pod info mismatch: got namespace=%s, name=%s, want namespace=%s, name=%s",
						podStats.Namespace, podStats.Name, tt.namespace, tt.podName)
				}

				// 3. 验证LastUpdate是否已更新（应该大于初始值）
				if podStats.LastUpdate <= initialLastUpdate && initialLastUpdate > 0 {
					t.Errorf("Pod LastUpdate not updated: got %d, initial was %d",
						podStats.LastUpdate, initialLastUpdate)
				}

				// 4. 验证vGPU是否存在
				vgpuStats, exists := podStats.VGPUs[tt.vgpuIndex]
				if !exists {
					t.Errorf("vGPU stats not found for index %d", tt.vgpuIndex)
					return
				}

				// 5. 验证vGPU信息是否正确
				if vgpuStats.UUID != tt.vgpuUUID || vgpuStats.Index != tt.vgpuIndex {
					t.Errorf("vGPU info mismatch: got UUID=%s, index=%d, want UUID=%s, index=%d",
						vgpuStats.UUID, vgpuStats.Index, tt.vgpuUUID, tt.vgpuIndex)
				}

				// 6. 验证vGPU利用率数据是否正确
				if vgpuStats.Utilization.DecUtil != tt.util.DecUtil ||
					vgpuStats.Utilization.EncUtil != tt.util.EncUtil ||
					vgpuStats.Utilization.SmUtil != tt.util.SmUtil {
					t.Errorf("vGPU utilization mismatch: got %+v, want DecUtil=%d, EncUtil=%d, SmUtil=%d",
						vgpuStats.Utilization, tt.util.DecUtil, tt.util.EncUtil, tt.util.SmUtil)
				}

				// 7. 验证共享内存中的进程是否正确更新（不过这只有在有进程时才会更新）
				for i := 0; i < int(s.sr.procnum); i++ {
					if s.sr.procs[i].status == 1 && s.sr.procs[i].pid > 0 {
						devUtil := s.sr.procs[i].deviceUtil[tt.vgpuIndex]
						if devUtil.DecUtil != tt.util.DecUtil ||
							devUtil.EncUtil != tt.util.EncUtil ||
							devUtil.SmUtil != tt.util.SmUtil {
							t.Errorf("Shared memory not updated correctly for process %d: got %+v, want %+v",
								s.sr.procs[i].pid, devUtil, tt.util)
						}
					}
				}
			}
		})
	}
}
