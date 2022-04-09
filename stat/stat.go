package stat

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type NodeStat struct {
	PeerId      string
	HostInfo    *host.InfoStat
	Swapmem     *mem.SwapMemoryStat // 交换内存信息
	Mem         *mem.VirtualMemoryStat
	CpuInfo     []cpu.InfoStat
	DiskUseInfo *disk.UsageStat
	Interfaces  []net.InterfaceStat
}

func GetNodeStat() *NodeStat {
	vm, svm := MemInfo()
	return &NodeStat{
		CpuInfo:     CpuInfo(),
		Mem:         vm,
		Swapmem:     svm,
		DiskUseInfo: DiskUsage(),
		Interfaces:  Interfaces(),
		HostInfo:    HostInfos(),
	}
}

func CpuInfo() []cpu.InfoStat {
	r, _ := cpu.Info()
	return r
}

func MemInfo() (vm *mem.VirtualMemoryStat, svm *mem.SwapMemoryStat) {
	vm, _ = mem.VirtualMemory()
	svm, _ = mem.SwapMemory()
	return
}

func DiskUsage() (r *disk.UsageStat) {
	r, _ = disk.Usage("/")
	return
}

func Interfaces() (r []net.InterfaceStat) {
	r, _ = net.Interfaces()
	return
}

func HostInfos() (r *host.InfoStat) {
	r, _ = host.Info()
	return
}
