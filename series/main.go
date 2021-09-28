package series

import (
	"fmt"
	"time"

	"gonum.org/v1/plot/plotter"
)

// type Datapoint struct {
// 	Timestamp time.Time
// 	Value     float64
// }

// func (d Datapoint) Print() {
// 	fmt.Printf("%s: %f\n", d.Timestamp, d.Value)
// }

type Datapoints struct {
	values   plotter.XYs
	Capacity int
}

func (d *Datapoints) PushXY(xy plotter.XY) error {
	if len(d.values) == d.Capacity {
		d.values = d.values[1:d.Capacity]
	}

	d.values = append(d.values, xy)

	return nil
}

func (d *Datapoints) Push(ts time.Time, v float64) error {
	d.PushXY(plotter.XY{float64(ts.Unix()), v})
	return nil
}

func (d *Datapoints) All() plotter.XYs {
	return d.values
}

func (d *Datapoints) Last(n int) plotter.XYs {

	if len(d.values) <= n {
		return d.values
	}
	return d.values[len(d.values)-n:]
}

func (d *Datapoints) Earliest() plotter.XY {
	if len(d.values) == 0 {
		return plotter.XY{}
	}

	return d.values[0]
}

func (d *Datapoints) Latest() plotter.XY {
	if len(d.values) == 0 {
		return plotter.XY{}
	}
	return d.values[len(d.values)-1]
}

type Series struct {
	Name       string
	Datapoints *Datapoints
}

func NewSeries(name string, capacity int) *Series {
	s := new(Series)
	s.Name = name

	s.Datapoints = &Datapoints{
		values:   make(plotter.XYs, 0, capacity),
		Capacity: capacity,
	}

	return s
}

func (s *Series) Print() {
	fmt.Println()
	for i, dp := range s.Datapoints.All() {
		fmt.Printf("[%d] %.0f: %f\n", i, dp.X, dp.Y)
	}
}

// type WorkerFunc func(<-chan time.Time)

// var (
// 	ASeries     *Series = NewSeries("ASeries", 10)
// 	BSeries     *Series = NewSeries("VSeries", 10)
// 	tickerMutex         = new(sync.Mutex)
// )

// func ASeriesWorker(c <-chan time.Time) {
// 	fs, _ := procfs.NewFS("/proc")
// 	pstat, _ := fs.Stat()

// 	for t := range c {
// 		// tickerMutex.Lock()
// 		ASeries.Datapoints.Push(Datapoint{t, float64(pstat.ProcessesRunning)})
// 		// tickerMutex.Unlock()
// 	}
// }

// func BSeriesWorker(c <-chan time.Time) {
// 	v := float64(0)
// 	for t := range c {
// 		// tickerMutex.Lock()
// 		BSeries.Datapoints.Push(Datapoint{t, v})
// 		v += 1
// 		// tickerMutex.Unlock()
// 	}
// }

// func StartWorker(wf WorkerFunc) {
// 	c := make(chan time.Time)

// }
