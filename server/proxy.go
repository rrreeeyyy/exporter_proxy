package server

import (
	"fmt"
	"github.com/rrreeeyyy/exporter_proxy/accesslogger"
	"github.com/rrreeeyyy/exporter_proxy/config"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type ExporterProxy struct {
	ServePath *string
	Target    *url.URL
	AccessLog accesslogger.AccessLogger
	*httputil.ReverseProxy
}

func NewExporterProxy(c *config.ExporterConfig, al accesslogger.AccessLogger, el *log.Logger) (*ExporterProxy, error) {
	target, err := url.Parse(*c.URL)
	if err != nil {
		return nil, err
	}
	p := &ExporterProxy{c.Path, target, al, &httputil.ReverseProxy{ErrorLog: el, Director: createDirector(target)}}
	return p, nil
}

func (p *ExporterProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	record := accesslogger.AccessLogRecord{}
	logRW := wrapLogResponseWriter(rw)

	start := time.Now()
	p.ReverseProxy.ServeHTTP(logRW, req)
	end := time.Now()

	if p.AccessLog != nil {
		record["time"] = end.Format(time.RFC3339)
		record["time_nsec"] = fmt.Sprintf("%d", end.UnixNano())
		record["reqtime_nsec"] = fmt.Sprintf("%d", end.UnixNano()-start.UnixNano())
		record["status"] = fmt.Sprintf("%d", logRW.Status())
		record["size"] = fmt.Sprintf("%d", logRW.Size())
		record["path"] = req.URL.Path
		record["query"] = req.URL.RawQuery
		record["method"] = req.Method
		p.AccessLog.Log(record)
	}
}

func createDirector(target *url.URL) func(*http.Request) {
	targetQuery := target.RawQuery
	return func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
}
