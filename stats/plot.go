package stats

import (
	"bytes"
	"image/color"

	"github.com/pbaettig/raspi-dash/series"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
)

var (
	TemperaturePlot CPUTemperaturePlot = CPUTemperaturePlot{series.NewSeries("CPU Temperature", 3600)}
	NetworkRxTxPlot NetworkPlot        = NetworkPlot{series.NewSeries("eth0 Rx", 3600), series.NewSeries("eth0 Tx", 3600)}
	LoadAvgPlot     LoadPlot           = LoadPlot{
		Avg1:  series.NewSeries("avg1", 3600),
		Avg5:  series.NewSeries("avg5", 3600),
		Avg15: series.NewSeries("avg15", 3600),
	}
)

type StatPlotter interface {
	Plot(n int) ([]byte, error)
}

type CPUTemperaturePlot struct {
	*series.Series
}

func (ctp CPUTemperaturePlot) PNG(n int) ([]byte, error) {
	p := setupPlot("CPU Temperature")

	p.Y.Min = 0
	p.Y.Max = 120

	// linePoints := make(plotter.XYs, 0)
	// var datapoints []series.Datapoint
	// if n == -1 {
	// 	datapoints = ctp.Datapoints.All()
	// } else {
	// 	datapoints = ctp.Datapoints.Last(n)
	// }

	// for _, dp := range datapoints {
	// 	linePoints = append(linePoints, plotter.XY{float64(dp.Timestamp.Unix()), dp.Value})
	// }

	line, err := plotter.NewLine(ctp.Series.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	line.LineStyle = draw.LineStyle{
		Color: color.RGBA{255, 128, 0, 255},
		Width: 1.0,
	}

	p.X.Tick.Marker = plot.TimeTicks{}

	p.Add(line)

	// Save the plot to a PNG file.
	wt, err := p.WriterTo(640, 200, "png")
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	wt.WriteTo(buf)

	return buf.Bytes(), nil
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
		Color: color.RGBA{52, 92, 89, 255},
		Width: 1.0,
	}

	txLine, err := plotter.NewLine(np.Tx.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	txLine.LineStyle = draw.LineStyle{
		Color: color.RGBA{197, 72, 199, 255},
		Width: 1.0,
	}

	p := setupPlot("Network")
	p.Add(rxLine, txLine)
	p.Legend.Add("Rx", rxLine)
	p.Legend.Add("Tx", txLine)
	return plotToPng(p)
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
		Width: 1.0,
	}

	avg5Line, err := plotter.NewLine(lp.Avg5.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	avg5Line.LineStyle = draw.LineStyle{
		Color: color.RGBA{144, 186, 173, 255},
		Width: 1.0,
	}

	avg15Line, err := plotter.NewLine(lp.Avg15.Datapoints.Last(n))
	if err != nil {
		return []byte{}, err
	}
	avg15Line.LineStyle = draw.LineStyle{
		Color: color.RGBA{173, 246, 177, 255},
		Width: 1.0,
	}

	p := setupPlot("LoadAvg")
	p.Add(avg1Line, avg5Line, avg15Line)
	p.Legend.Add("Avg1", avg1Line)
	p.Legend.Add("Avg5", avg5Line)
	p.Legend.Add("Avg15", avg15Line)

	return plotToPng(p)
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
