package cli

import (
	"fmt"
	"github.com/rrreeeyyy/exporter_proxy/handler"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
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

	config, err := LoadConfigFromYAML(options.Config)
	if err != nil {
		log.Fatal(err)
	}

	err = config.Validate()
	if err != nil {
		log.Fatal(err)
	}

	start(config)
}

func start(config *Config) {
	exporters, err := config.BuildExporters()
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range exporters {
		handler := new(handler.ExporterHandler)
		handler.URL = e.URL
		director := handler.CreateDirector()
		rproxy := &httputil.ReverseProxy{Director: director}
		http.Handle(*e.Path, rproxy)
	}

	err = http.ListenAndServe(*config.Listen, nil)
	if err != nil {
		log.Fatal(err)
	}
}
