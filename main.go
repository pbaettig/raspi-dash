package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/pbaettig/raspi-dash/assets"
	"github.com/pbaettig/raspi-dash/stats"
	"github.com/pbaettig/raspi-dash/templates"
)

var (
	sigs   chan os.Signal = make(chan os.Signal, 1)
	server *http.Server
)

func interuptSignalHandler() {
	<-sigs
	log.Println("signal received, shutting down.")
	server.Close()
}

func writePlot(p stats.StatPlotter, n int, w http.ResponseWriter) {
	b, err := p.PNG(n)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Cache-Control", "no-cache, must-revalidate")
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.IndexPage.Execute(w, stats.PrepareIndexPageData())
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func main() {
	// a, err := borg.PhotosRepo.NewestBackupArchive()
	// if err != nil {
	// 	log.Fatalln(err.Error())
	// }
	// fmt.Printf("%s,%s,%s\n", a.Name, a.Created, a.ID)
	// fmt.Println(time.Since(time.Time(a.Created)))

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go interuptSignalHandler()

	r := mux.NewRouter()
	r.HandleFunc("/plot/{name}", plotHandler)
	r.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(http.FS(assets.FS))))
	r.HandleFunc("/", indexHandler)

	server = &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalln(err.Error())
	}

	os.Exit(0)
}
