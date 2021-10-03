package router

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pbaettig/raspi-dash/config"
	"github.com/pbaettig/raspi-dash/docs"
	"github.com/pbaettig/raspi-dash/stats"
	"github.com/pbaettig/raspi-dash/templates"
)

var (
	docfs = docs.NewDocuments(config.DocumentsPath)
)

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

func docByIdHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		buf []byte
	)

	vars := mux.Vars(r)
	id := vars["id"]
	if m, _ := regexp.MatchString(`^\d{3}$`, id); !m {
		http.Error(w, "id not recognized", http.StatusBadRequest)
		return
	}

	found, err := docfs.FindById(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("no document with id %s found", id), http.StatusNotFound)
		return
	}

	if len(found) == 1 {
		buf, err = docfs.ReadFile(found[0])
	} else {
		buf, err = docfs.ZipFiles(found)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(buf)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.IndexPage.Execute(w, stats.PrepareIndexPageData())
	if err != nil {
		log.Fatalln(err.Error())
	}
}
