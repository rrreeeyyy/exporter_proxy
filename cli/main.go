package cli

import (
	"fmt"
	"github.com/rrreeeyyy/exporter_proxy/config"
	"github.com/rrreeeyyy/exporter_proxy/server"
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

	err := http.ListenAndServe(*config.Listen, nil)
	if err != nil {
		log.Fatal(err)
	}
}
