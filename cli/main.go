package cli

import (
	"fmt"
	"log"
	"net/http"
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

func (handler *ExporterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", handler.to)
}

type ExporterHandler struct {
	to string
}

func start(config *Config) {
	exporters, err := config.BuildExporters()
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range exporters {
		handler := new(ExporterHandler)
		handler.to = e.URL.Host
		http.Handle(*e.Path, handler)
	}

	err = http.ListenAndServe(*config.Listen, nil)
	if err != nil {
		log.Fatal(err)
	}
}
