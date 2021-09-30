package templates

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/pbaettig/raspi-dash/borg"
	"github.com/prometheus/procfs"
)

type IndexPageData struct {
	LoadAvg1       string
	LoadAvg5       string
	LoadAvg15      string
	CPUTemp        string
	Plots          map[string]string
	RaidStats      []procfs.MDStat
	Backups        map[string][]borg.Archive
	RangeSliderMin int
	RangeSliderMax int
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%dh %dm", h, m)
}

//go:embed *.tmpl
var fs embed.FS
var IndexPage *template.Template

func init() {
	var err error
	IndexPage, err = template.New("index.html.tmpl").Funcs(
		template.FuncMap{
			"fmtDuration": fmtDuration,
		}).ParseFS(fs, "index.html.tmpl")

	if err != nil {
		log.Fatalf("cannot parse template: %s", err.Error())
	}
}
