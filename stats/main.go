package stats

import (
	"fmt"
	"log"
	"time"

	"github.com/pbaettig/raspi-dash/borg"
	"github.com/pbaettig/raspi-dash/config"
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
	proc                   procfs.FS
	DocumentsNewestBackups []borg.Archive
	PhotosNewestBackups    []borg.Archive
	Backups                map[string][]borg.Archive = map[string][]borg.Archive{
		borg.DocumentsRepo.Name: {},
		borg.PhotosRepo.Name:    {},
	}

	prevNetworkRx float64
	prevNetworkTx float64
)

func init() {
	var err error
	proc, err = procfs.NewDefaultFS()
	if err != nil {
		log.Fatalln(err.Error())
	}

	go updateTicker()
}

func updateBackups() {
	dbs, err := borg.DocumentsRepo.ListBackupArchives()
	if err == nil {
		// DocumentsNewestBackups = dbs[:5]
		Backups[borg.DocumentsRepo.Name] = dbs[:5]
	}
	pbs, err := borg.PhotosRepo.ListBackupArchives()
	if err == nil {
		// PhotosNewestBackups = pbs[:5]
		Backups[borg.PhotosRepo.Name] = pbs[:5]
	}
}

func updatePlots(t time.Time) {
	temp, _ := CPUTemperature()

	CPUTemperaturePlot.Value.Datapoints.Push(t, temp)

	avg, _ := LoadAvg()
	LoadAvgPlot.Avg1.Datapoints.Push(t, avg.Load1)
	LoadAvgPlot.Avg5.Datapoints.Push(t, avg.Load5)
	LoadAvgPlot.Avg15.Datapoints.Push(t, avg.Load15)

	mi, _ := Meminfo()
	MemoryUsedPlot.Value.Datapoints.Push(t, (float64(*mi.MemTotal)-float64(*mi.MemFree))*100/float64(*mi.MemTotal))

	fs, err := Filesystems()
	if err != nil {
		log.Fatalln(err.Error())
	}

	DiskUsagePlot.Boot.Datapoints.Push(t, fs["/boot"].UsagePercent)
	DiskUsagePlot.Data.Datapoints.Push(t, fs["/data"].UsagePercent)
	DiskUsagePlot.Root.Datapoints.Push(t, fs["/"].UsagePercent)

	n, _ := NetworkEth0()

	rxR := 0.0
	txR := 0.0

	if prevNetworkRx != 0 {
		rxR = (float64(n.RxBytes) - prevNetworkRx) / (1024 * 1024)
	}
	if prevNetworkTx != 0 {
		txR = (float64(n.TxBytes) - prevNetworkTx) / (1024 * 1024)
	}

	NetworkRxTxPlot.Rx.Datapoints.Push(t, rxR)
	NetworkRxTxPlot.Tx.Datapoints.Push(t, txR)
	prevNetworkRx = float64(n.RxBytes)
	prevNetworkTx = float64(n.TxBytes)
}

func updateTicker() {
	backupTicker := time.NewTicker(config.BackupUpdateInterval)
	plotTicker := time.NewTicker(config.PlotUpdateInterval)

	go updateBackups()

	for {
		select {
		case t := <-plotTicker.C:
			go updatePlots(t)
		case <-backupTicker.C:
			go updateBackups()
		}
	}
}

func PrepareIndexPageData() templates.IndexPageData {
	ipd := templates.IndexPageData{Plots: make(map[string]string)}

	avg1 := LoadAvgPlot.Avg1.Datapoints.Latest()
	avg5 := LoadAvgPlot.Avg5.Datapoints.Latest()
	avg15 := LoadAvgPlot.Avg15.Datapoints.Latest()

	ipd.LoadAvg1 = fmt.Sprintf("%.2f", avg1.Y)
	ipd.LoadAvg5 = fmt.Sprintf("%.2f", avg5.Y)
	ipd.LoadAvg15 = fmt.Sprintf("%.2f", avg15.Y)

	temp := CPUTemperaturePlot.Value.Datapoints.Latest()
	ipd.CPUTemp = fmt.Sprintf("%.1f", temp.Y)

	for n := range AllPlots {
		ipd.Plots[n] = fmt.Sprintf("/plot/%s?range=-1", n)
	}

	ipd.RaidStats, _ = MDStats()

	// update Backup Age
	for k := range Backups {
		for i := range Backups[k] {
			c := time.Time(Backups[k][i].Created)
			Backups[k][i].Age = time.Since(c.In(time.Local))
		}
	}

	ipd.Backups = Backups

	ipd.RangeSliderMin = 60
	ipd.RangeSliderMax = config.PlotDatapoints

	return ipd
}
