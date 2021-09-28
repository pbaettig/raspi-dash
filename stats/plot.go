package stats

import (
	"bytes"
	"image/color"
	"time"

	"github.com/pbaettig/raspi-dash/series"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
)

var (
	TemperaturePlot SingleValuePlot = SingleValuePlot{
		Value: series.NewSeries("cpuTemp", 3600),
		Name:  "CPUTemperature",
		YMin:  0,
		YMax:  100,
		LineStyle: &draw.LineStyle{
			Color: color.RGBA{216, 87, 42, 255},
			Width: 1.5,
		},
	}
	NetworkRxTxPlot NetworkPlot = NetworkPlot{
		Rx: series.NewSeries("eth0 Rx", 3600),
		Tx: series.NewSeries("eth0 Tx", 3600),
	}
	LoadAvgPlot LoadPlot = LoadPlot{
		Avg1:  series.NewSeries("avg1", 3600),
		Avg5:  series.NewSeries("avg5", 3600),
		Avg15: series.NewSeries("avg15", 3600),
	}
	MemoryUsedPlot SingleValuePlot = SingleValuePlot{
		Value: series.NewSeries("used", 3600),
		Name:  "MemoryUsed",
		YMin:  0,
		YMax:  100,
		LineStyle: &draw.LineStyle{
			Color: color.RGBA{37, 137, 189, 255},
			Width: 1.2,
		},
	}
	AllPlots map[string]StatPlotter = map[string]StatPlotter{
		"cpuTemp":    &TemperaturePlot,
		"network":    &NetworkRxTxPlot,
		"loadAvg":    &LoadAvgPlot,
		"memoryUsed": &MemoryUsedPlot,
	}
)

type StatPlotter interface {
	PNG(n int) ([]byte, error)
	AddPoint(t time.Time, v ...float64)
}

// type CPUTemperaturePlot struct {
// 	*series.Series
// }

// func (ctp CPUTemperaturePlot) PNG(n int) ([]byte, error) {
// 	line, err := plotter.NewLine(ctp.Series.Datapoints.Last(n))
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	line.LineStyle = draw.LineStyle{
// 		Color: color.RGBA{202, 46, 85, 255},
// 		Width: 1.0,
// 	}
// 	p := setupPlot("CPU Temperature")
// 	p.Y.Min = 0
// 	p.Y.Max = 120
// 	p.X.Tick.Marker = plot.TimeTicks{}
// 	p.Add(line)

// 	return plotToPng(p)
// }

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

	p := setupPlot("LoadAvg")
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

// type MemPlot struct {
// 	Used, Available *series.Series
// }

// func (mp MemPlot) PNG(n int) ([]byte, error) {
// 	usedLine, err := plotter.NewLine(mp.Used.Datapoints.Last(n))
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	usedLine.LineStyle = draw.LineStyle{
// 		Color: color.RGBA{238, 118, 116, 255},
// 		Width: 1.0,
// 	}
// 	usedLine.FillColor = color.RGBA{238, 118, 116, 255}

// 	availableLine, err := plotter.NewLine(mp.Available.Datapoints.Last(n))
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	availableLine.LineStyle = draw.LineStyle{
// 		Color: color.RGBA{51, 103, 59, 255},
// 		Width: 1.0,
// 	}
// 	availableLine.FillColor = color.RGBA{51, 103, 59, 255}

// 	p := setupPlot("Memory")
// 	p.Add(usedLine, availableLine)
// 	p.Legend.Add("used", usedLine)
// 	p.Legend.Add("available", availableLine)

// 	return plotToPng(p)
// }

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
	p.Title.Text = title
	p.X.Tick.Marker = plot.TimeTicks{}
	p.Add(plotter.NewGrid())

	return p
}

func plotToPng(p *plot.Plot) ([]byte, error) {
	// Save the plot to a PNG file.
	wt, err := p.WriterTo(640, 200, "png")
	if err != nil {
		return []byte{}, err
	}

	buf := new(bytes.Buffer)
	wt.WriteTo(buf)

	return buf.Bytes(), nil
}
