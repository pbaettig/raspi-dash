package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pbaettig/raspi-dash/config"
	"github.com/pbaettig/raspi-dash/letsencrypt"
	"github.com/pbaettig/raspi-dash/router"
)

var (
	sigs        chan os.Signal = make(chan os.Signal, 1)
	httpServer  *http.Server
	httpsServer *http.Server
)

func interuptSignalHandler() {
	<-sigs
	log.Println("signal received, shutting down.")
	httpServer.Close()
	httpsServer.Close()
}

func main() {
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go interuptSignalHandler()

	cs := letsencrypt.Certs{
		Domains:        []string{config.ExternalDomainName},
		CertFilePath:   config.TLSCertificateFilePath,
		PrivateKeyPath: config.TLSPrivateKeyFilePath,
		// CADirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
		CADirectoryURL: "https://acme-v02.api.letsencrypt.org/directory",
	}
	if err := cs.RequestCerts(config.LetsEncryptUserName); err != nil {
		if err != letsencrypt.ErrCertFilesAlreadyExist {
			log.Fatalln(err.Error())
		}
		fmt.Println("cert files already exist")
	}

	httpServer = &http.Server{
		Addr:           ":8080",
		Handler:        router.PermanentRedirectHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err.Error())
		}
	}()

	httpsServer = &http.Server{
		Addr:           ":8443",
		Handler:        router.New(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	time.Sleep(5 * time.Second)

	if err := httpsServer.ListenAndServeTLS(cs.CertFilePath, cs.PrivateKeyPath); err != nil && err != http.ErrServerClosed {
		log.Fatalln(err.Error())
	}

	os.Exit(0)
}
