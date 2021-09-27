package stats

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/pbaettig/raspi-dash/templates"
	"github.com/prometheus/procfs"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
)

type vcgencmdOutput struct {
	Key   string
	Value string
}

type throttleStatus struct {
	UnderVoltage                      bool
	CurrentlyThrottled                bool
	ArmFrequencyCapped                bool
	SoftTemperatureReached            bool
	UnderVoltageSinceReboot           bool
	ThrottledSinceReboot              bool
	ArmFrequencyCappedSinceReboot     bool
	SoftTemperatureReachedSinceReboot bool
}

var (
	proc procfs.FS
)

func init() {
	var err error
	proc, err = procfs.NewDefaultFS()
	if err != nil {
		log.Fatalln(err.Error())
	}
}

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

func PlotPNG() []byte {
	rand.Seed(int64(0))

	p := plot.New()

	p.Title.Text = "Plotutil example"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Temp"

	p.Y.Min = 0
	p.Y.Max = 100
	// points := make(plotter.XYs, 4)
	// points[0] = plotter.XY{1, 55}
	// points[1] = plotter.XY{2, 57}
	// points[2] = plotter.XY{3, 75}
	// points[3] = plotter.XY{4, 89}

	ps := plotter.XYs{
		plotter.XY{0, 55},
		plotter.XY{1, 57},
		plotter.XY{2, 75},
		plotter.XY{3, 89},
		plotter.XY{4, 89},
		plotter.XY{5, 75},
		plotter.XY{6, 72},
		plotter.XY{7, 66},
		plotter.XY{8, 55},
		plotter.XY{9, 50},
	}

	p.X.Min = 0
	p.X.Max = float64(len(ps) - 1)

	err := plotutil.AddLinePoints(p, "First", ps)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	wt, err := p.WriterTo(640, 200, "png")
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	wt.WriteTo(buf)

	return buf.Bytes()
}

func CollectIndexPageData() templates.IndexPageData {

	ipd := templates.IndexPageData{}
	la, err := LoadAvg()
	if err == nil {
		ipd.LoadAvg1 = fmt.Sprintf("%.2f", la.Load1)
		ipd.LoadAvg5 = fmt.Sprintf("%.2f", la.Load5)
		ipd.LoadAvg15 = fmt.Sprintf("%.2f", la.Load15)
	}

	temp, err := CPUTemperature()
	if err == nil {
		ipd.CPUTemp = fmt.Sprintf("%.1f", temp)
	}

	ipd.PlotB64 = base64.StdEncoding.EncodeToString(PlotPNG())

	return ipd

}
