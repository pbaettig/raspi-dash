package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pbaettig/raspi-dash/router"
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

func main() {
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go interuptSignalHandler()

	server = &http.Server{
		Addr:           ":8080",
		Handler:        router.New(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalln(err.Error())
	}

	os.Exit(0)
}
