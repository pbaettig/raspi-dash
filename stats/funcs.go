package stats

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/procfs"
	"golang.org/x/sys/unix"
)

func runVcgencmd(cmd string) (vcgencmdOutput, error) {
	c := exec.Command("vcgencmd", cmd)
	out, err := c.Output()
	if err != nil {
		return vcgencmdOutput{}, err
	}

	split := strings.SplitN(string(out), "=", 2)
	if len(split) != 2 {
		return vcgencmdOutput{}, fmt.Errorf("cannot parse vcgencmd output: %s", string(out))
	}
	return vcgencmdOutput{split[0], split[1]}, nil
}

func CPUTemperature() (float64, error) {
	t, err := runVcgencmd("measure_temp")
	if err != nil {
		return 0, err
	}

	p := regexp.MustCompile(`^([\d\.]+)`)
	sm := p.FindStringSubmatch(t.Value)
	if len(sm) == 0 {
		return 0, fmt.Errorf("cannot parse temperature")
	}

	temp, err := strconv.ParseFloat(sm[0], 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert temperature: %w", err)
	}

	return temp, nil
}

func CPUThrottlingStatus() (throttleStatus, error) {
	r, err := runVcgencmd("get_throttled")
	if err != nil {
		return throttleStatus{}, err
	}

	// r := vcgencmdOutput{"", "0x20002"}

	t, err := strconv.ParseInt(r.Value, 0, 64)
	if err != nil {
		return throttleStatus{}, err
	}

	ts := throttleStatus{}
	ts.UnderVoltage = t&(1<<0) > 0
	ts.CurrentlyThrottled = t&(1<<1) > 0
	ts.ArmFrequencyCapped = t&(1<<2) > 0
	ts.SoftTemperatureReached = t&(1<<3) > 0

	ts.ThrottledSinceReboot = t&(1<<17) > 0
	ts.UnderVoltageSinceReboot = t&(1<<18) > 0
	ts.ArmFrequencyCappedSinceReboot = t&(1<<19) > 0
	ts.SoftTemperatureReachedSinceReboot = t&(1<<20) > 0

	return ts, nil
}

func MDStats() ([]procfs.MDStat, error) {
	return proc.MDStat()
}

func Meminfo() (procfs.Meminfo, error) {
	return proc.Meminfo()
}

func LoadAvg() (*procfs.LoadAvg, error) {

	return proc.LoadAvg()
}

func NetworkEth0() (procfs.NetDevLine, error) {
	ndl, err := proc.NetDev()
	return ndl["eth0"], err
}

func Mounts() ([]*procfs.MountInfo, error) {
	filtered := make([]*procfs.MountInfo, 0)

	mis, err := procfs.GetMounts()
	if err != nil {
		return filtered, err
	}

	for _, mi := range mis {
		if strings.HasPrefix(mi.FSType, "ext") {
			filtered = append(filtered, mi)
			continue
		}

		if strings.Contains(mi.FSType, "fat") {
			filtered = append(filtered, mi)
			continue
		}
	}

	return filtered, nil
}

type filesystem struct {
	Type         string
	SizeBytes    uint64
	FreeBytes    uint64
	UsedBytes    uint64
	UsagePercent float64
}

func Filesystems() (map[string]filesystem, error) {
	fs := make(map[string]filesystem)

	mounts, err := Mounts()
	if err != nil {
		return fs, err
	}

	for _, m := range mounts {
		var stat unix.Statfs_t
		err := unix.Statfs(m.MountPoint, &stat)
		if err != nil {
			continue
		}
		size := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bavail * uint64(stat.Bsize)
		used := size - free
		f := new(filesystem)
		f.Type = m.FSType
		f.SizeBytes = size
		f.FreeBytes = free
		f.UsedBytes = used
		f.UsagePercent = float64(used) * 100 / float64(size)

		fs[m.MountPoint] = filesystem{
			Type:         m.FSType,
			SizeBytes:    size,
			FreeBytes:    free,
			UsedBytes:    used,
			UsagePercent: float64(used) * 100 / float64(size),
		}
	}
	return fs, nil
}
