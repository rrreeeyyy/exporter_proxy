package cli

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rrreeeyyy/exporter_proxy/config"
	"github.com/rrreeeyyy/exporter_proxy/listener"
	"github.com/rrreeeyyy/exporter_proxy/server"
)

func Start(args []string) {
	options := CommandLineOptions{}
	flagSet := SetupFlagSet(args[0], &options)

	err := flagSet.Parse(args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if options.ShowVersion {
		fmt.Printf("exporter_proxy v%s\n", Version)
		os.Exit(0)
	}

	if options.Config == "" {
		log.Fatal("ERROR: option -config is mandatory")
	}

	config, err := config.LoadConfigFromYAML(options.Config)
	if err != nil {
		log.Fatal(err)
	}

	err = config.Validate()
	if err != nil {
		log.Fatal(err)
	}

	start(config)
}

func start(config *config.Config) {
	for _, e := range config.ExporterConfigs {
		proxy, err := server.NewExporterProxy(&e)
		if err != nil {
			log.Fatal(err)
		}
		http.Handle(*e.Path, proxy)
	}

	lsn, err := listener.Listen(*config.Listen)
	if err != nil {
		log.Fatal(err)
	}
	defer lsn.Close()

	srv := &http.Server{
		Handler: http.DefaultServeMux,
	}

	err = server.ServeHTTPAndHandleSignal(lsn, *srv, *config.ShutDownTimeout)
	if err != nil {
		log.Fatal(err)
	}
}
