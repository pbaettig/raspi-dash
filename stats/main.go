package stats

import (
	"encoding/base64"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pbaettig/raspi-dash/templates"
	"github.com/prometheus/procfs"
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

// TemperatureSeries *series.Series = series.NewSeries("CPU Temperature", 300)
)

func init() {
	var err error
	proc, err = procfs.NewDefaultFS()
	if err != nil {
		log.Fatalln(err.Error())
	}

	go plotSeriesCollector()
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

func plotSeriesCollector() {
	ticker := time.NewTicker(1 * time.Second)
	prevRxB := uint64(0)
	prevTxB := uint64(0)
	rxR := 0.0
	txR := 0.0
	for {
		t := <-ticker.C
		temp, _ := CPUTemperature()
		TemperaturePlot.Series.Datapoints.Push(t, temp)

		n, _ := NetworkEth0()
		if prevRxB != 0 && prevTxB != 0 {
			rxR = float64(n.RxBytes - prevRxB)
			txR = float64(n.TxBytes - prevTxB)
		}

		NetworkRxTxPlot.Rx.Datapoints.Push(t, rxR)
		NetworkRxTxPlot.Tx.Datapoints.Push(t, txR)
		prevRxB = n.RxBytes
		prevTxB = n.TxBytes

		avg, _ := LoadAvg()
		LoadAvgPlot.Avg1.Datapoints.Push(t, avg.Load1)
		LoadAvgPlot.Avg5.Datapoints.Push(t, avg.Load5)
		LoadAvgPlot.Avg15.Datapoints.Push(t, avg.Load15)
	}

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

// func PlotPNG() []byte {
// 	rand.Seed(int64(0))

// 	p := plot.New()

// 	p.Title.Text = "Plotutil example"
// 	p.X.Label.Text = "Time"
// 	p.Y.Label.Text = "Temp"

// 	p.Y.Min = 0
// 	p.Y.Max = 100
// 	// points := make(plotter.XYs, 4)
// 	// points[0] = plotter.XY{1, 55}
// 	// points[1] = plotter.XY{2, 57}
// 	// points[2] = plotter.XY{3, 75}
// 	// points[3] = plotter.XY{4, 89}

// 	ps := plotter.XYs{
// 		plotter.XY{0, 55},
// 		plotter.XY{1, 57},
// 		plotter.XY{2, 75},
// 		plotter.XY{3, 89},
// 		plotter.XY{4, 89},
// 		plotter.XY{5, 75},
// 		plotter.XY{6, 72},
// 		plotter.XY{7, 66},
// 		plotter.XY{8, 55},
// 		plotter.XY{9, 50},
// 	}

// 	p.X.Min = 0
// 	p.X.Max = float64(len(ps) - 1)

// 	err := plotutil.AddLinePoints(p, "First", ps)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Save the plot to a PNG file.
// 	wt, err := p.WriterTo(640, 200, "png")
// 	if err != nil {
// 		panic(err)
// 	}
// 	buf := new(bytes.Buffer)
// 	wt.WriteTo(buf)

// 	return buf.Bytes()
// }

func CollectIndexPageData() templates.IndexPageData {

	ipd := templates.IndexPageData{Plots: make(map[string]string)}
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

	tp, _ := TemperaturePlot.PNG(300)
	ipd.Plots["CPU Temp"] = base64.StdEncoding.EncodeToString(tp)

	np, _ := NetworkRxTxPlot.PNG(300)
	ipd.Plots["eth0 Rx/Tx"] = base64.StdEncoding.EncodeToString(np)

	lp, _ := LoadAvgPlot.PNG(300)
	ipd.Plots["LoadAvg"] = base64.StdEncoding.EncodeToString(lp)
	return ipd

}
