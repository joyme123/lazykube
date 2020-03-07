package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joyme123/lazykube/pkg/client"
	log "github.com/sirupsen/logrus"
)

// Version : app version
var Version string

func main() {
	version := flag.Bool("v", false, "lazykube version")

	var params client.WhSvrParameters
	flag.IntVar(&params.Port, "port", 443, "Webhook server port")
	flag.StringVar(&params.CertFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS")
	flag.StringVar(&params.KeyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")

	flag.Parse()

	if *version {
		fmt.Println()
		os.Exit(0)
	}

	s, err := client.NewWebhookServer(&params)
	if err != nil {
		log.Errorf("new webhook server error: %v", err)
	}

	go func() {
		err := s.Start()
		if err != nil {
			log.Errorf("start webhook server error: %v", err)
			os.Exit(-1)
		}
	}()

	// listening OS shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	s.Shutdown()
}
