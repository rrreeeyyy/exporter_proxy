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

	cfg, err := config.LoadConfigFromYAML(options.Config)
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.Validate()
	if err != nil {
		log.Fatal(err)
	}

	start(cfg)
}

func openWritableFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func start(cfg *config.Config) {
	var err error
	var errorLogger *log.Logger
	if cfg.ErrorLogConfig != nil {
		ef, err := openWritableFile(*cfg.ErrorLogConfig.Path)
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
	if cfg.AccessLogConfig != nil {
		if cfg.AccessLogConfig.Path != nil {
			af, err := openWritableFile(*cfg.AccessLogConfig.Path)
			if err != nil {
				errorLogger.Fatal(err)
			}
			defer af.Close()
			accessLogger, err = accesslogger.New(*cfg.AccessLogConfig.Format, af, cfg.AccessLogConfig.Fields)
			if err != nil {
				errorLogger.Fatal(err)
			}
		} else {
			accessLogger, err = accesslogger.New(*cfg.AccessLogConfig.Format, os.Stdout, cfg.AccessLogConfig.Fields)
			if err != nil {
				errorLogger.Fatal(err)
			}
		}
	}

	for _, e := range cfg.ExporterConfigs {
		proxy, err := server.NewExporterProxy(&e, accessLogger, errorLogger)
		if err != nil {
			errorLogger.Fatal(err)
		}
		http.Handle(*proxy.ServePath, proxy)
	}

	lsn, err := listener.Listen(*cfg.Listen)
	if err != nil {
		errorLogger.Fatal(err)
	}
	defer lsn.Close()

	srv := &http.Server{
		Handler: http.DefaultServeMux,
	}

	var tlsConfig config.TLSConfig
	if cfg.TLSConfig != nil {
		tlsConfig = *cfg.TLSConfig
	}

	err = server.ServeHTTPAndHandleSignal(lsn, *srv, *cfg.ShutDownTimeout, tlsConfig)
	if err != nil {
		errorLogger.Fatal(err)
	}
}
