package cli

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rrreeeyyy/exporter_proxy/accesslogger"
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
	var err error
	var errorLogger *log.Logger
	if config.ErrorLogConfig != nil {
		ef, err := openWritableFile(*config.ErrorLogConfig.Path)
		if err != nil {
			log.Fatal(err)
		}
		defer ef.Close()
		errorLogger = log.New(ef, "", log.LstdFlags)
	} else {
		errorLogger = log.New(os.Stderr, "", log.LstdFlags)
	}

	errorLogger.Printf("INFO: ExporterProxy v%s", Version)

	var accessLogger accesslogger.AccessLogger
	if config.AccessLogConfig != nil {
		if config.AccessLogConfig.Path != nil {
			af, err := openWritableFile(*config.AccessLogConfig.Path)
			if err != nil {
				errorLogger.Fatal(err)
			}
			defer af.Close()
			accessLogger, err = accesslogger.New(*config.AccessLogConfig.Format, af, config.AccessLogConfig.Fields)
			if err != nil {
				errorLogger.Fatal(err)
			}
		} else {
			accessLogger, err = accesslogger.New(*config.AccessLogConfig.Format, os.Stdout, config.AccessLogConfig.Fields)
			if err != nil {
				errorLogger.Fatal(err)
			}
		}
	}

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

	err = server.ServeHTTPAndHandleSignal(lsn, *srv, *config.ShutDownTimeout, *config.TLSConfig)
	if err != nil {
		errorLogger.Fatal(err)
	}
}
