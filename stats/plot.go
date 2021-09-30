package stats

import (
	"bytes"
	"image/color"
	"log"
	"time"

	"github.com/pbaettig/raspi-dash/config"
	"github.com/pbaettig/raspi-dash/series"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

var (
	CPUTemperaturePlot SingleValuePlot = SingleValuePlot{
		Value: series.NewSeries("cpuTemp", config.PlotDatapoints),
		Name:  "CPU Temperature",
		YMin:  0,
		YMax:  100,
		LineStyle: &draw.LineStyle{
			Color: color.RGBA{216, 87, 42, 255},
			Width: 1.5,
		},
	}
	NetworkRxTxPlot NetworkPlot = NetworkPlot{
		Rx: series.NewSeries("eth0 Rx", config.PlotDatapoints),
		Tx: series.NewSeries("eth0 Tx", config.PlotDatapoints),
	}
	LoadAvgPlot LoadPlot = LoadPlot{
		Avg1:  series.NewSeries("avg1", config.PlotDatapoints),
		Avg5:  series.NewSeries("avg5", config.PlotDatapoints),
		Avg15: series.NewSeries("avg15", config.PlotDatapoints),
	}
	MemoryUsedPlot SingleValuePlot = SingleValuePlot{
		Value: series.NewSeries("used", config.PlotDatapoints),
		Name:  "Memory Usage",
		YMin:  0,
		YMax:  100,
		LineStyle: &draw.LineStyle{
			Color: color.RGBA{37, 137, 189, 255},
			Width: 1.2,
		},
	}
	DiskUsagePlot DiskPlot = DiskPlot{
		Data: series.NewSeries("data", config.PlotDatapoints),
		Root: series.NewSeries("root", config.PlotDatapoints),
		Boot: series.NewSeries("boot", config.PlotDatapoints),
	}
	AllPlots map[string]StatPlotter = map[string]StatPlotter{
		"cpuTemp":     &CPUTemperaturePlot,
		"network":     &NetworkRxTxPlot,
		"loadAvg":     &LoadAvgPlot,
		"memoryUsage": &MemoryUsedPlot,
		"diskUsage":   &DiskUsagePlot,
	}
)

type StatPlotter interface {
	PNG(n int) ([]byte, error)
	// AddPoint(t time.Time, v ...float64)
}

type DiskPlot struct {
	Data *series.Series
	Root *series.Series
	Boot *series.Series
}

func (dp DiskPlot) PNG(n int) ([]byte, error) {
	dataLine, err := plotter.NewLine(dp.Data.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	dataLine.LineStyle = draw.LineStyle{
		Color: color.RGBA{43, 89, 195, 255},
		Width: 1.2,
	}

	rootLine, err := plotter.NewLine(dp.Root.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	rootLine.LineStyle = draw.LineStyle{
		Color: color.RGBA{211, 101, 130, 255},
		Width: 1.2,
	}

	bootLine, err := plotter.NewLine(dp.Boot.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	bootLine.LineStyle = draw.LineStyle{
		Color: color.RGBA{187, 219, 155, 255},
		Width: 1.2,
	}

	p := setupPlot("Disk Usage")
	p.Y.Min = 0
	p.Y.Max = 100
	p.Add(bootLine, dataLine, rootLine)

	p.Legend.Add("boot", bootLine)
	p.Legend.Add("data", dataLine)
	p.Legend.Add("root", rootLine)
	return plotToPng(p)
}

type NetworkPlot struct {
	Rx *series.Series
	Tx *series.Series
}

func (np NetworkPlot) PNG(n int) ([]byte, error) {
	rxLine, err := plotter.NewLine(np.Rx.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	rxLine.LineStyle = draw.LineStyle{
		Color: color.RGBA{232, 141, 103, 255},
		Width: 1.2,
	}

	txLine, err := plotter.NewLine(np.Tx.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	txLine.LineStyle = draw.LineStyle{
		Color: color.RGBA{153, 153, 195, 255},
		Width: 1.2,
	}

	p := setupPlot("Network")
	p.Y.Min = 0

	p.Add(rxLine, txLine)
	p.Legend.Add("Rx", rxLine)
	p.Legend.Add("Tx", txLine)
	return plotToPng(p)
}

// AddPoint Rx, TX
func (np NetworkPlot) AddPoint(t time.Time, v ...float64) {
	if len(v) != 2 {
		panic("not enough values to add point for NetworkPlot")
	}
	np.Rx.Datapoints.Push(t, v[0])
	np.Tx.Datapoints.Push(t, v[1])
}

type LoadPlot struct {
	Avg1, Avg5, Avg15 *series.Series
}

func (lp LoadPlot) PNG(n int) ([]byte, error) {
	avg1Line, err := plotter.NewLine(lp.Avg1.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	avg1Line.LineStyle = draw.LineStyle{
		Color: color.RGBA{113, 124, 137, 255},
		Width: 1.2,
	}

	avg5Line, err := plotter.NewLine(lp.Avg5.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	avg5Line.LineStyle = draw.LineStyle{
		Color: color.RGBA{144, 186, 173, 255},
		Width: 1.2,
	}

	avg15Line, err := plotter.NewLine(lp.Avg15.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	avg15Line.LineStyle = draw.LineStyle{
		Color: color.RGBA{173, 246, 177, 255},
		Width: 1.2,
	}

	p := setupPlot("Load Average")
	p.Add(avg1Line, avg5Line, avg15Line)
	p.Legend.Add("Avg1", avg1Line)
	p.Legend.Add("Avg5", avg5Line)
	p.Legend.Add("Avg15", avg15Line)

	return plotToPng(p)
}

// AddPoint avg1, avg5, avg15
func (lp LoadPlot) AddPoint(t time.Time, v ...float64) {
	if len(v) != 3 {
		panic("not enough values to add point for LoadPlot")
	}
	lp.Avg1.Datapoints.Push(t, v[0])
	lp.Avg5.Datapoints.Push(t, v[1])
	lp.Avg15.Datapoints.Push(t, v[2])
}

type SingleValuePlot struct {
	Name       string
	Value      *series.Series
	LineStyle  *draw.LineStyle
	YMin, YMax float64
}

func (sp SingleValuePlot) PNG(n int) ([]byte, error) {
	line, err := plotter.NewLine(sp.Value.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	if sp.LineStyle != nil {
		line.LineStyle = *sp.LineStyle
	}

	p := setupPlot(sp.Name)
	if sp.YMax > 0 {
		p.Y.Min = 0
		p.Y.Max = 100
	}

	p.Add(line)

	return plotToPng(p)
}

// AddPoint avg1, avg5, avg15
func (sp SingleValuePlot) AddPoint(t time.Time, v ...float64) {
	if len(v) != 1 {
		panic("not enough values to add point for SingleValuePlot")
	}
	sp.Value.Datapoints.Push(t, v[0])
}

func setupPlot(title string) *plot.Plot {
	p := plot.New()

	p.Title.TextStyle.Color = config.PlotTitleFontColor
	p.Title.TextStyle.Font.Style = config.PlotTitleFontStyle
	p.Title.TextStyle.Font.Weight = config.PlotTitleFontWeight
	p.Title.TextStyle.Font.Size = font.Length(config.PlotTitleFontSize)
	p.Title.Padding = vg.Points(5)
	p.Title.Text = title

	p.X.Tick.Marker = plot.TimeTicks{
		Ticker: nil,
		Format: "15:04:05",
		Time: func(t float64) time.Time {
			return time.Unix(int64(t), 0)
		},
	}

	p.Add(plotter.NewGrid())

	return p
}

func plotToPng(p *plot.Plot) ([]byte, error) {
	if p == nil {
		log.Fatalln("oops")
	}

	wt, err := p.WriterTo(config.PlotSizeWidth, config.PlotSizeHeight, "png")
	if err != nil {
		return []byte{}, err
	}

	buf := new(bytes.Buffer)
	wt.WriteTo(buf)

	return buf.Bytes(), nil
}
