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

func openWritableFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func start(config *config.Config) {
	ef, err := openWritableFile(*config.ErrorLogConfig.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer ef.Close()

	errorLogger := log.New(ef, "", log.LstdFlags)
	errorLogger.Printf("INFO: ExporterProxy v%s", Version)

	af, err := openWritableFile(*config.AccessLogConfig.Path)
	if err != nil {
		errorLogger.Fatal(err)
	}
	defer af.Close()

	accessLogger := log.New(af, "", log.LstdFlags)

	for _, e := range config.ExporterConfigs {
		proxy, err := server.NewExporterProxy(&e, accessLogger, errorLogger)
		if err != nil {
			errorLogger.Fatal(err)
		}
		http.Handle(*proxy.ServePath, proxy)
	}

	lsn, err := listener.Listen(*config.Listen)
	if err != nil {
		errorLogger.Fatal(err)
	}
	defer lsn.Close()

	srv := &http.Server{
		Handler: http.DefaultServeMux,
	}

	err = server.ServeHTTPAndHandleSignal(lsn, *srv, *config.ShutDownTimeout)
	if err != nil {
		errorLogger.Fatal(err)
	}
}
