package series

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/procfs"
)

type Datapoint struct {
	Timestamp time.Time
	Value     float64
}

func (d Datapoint) Print() {
	fmt.Printf("%s: %f\n", d.Timestamp, d.Value)
}

type Datapoints struct {
	values    []Datapoint
	maxValues int
}

func (d *Datapoints) Push(dp Datapoint) error {
	if len(d.values) == d.maxValues {
		d.values = d.values[1:d.maxValues]
	}

	d.values = append(d.values, dp)

	return nil
}

func (d *Datapoints) All() []Datapoint {
	return d.values
}

func (d *Datapoints) Earliest() Datapoint {
	if len(d.values) == 0 {
		return Datapoint{}
	}

	return d.values[0]
}

func (d *Datapoints) Latest() Datapoint {
	if len(d.values) == 0 {
		return Datapoint{}
	}
	return d.values[len(d.values)-1]
}

type Series struct {
	Name          string
	Datapoints    *Datapoints
	maxDatapoints int
}

func NewSeries(name string, capacity int) *Series {
	s := new(Series)
	s.Name = name

	s.Datapoints = &Datapoints{
		values:    make([]Datapoint, 0, capacity),
		maxValues: capacity,
	}

	return s
}

func (s *Series) Print() {
	fmt.Println()
	for i, dp := range s.Datapoints.All() {
		fmt.Printf("[%d] %s: %f\n", i, dp.Timestamp, dp.Value)
	}
}

type WorkerFunc func(<-chan time.Time)

var (
	ASeries     *Series = NewSeries("ASeries", 10)
	BSeries     *Series = NewSeries("VSeries", 10)
	tickerMutex         = new(sync.Mutex)
)

func ASeriesWorker(c <-chan time.Time) {
	fs, _ := procfs.NewFS("/proc")
	pstat, _ := fs.Stat()

	for t := range c {
		// tickerMutex.Lock()
		ASeries.Datapoints.Push(Datapoint{t, float64(pstat.ProcessesRunning)})
		// tickerMutex.Unlock()
	}
}

func BSeriesWorker(c <-chan time.Time) {
	v := float64(0)
	for t := range c {
		// tickerMutex.Lock()
		BSeries.Datapoints.Push(Datapoint{t, v})
		v += 1
		// tickerMutex.Unlock()
	}
}

// func StartWorker(wf WorkerFunc) {
// 	c := make(chan time.Time)

// }
