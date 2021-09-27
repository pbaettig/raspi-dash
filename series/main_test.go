package series

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestSeries_AddDatapoint(t *testing.T) {
	type fields struct {
		Name          string
		Datapoints    []Datapoint
		maxDatapoints int
	}
	type args struct {
		dp Datapoint
	}
	capacities := []int{1, 3, 4, 6, 19}
	for _, c := range capacities {
		n := fmt.Sprintf("test-capacity-%d", c)

		s := NewSeries(n, c)
		tsStart := time.Date(2021, time.September, 9, 11, 0, 0, 0, time.UTC)

		for i := 0; i < c*2; i++ {
			v := rand.Float64()
			ts := tsStart.Add(time.Duration(float64(i)) * time.Second)
			s.Datapoints.Push(Datapoint{ts, v})

			var idx int
			if i < c {
				idx = i
			} else {
				idx = c - 1
			}

			s.Print()
			d := s.Datapoints.Latest()

			if d.Value != v {
				t.Errorf("iteration %d: value [%d] is %f but should be %f", i, idx, d.Value, v)
			}
			if !d.Timestamp.Equal(ts) {
				t.Errorf("iteration %d: timestamp [%d] is %s but should be %s", i, idx, d.Timestamp, ts)
			}
		}
	}
}
