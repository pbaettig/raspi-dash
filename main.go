package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/pbaettig/raspi-dash/stats"
	"github.com/pbaettig/raspi-dash/templates"
)

func writePlot(p stats.StatPlotter, n int, w http.ResponseWriter) {
	b, err := p.PNG(n)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func plotHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	rp := r.FormValue("range")
	if rp == "" {
		rp = "-1"
	}

	rv, err := strconv.Atoi(rp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	n := vars["name"]

	p, ok := stats.AllPlots[n]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writePlot(p, rv, w)
}

func main() {
	mds, _ := stats.MDStats()
	for _, md := range mds {
		fmt.Printf("RAID stats for %s (%s):\n", md.Name, md.ActivityState)
		fmt.Printf("  Array Size: %.f GB\n", float64(md.BlocksTotal)/1024/1024)
		fmt.Printf("  Devices: (%d active / %d failed / %d down / %d spare)\n", md.DisksActive, md.DisksFailed, md.DisksDown, md.DisksSpare)
		for _, d := range md.Devices {
			fmt.Printf("    - %s\n", d)
		}
	}

	fmt.Println()

	t, _ := stats.CPUTemperature()
	fmt.Printf("current Temp: %.1fC\n", t)

	fmt.Println()

	ts, _ := stats.CPUThrottlingStatus()
	fmt.Printf("currently throttled: %v\n", ts.CurrentlyThrottled)
	fmt.Printf("currently under-volted: %v\n", ts.UnderVoltage)

	mi, _ := stats.Meminfo()
	fmt.Printf("memory: %d total / %d available / %d free\n", *mi.MemTotal, *mi.MemAvailable, *mi.MemFree)

	la, _ := stats.LoadAvg()
	fmt.Printf("load average: %.2f/%.2f/%.2f\n", la.Load1, la.Load5, la.Load15)

	eth0, _ := stats.NetworkEth0()
	fmt.Printf("%+v\n", eth0)

	r := mux.NewRouter()
	r.HandleFunc("/plot/{name}", plotHandler)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := templates.IndexPage.Execute(w, stats.CollectIndexPageData())
		if err != nil {
			log.Fatalln(err.Error())
		}
	})

	// http.HandleFunc("/")
	// http.HandleFunc("/plot/temp", func(w http.ResponseWriter, r *http.Request) {
	// 	b, err := stats.TemperaturePlot.PNG(300)
	// 	if err != nil {
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		return
	// 	}

	// 	w.Write(b)
	// })

	srv := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatalln(srv.ListenAndServe())

}
