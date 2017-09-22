package server

import (
	"github.com/rrreeeyyy/exporter_proxy/config"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ExporterProxy struct {
	Path   *string
	Target *url.URL
	*httputil.ReverseProxy
}

func NewExporterProxy(c *config.ExporterConfig) (*ExporterProxy, error) {
	target, err := url.Parse(*c.URL)
	if err != nil {
		return nil, err
	}
	p := &ExporterProxy{c.Path, target, &httputil.ReverseProxy{}}
	p.setDirector()
	return p, nil
}

func (p *ExporterProxy) setDirector() {
	p.Director = func(r *http.Request) {
		r.URL = p.Target
	}
}
