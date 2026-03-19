package monitor

import (
	"fmt"
	"os"
	"strings"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type GPUProcess struct {
	GPU  int
	PID  uint32
	User string
	Mem  uint64 // MiB
}

type GPUInfo struct {
	Index    int
	Name     string
	MemTotal uint64
	MemUsed  uint64
}

func Init() error {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return fmt.Errorf("nvml init: %v", nvml.ErrorString(ret))
	}
	return nil
}

func Shutdown() {
	nvml.Shutdown()
}

func ListGPUs() ([]GPUInfo, error) {
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("device count: %v", nvml.ErrorString(ret))
	}
	gpus := make([]GPUInfo, 0, count)
	for i := 0; i < count; i++ {
		dev, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			continue
		}
		name, _ := dev.GetName()
		mem, _ := dev.GetMemoryInfo()
		gpus = append(gpus, GPUInfo{
			Index:    i,
			Name:     name,
			MemTotal: mem.Total / 1024 / 1024,
			MemUsed:  mem.Used / 1024 / 1024,
		})
	}
	return gpus, nil
}

func ListProcesses() ([]GPUProcess, error) {
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("device count: %v", nvml.ErrorString(ret))
	}
	var procs []GPUProcess
	for i := 0; i < count; i++ {
		dev, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			continue
		}
		infos, ret := dev.GetComputeRunningProcesses()
		if ret != nvml.SUCCESS {
			continue
		}
		for _, info := range infos {
			user := pidToUser(info.Pid)
			procs = append(procs, GPUProcess{
				GPU:  i,
				PID:  info.Pid,
				User: user,
				Mem:  info.UsedGpuMemory / 1024 / 1024,
			})
		}
	}
	return procs, nil
}

// UserGPUMap returns map[username][]gpuIndex (deduplicated)
func UserGPUMap() (map[string][]int, error) {
	procs, err := ListProcesses()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]map[int]bool)
	for _, p := range procs {
		if seen[p.User] == nil {
			seen[p.User] = make(map[int]bool)
		}
		seen[p.User][p.GPU] = true
	}
	result := make(map[string][]int)
	for user, gpus := range seen {
		for gpu := range gpus {
			result[user] = append(result[user], gpu)
		}
	}
	return result, nil
}

func pidToUser(pid uint32) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return uidToName(fields[1])
			}
		}
	}
	return "unknown"
}

func uidToName(uid string) string {
	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return uid
	}
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Split(line, ":")
		if len(parts) >= 3 && parts[2] == uid {
			return parts[0]
		}
	}
	return uid
}
