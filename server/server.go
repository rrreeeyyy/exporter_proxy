package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rrreeeyyy/exporter_proxy/config"
)

func ServeHTTPAndHandleSignal(listener net.Listener, server http.Server, timeout time.Duration, tlsconfig config.TLSConfig) error {
	if &tlsconfig != nil {
		go func() {
			server.ServeTLS(listener, *tlsconfig.CertFile, *tlsconfig.KeyFile)
		}()
	} else {
		go func() {
			server.Serve(listener)
		}()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	<-sigCh

	ctx := context.Background()
	if timeout != time.Duration(0) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
